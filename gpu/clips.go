// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"image"

	"gioui.org/f32"
	if32 "gioui.org/internal/f32"
)

type prunedPath struct{ path, parent *pathOp }

func transformedPrimitive(p analyticPrimitive, t f32.Affine2D) analyticPrimitive {
	sx, hx, _, hy, sy, _ := t.Elems()
	if p.kind == 0 || sx <= 0 || sx != sy || hx != 0 || hy != 0 {
		return analyticPrimitive{}
	}
	p.bounds.Min = t.Transform(p.bounds.Min)
	p.bounds.Max = t.Transform(p.bounds.Max)
	for i := range p.radii {
		p.radii[i] *= sx
	}
	return p
}
func (d *drawOps) simplifyClips() {
	clear(d.prunedPaths)
	for i := range d.imageOps {
		img := &d.imageOps[i]
		if !img.culled {
			img.path = d.prunePath(img.path, img.clip)
		}
	}
}
func (d *drawOps) prunePath(p *pathOp, clip image.Rectangle) *pathOp {
	if p == nil {
		return nil
	}
	parent := d.prunePath(p.parent, clip)
	if p.path && p.primitive.kind != 0 && p.primitive.containsRect(if32.FRect(clip).Sub(p.off), 1) {
		return parent
	}
	if parent == p.parent {
		return p
	}
	key := prunedPath{p, parent}
	if copy := d.prunedPaths[key]; copy != nil {
		return copy
	}
	copy := d.newPathOp()
	*copy = *p
	copy.parent = parent
	copy.rect = !copy.path && (parent == nil || parent.rect)
	copy.clip = image.Rectangle{}
	if copy.path {
		d.pathOps = append(d.pathOps, copy)
	}
	if d.prunedPaths == nil {
		d.prunedPaths = make(map[prunedPath]*pathOp)
	}
	d.prunedPaths[key] = copy
	return copy
}
func (d *drawOps) useAnalyticPrimitives() {
	for i := range d.imageOps {
		img := &d.imageOps[i]
		if img.culled || img.material.material == materialEmbed {
			continue
		}
		var shape *pathOp
		multiple := false
		for p := img.path; p != nil; p = p.parent {
			if p.path {
				if shape != nil {
					multiple = true
					break
				}
				shape = p
			}
		}
		if multiple || shape == nil || shape.primitive.kind == 0 {
			continue
		}
		p := shape.primitive
		// Large stable surfaces are cheaper to sample from retained masks on
		// mobile GPUs; direct coverage is reserved for small UI primitives.
		if p.bounds.Dx()*p.bounds.Dy() > 4096 {
			continue
		}
		// Keep direct quadratic coverage within the validated UI-size range;
		// tiny radii and unsupported geometry retain the path renderer.
		if p.kind == analyticRRect {
			valid := true
			for _, r := range p.radii {
				if r != 0 && (r < 2 || r > 64 || r > p.bounds.Dx()/2 || r > p.bounds.Dy()/2) {
					valid = false
				}
			}
			if !valid {
				continue
			}
		} else if p.bounds.Dx() != p.bounds.Dy() || p.bounds.Dx() > 128 {
			continue
		}
		p.bounds = p.bounds.Add(shape.off)
		img.analytic = p
		img.path = nil
	}
}
