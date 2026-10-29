// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"image"
	"math"
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
	}
	// Collection adds direct paints to a layer. Its complete bounds must also
	// include children before testing whether an outside rectangle hides it.
	for i := len(d.layers) - 1; i >= 0; i-- {
		l := &d.layers[i]
		if l.parent >= 0 {
			d.layers[l.parent].clip = d.layers[l.parent].clip.Union(l.clip)
		}
	}
	var hidden [maxCovers]bool
	var covers [maxCovers]image.Rectangle
	ncovers := 0
	for i := len(d.imageOps) - 1; i >= 0; i-- {
		// Leave later scopes before entering an earlier adjacent scope.
		for j, l := range d.layers {
			if !hidden[j] && l.opStart < l.opEnd && l.opStart == i+1 {
				ncovers = 0
			}
		}
		jump := false
		// Layers are recorded in preorder: test parents before children that
		// share the same final paint.
		for j, l := range d.layers {
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
