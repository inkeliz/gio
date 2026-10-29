// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"encoding/binary"
	"math"

	"gioui.org/internal/f32"
	"gioui.org/internal/ops"
	"gioui.org/internal/scene"
)

type analyticPrimitiveKind uint8

const (
	analyticRRect analyticPrimitiveKind = iota + 1
	analyticEllipse
)

// analyticPrimitive describes only the canonical paths emitted by clip.RRect
// and clip.Ellipse. Radii are ordered NW, NE, SE, SW; ellipse radii follow from
// bounds. Recognition does not authorize replacing their cubic circle profiles
// with a different coverage approximation.
type analyticPrimitive struct {
	kind   analyticPrimitiveKind
	bounds f32.Rectangle
	radii  [4]float32
}

// recognizeAnalyticPrimitive recognizes untransformed, filled path commands.
// Callers must exclude strokes and unsupported transformations. A bounded
// command count and exact reconstruction avoid scanning arbitrary large paths,
// accepting approximate matches, or adding metadata to the operations protocol.
func recognizeAnalyticPrimitive(data []byte) (analyticPrimitive, bool) {
	const stride = scene.CommandSize + 4
	if len(data)%stride != 0 || len(data) < 4*stride || len(data) > 8*stride {
		return analyticPrimitive{}, false
	}
	var storage [8]scene.Command
	commands := storage[:len(data)/stride]
	contour := binary.LittleEndian.Uint32(data)
	inf := float32(math.Inf(1))
	bounds := f32.Rectangle{Min: f32.Pt(inf, inf), Max: f32.Pt(-inf, -inf)}
	for i := range commands {
		record := data[i*stride : (i+1)*stride]
		if binary.LittleEndian.Uint32(record) != contour {
			return analyticPrimitive{}, false
		}
		cmd := ops.DecodeCommand(record[4:])
		commands[i] = cmd
		var end int
		switch cmd.Op() {
		case scene.OpLine:
			end = 4
		case scene.OpCubic:
			end = 8
		default:
			return analyticPrimitive{}, false
		}
		for j := 1; j < end; j += 2 {
			x, y := math.Float32frombits(cmd[j]), math.Float32frombits(cmd[j+1])
			if !(x >= -math.MaxFloat32 && x <= math.MaxFloat32 && y >= -math.MaxFloat32 && y <= math.MaxFloat32) {
				return analyticPrimitive{}, false
			}
			bounds.Min.X, bounds.Min.Y = min(bounds.Min.X, x), min(bounds.Min.Y, y)
			bounds.Max.X, bounds.Max.Y = max(bounds.Max.X, x), max(bounds.Max.Y, y)
		}
	}
	if bounds.Empty() {
		return analyticPrimitive{}, false
	}
	if p, ok := recognizeEllipse(commands, bounds); ok {
		return p, true
	}
	return recognizeRRect(commands, bounds)
}

func recognizeEllipse(commands []scene.Command, bounds f32.Rectangle) (analyticPrimitive, bool) {
	if len(commands) != 4 {
		return analyticPrimitive{}, false
	}
	// clip.Ellipse takes integer bounds. Its float32 aspect-ratio operations
	// can move extrema by a few ULPs; reconstruct from the nearest integers
	// and still require every resulting cubic coordinate to match exactly.
	w := float32(math.Round(float64(bounds.Min.X)))
	n := float32(math.Round(float64(bounds.Min.Y)))
	e := float32(math.Round(float64(bounds.Max.X)))
	s := float32(math.Round(float64(bounds.Max.Y)))
	if !(w < e && n < s) {
		return analyticPrimitive{}, false
	}
	const q = 4 * (math.Sqrt2 - 1) / 3
	center := f32.Pt(w+e, n+s).Mul(.5)
	r := (e - w) * .5
	scale := (s - n) / (e - w)
	curve := r * q
	top := f32.Pt(center.X, center.Y-r*scale)
	right := f32.Pt(center.X+r, center.Y)
	bottom := f32.Pt(center.X, center.Y+r*scale)
	left := f32.Pt(center.X-r, center.Y)
	expected := [...]scene.Command{
		scene.Cubic(top, f32.Pt(center.X+curve, center.Y-r*scale), f32.Pt(center.X+r, center.Y-curve*scale), right),
		scene.Cubic(right, f32.Pt(center.X+r, center.Y+curve*scale), f32.Pt(center.X+curve, center.Y+r*scale), bottom),
		scene.Cubic(bottom, f32.Pt(center.X-curve, center.Y+r*scale), f32.Pt(center.X-r, center.Y+curve*scale), left),
		scene.Cubic(left, f32.Pt(center.X-r, center.Y-curve*scale), f32.Pt(center.X-curve, center.Y-r*scale), top),
	}
	if !matchPrimitiveCommands(commands, expected[:]) {
		return analyticPrimitive{}, false
	}
	return analyticPrimitive{kind: analyticEllipse, bounds: f32.Rectangle{Min: f32.Pt(w, n), Max: f32.Pt(e, s)}}, true
}

func recognizeRRect(commands []scene.Command, bounds f32.Rectangle) (analyticPrimitive, bool) {
	w, n, e, s := bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y
	var radii [4]float32
	for _, cmd := range commands {
		if cmd.Op() != scene.OpCubic {
			continue
		}
		from, _, _, to := scene.DecodeCubic(cmd)
		var corner int
		var rx, ry float32
		switch {
		case from.Y == n && to.X == e && from.X < e && to.Y > n:
			corner, rx, ry = 1, e-from.X, to.Y-n
		case from.X == e && to.Y == s && from.Y < s && to.X < e:
			corner, rx, ry = 2, e-to.X, s-from.Y
		case from.Y == s && to.X == w && from.X > w && to.Y < s:
			corner, rx, ry = 3, from.X-w, s-to.Y
		case from.X == w && to.Y == n && from.Y > n && to.X > w:
			corner, rx, ry = 0, to.X-w, from.Y-n
		default:
			return analyticPrimitive{}, false
		}
		if rx != ry || radii[corner] != 0 {
			return analyticPrimitive{}, false
		}
		radii[corner] = rx
	}
	nw, ne, se, sw := radii[0], radii[1], radii[2], radii[3]
	if nw+ne > e-w || sw+se > e-w || nw+sw > s-n || ne+se > s-n {
		return analyticPrimitive{}, false
	}
	const q = 4 * (math.Sqrt2 - 1) / 3
	const iq = 1 - q
	topLeft, topRight := f32.Pt(w+nw, n), f32.Pt(e-ne, n)
	rightTop, rightBottom := f32.Pt(e, n+ne), f32.Pt(e, s-se)
	bottomRight, bottomLeft := f32.Pt(e-se, s), f32.Pt(w+sw, s)
	leftBottom, leftTop := f32.Pt(w, s-sw), f32.Pt(w, n+nw)
	expected := [...]scene.Command{
		scene.Line(topLeft, topRight),
		scene.Cubic(topRight, f32.Pt(e-ne*iq, n), f32.Pt(e, n+ne*iq), rightTop),
		scene.Line(rightTop, rightBottom),
		scene.Cubic(rightBottom, f32.Pt(e, s-se*iq), f32.Pt(e-se*iq, s), bottomRight),
		scene.Line(bottomRight, bottomLeft),
		scene.Cubic(bottomLeft, f32.Pt(w+sw*iq, s), f32.Pt(w, s-sw*iq), leftBottom),
		scene.Line(leftBottom, leftTop),
		scene.Cubic(leftTop, f32.Pt(w, n+nw*iq), f32.Pt(w+nw*iq, n), topLeft),
	}
	if !matchPrimitiveCommands(commands, expected[:]) {
		return analyticPrimitive{}, false
	}
	return analyticPrimitive{kind: analyticRRect, bounds: bounds, radii: radii}, true
}

func matchPrimitiveCommands(commands, expected []scene.Command) bool {
	i := 0
	for _, cmd := range expected {
		// Path.LineTo and Path.CubeTo omit these exact degeneracies.
		switch cmd.Op() {
		case scene.OpLine:
			from, to := scene.DecodeLine(cmd)
			if from == to {
				continue
			}
		case scene.OpCubic:
			from, c0, c1, to := scene.DecodeCubic(cmd)
			if from == c0 && from == c1 && from == to {
				continue
			}
		}
		if i == len(commands) || commands[i] != cmd {
			return false
		}
		i++
	}
	return i == len(commands)
}

// containsRect reports whether r lies in a conservative opaque central region.
// inset is the antialiasing/support margin in the same coordinate system. The
// caller must account for transformations when converting a pixel AA margin.
func (p analyticPrimitive) containsRect(r f32.Rectangle, inset float32) bool {
	if r.Empty() || !(inset >= 0) {
		return false
	}
	inside := func(minx, miny, maxx, maxy float32) bool {
		return r.Min.X >= minx+inset && r.Min.Y >= miny+inset && r.Max.X <= maxx-inset && r.Max.Y <= maxy-inset
	}
	b := p.bounds
	switch p.kind {
	case analyticRRect:
		nw, ne, se, sw := p.radii[0], p.radii[1], p.radii[2], p.radii[3]
		return inside(b.Min.X, b.Min.Y+max(nw, ne), b.Max.X, b.Max.Y-max(sw, se)) ||
			inside(b.Min.X+max(nw, sw), b.Min.Y, b.Max.X-max(ne, se), b.Max.Y)
	case analyticEllipse:
		// The half-radius central box stays well inside the cubic ellipse,
		// including non-circular aspect ratios; the AA margin is additional.
		x, y := (b.Min.X+b.Max.X)*.5, (b.Min.Y+b.Max.Y)*.5
		rx, ry := b.Dx()*.25, b.Dy()*.25
		return inside(x-rx, y-ry, x+rx, y+ry)
	default:
		return false
	}
}
