(() => {

    // webgl handles the WebGL context from canvas.
    // Type: WebGLRenderingContext or WebGL2RenderingContext or 0 if not initialized.
    let webgl = 0;

    // textDecoder holds the TextDecoder used for encode string
    let textDecoder = new TextDecoder("utf-8");

    // invalidateBuffer is re-use when you call invalidateBuffer().
    let invalidateBuffer = new Int32Array(1);

    // Offset* is the byte-size of each type (matches with Reflect.Sizeof()).
    const OffsetInt64 = 8;
    const OffsetFloat64 = 8;
    const OffsetJSValue = 16; // We receive the `js.Value` instead of `js.Value.ref`.
    const OffsetString = 16;
    const OffsetSlice = 24;

    const gioLoadInt64 = (addr) => {
        // @TODO Why not use getBigUint64 from the DataView directly?!
        return go.mem.getUint32(addr + 8, true) + go.mem.getInt32(addr + 12, true) * 4294967296;
    }
    const gioLoadInt32 = (addr) => {
        return go.mem.getUint32(addr + 8, true);
    }
    const gioLoadJSValue = (addr) => {
        const f = go.mem.getFloat64(addr + 8, true);
        if (f === 0) {
            return undefined;
        }
        if (!isNaN(f)) {
            return f;
        }

        return go._values[go.mem.getUint32(addr + 8, true)];
    }
    const gioLoadString = (addr) => {
        return textDecoder.decode(new DataView(go._inst.exports.mem.buffer, gioLoadInt64(addr), gioLoadInt64(addr + 8)));
    }
    const gioLoadSlice = (addr) => {
        const s = new Uint8Array(go._inst.exports.mem.buffer, gioLoadInt64(addr), gioLoadInt64(addr + 8))
        if (s.byteLength === 0) {
            return null
        }
        return s
    }
    const gioLoadFloat64 = (addr) => {
        return go.mem.getFloat64(addr + 8, true);
    }

    const gioSetValue = (addr, v) => {
        addr += 8

        if (typeof v === "number" && v !== 0) {
            if (isNaN(v)) {
                go.mem.setUint32(addr + 4, 0x7FF80000, true);
                go.mem.setUint32(addr, 0, true);
                return;
            }
            go.mem.setFloat64(addr, v, true);
            return;
        }

        if (v === undefined) {
            go.mem.setFloat64(addr, 0, true);
            return;
        }

        let id = go._ids.get(v);
        if (id === undefined) {
            id = go._idPool.pop();
            if (id === undefined) {
                id = go._values.length;
            }
            go._values[id] = v;
            go._goRefCounts[id] = 0;
            go._ids.set(v, id);
        }
        go._goRefCounts[id]++;
        let typeFlag = 0;
        switch (typeof v) {
            case "object":
                if (v !== null) {
                    typeFlag = 1;
                }
                break;
            case "string":
                typeFlag = 2;
                break;
            case "symbol":
                typeFlag = 3;
                break;
            case "function":
                typeFlag = 4;
                break;
        }
        go.mem.setUint32(addr + 4, 0x7FF80000 | typeFlag, true);
        go.mem.setUint32(addr, id, true);
    }

    Object.assign(go.importObject.go, {
        // init(ctx js.Value)
        "gioui.org/internal/glimpl.initGL": (sp) => {
            webgl = gioLoadJSValue(sp)
        },
        // activeTexture(t Enum)
        "gioui.org/internal/glimpl.activeTexture": (sp) => {
            webgl.activeTexture(
                gioLoadInt64(sp),
            );
        },
        // attachShader(p Program, s Shader)
        "gioui.org/internal/glimpl.attachShader": (sp) => {
            webgl.attachShader(
                gioLoadJSValue(sp),
                gioLoadJSValue(sp + OffsetJSValue),
            );
        },
        // beginQuery(target Enum, b Buffer)
        "gioui.org/internal/glimpl.beginQuery": (sp) => {
            if (gioLoadInt64(sp) === 1) {
                gioLoadJSValue(sp + OffsetInt64).beginQuery(
                    gioLoadInt64(sp + OffsetInt64 + OffsetJSValue),
                    gioLoadJSValue(sp + OffsetInt64 + OffsetJSValue + OffsetInt64),
                );
            } else {
                gioLoadJSValue(sp + OffsetInt64).beginQueryEXT(
                    gioLoadInt64(sp + OffsetInt64 + OffsetJSValue),
                    gioLoadJSValue(sp + OffsetInt64 + OffsetJSValue + OffsetInt64),
                );
            }
        },
        // bindAttribLocation(p Program, a Attrib, name string)
        "gioui.org/internal/glimpl.bindAttribLocation": (sp) => {
            webgl.bindAttribLocation(
                gioLoadJSValue(sp),
                gioLoadInt64(sp + OffsetJSValue),
                gioLoadString(sp + OffsetJSValue + OffsetInt64),
            );
        },
        // bindBuffer(target Enum, b Buffer)
        "gioui.org/internal/glimpl.bindBuffer": (sp) => {
            webgl.bindBuffer(
                gioLoadInt64(sp),
                gioLoadJSValue(sp + OffsetInt64),
            );
        },
        // bindBufferBase(target Enum, index int, b Buffer)
        "gioui.org/internal/glimpl.bindBufferBase": (sp) => {
            webgl.bindBufferBase(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadJSValue(sp + OffsetInt64 + OffsetInt64),
            );
        },
        // bindFramebuffer(target Enum, fb Framebuffer)
        "gioui.org/internal/glimpl.bindFramebuffer": (sp) => {
            webgl.bindFramebuffer(
                gioLoadInt64(sp),
                gioLoadJSValue(sp + OffsetInt64),
            );
        },
        // bindRenderbuffer(target Enum, rb Renderbuffer)
        "gioui.org/internal/glimpl.bindRenderbuffer": (sp) => {
            webgl.bindRenderbuffer(
                gioLoadInt64(sp),
                gioLoadJSValue(sp + OffsetInt64),
            );
        },
        // bindTexture(target Enum, t Texture)
        "gioui.org/internal/glimpl.bindTexture": (sp) => {
            webgl.bindTexture(
                gioLoadInt64(sp),
                gioLoadJSValue(sp + OffsetInt64),
            );
        },
        // blendEquation(mode Enum)
        "gioui.org/internal/glimpl.blendEquation": (sp) => {
            webgl.blendEquation(
                gioLoadInt64(sp),
            );
        },
        // blendFunc(sfactor, dfactor Enum)
        "gioui.org/internal/glimpl.blendFunc": (sp) => {
            webgl.blendFunc(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
            );
        },
        // bufferData (target Enum, src []byte, usage Enum)
        "gioui.org/internal/glimpl.bufferData": (sp) => {
            webgl.bufferData(
                gioLoadInt64(sp),
                gioLoadSlice(sp + OffsetInt64),
                gioLoadInt64(sp + OffsetInt64 + OffsetSlice),
            );
        },
        // checkFramebufferStatus(target Enum) int64
        "gioui.org/internal/glimpl.checkFramebufferStatus": (sp) => {
            const result = webgl.checkFramebufferStatus(
                gioLoadInt64(sp),
            );
            gioSetValue(sp + OffsetInt64, result)
        },
        // clear(mask Enum)
        "gioui.org/internal/glimpl.clear": (sp) => {
            webgl.clear(
                gioLoadInt64(sp),
            );
        },
        // clearColor(red, green, blue, alpha float32)
        "gioui.org/internal/glimpl.clearColor": (sp) => {
            webgl.clearColor(
                gioLoadFloat64(sp),
                gioLoadFloat64(sp + OffsetFloat64),
                gioLoadFloat64(sp + (OffsetFloat64 * 2)),
                gioLoadFloat64(sp + (OffsetFloat64 * 3)),
            );
        },
        // clearDepthf(d float32)
        "gioui.org/internal/glimpl.clearDepthf": (sp) => {
            webgl.clearDepth(
                gioLoadFloat64(sp),
            );
        },
        // compileShader(s Shader)
        "gioui.org/internal/glimpl.compileShader": (sp) => {
            webgl.compileShader(
                gioLoadJSValue(sp),
            );
        },
        // createBuffer() uint64
        "gioui.org/internal/glimpl.createBuffer": (sp) => {
            const result = webgl.createBuffer();
            gioSetValue(sp, result)
        },
        // createFramebuffer() uint64
        "gioui.org/internal/glimpl.createFramebuffer": (sp) => {
            const result = webgl.createFramebuffer();
            gioSetValue(sp, result)
        },
        // createProgram() uint64
        "gioui.org/internal/glimpl.createProgram": (sp) => {
            const result = webgl.createProgram();
            gioSetValue(sp, result)
        },
        // createQuery() uint64
        "gioui.org/internal/glimpl.createQuery": (sp) => {
            const result = webgl.createQuery();
            gioSetValue(sp, result)
        },
        // createRenderbuffer() uint64
        "gioui.org/internal/glimpl.createRenderbuffer": (sp) => {
            const result = webgl.createRenderbuffer();
            gioSetValue(sp, result)
        },
        // createShaders(ty Enum) uint64
        "gioui.org/internal/glimpl.createShaders": (sp) => {
            const result = webgl.createShader(
                gioLoadInt64(sp),
            );
            gioSetValue(sp + OffsetInt64, result)
        },
        // createTexture() uint64
        "gioui.org/internal/glimpl.createTexture": (sp) => {
            const result = webgl.createTexture();
            gioSetValue(sp, result)
        },
        // deleteBuffer(v Buffer)
        "gioui.org/internal/glimpl.deleteBuffer": (sp) => {
            webgl.deleteBuffer(
                gioLoadJSValue(sp),
            );
        },
        // deleteFramebuffer(v Framebuffer)
        "gioui.org/internal/glimpl.deleteFramebuffer": (sp) => {
            webgl.deleteFramebuffer(
                gioLoadJSValue(sp),
            );
        },
        // deleteProgram(p Program)
        "gioui.org/internal/glimpl.deleteProgram": (sp) => {
            webgl.deleteProgram(
                gioLoadJSValue(sp),
            );
        },
        // deleteQuery(method int, ctx js.Value, query Query)
        "gioui.org/internal/glimpl.deleteQuery": (sp) => {
            if (gioLoadInt64(sp) === 1) {
                gioLoadJSValue(sp + OffsetInt64).deleteQuery(
                    gioLoadJSValue(sp + OffsetInt64 + OffsetJSValue),
                );
            } else {
                gioLoadJSValue(sp + OffsetInt64).deleteQueryEXT(
                    gioLoadJSValue(sp + OffsetInt64 + OffsetJSValue),
                );
            }
        },
        // deleteShader(s Shader)
        "gioui.org/internal/glimpl.deleteShader": (sp) => {
            webgl.deleteShader(
                gioLoadJSValue(sp),
            );
        },
        // deleteRenderbuffer(v Renderbuffer)
        "gioui.org/internal/glimpl.deleteRenderbuffer": (sp) => {
            webgl.deleteRenderbuffer(
                gioLoadJSValue(sp),
            );
        },
        // deleteTexture(v Texture)
        "gioui.org/internal/glimpl.deleteTexture": (sp) => {
            webgl.deleteTexture(
                gioLoadJSValue(sp),
            );
        },
        // depthFunc(fn Enum)
        "gioui.org/internal/glimpl.depthFunc": (sp) => {
            webgl.depthFunc(
                gioLoadInt64(sp),
            );
        },
        // depthMask(mask bool)
        "gioui.org/internal/glimpl.depthMask": (sp) => {
            webgl.depthMask(
                gioLoadInt64(sp),
            );
        },
        // disableVertexAttribArray(a Attrib)
        "gioui.org/internal/glimpl.disableVertexAttribArray": (sp) => {
            webgl.disableVertexAttribArray(
                gioLoadInt64(sp),
            );
        },
        // disable(cap Enum)
        "gioui.org/internal/glimpl.disable": (sp) => {
            webgl.disable(
                gioLoadInt64(sp),
            );
        },
        // drawArrays(mode Enum, first, count int)
        "gioui.org/internal/glimpl.drawArrays": (sp) => {
            webgl.drawArrays(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2)),
            )
        },
        // drawElements(mode Enum, count int, ty Enum, offset int)
        "gioui.org/internal/glimpl.drawElements": (sp) => {
            webgl.drawElements(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2)),
                gioLoadInt64(sp + (OffsetInt64 * 3))
            )
        },
        // enable(cap Enum)
        "gioui.org/internal/glimpl.enable": (sp) => {
            webgl.enable(
                gioLoadInt64(sp),
            );
        },
        // enableVertexAttribArray(a Attrib)
        "gioui.org/internal/glimpl.enableVertexAttribArray": (sp) => {
            webgl.enableVertexAttribArray(
                gioLoadInt64(sp),
            );
        },
        // endQuery(method int, ctx js.Value, target Enum)
        "gioui.org/internal/glimpl.endQuery": (sp) => {
            if (gioLoadInt64(sp) === 1) {
                gioLoadJSValue(sp + OffsetInt64).endQuery(
                    gioLoadInt64(sp + OffsetInt64 + OffsetJSValue),
                );
            } else {
                gioLoadJSValue(sp + OffsetInt64).endQueryEXT(
                    gioLoadInt64(sp + OffsetInt64 + OffsetJSValue),
                );
            }
        },
        // finish()
        "gioui.org/internal/glimpl.finish": (sp) => {
            webgl.finish();
        },
        // framebufferRenderbuffer(target, attachment, renderbuffertarget Enum, renderbuffer Renderbuffer)
        "gioui.org/internal/glimpl.framebufferRenderbuffer": (sp) => {
            webgl.framebufferRenderbuffer(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2)),
                gioLoadJSValue(sp + (OffsetInt64 * 3)),
            );
        },
        // framebufferTexture2D(target, attachment, texTarget Enum, t Texture, level int)
        "gioui.org/internal/glimpl.framebufferTexture2D": (sp) => {
            webgl.framebufferTexture2D(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2)),
                gioLoadJSValue(sp + (OffsetInt64 * 3)),
                gioLoadInt64(sp + (OffsetInt64 * 3) + OffsetJSValue),
            );
        },
        // getRenderbufferParameteri(target, pname Enum) uint64
        "gioui.org/internal/glimpl.getRenderbufferParameteri": (sp) => {
            const result = webgl.getRenderbufferParameteri(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
            );
            gioSetValue(sp + (OffsetInt64 * 2), result)
        },
        // getFramebufferAttachmentParameteri(target, attachment, pname Enum) uint64
        "gioui.org/internal/glimpl.getFramebufferAttachmentParameteri": (sp) => {
            const result = webgl.getFramebufferAttachmentParameter(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2)),
            );
            gioSetValue(sp + (OffsetInt64 * 3), result)
        },
        // getBinding(pname Enum) uint64
        "gioui.org/internal/glimpl.getBinding": (sp) => {
            const result = webgl.getParameter(
                gioLoadInt64(sp),
            );
            gioSetValue(sp + OffsetInt64, result)
        },
        // getInteger(pname Enum) uint64
        "gioui.org/internal/glimpl.getInteger": (sp) => {
            const result = webgl.getParameter(
                gioLoadInt64(sp),
            );
            gioSetValue(sp + OffsetInt64, result)
        },
        // getProgrami(p Program, pname Enum) uint64
        "gioui.org/internal/glimpl.getProgrami": (sp) => {
            const result = webgl.getProgramParameter(
                gioLoadJSValue(sp),
                gioLoadInt64(sp + OffsetJSValue)
            );
            gioSetValue(sp + OffsetJSValue + OffsetInt64, result)
        },
        // getProgramInfoLog(p Program) uint64
        "gioui.org/internal/glimpl.getProgramInfoLog": (sp) => {
            const result = webgl.getProgramInfoLog(
                gioLoadJSValue(sp),
            );
            gioSetValue(sp + OffsetJSValue, result)
        },
        // getQueryObjectuiv(method int, ctx js.Value, query Query, pname Enum) uint64
        "gioui.org/internal/glimpl.getQueryObjectuiv": (sp) => {
            var result = 0;
            if (gioLoadInt64(sp) === 1) {
                result = gioLoadJSValue(sp + OffsetInt64).getQueryParameter(
                    gioLoadJSValue(sp + OffsetInt64 + OffsetJSValue),
                    gioLoadInt64(sp + OffsetInt64 + (OffsetJSValue * 2)),
                );
            } else {
                result = gioLoadJSValue(sp + OffsetInt64).getQueryObjectEXT(
                    gioLoadJSValue(sp + OffsetInt64 + OffsetJSValue),
                    gioLoadInt64(sp + OffsetInt64 + (OffsetJSValue * 2)),
                );
            }
            gioSetValue(sp + (OffsetInt64 * 2) + (OffsetJSValue * 2), result);
        },
        // getShaderi(s Shader, pname Enum) uint
        "gioui.org/internal/glimpl.getShaderi": (sp) => {
            const result = webgl.getShaderParameter(
                gioLoadJSValue(sp),
                gioLoadInt64(sp + OffsetJSValue),
            );
            gioSetValue(sp + OffsetJSValue + OffsetInt64, result)
        },
        // getShaderInfoLog(s Shader) uint64
        "gioui.org/internal/glimpl.getShaderInfoLog": (sp) => {
            const result = webgl.getShaderInfoLog(
                gioLoadJSValue(sp),
            );
            gioSetValue(sp + OffsetJSValue, result)
        },
        // getString(method int, pname Enum) uint64
        "gioui.org/internal/glimpl.getString": (sp) => {
            var result = 0;
            if (gioLoadInt64(sp) === 1) {
                result = webgl.getSupportedExtensions();
            } else {
                result = webgl.getParameter(gioLoadInt64(sp + OffsetInt64));
            }
            gioSetValue(sp + (OffsetInt64 * 2), result);
        },
        // getUniformBlockIndex(p Program, name string) uint64
        "gioui.org/internal/glimpl.getUniformBlockIndex": (sp) => {
            const result = webgl.getUniformBlockIndex(
                gioLoadJSValue(sp),
                gioLoadString(sp + OffsetJSValue),
            )
            gioSetValue(sp + OffsetJSValue + OffsetString, result)
        },
        // getUniformLocation(p Program, name string) uint64
        "gioui.org/internal/glimpl.getUniformLocation": (sp) => {
            const result = webgl.getUniformLocation(
                gioLoadJSValue(sp),
                gioLoadString(sp + OffsetJSValue),
            )
            gioSetValue(sp + OffsetJSValue + OffsetString, result)
        },
        // invalidateFramebuffer(target, attachment Enum)
        "gioui.org/internal/glimpl.invalidateFramebuffer": (sp) => {
            if (typeof webgl.invalidateFramebuffer === "undefined") {
                return
            }
            invalidateBuffer.set([gioLoadInt64(sp + OffsetInt64)]);
            webgl.invalidateFramebuffer(
                gioLoadInt64(sp),
                invalidateBuffer,
            );
        },
        // linkProgram(p Program)
        "gioui.org/internal/glimpl.linkProgram": (sp) => {
            webgl.linkProgram(
                gioLoadJSValue(sp),
            );
        },
        // pixelStorei(pname Enum, param int32)
        "gioui.org/internal/glimpl.pixelStorei": (sp) => {
            webgl.pixelStorei(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
            );
        },
        // renderbufferStorage(target, internalformat Enum, width, height int)
        "gioui.org/internal/glimpl.renderbufferStorage": (sp) => {
            webgl.renderbufferStorage(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2)),
                gioLoadInt64(sp + (OffsetInt64 * 3)),
            );
        },
        // readPixels(x, y, width, height int, format, ty Enum, data []byte)
        "gioui.org/internal/glimpl.readPixels": (sp) => {
            webgl.readPixels(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2)),
                gioLoadInt64(sp + (OffsetInt64 * 3)),
                gioLoadInt64(sp + (OffsetInt64 * 4)),
                gioLoadInt64(sp + (OffsetInt64 * 5)),
                gioLoadSlice(sp + (OffsetInt64 * 6)),
            )
        },
        // scissor(x, y, width, height int32)
        "gioui.org/internal/glimpl.scissor": (sp) => {
            webgl.scissor(
                gioLoadInt32(sp),
                gioLoadInt32(sp + OffsetInt64),
                gioLoadInt32(sp + (OffsetInt64 * 2)),
                gioLoadInt32(sp + (OffsetInt64 * 3)),
            )
        },
        // shaderSource(s Shader, src string)
        "gioui.org/internal/glimpl.shaderSource": (sp) => {
            webgl.shaderSource(
                gioLoadJSValue(sp),
                gioLoadString(sp + OffsetJSValue),
            )
        },
        // texImage2D(target Enum, level int, internalFormat int, width, height int, format, ty Enum, data []byte)
        "gioui.org/internal/glimpl.texImage2D": (sp) => {
            webgl.texImage2D(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2)),
                gioLoadInt64(sp + (OffsetInt64 * 3)),
                gioLoadInt64(sp + (OffsetInt64 * 4)),
                0,
                gioLoadInt64(sp + (OffsetInt64 * 5)),
                gioLoadInt64(sp + (OffsetInt64 * 6)),
                gioLoadSlice(sp + (OffsetInt64 * 7)),
            );
        },
        // texSubImage2D(target Enum, level int, x, y, width, height int, format, ty Enum, data []byte)
        "gioui.org/internal/glimpl.texSubImage2D": (sp) => {
            webgl.texSubImage2D(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2)),
                gioLoadInt64(sp + (OffsetInt64 * 3)),
                gioLoadInt64(sp + (OffsetInt64 * 4)),
                gioLoadInt64(sp + (OffsetInt64 * 5)),
                gioLoadInt64(sp + (OffsetInt64 * 6)),
                gioLoadInt64(sp + (OffsetInt64 * 7)),
                gioLoadSlice(sp + (OffsetInt64 * 8)),
            );
        },
        // texParameteri(target, pname Enum, param int)
        "gioui.org/internal/glimpl.texParameteri": (sp) => {
            webgl.texParameteri(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2))
            )
        },
        // uniformBlockBinding(p Program, uniformBlockIndex uint, uniformBlockBinding uint)
        "gioui.org/internal/glimpl.uniformBlockBinding": (sp) => {
            webgl.uniformBlockBinding(
                gioLoadJSValue(sp),
                gioLoadInt64(sp + OffsetJSValue),
                gioLoadInt64(sp + OffsetJSValue + OffsetInt64),
            )
        },
        // uniform1i(dst Uniform, v int)
        "gioui.org/internal/glimpl.uniform1i": (sp) => {
            webgl.uniform1i(
                gioLoadJSValue(sp),
                gioLoadInt64(sp + OffsetJSValue),
            )
        },
        // uniformXf(x int, dst Uniform, v0, v1, v2, v3 float32)
        "gioui.org/internal/glimpl.uniformXf": (sp) => {
            switch (gioLoadInt32(sp)) {
                case 1:
                    webgl.uniform1f(
                        gioLoadJSValue(sp + OffsetInt64),
                        gioLoadFloat64(sp + OffsetInt64 + OffsetJSValue),
                    );
                    return
                case 2:
                    webgl.uniform2f(
                        gioLoadJSValue(sp + OffsetInt64),
                        gioLoadFloat64(sp + OffsetInt64 + OffsetJSValue),
                        gioLoadFloat64(sp + OffsetInt64 + OffsetJSValue + OffsetFloat64),
                    );
                    return
                case 3:
                    webgl.uniform3f(
                        gioLoadJSValue(sp + OffsetInt64),
                        gioLoadFloat64(sp + OffsetInt64 + OffsetJSValue),
                        gioLoadFloat64(sp + OffsetInt64 + OffsetJSValue + OffsetFloat64),
                        gioLoadFloat64(sp + OffsetInt64 + OffsetJSValue + (OffsetFloat64 * 2)),
                    );
                    return
                case 4:
                    webgl.uniform4f(
                        gioLoadJSValue(sp + OffsetInt64),
                        gioLoadFloat64(sp + OffsetInt64 + OffsetJSValue),
                        gioLoadFloat64(sp + OffsetInt64 + OffsetJSValue + OffsetFloat64),
                        gioLoadFloat64(sp + OffsetInt64 + OffsetJSValue + (OffsetFloat64 * 2)),
                        gioLoadFloat64(sp + OffsetInt64 + OffsetJSValue + (OffsetFloat64 * 3)),
                    );
                    return
            }
        },

        // useProgram(p Program)
        "gioui.org/internal/glimpl.useProgram": (sp) => {
            webgl.useProgram(
                gioLoadJSValue(sp),
            )
        },
        // vertexAttribPointer(dst Attrib, size int, ty Enum, normalized bool, stride, offset int)
        "gioui.org/internal/glimpl.vertexAttribPointer": (sp) => {
            webgl.vertexAttribPointer(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2)),
                gioLoadInt64(sp + (OffsetInt64 * 3)),
                gioLoadInt64(sp + (OffsetInt64 * 4)),
                gioLoadInt64(sp + (OffsetInt64 * 5)),
                gioLoadInt64(sp + (OffsetInt64 * 6)),
            );
        },
        // viewport(x, y, width, height int)
        "gioui.org/internal/glimpl.viewport": (sp) => {
            webgl.viewport(
                gioLoadInt32(sp),
                gioLoadInt32(sp + OffsetInt64),
                gioLoadInt32(sp + (OffsetInt64 * 2)),
                gioLoadInt32(sp + (OffsetInt64 * 3)),
            );
        },
    })
})();