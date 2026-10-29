// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"image"
	"slices"

	"gioui.org/gpu/internal/driver"
)

type intersectionSlot struct {
	place placement
	size  image.Point
}

type intersectionSource struct {
	mask    stencilKey
	texture driver.Texture
	clip    image.Rectangle
}

type intersectionEntry struct {
	slot       intersectionSlot
	texture    driver.Texture
	start, end int
}

type intersectionCache struct {
	entries, previousEntries []intersectionEntry
	sources, previousSources []intersectionSource
	index                    map[intersectionSlot]int
	dirty, clearPages        []bool
}

func (r *renderer) packIntersectionsIncremental(cache *opCache, ops []imageOp, disabled bool) {
	r.packIntersections(ops)
	if len(r.intersections.sizes) > 0 {
		r.pather.stenciler.beginIntersect(r.intersections.sizes)
	}
	c := &r.intersectionCache
	if c.index == nil {
		c.index = make(map[intersectionSlot]int)
	}
	clear(c.entries)
	c.entries = c.entries[:0]
	clear(c.sources)
	c.sources = c.sources[:0]
	c.dirty = c.dirty[:0]
	c.clearPages = c.clearPages[:0]
	for range r.intersections.sizes {
		c.clearPages = append(c.clearPages, true)
	}
	for _, img := range ops {
		if img.clipType != clipTypeIntersection {
			c.dirty = append(c.dirty, false)
			continue
		}
		start := len(c.sources)
		for p := img.path; p != nil; p = p.parent {
			if p.path {
				key := maskFor(cache, p)
				tex := r.pather.stenciler.cover(p.place.Idx).tex
				c.sources = append(c.sources, intersectionSource{stencilKey{key.digest, key.clip, p.place}, tex, img.clip.Sub(p.clip.Min)})
			}
		}
		slices.Reverse(c.sources[start:])
		entry := intersectionEntry{
			slot:    intersectionSlot{img.place, img.clip.Size()},
			texture: r.pather.stenciler.intersections.fbos[img.place.Idx].tex,
			start:   start,
			end:     len(c.sources),
		}
		dirty := true
		if previous, found := c.index[entry.slot]; found && !disabled {
			old := c.previousEntries[previous]
			dirty = old.texture != entry.texture || !slices.Equal(c.sources[start:], c.previousSources[old.start:old.end])
		}
		c.dirty = append(c.dirty, dirty)
		if !dirty {
			c.clearPages[img.place.Idx] = false
		}
		c.entries = append(c.entries, entry)
	}
	c.previousEntries, c.entries = c.entries, c.previousEntries
	c.previousSources, c.sources = c.sources, c.previousSources
	clear(c.index)
	for i, entry := range c.previousEntries {
		c.index[entry.slot] = i
	}
}

func (r *renderer) prepareIntersectionsIncremental(ops []imageOp) {
	for i, img := range ops {
		if !r.intersectionCache.dirty[i] {
			continue
		}
		for p := img.path; p != nil; p = p.parent {
			if p.path {
				r.ctx.PrepareTexture(r.pather.stenciler.cover(p.place.Idx).tex)
			}
		}
	}
}

func (r *renderer) intersectIncremental(ops []imageOp) {
	fbo := -1
	fullPage := false
	for i, img := range ops {
		if !r.intersectionCache.dirty[i] {
			continue
		}
		if fbo != img.place.Idx {
			if fbo != -1 {
				r.ctx.EndRenderPass()
			}
			fbo = img.place.Idx
			fullPage = r.intersectionCache.clearPages[fbo]
			desc := driver.LoadDesc{Action: driver.LoadActionKeep}
			if fullPage {
				desc.Action, desc.ClearColor.R = driver.LoadActionClear, 1
				// A later visit must preserve targets already written to this page.
				r.intersectionCache.clearPages[fbo] = false
			}
			r.ctx.BeginRenderPass(r.pather.stenciler.intersections.fbos[fbo].tex, desc)
		}
		if !fullPage {
			r.clearFloatSlot(img.place, img.clip.Size(), 1)
		}
		r.ctx.BindPipeline(r.pather.stenciler.ipipeline.pipeline.pipeline)
		r.ctx.BindVertexBuffer(r.blitter.quadVerts, 0)
		r.ctx.Viewport(img.place.Pos.X, img.place.Pos.Y, img.clip.Dx(), img.clip.Dy())
		r.intersectPath(img.path, img.clip)
	}
	if fbo != -1 {
		r.ctx.EndRenderPass()
	}
}
