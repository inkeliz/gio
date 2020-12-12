// SPDX-License-Identifier: Unlicense OR MIT

package glimpl

import (
	"errors"
	"strings"
	"syscall/js"
	_ "unsafe"
)

type Functions struct {
	Ctx Context
	// Ref is the index/reference of the Ctx on the Javascript array.
	Ref uint32
}

type Context js.Value

func NewFunctions(ctx Context) (*Functions, error) {
	f := &Functions{Ctx: ctx}
	if err := f.Init(); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *Functions) Init() error {
	response := initGL(f.Ctx)
	// Javascript can't return multiple values and we can't pass object from JS to wasm. So,
	// one single uint64 is used to store two information:
	// `ptr` is the index/"pointer" to the JS context.
	// `err` is the error code.
	ref, err := response>>32, response&0xFFFFFFFF
	switch err {
	case 0: // Success
		f.Ref = uint32(ref)
		return nil
	default:
		return errors.New("gl: unknown error")
	case 1:
		return errors.New("gl: no support for neither EXT_color_buffer_half_float nor EXT_color_buffer_float")
	case 2:
		return errors.New("gl: no support for neither OES_texture_half_float nor OES_texture_float")
	case 3:
		return errors.New("gl: no support for EXT_sRGB")
	}
}
func (f *Functions) ActiveTexture(t Enum) {
	activeTexture(f.Ref, t)
}
func (f *Functions) AttachShader(p Program, s Shader) {
	attachShader(f.Ref, p, s)
}
func (f *Functions) BeginQuery(target Enum, query Query) {
	beginQuery(f.Ref, target, query)
}
func (f *Functions) BindAttribLocation(p Program, a Attrib, name string) {
	bindAttribLocation(f.Ref, p, a, name)
}
func (f *Functions) BindBuffer(target Enum, b Buffer) {
	bindBuffer(f.Ref, target, b)
}
func (f *Functions) BindBufferBase(target Enum, index int, b Buffer) {
	bindBufferBase(f.Ref, target, index, b)
}
func (f *Functions) BindFramebuffer(target Enum, fb Framebuffer) {
	bindFramebuffer(f.Ref, target, fb)
}
func (f *Functions) BindRenderbuffer(target Enum, rb Renderbuffer) {
	bindRenderbuffer(f.Ref, target, rb)
}
func (f *Functions) BindTexture(target Enum, t Texture) {
	bindTexture(f.Ref, target, t)
}
func (f *Functions) BlendEquation(mode Enum) {
	blendEquation(f.Ref, mode)
}
func (f *Functions) BlendFunc(sfactor, dfactor Enum) {
	blendFunc(f.Ref, sfactor, dfactor)
}
func (f *Functions) BufferData(target Enum, src []byte, usage Enum) {
	bufferData(f.Ref, target, src, usage)
}
func (f *Functions) CheckFramebufferStatus(target Enum) Enum {
	return Enum(makeValue(checkFramebufferStatus(f.Ref, target)).Int())
}
func (f *Functions) Clear(mask Enum) {
	clear(f.Ref, mask)
}
func (f *Functions) ClearColor(red, green, blue, alpha float32) {
	clearColor(f.Ref, float64(red), float64(green), float64(blue), float64(alpha))
}
func (f *Functions) ClearDepthf(d float32) {
	clearDepthf(f.Ref, float64(d))
}
func (f *Functions) CompileShader(s Shader) {
	compileShader(f.Ref, s)
}
func (f *Functions) CreateBuffer() Buffer {
	return Buffer(makeValue(createBuffer(f.Ref)))
}
func (f *Functions) CreateFramebuffer() Framebuffer {
	return Framebuffer(makeValue(createFramebuffer(f.Ref)))
}
func (f *Functions) CreateProgram() Program {
	return Program(makeValue(createProgram(f.Ref)))
}
func (f *Functions) CreateQuery() Query {
	return Query(makeValue(createQuery(f.Ref)))
}
func (f *Functions) CreateRenderbuffer() Renderbuffer {
	return Renderbuffer(makeValue(createRenderbuffer(f.Ref)))
}
func (f *Functions) CreateShader(ty Enum) Shader {
	return Shader(makeValue(createShaders(f.Ref, ty)))
}
func (f *Functions) CreateTexture() Texture {
	return Texture(makeValue(createTexture(f.Ref)))
}
func (f *Functions) DeleteBuffer(v Buffer) {
	deleteBuffer(f.Ref, v)
}
func (f *Functions) DeleteFramebuffer(v Framebuffer) {
	deleteFramebuffer(f.Ref, v)
}
func (f *Functions) DeleteProgram(p Program) {
	deleteProgram(f.Ref, p)
}
func (f *Functions) DeleteQuery(query Query) {
	deleteQuery(f.Ref, query)
}
func (f *Functions) DeleteShader(s Shader) {
	deleteShader(f.Ref, s)
}
func (f *Functions) DeleteRenderbuffer(v Renderbuffer) {
	deleteRenderbuffer(f.Ref, v)
}
func (f *Functions) DeleteTexture(v Texture) {
	deleteTexture(f.Ref, v)
}
func (f *Functions) DepthFunc(fn Enum) {
	depthFunc(f.Ref, fn)
}
func (f *Functions) DepthMask(mask bool) {
	depthMask(f.Ref, mask)
}
func (f *Functions) DisableVertexAttribArray(a Attrib) {
	disableVertexAttribArray(f.Ref, a)
}
func (f *Functions) Disable(cap Enum) {
	disable(f.Ref, cap)
}
func (f *Functions) DrawArrays(mode Enum, first, count int) {
	drawArrays(f.Ref, mode, first, count)
}
func (f *Functions) DrawElements(mode Enum, count int, ty Enum, offset int) {
	drawElements(f.Ref, mode, count, ty, offset)
}
func (f *Functions) Enable(cap Enum) {
	enable(f.Ref, cap)
}
func (f *Functions) EnableVertexAttribArray(a Attrib) {
	enableVertexAttribArray(f.Ref, a)
}
func (f *Functions) EndQuery(target Enum) {
	endQuery(f.Ref, target)
}
func (f *Functions) Finish() {
	finish(f.Ref)
}
func (f *Functions) FramebufferRenderbuffer(target, attachment, renderbuffertarget Enum, renderbuffer Renderbuffer) {
	framebufferRenderbuffer(f.Ref, target, attachment, renderbuffertarget, renderbuffer)
}
func (f *Functions) FramebufferTexture2D(target, attachment, texTarget Enum, t Texture, level int) {
	framebufferTexture2D(f.Ref, target, attachment, texTarget, t, level)
}
func (f *Functions) GetError() Enum {
	// Avoid slow getError calls. See gio#179.
	return 0
}
func (f *Functions) GetRenderbufferParameteri(target, pname Enum) int {
	return paramVal(makeValue(getRenderbufferParameteri(f.Ref, target, pname)))
}
func (f *Functions) GetFramebufferAttachmentParameteri(target, attachment, pname Enum) int {
	return paramVal(makeValue(getFramebufferAttachmentParameteri(f.Ref, target, attachment, pname)))
}
func (f *Functions) GetBinding(pname Enum) Object {
	return Object(makeValue(getBinding(f.Ref, pname)))
}
func (f *Functions) GetInteger(pname Enum) int {
	return paramVal(makeValue(getInteger(f.Ref, pname)))
}
func (f *Functions) GetProgrami(p Program, pname Enum) int {
	return paramVal(makeValue(getProgrami(f.Ref, p, pname)))
}
func (f *Functions) GetProgramInfoLog(p Program) string {
	return makeValue(getProgramInfoLog(f.Ref, p)).String()
}
func (f *Functions) GetQueryObjectuiv(query Query, pname Enum) uint {
	return uint(paramVal(makeValue(getQueryObjectuiv(f.Ref, query, pname))))
}
func (f *Functions) GetShaderi(s Shader, pname Enum) int {
	return paramVal(makeValue(getShaderi(f.Ref, s, pname)))
}
func (f *Functions) GetShaderInfoLog(s Shader) string {
	return makeValue(getShaderInfoLog(f.Ref, s)).String()
}
func (f *Functions) GetString(pname Enum) string {
	switch pname {
	case EXTENSIONS:
		extsjs := makeValue(getString(f.Ref, 1, 0))
		var exts []string
		for i := 0; i < extsjs.Length(); i++ {
			exts = append(exts, "GL_"+extsjs.Index(i).String())
		}
		return strings.Join(exts, " ")
	default:
		return makeValue(getString(f.Ref, 0, pname)).String()
	}
}
func (f *Functions) GetUniformBlockIndex(p Program, name string) uint {
	return uint(paramVal(makeValue(getUniformBlockIndex(f.Ref, p, name))))
}
func (f *Functions) GetUniformLocation(p Program, name string) Uniform {
	return Uniform(makeValue(getUniformLocation(f.Ref, p, name)))
}
func (f *Functions) InvalidateFramebuffer(target, attachment Enum) {
	invalidateFramebuffer(f.Ref, target, attachment)
}
func (f *Functions) LinkProgram(p Program) {
	linkProgram(f.Ref, p)
}
func (f *Functions) PixelStorei(pname Enum, param int32) {
	pixelStorei(f.Ref, pname, param)
}
func (f *Functions) RenderbufferStorage(target, internalformat Enum, width, height int) {
	renderbufferStorage(f.Ref, target, internalformat, width, height)
}
func (f *Functions) ReadPixels(x, y, width, height int, format, ty Enum, data []byte) {
	readPixels(f.Ref, x, y, width, height, format, ty, data)
}
func (f *Functions) Scissor(x, y, width, height int32) {
	scissor(f.Ref, x, y, width, height)
}
func (f *Functions) ShaderSource(s Shader, src string) {
	shaderSource(f.Ref, s, src)
}
func (f *Functions) TexImage2D(target Enum, level int, internalFormat int, width, height int, format, ty Enum, data []byte) {
	texImage2D(f.Ref, target, level, internalFormat, width, height, format, ty, data)
}
func (f *Functions) TexSubImage2D(target Enum, level int, x, y, width, height int, format, ty Enum, data []byte) {
	texSubImage2D(f.Ref, target, level, x, y, width, height, format, ty, data)
}
func (f *Functions) TexParameteri(target, pname Enum, param int) {
	texParameteri(f.Ref, target, pname, param)
}
func (f *Functions) UniformBlockBinding(p Program, blockIndex uint, blockBinding uint) {
	uniformBlockBinding(f.Ref, p, blockIndex, blockBinding)
}
func (f *Functions) Uniform1i(dst Uniform, v int) {
	uniform1i(f.Ref, dst, v)
}
func (f *Functions) Uniform1f(dst Uniform, v0 float32) {
	uniformXf(f.Ref, 1, dst, float64(v0), 0, 0, 0)
}
func (f *Functions) Uniform2f(dst Uniform, v0, v1 float32) {
	uniformXf(f.Ref, 2, dst, float64(v0), float64(v1), 0, 0)
}
func (f *Functions) Uniform3f(dst Uniform, v0, v1, v2 float32) {
	uniformXf(f.Ref, 3, dst, float64(v0), float64(v1), float64(v2), 0)
}
func (f *Functions) Uniform4f(dst Uniform, v0, v1, v2, v3 float32) {
	uniformXf(f.Ref, 4, dst, float64(v0), float64(v1), float64(v2), float64(v3))
}
func (f *Functions) UseProgram(p Program) {
	useProgram(f.Ref, p)
}
func (f *Functions) VertexAttribPointer(dst Attrib, size int, ty Enum, normalized bool, stride, offset int) {
	vertexAttribPointer(f.Ref, dst, size, ty, normalized, stride, offset)
}
func (f *Functions) Viewport(x, y, width, height int) {
	viewport(f.Ref, x, y, width, height)
}

func paramVal(v js.Value) int {
	switch v.Type() {
	case js.TypeBoolean:
		if b := v.Bool(); b {
			return 1
		} else {
			return 0
		}
	case js.TypeNumber:
		return v.Int()
	default:
		panic("unknown parameter type")
	}
}

//go:linkname makeValue syscall/js.makeValue
func makeValue(r uint64) js.Value

// Those functions are defined in `gl_js.js` and `gl_js.s`.
func initGL(ctx Context) uint64
func activeTexture(ref uint32, t Enum)
func attachShader(ref uint32, p Program, s Shader)
func beginQuery(ref uint32, target Enum, query Query)
func bindAttribLocation(ref uint32, p Program, a Attrib, name string)
func bindBuffer(ref uint32, target Enum, b Buffer)
func bindBufferBase(ref uint32, target Enum, index int, b Buffer)
func bindFramebuffer(ref uint32, target Enum, fb Framebuffer)
func bindRenderbuffer(ref uint32, target Enum, rb Renderbuffer)
func bindTexture(ref uint32, target Enum, t Texture)
func blendEquation(ref uint32, mode Enum)
func blendFunc(ref uint32, sfactor, dfactor Enum)
func bufferData(ref uint32, target Enum, src []byte, usage Enum)
func checkFramebufferStatus(ref uint32, target Enum) uint64
func clear(ref uint32, mask Enum)
func clearColor(ref uint32, red, green, blue, alpha float64)
func clearDepthf(ref uint32, d float64)
func compileShader(ref uint32, s Shader)
func createBuffer(ref uint32) uint64
func createFramebuffer(ref uint32) uint64
func createProgram(ref uint32) uint64
func createQuery(ref uint32) uint64
func createRenderbuffer(ref uint32) uint64
func createShaders(ref uint32, ty Enum) uint64
func createTexture(ref uint32) uint64
func deleteBuffer(ref uint32, v Buffer)
func deleteFramebuffer(ref uint32, v Framebuffer)
func deleteProgram(ref uint32, p Program)
func deleteQuery(ref uint32, query Query)
func deleteShader(ref uint32, s Shader)
func deleteRenderbuffer(ref uint32, v Renderbuffer)
func deleteTexture(ref uint32, v Texture)
func depthFunc(ref uint32, fn Enum)
func depthMask(ref uint32, mask bool)
func disableVertexAttribArray(ref uint32, a Attrib)
func disable(ref uint32, cap Enum)
func drawArrays(ref uint32, mode Enum, first, count int)
func drawElements(ref uint32, mode Enum, count int, ty Enum, offset int)
func enable(ref uint32, cap Enum)
func enableVertexAttribArray(ref uint32, a Attrib)
func endQuery(ref uint32, target Enum)
func finish(ref uint32)
func framebufferRenderbuffer(ref uint32, target, attachment, renderbuffertarget Enum, renderbuffer Renderbuffer)
func framebufferTexture2D(ref uint32, target, attachment, texTarget Enum, t Texture, level int)
func getRenderbufferParameteri(ref uint32, target, pname Enum) uint64
func getFramebufferAttachmentParameteri(ref uint32, target, attachment, pname Enum) uint64
func getBinding(ref uint32, pname Enum) uint64
func getInteger(ref uint32, pname Enum) uint64
func getProgrami(ref uint32, p Program, pname Enum) uint64
func getProgramInfoLog(ref uint32, p Program) uint64
func getQueryObjectuiv(ref uint32, query Query, pname Enum) uint64
func getShaderi(ref uint32, s Shader, pname Enum) uint64
func getShaderInfoLog(ref uint32, s Shader) uint64
func getString(ref uint32, method int, pname Enum) uint64
func getUniformBlockIndex(ref uint32, p Program, name string) uint64
func getUniformLocation(ref uint32, p Program, name string) uint64
func invalidateFramebuffer(ref uint32, target, attachment Enum)
func linkProgram(ref uint32, p Program)
func pixelStorei(ref uint32, pname Enum, param int32)
func renderbufferStorage(ref uint32, target, internalformat Enum, width, height int)
func readPixels(ref uint32, x, y, width, height int, format, ty Enum, data []byte)
func scissor(ref uint32, x, y, width, height int32)
func shaderSource(ref uint32, s Shader, src string)
func texImage2D(ref uint32, target Enum, level int, internalFormat int, width, height int, format, ty Enum, data []byte)
func texSubImage2D(ref uint32, target Enum, level int, x, y, width, height int, format, ty Enum, data []byte)
func texParameteri(ref uint32, target, pname Enum, param int)
func uniformBlockBinding(ref uint32, p Program, uniformBlockIndex uint, uniformBlockBinding uint)
func uniform1i(ref uint32, dst Uniform, v int)
func uniformXf(ref uint32, x int, dst Uniform, v0, v1, v2, v3 float64)
func useProgram(ref uint32, p Program)
func vertexAttribPointer(ref uint32, dst Attrib, size int, ty Enum, normalized bool, stride, offset int)
func viewport(ref uint32, x, y, width, height int)
