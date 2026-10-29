// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"image"
	"math/bits"
	"unsafe"

	"gioui.org/f32"
	"gioui.org/gpu/internal/driver"
	"gioui.org/internal/byteslice"
	if32 "gioui.org/internal/f32"
	"gioui.org/internal/f32color"
	"gioui.org/shader"
	"gioui.org/shader/gio"
)

const batchQuads = 256

type batchVertex struct {
	pos, uv, cover [2]float32
	color, color2  [4]float32
	shape, half    [2]float32
	radii          [4]float32
}
type batchUniforms struct {
	transform                  [4]float32
	fbo, material, clipMode, _ float32
}
type batchKey struct {
	texture, cover driver.Texture
	material       materialType
	clipMode       uint8
	opaque         bool
}
type drawBatch struct {
	ctx       driver.Device
	pipelines [2][2][2]*pipeline
	uniforms  *batchUniforms
	indices   driver.Buffer
	white     driver.Texture
	vertices  []batchVertex
	buffers   [9][]driver.Buffer
	used      [9]int
	count     int
	key       batchKey
	first     [1]imageOp
}

// Keep other backends on their existing pipelines until their shader assets
// are generated and verified. This optional capability has no public API cost.
func newDrawBatch(ctx driver.Device) *drawBatch {
	supported, ok := ctx.(interface{ SupportsRasterBatch() bool })
	if !ok || !supported.SupportsRasterBatch() {
		return nil
	}
	b := &drawBatch{ctx: ctx, uniforms: new(batchUniforms), vertices: make([]batchVertex, batchQuads*4)}
	layout := driver.VertexLayout{Stride: int(unsafe.Sizeof(batchVertex{}))}
	for _, input := range []struct{ size, offset int }{
		{2, int(unsafe.Offsetof(batchVertex{}.pos))}, {2, int(unsafe.Offsetof(batchVertex{}.uv))},
		{2, int(unsafe.Offsetof(batchVertex{}.cover))}, {4, int(unsafe.Offsetof(batchVertex{}.color))},
		{4, int(unsafe.Offsetof(batchVertex{}.color2))}, {2, int(unsafe.Offsetof(batchVertex{}.shape))},
		{2, int(unsafe.Offsetof(batchVertex{}.half))}, {4, int(unsafe.Offsetof(batchVertex{}.radii))},
	} {
		layout.Inputs = append(layout.Inputs, driver.InputDesc{Type: shader.DataTypeFloat, Size: input.size, Offset: input.offset})
	}
	for analytic, source := range []shader.Sources{gio.Shader_batch_plain_frag, gio.Shader_batch_frag} {
		vs, fs, err := newShaders(ctx, gio.Shader_batch_vert, source)
		if err != nil {
			b.release()
			return nil
		}
		defer vs.Release()
		defer fs.Release()
		for i, format := range []driver.TextureFormat{driver.TextureFormatOutput, driver.TextureFormatSRGBA} {
			for opaque := 0; opaque < 2; opaque++ {
				blend := driver.BlendDesc{SrcFactor: driver.BlendFactorOne, DstFactor: driver.BlendFactorOneMinusSrcAlpha}
				if opaque != 0 {
					blend.DstFactor = driver.BlendFactorZero
				}
				p, err := ctx.NewPipeline(driver.PipelineDesc{VertexShader: vs, FragmentShader: fs, VertexLayout: layout, BlendDesc: blend, PixelFormat: format, Topology: driver.TopologyTriangles})
				if err != nil {
					b.release()
					return nil
				}
				b.pipelines[analytic][i][opaque] = &pipeline{p, newUniformBuffer(ctx, byteslice.View(b.uniforms))}
			}
		}
	}
	var err error
	indices := make([]uint16, batchQuads*6)
	for i := range batchQuads {
		v := uint16(i * 4)
		copy(indices[i*6:], []uint16{v, v + 1, v + 2, v + 2, v + 1, v + 3})
	}
	b.indices, err = ctx.NewImmutableBuffer(driver.BufferBindingIndices, byteslice.Slice(indices))
	if err != nil {
		b.release()
		return nil
	}
	b.white, err = ctx.NewTexture(driver.TextureFormatSRGBA, 1, 1, driver.FilterNearest, driver.FilterNearest, driver.BufferBindingTexture)
	if err != nil {
		b.release()
		return nil
	}
	b.white.Upload(image.Point{}, image.Pt(1, 1), []byte{255, 255, 255, 255}, 4)
	return b
}
func (b *drawBatch) beginFrame() {
	if b != nil {
		clear(b.used[:])
		clear(b.first[:])
	}
}
func (b *drawBatch) release() {
	if b == nil {
		return
	}
	for _, formats := range b.pipelines {
		for _, row := range formats {
			for _, p := range row {
				if p != nil {
					p.Release()
				}
			}
		}
	}
	for _, pool := range b.buffers {
		for _, v := range pool {
			v.Release()
		}
	}
	if b.indices != nil {
		b.indices.Release()
	}
	if b.white != nil {
		b.white.Release()
	}
}
func (b *drawBatch) flush(r *renderer, fbo bool, offset, viewport image.Point) {
	if b.count == 0 {
		return
	}
	// A lone ordinary quad already has a cheaper static-vertex pipeline.
	if b.count == 1 && b.key.clipMode <= 1 {
		r.drawOpsUnbatched(fbo, offset, viewport, b.first[:])
		b.count = 0
		return
	}
	bucket := bits.Len(uint(b.count - 1))
	used := b.used[bucket]
	capacity := 1 << bucket
	if used == len(b.buffers[bucket]) {
		v, err := b.ctx.NewBuffer(driver.BufferBindingVertices, capacity*4*int(unsafe.Sizeof(batchVertex{})))
		if err != nil {
			panic(err)
		}
		b.buffers[bucket] = append(b.buffers[bucket], v)
	}
	buf := b.buffers[bucket][used]
	b.used[bucket]++
	// Upload the complete bucket to orphan its previous storage, avoiding a
	// mid-frame BufferSubData wait on vertices still consumed by the GPU.
	buf.Upload(byteslice.Slice(b.vertices[:capacity*4]))
	fi, opaque := 0, 0
	if fbo {
		fi = 1
	}
	if b.key.opaque {
		opaque = 1
	}
	b.uniforms.material = float32(b.key.material)
	b.uniforms.clipMode = float32(b.key.clipMode)
	analytic := 0
	if b.key.clipMode > 1 {
		analytic = 1
	}
	p := b.pipelines[analytic][fi][opaque]
	b.ctx.BindPipeline(p.pipeline)
	p.UploadUniforms(b.ctx)
	tex, cover := b.key.texture, b.key.cover
	if tex == nil {
		tex = b.white
	}
	if cover == nil {
		cover = b.white
	}
	b.ctx.BindTexture(0, tex)
	b.ctx.BindTexture(1, cover)
	b.ctx.BindVertexBuffer(buf, 0)
	b.ctx.BindIndexBuffer(b.indices)
	b.ctx.DrawElements(0, b.count*6)
	b.count = 0
}
func rgba(c f32color.RGBA, opacity float32) [4]float32 {
	return [4]float32{c.R * opacity, c.G * opacity, c.B * opacity, c.A * opacity}
}
func (b *drawBatch) draw(r *renderer, fbo bool, offset, viewport image.Point, paints []imageOp) {
	b.count = 0
	b.uniforms.transform = [4]float32{2 / float32(viewport.X), 2 / float32(viewport.Y), -1, -1}
	b.uniforms.fbo = 0
	if fbo {
		b.uniforms.fbo = 1
	}
	for i := 0; i < len(paints); i++ {
		img := paints[i]
		i += img.layerOps
		if img.culled || img.clip.Empty() {
			continue
		}
		m := img.material
		if img.analytic.kind == 0 && m.material != materialColor {
			b.flush(r, fbo, offset, viewport)
			b.first[0] = img
			r.drawOpsUnbatched(fbo, offset, viewport, b.first[:])
			continue
		}
		key := batchKey{material: m.material, texture: m.tex, opaque: img.analytic.kind == 0 && img.clipType == clipTypeNone && m.opaque && m.opacity == 1}
		var scale, off f32.Point
		if img.analytic.kind != 0 {
			key.clipMode = 2
			if img.analytic.kind == analyticEllipse {
				key.clipMode = 3
			}
		} else if img.clipType != clipTypeNone {
			var f FBO
			if img.clipType == clipTypePath {
				f = r.pather.stenciler.cover(img.place.Idx)
			} else {
				f = r.pather.stenciler.intersections.fbos[img.place.Idx]
			}
			key.cover = f.tex
			key.clipMode = 1
			uv := image.Rectangle{Min: img.place.Pos, Max: img.place.Pos.Add(img.clip.Size())}
			scale, off = texSpaceTransform(if32.FRect(uv), f.size)
		}
		if m.material == materialEmbed {
			key.material = materialColor
			key.opaque = true
		}
		if b.count != 0 && (key != b.key || b.count == batchQuads) {
			b.flush(r, fbo, offset, viewport)
		}
		b.key = key
		if b.count == 0 {
			b.first[0] = img
		}
		col, col2 := m.color, m.color2
		if m.material == materialLinearGradient {
			col = m.color1
		}
		if m.material == materialTexture {
			col = f32color.RGBA{R: 1, G: 1, B: 1, A: 1}
		}
		if m.material == materialEmbed {
			col = f32color.RGBA{}
		}
		rect := img.clip.Add(offset)
		center := img.analytic.bounds.Min.Add(img.analytic.bounds.Max).Mul(.5)
		half := img.analytic.bounds.Size().Mul(.5)
		for corner := 0; corner < 4; corner++ {
			u, v := float32(corner&1), float32((corner>>1)&1)
			x, y := float32(rect.Min.X)+u*float32(rect.Dx()), float32(rect.Min.Y)+v*float32(rect.Dy())
			uv := m.uvTrans.Transform(f32.Pt(u, v))
			b.vertices[b.count*4+corner] = batchVertex{pos: [2]float32{x, y}, uv: [2]float32{uv.X, uv.Y}, cover: [2]float32{u*scale.X + off.X, v*scale.Y + off.Y}, color: rgba(col, m.opacity), color2: rgba(col2, m.opacity), shape: [2]float32{x - float32(offset.X) - center.X, y - float32(offset.Y) - center.Y}, half: [2]float32{half.X, half.Y}, radii: img.analytic.radii}
		}
		b.count++
	}
	b.flush(r, fbo, offset, viewport)
}
