// SPDX-License-Identifier: Unlicense OR MIT

package glimpl

import (
	"errors"
	"strings"
	"syscall/js"
	_ "unsafe"
)

type Functions struct {
	Ctx                             js.Value
	EXT_disjoint_timer_query        js.Value
	EXT_disjoint_timer_query_webgl2 js.Value
}

type Context js.Value

func NewFunctions(ctx Context) (*Functions, error) {
	f := &Functions{Ctx: js.Value(ctx)}
	if err := f.Init(); err != nil {
		return nil, err
	}
	return f, nil
}
func initGL(ctx js.Value)
func (f *Functions) Init() error {
	webgl2Class := js.Global().Get("WebGL2RenderingContext")
	iswebgl2 := !webgl2Class.IsUndefined() && f.Ctx.InstanceOf(webgl2Class)
	if !iswebgl2 {
		f.EXT_disjoint_timer_query = f.getExtension("EXT_disjoint_timer_query")
		if f.getExtension("OES_texture_half_float").IsNull() && f.getExtension("OES_texture_float").IsNull() {
			return errors.New("gl: no support for neither OES_texture_half_float nor OES_texture_float")
		}
		if f.getExtension("EXT_sRGB").IsNull() {
			return errors.New("gl: EXT_sRGB not supported")
		}
	} else {
		// WebGL2 extensions.
		f.EXT_disjoint_timer_query_webgl2 = f.getExtension("EXT_disjoint_timer_query_webgl2")
		if f.getExtension("EXT_color_buffer_half_float").IsNull() && f.getExtension("EXT_color_buffer_float").IsNull() {
			return errors.New("gl: no support for neither EXT_color_buffer_half_float nor EXT_color_buffer_float")
		}
	}
	initGL(f.Ctx)
	return nil
}
func (f *Functions) getExtension(name string) js.Value {
	return f.Ctx.Call("getExtension", name)
}
func activeTexture(t Enum)
func (f *Functions) ActiveTexture(t Enum) {
	activeTexture(t)
}
func attachShader(p Program, s Shader)
func (f *Functions) AttachShader(p Program, s Shader) {
	attachShader(p, s)
}
func beginQuery(method int, ctx js.Value, target Enum, query Query)
func (f *Functions) BeginQuery(target Enum, query Query) {
	if !f.EXT_disjoint_timer_query_webgl2.IsNull() {
		beginQuery(1, f.Ctx, target, query)
	} else {
		beginQuery(0, f.EXT_disjoint_timer_query, target, query)
	}
}
func bindAttribLocation(p Program, a Attrib, name string)
func (f *Functions) BindAttribLocation(p Program, a Attrib, name string) {
	bindAttribLocation(p, a, name)
}
func bindBuffer(target Enum, b Buffer)
func (f *Functions) BindBuffer(target Enum, b Buffer) {
	bindBuffer(target, b)
}
func bindBufferBase(target Enum, index int, b Buffer)
func (f *Functions) BindBufferBase(target Enum, index int, b Buffer) {
	bindBufferBase(target, index, b)
}
func bindFramebuffer(target Enum, fb Framebuffer)
func (f *Functions) BindFramebuffer(target Enum, fb Framebuffer) {
	bindFramebuffer(target, fb)
}
func bindRenderbuffer(target Enum, rb Renderbuffer)
func (f *Functions) BindRenderbuffer(target Enum, rb Renderbuffer) {
	bindRenderbuffer(target, rb)
}
func bindTexture(target Enum, t Texture)
func (f *Functions) BindTexture(target Enum, t Texture) {
	bindTexture(target, t)
}
func blendEquation(mode Enum)
func (f *Functions) BlendEquation(mode Enum) {
	blendEquation(mode)
}
func blendFunc(sfactor, dfactor Enum)
func (f *Functions) BlendFunc(sfactor, dfactor Enum) {
	blendFunc(sfactor, dfactor)
}
func bufferData(target Enum, src []byte, usage Enum)
func (f *Functions) BufferData(target Enum, src []byte, usage Enum) {
	bufferData(target, src, usage)
}
func checkFramebufferStatus(target Enum) uint64
func (f *Functions) CheckFramebufferStatus(target Enum) Enum {
	return Enum(makeValue(checkFramebufferStatus(target)).Int())
}
func clear(mask Enum)
func (f *Functions) Clear(mask Enum) {
	clear(mask)
}
func clearColor(red, green, blue, alpha float64)
func (f *Functions) ClearColor(red, green, blue, alpha float32) {
	clearColor(float64(red), float64(green), float64(blue), float64(alpha))
}
func clearDepthf(d float64)
func (f *Functions) ClearDepthf(d float32) {
	clearDepthf(float64(d))
}
func compileShader(s Shader)
func (f *Functions) CompileShader(s Shader) {
	compileShader(s)
}
func createBuffer() uint64
func (f *Functions) CreateBuffer() Buffer {
	return Buffer(makeValue(createBuffer()))
}
func createFramebuffer() uint64
func (f *Functions) CreateFramebuffer() Framebuffer {
	return Framebuffer(makeValue(createFramebuffer()))
}
func createProgram() uint64
func (f *Functions) CreateProgram() Program {
	return Program(makeValue(createProgram()))
}
func createQuery() uint64
func (f *Functions) CreateQuery() Query {
	return Query(makeValue(createQuery()))
}
func createRenderbuffer() uint64
func (f *Functions) CreateRenderbuffer() Renderbuffer {
	return Renderbuffer(makeValue(createRenderbuffer()))
}
func createShaders(ty Enum) uint64
func (f *Functions) CreateShader(ty Enum) Shader {
	return Shader(makeValue(createShaders(ty)))
}
func createTexture() uint64
func (f *Functions) CreateTexture() Texture {
	return Texture(makeValue(createTexture()))
}
func deleteBuffer(v Buffer)
func (f *Functions) DeleteBuffer(v Buffer) {
	deleteBuffer(v)
}
func deleteFramebuffer(v Framebuffer)
func (f *Functions) DeleteFramebuffer(v Framebuffer) {
	deleteFramebuffer(v)
}
func deleteProgram(p Program)
func (f *Functions) DeleteProgram(p Program) {
	deleteProgram(p)
}
func deleteQuery(method int, ctx js.Value, query Query)
func (f *Functions) DeleteQuery(query Query) {
	if !f.EXT_disjoint_timer_query_webgl2.IsNull() {
		deleteQuery(1, f.Ctx, query)
	} else {
		deleteQuery(0, f.EXT_disjoint_timer_query, query)
	}
}
func deleteShader(s Shader)
func (f *Functions) DeleteShader(s Shader) {
	deleteShader(s)
}
func deleteRenderbuffer(v Renderbuffer)
func (f *Functions) DeleteRenderbuffer(v Renderbuffer) {
	deleteRenderbuffer(v)
}
func deleteTexture(v Texture)
func (f *Functions) DeleteTexture(v Texture) {
	deleteTexture(v)
}
func depthFunc(fn Enum)
func (f *Functions) DepthFunc(fn Enum) {
	depthFunc(fn)
}
func depthMask(mask bool)
func (f *Functions) DepthMask(mask bool) {
	depthMask(mask)
}
func disableVertexAttribArray(a Attrib)
func (f *Functions) DisableVertexAttribArray(a Attrib) {
	disableVertexAttribArray(a)
}
func disable(cap Enum)
func (f *Functions) Disable(cap Enum) {
	disable(cap)
}
func drawArrays(mode Enum, first, count int)
func (f *Functions) DrawArrays(mode Enum, first, count int) {
	drawArrays(mode, first, count)
}
func drawElements(mode Enum, count int, ty Enum, offset int)
func (f *Functions) DrawElements(mode Enum, count int, ty Enum, offset int) {
	drawElements(mode, count, ty, offset)
}
func enable(cap Enum)
func (f *Functions) Enable(cap Enum) {
	enable(cap)
}
func enableVertexAttribArray(a Attrib)
func (f *Functions) EnableVertexAttribArray(a Attrib) {
	enableVertexAttribArray(a)
}
func endQuery(method int, ctx js.Value, target Enum)
func (f *Functions) EndQuery(target Enum) {
	if !f.EXT_disjoint_timer_query_webgl2.IsNull() {
		endQuery(1, f.Ctx, target)
	} else {
		endQuery(0, f.EXT_disjoint_timer_query, target)
	}
}
func finish()
func (f *Functions) Finish() {
	finish()
}
func framebufferRenderbuffer(target, attachment, renderbuffertarget Enum, renderbuffer Renderbuffer)
func (f *Functions) FramebufferRenderbuffer(target, attachment, renderbuffertarget Enum, renderbuffer Renderbuffer) {
	framebufferRenderbuffer(target, attachment, renderbuffertarget, renderbuffer)
}
func framebufferTexture2D(target, attachment, texTarget Enum, t Texture, level int)
func (f *Functions) FramebufferTexture2D(target, attachment, texTarget Enum, t Texture, level int) {
	framebufferTexture2D(target, attachment, texTarget, t, level)
}
func (f *Functions) GetError() Enum {
	// Avoid slow getError calls. See gio#179.
	return 0
}
func getRenderbufferParameteri(target, pname Enum) uint64
func (f *Functions) GetRenderbufferParameteri(target, pname Enum) int {
	return paramVal(makeValue(getRenderbufferParameteri(target, pname)))
}
func getFramebufferAttachmentParameteri(target, attachment, pname Enum) uint64
func (f *Functions) GetFramebufferAttachmentParameteri(target, attachment, pname Enum) int {
	return paramVal(makeValue(getFramebufferAttachmentParameteri(target, attachment, pname)))
}
func getBinding(pname Enum) uint64
func (f *Functions) GetBinding(pname Enum) Object {
	return Object(makeValue(getBinding(pname)))
}
func getInteger(pname Enum) uint64
func (f *Functions) GetInteger(pname Enum) int {
	return paramVal(makeValue(getInteger(pname)))
}
func getProgrami(p Program, pname Enum) uint64
func (f *Functions) GetProgrami(p Program, pname Enum) int {
	return paramVal(makeValue(getProgrami(p, pname)))
}
func getProgramInfoLog(p Program) uint64
func (f *Functions) GetProgramInfoLog(p Program) string {
	return makeValue(getProgramInfoLog(p)).String()
}
func getQueryObjectuiv(method int, ctx js.Value, query Query, pname Enum) uint64
func (f *Functions) GetQueryObjectuiv(query Query, pname Enum) uint {
	if !f.EXT_disjoint_timer_query_webgl2.IsNull() {
		return uint(paramVal(makeValue(getQueryObjectuiv(1, f.Ctx, query, pname))))
	} else {
		return uint(paramVal(makeValue(getQueryObjectuiv(0, f.EXT_disjoint_timer_query, query, pname))))
	}
}
func getShaderi(s Shader, pname Enum) uint64
func (f *Functions) GetShaderi(s Shader, pname Enum) int {
	return paramVal(makeValue(getShaderi(s, pname)))
}
func getShaderInfoLog(s Shader) uint64
func (f *Functions) GetShaderInfoLog(s Shader) string {
	return makeValue(getShaderInfoLog(s)).String()
}
func getString(method int, pname Enum) uint64
func (f *Functions) GetString(pname Enum) string {
	switch pname {
	case EXTENSIONS:
		extsjs := makeValue(getString(1, 0))
		var exts []string
		for i := 0; i < extsjs.Length(); i++ {
			exts = append(exts, "GL_"+extsjs.Index(i).String())
		}
		return strings.Join(exts, " ")
	default:
		return makeValue(getString(0, pname)).String()
	}
}
func getUniformBlockIndex(p Program, name string) uint64
func (f *Functions) GetUniformBlockIndex(p Program, name string) uint {
	return uint(paramVal(makeValue(getUniformBlockIndex(p, name))))
}
func getUniformLocation(p Program, name string) uint64
func (f *Functions) GetUniformLocation(p Program, name string) Uniform {
	return Uniform(makeValue(getUniformLocation(p, name)))
}
func invalidateFramebuffer(target, attachment Enum)
func (f *Functions) InvalidateFramebuffer(target, attachment Enum) {
	invalidateFramebuffer(target, attachment)
}
func linkProgram(p Program)
func (f *Functions) LinkProgram(p Program) {
	linkProgram(p)
}
func pixelStorei(pname Enum, param int32)
func (f *Functions) PixelStorei(pname Enum, param int32) {
	pixelStorei(pname, param)
}
func renderbufferStorage(target, internalformat Enum, width, height int)
func (f *Functions) RenderbufferStorage(target, internalformat Enum, width, height int) {
	renderbufferStorage(target, internalformat, width, height)
}
func readPixels(x, y, width, height int, format, ty Enum, data []byte)
func (f *Functions) ReadPixels(x, y, width, height int, format, ty Enum, data []byte) {
	readPixels(x, y, width, height, format, ty, data)
}
func scissor(x, y, width, height int32)
func (f *Functions) Scissor(x, y, width, height int32) {
	scissor(x, y, width, height)
}
func shaderSource(s Shader, src string)
func (f *Functions) ShaderSource(s Shader, src string) {
	shaderSource(s, src)
}
func texImage2D(target Enum, level int, internalFormat int, width, height int, format, ty Enum, data []byte)
func (f *Functions) TexImage2D(target Enum, level int, internalFormat int, width, height int, format, ty Enum, data []byte) {
	texImage2D(target, level, internalFormat, width, height, format, ty, data)
}
func texSubImage2D(target Enum, level int, x, y, width, height int, format, ty Enum, data []byte)
func (f *Functions) TexSubImage2D(target Enum, level int, x, y, width, height int, format, ty Enum, data []byte) {
	texSubImage2D(target, level, x, y, width, height, format, ty, data)
}
func texParameteri(target, pname Enum, param int)
func (f *Functions) TexParameteri(target, pname Enum, param int) {
	texParameteri(target, pname, param)
}
func uniformBlockBinding(p Program, uniformBlockIndex uint, uniformBlockBinding uint)
func (f *Functions) UniformBlockBinding(p Program, blockIndex uint, blockBinding uint) {
	uniformBlockBinding(p, blockIndex, blockBinding)
}
func uniform1i(dst Uniform, v int)
func (f *Functions) Uniform1i(dst Uniform, v int) {
	uniform1i(dst, v)
}
func uniformXf(x int, dst Uniform, v0, v1, v2, v3 float64)
func (f *Functions) Uniform1f(dst Uniform, v0 float32) {
	uniformXf(1, dst, float64(v0), 0, 0, 0)
}
func (f *Functions) Uniform2f(dst Uniform, v0, v1 float32) {
	uniformXf(2, dst, float64(v0), float64(v1), 0, 0)
}
func (f *Functions) Uniform3f(dst Uniform, v0, v1, v2 float32) {
	uniformXf(3, dst, float64(v0), float64(v1), float64(v2), 0)
}
func (f *Functions) Uniform4f(dst Uniform, v0, v1, v2, v3 float32) {
	uniformXf(4, dst, float64(v0), float64(v1), float64(v2), float64(v3))
}
func useProgram(p Program)
func (f *Functions) UseProgram(p Program) {
	useProgram(p)
}
func vertexAttribPointer(dst Attrib, size int, ty Enum, normalized bool, stride, offset int)
func (f *Functions) VertexAttribPointer(dst Attrib, size int, ty Enum, normalized bool, stride, offset int) {
	vertexAttribPointer(dst, size, ty, normalized, stride, offset)
}
func viewport(x, y, width, height int)
func (f *Functions) Viewport(x, y, width, height int) {
	viewport(x, y, width, height)
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
