(() => {

    // webgl is the array which handles the WebGL context. See InitGL function.
    let webgl = [];

    // textDecoder holds the TextDecoder used for encode string.
    let textDecoder = new TextDecoder("utf-8");

    // invalidateBuffer is re-use when you call invalidateBuffer().
    let invalidateBuffer = new Int32Array(1);

    // Offset* is the byte-size of each type (matches with Reflect.Sizeof()).
    const OffsetContextIndex = 8;
    const OffsetInt64 = 8;
    const OffsetFloat64 = 8;
    const OffsetJSValue = 16; // We receive the `js.Value` instead of `js.Value.ref`.
    const OffsetString = 16;
    const OffsetSlice = 24;

    const gioLoadContext = (addr) => {
        return webgl[go.mem.getUint32(addr, true)];
    }
    const gioLoadInt64 = (addr) => {
        // bigInt doesn't work
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

    const gioSetJSValue = (addr, v) => {
        const nanHead = 0x7FF80000;
        addr += 8

        if (typeof v === "number" && v !== 0) {
            if (isNaN(v)) {
                go.mem.setUint32(addr + 4, nanHead, true);
                go.mem.setUint32(addr, 0, true);
                return;
            }
            go.mem.setFloat64(addr, v, true);
            return;
        }

        switch (v) {
            case undefined:
                // valueUndefined
                go.mem.setFloat64(addr, 0, true);
                return;
            case null:
                // valueNull
                go.mem.setUint32(addr + 4, nanHead, true);
                go.mem.setUint32(addr, 2, true);
                return;
            case true:
                // valueTrue
                go.mem.setUint32(addr + 4, nanHead, true);
                go.mem.setUint32(addr, 3, true);
                return;
            case false:
                // valueFalse
                go.mem.setUint32(addr + 4, nanHead, true);
                go.mem.setUint32(addr, 4, true);
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
        go.mem.setUint32(addr + 4, nanHead | typeFlag, true);
        go.mem.setUint32(addr, id, true);
    }
    const gioSetInt32Value = (addr, v) => {
        go.mem.setUint32(addr + 8, v, true);
    }

    Object.assign(go.importObject.go, {
        // init(ctx js.Value)
        "gioui.org/internal/glimpl.initGL": (sp) => {
            sp = (sp >>> 0);

            const context = gioLoadJSValue(sp);
            let error = 0
            if (typeof window.WebGL2RenderingContext !== "undefined" && context instanceof window.WebGL2RenderingContext) {
                const ext1 = context.getExtension("EXT_color_buffer_half_float");
                const ext2 = context.getExtension("EXT_color_buffer_float");
                if (ext1 === null && ext2 === null) {
                    error = 1;
                }
            } else {
                const ext1 = context.getExtension("OES_texture_half_float");
                const ext2 = context.getExtension("OES_texture_float");
                const ext3 = context.getExtension("EXT_sRGB");
                if (ext1 === null && ext2 === null) {
                    error = 2;
                }
                if (ext3 === null) {
                    error = 3;
                }
            }

            const ref = webgl.push({
                ctx: context,
                EXT_disjoint_timer_query: context.getExtension("EXT_disjoint_timer_query"),
                EXT_disjoint_timer_query_webgl2: context.getExtension("EXT_disjoint_timer_query_webgl2"),
            });

            sp = (go._inst.exports.getsp() >>> 0);
            gioSetInt32Value(sp + OffsetJSValue, error);
            gioSetInt32Value(sp + 4 + OffsetJSValue, ref - 1);
        },
        // activeTexture(t Enum)
        "gioui.org/internal/glimpl.activeTexture": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.activeTexture(
                gioLoadInt64(sp),
            );
        },
        // attachShader(p Program, s Shader)
        "gioui.org/internal/glimpl.attachShader": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.attachShader(
                gioLoadJSValue(sp),
                gioLoadJSValue(sp + OffsetJSValue),
            );
        },
        // beginQuery(target Enum, b Buffer)
        "gioui.org/internal/glimpl.beginQuery": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            if (webgl.EXT_disjoint_timer_query_webgl2 !== null) {
                webgl.ctx.beginQuery(
                    gioLoadInt64(sp),
                    gioLoadJSValue(sp + OffsetInt64),
                );
            } else {
                webgl.EXT_disjoint_timer_query.beginQueryEXT(
                    gioLoadInt64(sp),
                    gioLoadJSValue(sp + OffsetInt64),
                );
            }
        },
        // bindAttribLocation(p Program, a Attrib, name string)
        "gioui.org/internal/glimpl.bindAttribLocation": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.bindAttribLocation(
                gioLoadJSValue(sp),
                gioLoadInt64(sp + OffsetJSValue),
                gioLoadString(sp + OffsetJSValue + OffsetInt64),
            );
        },
        // bindBuffer(target Enum, b Buffer)
        "gioui.org/internal/glimpl.bindBuffer": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.bindBuffer(
                gioLoadInt64(sp),
                gioLoadJSValue(sp + OffsetInt64),
            );
        },
        // bindBufferBase(target Enum, index int, b Buffer)
        "gioui.org/internal/glimpl.bindBufferBase": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.bindBufferBase(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadJSValue(sp + OffsetInt64 + OffsetInt64),
            );
        },
        // bindFramebuffer(target Enum, fb Framebuffer)
        "gioui.org/internal/glimpl.bindFramebuffer": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.bindFramebuffer(
                gioLoadInt64(sp),
                gioLoadJSValue(sp + OffsetInt64),
            );
        },
        // bindRenderbuffer(target Enum, rb Renderbuffer)
        "gioui.org/internal/glimpl.bindRenderbuffer": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.bindRenderbuffer(
                gioLoadInt64(sp),
                gioLoadJSValue(sp + OffsetInt64),
            );
        },
        // bindTexture(target Enum, t Texture)
        "gioui.org/internal/glimpl.bindTexture": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.bindTexture(
                gioLoadInt64(sp),
                gioLoadJSValue(sp + OffsetInt64),
            );
        },
        // blendEquation(mode Enum)
        "gioui.org/internal/glimpl.blendEquation": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.blendEquation(
                gioLoadInt64(sp),
            );
        },
        // blendFunc(sfactor, dfactor Enum)
        "gioui.org/internal/glimpl.blendFunc": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.blendFunc(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
            );
        },
        // bufferData (target Enum, src []byte, usage Enum)
        "gioui.org/internal/glimpl.bufferData": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.bufferData(
                gioLoadInt64(sp),
                gioLoadSlice(sp + OffsetInt64),
                gioLoadInt64(sp + OffsetInt64 + OffsetSlice),
            );
        },
        // checkFramebufferStatus(target Enum) int64
        "gioui.org/internal/glimpl.checkFramebufferStatus": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.checkFramebufferStatus(
                gioLoadInt64(sp),
            );
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp + OffsetInt64, result);
        },
        // clear(mask Enum)
        "gioui.org/internal/glimpl.clear": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.clear(
                gioLoadInt64(sp),
            );
        },
        // clearColor(red, green, blue, alpha float32)
        "gioui.org/internal/glimpl.clearColor": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.clearColor(
                gioLoadFloat64(sp),
                gioLoadFloat64(sp + OffsetFloat64),
                gioLoadFloat64(sp + (OffsetFloat64 * 2)),
                gioLoadFloat64(sp + (OffsetFloat64 * 3)),
            );
        },
        // clearDepthf(d float32)
        "gioui.org/internal/glimpl.clearDepthf": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.clearDepth(
                gioLoadFloat64(sp),
            );
        },
        // compileShader(s Shader)
        "gioui.org/internal/glimpl.compileShader": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.compileShader(
                gioLoadJSValue(sp),
            );
        },
        // createBuffer() uint64
        "gioui.org/internal/glimpl.createBuffer": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.createBuffer();
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp, result);
        },
        // createFramebuffer() uint64
        "gioui.org/internal/glimpl.createFramebuffer": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.createFramebuffer();
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp, result);
        },
        // createProgram() uint64
        "gioui.org/internal/glimpl.createProgram": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.createProgram();
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp, result);
        },
        // createQuery() uint64
        "gioui.org/internal/glimpl.createQuery": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.createQuery();
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp, result);
        },
        // createRenderbuffer() uint64
        "gioui.org/internal/glimpl.createRenderbuffer": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.createRenderbuffer();
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp, result);
        },
        // createShaders(ty Enum) uint64
        "gioui.org/internal/glimpl.createShaders": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.createShader(
                gioLoadInt64(sp),
            );
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp + OffsetInt64, result);
        },
        // createTexture() uint64
        "gioui.org/internal/glimpl.createTexture": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.createTexture();
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp, result);
        },
        // deleteBuffer(v Buffer)
        "gioui.org/internal/glimpl.deleteBuffer": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.deleteBuffer(
                gioLoadJSValue(sp),
            );
        },
        // deleteFramebuffer(v Framebuffer)
        "gioui.org/internal/glimpl.deleteFramebuffer": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.deleteFramebuffer(
                gioLoadJSValue(sp),
            );
        },
        // deleteProgram(p Program)
        "gioui.org/internal/glimpl.deleteProgram": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.deleteProgram(
                gioLoadJSValue(sp),
            );
        },
        // deleteQuery(query Query)
        "gioui.org/internal/glimpl.deleteQuery": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            if (webgl.EXT_disjoint_timer_query_webgl2 !== null) {
                webgl.ctx.deleteQuery(
                    gioLoadJSValue(sp),
                );
            } else {
                webgl.EXT_disjoint_timer_query_webgl.deleteQueryEXT(
                    gioLoadJSValue(sp),
                );
            }
        },
        // deleteShader(s Shader)
        "gioui.org/internal/glimpl.deleteShader": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.deleteShader(
                gioLoadJSValue(sp),
            );
        },
        // deleteRenderbuffer(v Renderbuffer)
        "gioui.org/internal/glimpl.deleteRenderbuffer": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.deleteRenderbuffer(
                gioLoadJSValue(sp),
            );
        },
        // deleteTexture(v Texture)
        "gioui.org/internal/glimpl.deleteTexture": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.deleteTexture(
                gioLoadJSValue(sp),
            );
        },
        // depthFunc(fn Enum)
        "gioui.org/internal/glimpl.depthFunc": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.depthFunc(
                gioLoadInt64(sp),
            );
        },
        // depthMask(mask bool)
        "gioui.org/internal/glimpl.depthMask": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.depthMask(
                gioLoadInt64(sp),
            );
        },
        // disableVertexAttribArray(a Attrib)
        "gioui.org/internal/glimpl.disableVertexAttribArray": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.disableVertexAttribArray(
                gioLoadInt64(sp),
            );
        },
        // disable(cap Enum)
        "gioui.org/internal/glimpl.disable": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.disable(
                gioLoadInt64(sp),
            );
        },
        // drawArrays(mode Enum, first, count int)
        "gioui.org/internal/glimpl.drawArrays": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.drawArrays(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2)),
            )
        },
        // drawElements(mode Enum, count int, ty Enum, offset int)
        "gioui.org/internal/glimpl.drawElements": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.drawElements(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2)),
                gioLoadInt64(sp + (OffsetInt64 * 3)),
            )
        },
        // enable(cap Enum)
        "gioui.org/internal/glimpl.enable": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.enable(
                gioLoadInt64(sp),
            );
        },
        // enableVertexAttribArray(a Attrib)
        "gioui.org/internal/glimpl.enableVertexAttribArray": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.enableVertexAttribArray(
                gioLoadInt64(sp),
            );
        },
        // endQuery(target Enum)
        "gioui.org/internal/glimpl.endQuery": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            if (webgl.EXT_disjoint_timer_query_webgl2 !== null) {
                webgl.ctx.endQuery(
                    gioLoadInt64(sp),
                );
            } else {
                webgl.EXT_disjoint_timer_query.endQueryEXT(
                    gioLoadInt64(sp),
                );
            }
        },
        // finish()
        "gioui.org/internal/glimpl.finish": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.finish();
        },
        // framebufferRenderbuffer(target, attachment, renderbuffertarget Enum, renderbuffer Renderbuffer)
        "gioui.org/internal/glimpl.framebufferRenderbuffer": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.framebufferRenderbuffer(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2)),
                gioLoadJSValue(sp + (OffsetInt64 * 3)),
            );
        },
        // framebufferTexture2D(target, attachment, texTarget Enum, t Texture, level int)
        "gioui.org/internal/glimpl.framebufferTexture2D": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.framebufferTexture2D(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2)),
                gioLoadJSValue(sp + (OffsetInt64 * 3)),
                gioLoadInt64(sp + (OffsetInt64 * 3) + OffsetJSValue),
            );
        },
        // getRenderbufferParameteri(target, pname Enum) uint64
        "gioui.org/internal/glimpl.getRenderbufferParameteri": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.getRenderbufferParameter(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
            );
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp + (OffsetInt64 * 2), result);
        },
        // getFramebufferAttachmentParameteri(target, attachment, pname Enum) uint64
        "gioui.org/internal/glimpl.getFramebufferAttachmentParameteri": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.getFramebufferAttachmentParameter(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2)),
            );
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp + (OffsetInt64 * 3), result);
        },
        // getBinding(pname Enum) uint64
        "gioui.org/internal/glimpl.getBinding": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.getParameter(
                gioLoadInt64(sp),
            );
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp + OffsetInt64, result);
        },
        // getInteger(pname Enum) uint64
        "gioui.org/internal/glimpl.getInteger": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.getParameter(
                gioLoadInt64(sp),
            );
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp + OffsetInt64, result);
        },
        // getProgrami(p Program, pname Enum) uint64
        "gioui.org/internal/glimpl.getProgrami": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.getProgramParameter(
                gioLoadJSValue(sp),
                gioLoadInt64(sp + OffsetJSValue)
            );
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp + OffsetJSValue + OffsetInt64, result);
        },
        // getProgramInfoLog(p Program) uint64
        "gioui.org/internal/glimpl.getProgramInfoLog": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.getProgramInfoLog(
                gioLoadJSValue(sp),
            );
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp + OffsetJSValue, result);
        },
        // getQueryObjectuiv(query Query, pname Enum) uint64
        "gioui.org/internal/glimpl.getQueryObjectuiv": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            let result = 0;
            if (webgl.EXT_disjoint_timer_query_webgl2 !== null) {
                result = webgl.ctx.getQueryParameter(
                    gioLoadJSValue(sp),
                    gioLoadInt64(sp + OffsetJSValue),
                );
            } else {
                result = webgl.EXT_disjoint_timer_query.getQueryObjectEXT(
                    gioLoadJSValue(sp),
                    gioLoadInt64(sp + OffsetJSValue),
                );
            }

            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp + OffsetInt64 + OffsetJSValue, result);
        },
        // getShaderi(s Shader, pname Enum) uint
        "gioui.org/internal/glimpl.getShaderi": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.getShaderParameter(
                gioLoadJSValue(sp),
                gioLoadInt64(sp + OffsetJSValue),
            );
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp + OffsetJSValue + OffsetInt64, result);
        },
        // getShaderInfoLog(s Shader) uint64
        "gioui.org/internal/glimpl.getShaderInfoLog": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.getShaderInfoLog(
                gioLoadJSValue(sp),
            );
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp + OffsetJSValue, result)
        },
        // getString(method int, pname Enum) uint64
        "gioui.org/internal/glimpl.getString": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            let result = 0;
            if (gioLoadInt64(sp) === 1) {
                result = webgl.ctx.getSupportedExtensions();
            } else {
                result = webgl.ctx.getParameter(gioLoadInt64(sp + OffsetInt64));
            }
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp + (OffsetInt64 * 2), result);
        },
        // getUniformBlockIndex(p Program, name string) uint64
        "gioui.org/internal/glimpl.getUniformBlockIndex": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.getUniformBlockIndex(
                gioLoadJSValue(sp),
                gioLoadString(sp + OffsetJSValue),
            )
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp + OffsetJSValue + OffsetString, result);
        },
        // getUniformLocation(p Program, name string) uint64
        "gioui.org/internal/glimpl.getUniformLocation": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            const result = webgl.ctx.getUniformLocation(
                gioLoadJSValue(sp),
                gioLoadString(sp + OffsetJSValue),
            )
            sp = (go._inst.exports.getsp() >>> 0) + OffsetContextIndex;
            gioSetJSValue(sp + OffsetJSValue + OffsetString, result);
        },
        // invalidateFramebuffer(target, attachment Enum)
        "gioui.org/internal/glimpl.invalidateFramebuffer": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            if (typeof webgl.ctx.invalidateFramebuffer === "undefined") {
                return
            }
            invalidateBuffer.set([gioLoadInt64(sp + OffsetInt64)]);
            webgl.ctx.invalidateFramebuffer(
                gioLoadInt64(sp),
                invalidateBuffer,
            );
        },
        // linkProgram(p Program)
        "gioui.org/internal/glimpl.linkProgram": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.linkProgram(
                gioLoadJSValue(sp),
            );
        },
        // pixelStorei(pname Enum, param int32)
        "gioui.org/internal/glimpl.pixelStorei": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.pixelStorei(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
            );
        },
        // renderbufferStorage(target, internalformat Enum, width, height int)
        "gioui.org/internal/glimpl.renderbufferStorage": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.renderbufferStorage(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2)),
                gioLoadInt64(sp + (OffsetInt64 * 3)),
            );
        },
        // readPixels(x, y, width, height int, format, ty Enum, data []byte)
        "gioui.org/internal/glimpl.readPixels": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.readPixels(
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
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.scissor(
                gioLoadInt32(sp),
                gioLoadInt32(sp + OffsetInt64),
                gioLoadInt32(sp + (OffsetInt64 * 2)),
                gioLoadInt32(sp + (OffsetInt64 * 3)),
            )
        },
        // shaderSource(s Shader, src string)
        "gioui.org/internal/glimpl.shaderSource": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.shaderSource(
                gioLoadJSValue(sp),
                gioLoadString(sp + OffsetJSValue),
            )
        },
        // texImage2D(target Enum, level int, internalFormat int, width, height int, format, ty Enum, data []byte)
        "gioui.org/internal/glimpl.texImage2D": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.texImage2D(
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
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.texSubImage2D(
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
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.texParameteri(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2))
            )
        },
        // uniformBlockBinding(p Program, uniformBlockIndex uint, uniformBlockBinding uint)
        "gioui.org/internal/glimpl.uniformBlockBinding": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.uniformBlockBinding(
                gioLoadJSValue(sp),
                gioLoadInt64(sp + OffsetJSValue),
                gioLoadInt64(sp + OffsetJSValue + OffsetInt64),
            )
        },
        // uniform1i(dst Uniform, v int)
        "gioui.org/internal/glimpl.uniform1i": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.uniform1i(
                gioLoadJSValue(sp),
                gioLoadInt64(sp + OffsetJSValue),
            )
        },
        // uniformXf(x int, dst Uniform, v0, v1, v2, v3 float32)
        "gioui.org/internal/glimpl.uniformXf": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            switch (gioLoadInt32(sp)) {
                case 1:
                    webgl.ctx.uniform1f(
                        gioLoadJSValue(sp + OffsetInt64),
                        gioLoadFloat64(sp + OffsetInt64 + OffsetJSValue),
                    );
                    return
                case 2:
                    webgl.ctx.uniform2f(
                        gioLoadJSValue(sp + OffsetInt64),
                        gioLoadFloat64(sp + OffsetInt64 + OffsetJSValue),
                        gioLoadFloat64(sp + OffsetInt64 + OffsetJSValue + OffsetFloat64),
                    );
                    return
                case 3:
                    webgl.ctx.uniform3f(
                        gioLoadJSValue(sp + OffsetInt64),
                        gioLoadFloat64(sp + OffsetInt64 + OffsetJSValue),
                        gioLoadFloat64(sp + OffsetInt64 + OffsetJSValue + OffsetFloat64),
                        gioLoadFloat64(sp + OffsetInt64 + OffsetJSValue + (OffsetFloat64 * 2)),
                    );
                    return
                case 4:
                    webgl.ctx.uniform4f(
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
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.useProgram(
                gioLoadJSValue(sp),
            )
        },
        // vertexAttribPointer(dst Attrib, size int, ty Enum, normalized bool, stride, offset int)
        "gioui.org/internal/glimpl.vertexAttribPointer": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.vertexAttribPointer(
                gioLoadInt64(sp),
                gioLoadInt64(sp + OffsetInt64),
                gioLoadInt64(sp + (OffsetInt64 * 2)),
                gioLoadInt64(sp + (OffsetInt64 * 3)),
                gioLoadInt64(sp + (OffsetInt64 * 4)),
                gioLoadInt64(sp + (OffsetInt64 * 5)),
            );
        },
        // viewport(x, y, width, height int)
        "gioui.org/internal/glimpl.viewport": (sp) => {
            sp = (sp >>> 0) + OffsetContextIndex;
            const webgl = gioLoadContext(sp);
            webgl.ctx.viewport(
                gioLoadInt32(sp),
                gioLoadInt32(sp + OffsetInt64),
                gioLoadInt32(sp + (OffsetInt64 * 2)),
                gioLoadInt32(sp + (OffsetInt64 * 3)),
            );
        },
    })
})();