// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"image"
	"math"
	"slices"
)

// cullOccluded leaves paint indices intact. Hidden opacity groups may be
// skipped as a unit; other group boundaries reset coverage so contents never
// occlude paints outside their own compositing scope.
func (d *drawOps) cullOccluded() {
	const maxCovers = 64
	if len(d.layers) > maxCovers || len(d.embedRegions.Views) != 0 || len(d.embedRegionsOld) != 0 {
		return
	}
	for i, l := range d.layers {
		if l.opStart < 0 || l.opEnd < l.opStart || l.opEnd > len(d.imageOps) || l.parent < -1 || l.parent >= i {
			return
		}
		if i > 0 && l.opStart < d.layers[i-1].opStart {
			return
		}
	}
	// Collection adds direct paints to a layer. Its complete bounds must also
	// include children before testing whether an outside rectangle hides it.
	for i := len(d.layers) - 1; i >= 0; i-- {
		l := &d.layers[i]
		if l.parent >= 0 {
			d.layers[l.parent].clip = d.layers[l.parent].clip.Union(l.clip)
		}
	}
	// Starts are already in preorder. Sort end boundaries once, preserving
	// preorder for parents and children that share their final paint.
	var ends [maxCovers]int
	for i := range d.layers {
		ends[i] = i
	}
	slices.SortStableFunc(ends[:len(d.layers)], func(a, b int) int {
		return d.layers[b].opEnd - d.layers[a].opEnd
	})
	start, end := len(d.layers)-1, 0
	var hidden [maxCovers]bool
	var covers [maxCovers]image.Rectangle
	ncovers := 0
	for i := len(d.imageOps) - 1; i >= 0; i-- {
		// Leave later scopes before entering an earlier adjacent scope.
		for start >= 0 && d.layers[start].opStart >= i+1 {
			j := start
			start--
			l := d.layers[j]
			if !hidden[j] && l.opStart < l.opEnd && l.opStart == i+1 {
				ncovers = 0
			}
		}
		jump := false
		// Layers are recorded in preorder: test parents before children that
		// share the same final paint.
		for end < len(d.layers) && d.layers[ends[end]].opEnd >= i+1 {
			j := ends[end]
			end++
			l := d.layers[j]
			if hidden[j] || l.opStart == l.opEnd || l.opEnd != i+1 {
				continue
			}
			if occludedBy(l.clip, covers[:ncovers]) {
				for k := l.opStart; k < l.opEnd; k++ {
					d.imageOps[k].culled = true
				}
				for k := j; k < len(d.layers); k++ {
					if k != j && d.layers[k].depth <= l.depth {
						break
					}
					hidden[k] = true
					d.layers[k].clip = image.Rectangle{}
				}
				i = l.opStart
				jump = true
				break
			}
			ncovers = 0
		}
		if jump {
			continue
		}
		img := &d.imageOps[i]
		if img.culled {
			continue
		}
		m := img.material
		if img.clip.Empty() || (m.material == materialColor && m.color.A == 0) || occludedBy(img.clip, covers[:ncovers]) {
			img.culled = true
			continue
		}
		// Keep one visible rectangle so opacity-layer indices stay intact and
		// trimming needs no extra draws, vertices, or texture-coordinate changes.
		if m.material == materialColor && (img.path == nil || img.path.rect) {
			for _, cover := range covers[:ncovers] {
				img.clip = subtractOccludedStrip(img.clip, cover)
			}
			if img.clip.Empty() {
				img.culled = true
				continue
			}
		}
		if !m.opaque || m.opacity != 1 || (m.material != materialColor && m.material != materialLinearGradient) {
			continue
		}
		cover := opaqueInnerRect(img)
		if cover.Empty() {
			continue
		}
		for j := 0; j < ncovers; {
			if covers[j].In(cover) {
				ncovers--
				covers[j] = covers[ncovers]
			} else {
				j++
			}
		}
		if ncovers < len(covers) {
			covers[ncovers] = cover
			ncovers++
		}
	}
}

func occludedBy(clip image.Rectangle, covers []image.Rectangle) bool {
	if clip.Empty() {
		return false
	}
	for _, cover := range covers {
		if clip.In(cover) {
			return true
		}
	}
	return false
}

// subtractOccludedStrip removes covered edge strips while retaining a single
// visible rectangle. Interior cuts and cuts requiring multiple draws stay intact.
func subtractOccludedStrip(clip, cover image.Rectangle) image.Rectangle {
	cut := clip.Intersect(cover)
	if cut.Empty() {
		return clip
	}
	if cut.Min.X == clip.Min.X && cut.Max.X == clip.Max.X {
		if cut.Min.Y == clip.Min.Y {
			clip.Min.Y = cut.Max.Y
		} else if cut.Max.Y == clip.Max.Y {
			clip.Max.Y = cut.Min.Y
		}
	} else if cut.Min.Y == clip.Min.Y && cut.Max.Y == clip.Max.Y {
		if cut.Min.X == clip.Min.X {
			clip.Min.X = cut.Max.X
		} else if cut.Max.X == clip.Max.X {
			clip.Max.X = cut.Min.X
		}
	}
	return clip
}

// opaqueInnerRect chooses one central strip, inset a pixel from its boundaries,
// and intersects it with every ancestor. Unsupported masks cannot occlude.
func opaqueInnerRect(img *imageOp) image.Rectangle {
	cover := img.clip
	for p := img.path; p != nil; p = p.parent {
		if !p.path {
			continue
		}
		q := p.primitive
		if q.kind != analyticRRect {
			return image.Rectangle{}
		}
		b := q.bounds.Add(p.off)
		nw, ne, se, sw := q.radii[0], q.radii[1], q.radii[2], q.radii[3]
		h := inwardOcclusionRect(cover, b.Min.X+1, b.Min.Y+max(nw, ne)+1, b.Max.X-1, b.Max.Y-max(sw, se)-1)
		v := inwardOcclusionRect(cover, b.Min.X+max(nw, sw)+1, b.Min.Y+1, b.Max.X-max(ne, se)-1, b.Max.Y-1)
		if h.Dx()*h.Dy() >= v.Dx()*v.Dy() {
			cover = h
		} else {
			cover = v
		}
		if cover.Empty() {
			return image.Rectangle{}
		}
	}
	return cover
}

func inwardOcclusionRect(clip image.Rectangle, minx, miny, maxx, maxy float32) image.Rectangle {
	minx, miny = max(minx, float32(clip.Min.X)), max(miny, float32(clip.Min.Y))
	maxx, maxy = min(maxx, float32(clip.Max.X)), min(maxy, float32(clip.Max.Y))
	if !(minx < maxx && miny < maxy) {
		return image.Rectangle{}
	}
	r := image.Rectangle{
		Min: image.Pt(int(math.Ceil(float64(minx))), int(math.Ceil(float64(miny)))),
		Max: image.Pt(int(math.Floor(float64(maxx))), int(math.Floor(float64(maxy)))),
	}
	if r.Empty() {
		return image.Rectangle{}
	}
	return r
}
