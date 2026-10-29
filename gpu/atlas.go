// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"image"
	"slices"
)

type maskIdentity struct {
	digest [32]byte
	clip   image.Rectangle
}

type atlasRect struct {
	page int
	rect image.Rectangle
}

// atlasAllocator preserves slots while bounded incremental growth can fit them.
type atlasAllocator struct {
	free, scratch []atlasRect
	sizes         []image.Point
}

func (a *atlasAllocator) reset(fbos []FBO) {
	a.free = a.free[:0]
	a.sizes = a.sizes[:0]
	for i, f := range fbos {
		a.sizes = append(a.sizes, f.size)
		if f.tex != nil {
			a.free = append(a.free, atlasRect{i, image.Rectangle{Max: f.size}})
		}
	}
}

func (a *atlasAllocator) reserve(slot atlasRect) bool {
	next := a.scratch[:0]
	area := 0
	for _, f := range a.free {
		cut := f.rect.Intersect(slot.rect)
		if f.page != slot.page || cut.Empty() {
			next = append(next, f)
			continue
		}
		area += cut.Dx() * cut.Dy()
		for _, part := range [4]image.Rectangle{
			{Min: f.rect.Min, Max: image.Pt(f.rect.Max.X, cut.Min.Y)},
			{Min: image.Pt(f.rect.Min.X, cut.Max.Y), Max: f.rect.Max},
			{Min: image.Pt(f.rect.Min.X, cut.Min.Y), Max: image.Pt(cut.Min.X, cut.Max.Y)},
			{Min: image.Pt(cut.Max.X, cut.Min.Y), Max: image.Pt(f.rect.Max.X, cut.Max.Y)},
		} {
			if !part.Empty() {
				next = append(next, atlasRect{f.page, part})
			}
		}
	}
	a.free, a.scratch = next, a.free[:0]
	return area == slot.rect.Dx()*slot.rect.Dy()
}

func (a *atlasAllocator) alloc(size image.Point) (placement, bool) {
	best, waste := -1, int(^uint(0)>>1)
	for i, f := range a.free {
		if size.X <= f.rect.Dx() && size.Y <= f.rect.Dy() {
			if n := f.rect.Dx()*f.rect.Dy() - size.X*size.Y; n < waste {
				best, waste = i, n
			}
		}
	}
	if best == -1 {
		return placement{}, false
	}
	f := a.free[best]
	p := placement{Idx: f.page, Pos: f.rect.Min}
	a.reserve(atlasRect{p.Idx, image.Rectangle{Min: p.Pos, Max: p.Pos.Add(size)}})
	return p, true
}

// Bound extra stable-placement growth instead of retaining arbitrarily large
// sparse atlases. Larger frames retain the ordinary compact packing fallback.
const maxStableAtlasBytes = 32 << 20

func (a *atlasAllocator) grow(size, viewport, maxDims image.Point, bytesPerPixel int) bool {
	if len(a.sizes) >= 8 || size.X > maxDims.X || size.Y > maxDims.Y {
		return false
	}
	// One small page gives new glyphs/masks room without resizing old textures.
	size.X = max(size.X, min(viewport.X, maxDims.X, 512))
	size.Y = max(size.Y, min(viewport.Y, maxDims.Y, 512))
	// fboSet.resize adds 5% on allocation; include that padding in the budget.
	padded := size.Mul(105).Div(100)
	pixels := int64(padded.X) * int64(padded.Y)
	for _, page := range a.sizes {
		pixels += int64(page.X) * int64(page.Y)
	}
	if bytesPerPixel <= 0 {
		bytesPerPixel = 16 // Conservative RGBA32F fallback for other backends.
	}
	if pixels*int64(bytesPerPixel) > maxStableAtlasBytes {
		return false
	}
	page := len(a.sizes)
	a.sizes = append(a.sizes, size)
	a.free = append(a.free, atlasRect{page, image.Rectangle{Max: size}})
	return true
}

type stencilAtlas struct {
	allocator         atlasAllocator
	current, previous map[maskIdentity]placement
	keptPages         int // Budget-approved growth, retained until compact packing.
}

func maskFor(cache *opCache, p *pathOp) maskIdentity {
	digest := cache.cache[p.cacheIndex-1].data.digest
	origin := image.Pt(int(p.off.X), int(p.off.Y))
	return maskIdentity{digest, p.clip.Sub(origin)}
}

func (r *renderer) packStencilsStable(cache *opCache, pops *[]*pathOp, disabled bool) {
	s := &r.stencilAtlas
	if s.current == nil {
		s.current, s.previous = make(map[maskIdentity]placement), make(map[maskIdentity]placement)
	}
	clear(s.current)
	ops := *pops
	for i := 0; i < len(ops); {
		p := ops[i]
		if p.clip.Empty() {
			ops[i], ops = ops[len(ops)-1], ops[:len(ops)-1]
			continue
		}
		full := p.clip.Union(p.bounds.Add(p.off).Round())
		size := full.Size()
		if size.X > 0 && size.Y > 0 && size.X <= r.blitter.viewport.X && size.Y <= r.blitter.viewport.Y && size.X <= r.packer.maxDims.X && size.Y <= r.packer.maxDims.Y {
			p.clip = full
		}
		i++
	}
	*pops = ops
	stable := !disabled && len(ops) != 0
	retained := stable
	if stable {
		// Existing slots are already disjoint; unchanged masks need no allocator.
		for _, p := range ops {
			key := maskFor(cache, p)
			place, found := s.previous[key]
			if !found {
				retained = false
				break
			}
			p.place, s.current[key] = place, place
		}
	}
	preserved := false
	if stable && !retained {
		clear(s.current)
		s.allocator.reset(r.pather.stenciler.fbos.fbos)
		// Protect all retained masks before assigning any newly visible mask.
		for _, p := range ops {
			key := maskFor(cache, p)
			if _, found := s.current[key]; found {
				continue
			}
			if place, found := s.previous[key]; found {
				if !s.allocator.reserve(atlasRect{place.Idx, image.Rectangle{Min: place.Pos, Max: place.Pos.Add(p.clip.Size())}}) {
					stable = false
					break
				}
				s.current[key] = place
				preserved = true
			}
		}
	}
	if stable && !retained {
		grown := false
		for _, p := range ops {
			key := maskFor(cache, p)
			if _, found := s.current[key]; !found {
				place, fits := s.allocator.alloc(p.clip.Size())
				if !fits && preserved && !grown && s.allocator.grow(p.clip.Size(), r.blitter.viewport, r.packer.maxDims, r.ctx.Caps().FloatRenderTargetBytes) {
					grown = true
					s.keptPages = len(s.allocator.sizes)
					place, fits = s.allocator.alloc(p.clip.Size())
				}
				if !fits {
					stable = false
					break
				}
				s.current[key] = place
			}
			p.place = s.current[key]
		}
	}
	if stable {
		pages := 0
		for _, place := range s.current {
			pages = max(pages, place.Idx+1)
		}
		pages = max(pages, s.keptPages)
		r.packer.clear()
		if retained {
			for _, f := range r.pather.stenciler.fbos.fbos[:pages] {
				r.packer.sizes = append(r.packer.sizes, f.size)
			}
		} else {
			r.packer.sizes = append(r.packer.sizes, s.allocator.sizes[:pages]...)
		}
	} else {
		// Compact normally when existing pages cannot accommodate this frame.
		s.keptPages = 0
		r.packStencils(pops)
		clear(s.current)
		for _, p := range *pops {
			key := maskFor(cache, p)
			if place, found := s.current[key]; found && !disabled {
				p.place = place
			} else {
				s.current[key] = p.place
			}
		}
	}
	// Masks occupy disjoint atlas slots, so page grouping does not change
	// their contents and avoids reopening a page after rendering another one.
	if !retained && len(r.packer.sizes) > 1 {
		slices.SortStableFunc(*pops, func(a, b *pathOp) int { return a.place.Idx - b.place.Idx })
	}
	s.previous, s.current = s.current, s.previous
}
