// SPDX-License-Identifier: Unlicense OR MIT

#include "go_asm.h"
#include "textflag.h"

TEXT ·initGL(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·activeTexture(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·attachShader(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·beginQuery(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·bindAttribLocation(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·bindBuffer(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·bindBufferBase(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·bindFramebuffer(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·bindRenderbuffer(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·bindTexture(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·blendEquation(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·blendFunc(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·bufferData(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·checkFramebufferStatus(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·clear(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·clearColor(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·clearDepthf(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·compileShader(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·createBuffer(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·createFramebuffer(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·createProgram(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·createQuery(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·createRenderbuffer(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·createShaders(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·createTexture(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·deleteBuffer(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·deleteFramebuffer(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·deleteProgram(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·deleteQuery(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·deleteShader(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·deleteRenderbuffer(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·deleteTexture(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·depthFunc(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·depthMask(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·disableVertexAttribArray(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·disable(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·drawArrays(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·drawElements(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·enable(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·enableVertexAttribArray(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·endQuery(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·finish(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·framebufferRenderbuffer(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·framebufferTexture2D(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·getUniformLocation(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·getUniformBlockIndex(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·getFramebufferAttachmentParameteri(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·getBinding(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·getInteger(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·getProgrami(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·getProgramInfoLog(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·getQueryObjectuiv(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·getShaderi(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·getShaderInfoLog(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·getString(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·getRenderbufferParameteri(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·invalidateFramebuffer(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·linkProgram(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·pixelStorei(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·renderbufferStorage(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·readPixels(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·scissor(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·shaderSource(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·texImage2D(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·texSubImage2D(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·texParameteri(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·uniformBlockBinding(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·uniform1i(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·uniformXf(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·useProgram(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·vertexAttribPointer(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·viewport(SB), NOSPLIT, $0
  CallImport
  RET

TEXT wasm_export_test(SB),NOSPLIT,$0
    I32Const $0
    Call glimpl·test(SB)

    Return
