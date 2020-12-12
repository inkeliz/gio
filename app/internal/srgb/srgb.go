// SPDX-License-Identifier: Unlicense OR MIT

package srgb

import (
	"fmt"
	"runtime"
	"strings"

	"gioui.org/internal/glimpl"
	"gioui.org/internal/unsafe"
)

// FBO implements an intermediate sRGB FBO
// for gamma-correct rendering on platforms without
// sRGB enabled native framebuffers.
type FBO struct {
	c             *glimpl.Functions
	width, height int
	frameBuffer   glimpl.Framebuffer
	depthBuffer   glimpl.Renderbuffer
	colorTex      glimpl.Texture
	blitted       bool
	quad          glimpl.Buffer
	prog          glimpl.Program
	gl3           bool
}

func New(ctx glimpl.Context) (*FBO, error) {
	f, err := glimpl.NewFunctions(ctx)
	if err != nil {
		return nil, err
	}
	var gl3 bool
	glVer := f.GetString(glimpl.VERSION)
	ver, _, err := glimpl.ParseGLVersion(glVer)
	if err != nil {
		return nil, err
	}
	if ver[0] >= 3 {
		gl3 = true
	} else {
		exts := f.GetString(glimpl.EXTENSIONS)
		if !strings.Contains(exts, "EXT_sRGB") {
			return nil, fmt.Errorf("no support for OpenGL ES 3 nor EXT_sRGB")
		}
	}
	s := &FBO{
		c:           f,
		gl3:         gl3,
		frameBuffer: f.CreateFramebuffer(),
		colorTex:    f.CreateTexture(),
		depthBuffer: f.CreateRenderbuffer(),
	}
	f.BindTexture(glimpl.TEXTURE_2D, s.colorTex)
	f.TexParameteri(glimpl.TEXTURE_2D, glimpl.TEXTURE_WRAP_S, glimpl.CLAMP_TO_EDGE)
	f.TexParameteri(glimpl.TEXTURE_2D, glimpl.TEXTURE_WRAP_T, glimpl.CLAMP_TO_EDGE)
	f.TexParameteri(glimpl.TEXTURE_2D, glimpl.TEXTURE_MAG_FILTER, glimpl.NEAREST)
	f.TexParameteri(glimpl.TEXTURE_2D, glimpl.TEXTURE_MIN_FILTER, glimpl.NEAREST)
	return s, nil
}

func (s *FBO) Blit() {
	if !s.blitted {
		prog, err := glimpl.CreateProgram(s.c, blitVSrc, blitFSrc, []string{"pos", "uv"})
		if err != nil {
			panic(err)
		}
		s.prog = prog
		s.c.UseProgram(prog)
		s.c.Uniform1i(s.c.GetUniformLocation(prog, "tex"), 0)
		s.quad = s.c.CreateBuffer()
		s.c.BindBuffer(glimpl.ARRAY_BUFFER, s.quad)
		s.c.BufferData(glimpl.ARRAY_BUFFER,
			unsafe.BytesView([]float32{
				-1, +1, 0, 1,
				+1, +1, 1, 1,
				-1, -1, 0, 0,
				+1, -1, 1, 0,
			}),
			glimpl.STATIC_DRAW)
		s.blitted = true
	}
	s.c.BindFramebuffer(glimpl.FRAMEBUFFER, glimpl.Framebuffer{})
	s.c.UseProgram(s.prog)
	s.c.BindTexture(glimpl.TEXTURE_2D, s.colorTex)
	s.c.BindBuffer(glimpl.ARRAY_BUFFER, s.quad)
	s.c.VertexAttribPointer(0 /* pos */, 2, glimpl.FLOAT, false, 4*4, 0)
	s.c.VertexAttribPointer(1 /* uv */, 2, glimpl.FLOAT, false, 4*4, 4*2)
	s.c.EnableVertexAttribArray(0)
	s.c.EnableVertexAttribArray(1)
	s.c.DrawArrays(glimpl.TRIANGLE_STRIP, 0, 4)
	s.c.BindTexture(glimpl.TEXTURE_2D, glimpl.Texture{})
	s.c.DisableVertexAttribArray(0)
	s.c.DisableVertexAttribArray(1)
	s.c.BindFramebuffer(glimpl.FRAMEBUFFER, s.frameBuffer)
	s.c.InvalidateFramebuffer(glimpl.FRAMEBUFFER, glimpl.COLOR_ATTACHMENT0)
	s.c.InvalidateFramebuffer(glimpl.FRAMEBUFFER, glimpl.DEPTH_ATTACHMENT)
	// The Android emulator requires framebuffer 0 bound at eglSwapBuffer time.
	// Bind the sRGB framebuffer again in afterPresent.
	s.c.BindFramebuffer(glimpl.FRAMEBUFFER, glimpl.Framebuffer{})
}

func (s *FBO) AfterPresent() {
	s.c.BindFramebuffer(glimpl.FRAMEBUFFER, s.frameBuffer)
}

func (s *FBO) Refresh(w, h int) error {
	s.width, s.height = w, h
	if w == 0 || h == 0 {
		return nil
	}
	s.c.BindTexture(glimpl.TEXTURE_2D, s.colorTex)
	if s.gl3 {
		s.c.TexImage2D(glimpl.TEXTURE_2D, 0, glimpl.SRGB8_ALPHA8, w, h, glimpl.RGBA, glimpl.UNSIGNED_BYTE, nil)
	} else /* EXT_sRGB */ {
		s.c.TexImage2D(glimpl.TEXTURE_2D, 0, glimpl.SRGB_ALPHA_EXT, w, h, glimpl.SRGB_ALPHA_EXT, glimpl.UNSIGNED_BYTE, nil)
	}
	currentRB := glimpl.Renderbuffer(s.c.GetBinding(glimpl.RENDERBUFFER_BINDING))
	s.c.BindRenderbuffer(glimpl.RENDERBUFFER, s.depthBuffer)
	s.c.RenderbufferStorage(glimpl.RENDERBUFFER, glimpl.DEPTH_COMPONENT16, w, h)
	s.c.BindRenderbuffer(glimpl.RENDERBUFFER, currentRB)
	s.c.BindFramebuffer(glimpl.FRAMEBUFFER, s.frameBuffer)
	s.c.FramebufferTexture2D(glimpl.FRAMEBUFFER, glimpl.COLOR_ATTACHMENT0, glimpl.TEXTURE_2D, s.colorTex, 0)
	s.c.FramebufferRenderbuffer(glimpl.FRAMEBUFFER, glimpl.DEPTH_ATTACHMENT, glimpl.RENDERBUFFER, s.depthBuffer)
	if st := s.c.CheckFramebufferStatus(glimpl.FRAMEBUFFER); st != glimpl.FRAMEBUFFER_COMPLETE {
		return fmt.Errorf("sRGB framebuffer incomplete (%dx%d), status: %#x error: %x", s.width, s.height, st, s.c.GetError())
	}

	if runtime.GOOS == "js" {
		// With macOS Safari, rendering to and then reading from a SRGB8_ALPHA8
		// texture result in twice gamma corrected colors. Using a plain RGBA
		// texture seems to work.
		s.c.ClearColor(.5, .5, .5, 1.0)
		s.c.Clear(glimpl.COLOR_BUFFER_BIT)
		var pixel [4]byte
		s.c.ReadPixels(0, 0, 1, 1, glimpl.RGBA, glimpl.UNSIGNED_BYTE, pixel[:])
		if pixel[0] == 128 { // Correct sRGB color value is ~188
			s.c.TexImage2D(glimpl.TEXTURE_2D, 0, glimpl.RGBA, w, h, glimpl.RGBA, glimpl.UNSIGNED_BYTE, nil)
			if st := s.c.CheckFramebufferStatus(glimpl.FRAMEBUFFER); st != glimpl.FRAMEBUFFER_COMPLETE {
				return fmt.Errorf("fallback RGBA framebuffer incomplete (%dx%d), status: %#x error: %x", s.width, s.height, st, s.c.GetError())
			}
		}
	}

	return nil
}

func (s *FBO) Release() {
	s.c.DeleteFramebuffer(s.frameBuffer)
	s.c.DeleteTexture(s.colorTex)
	s.c.DeleteRenderbuffer(s.depthBuffer)
	if s.blitted {
		s.c.DeleteBuffer(s.quad)
		s.c.DeleteProgram(s.prog)
	}
	s.c = nil
}

const (
	blitVSrc = `
#version 100

precision highp float;

attribute vec2 pos;
attribute vec2 uv;

varying vec2 vUV;

void main() {
    gl_Position = vec4(pos, 0, 1);
    vUV = uv;
}
`
	blitFSrc = `
#version 100

precision mediump float;

uniform sampler2D tex;
varying vec2 vUV;

vec3 gamma(vec3 rgb) {
	vec3 exp = vec3(1.055)*pow(rgb, vec3(0.41666)) - vec3(0.055);
	vec3 lin = rgb * vec3(12.92);
	bvec3 cut = lessThan(rgb, vec3(0.0031308));
	return vec3(cut.r ? lin.r : exp.r, cut.g ? lin.g : exp.g, cut.b ? lin.b : exp.b);
}

void main() {
    vec4 col = texture2D(tex, vUV);
	vec3 rgb = col.rgb;
	rgb = gamma(rgb);
	gl_FragColor = vec4(rgb, col.a);
}
`
)
