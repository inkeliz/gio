// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import "image"

type maskIdentity struct {
	digest [32]byte
	clip   image.Rectangle
}

type atlasRect struct {
	page int
	rect image.Rectangle
}

// atlasAllocator only reuses storage already owned by the renderer.
type atlasAllocator struct {
	free, scratch []atlasRect
}

func (a *atlasAllocator) reset(fbos []FBO) {
	a.free = a.free[:0]
	for i, f := range fbos {
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

type stencilAtlas struct {
	allocator         atlasAllocator
	current, previous map[maskIdentity]placement
}

func maskFor(cache *opCache, p *pathOp) maskIdentity {
	v, _ := cache.get(p.pathKey)
	origin := image.Pt(int(p.off.X), int(p.off.Y))
	return maskIdentity{v.data.digest, p.clip.Sub(origin)}
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
			}
		}
	}
	if stable && !retained {
		for _, p := range ops {
			key := maskFor(cache, p)
			if _, found := s.current[key]; !found {
				place, fits := s.allocator.alloc(p.clip.Size())
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
		r.packer.clear()
		for _, f := range r.pather.stenciler.fbos.fbos[:pages] {
			r.packer.sizes = append(r.packer.sizes, f.size)
		}
	} else {
		// Compact normally when existing pages cannot accommodate this frame.
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
	s.previous, s.current = s.current, s.previous
}
