// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"encoding/binary"
	"image/color"
	"math"

	"gioui.org/internal/f32"
	"gioui.org/internal/ops"
)

// Instructions retain local state, not viewport-dependent draw records. Keep
// the instruction stream compact; only the operations that need a payload
// read one from the corresponding typed slice.
type decodedOp struct {
	key  ops.Key
	arg  uint32
	typ  ops.OpType
	push bool
}

type decodedClip struct {
	bounds  f32.Rectangle
	outline bool
}

type macroProgram struct {
	ops        []decodedOp
	transforms []f32.Affine2D
	clips      []decodedClip
	gradients  []linearGradientOpData
	images     []imageOpData
	paths      [][]byte
}

func (p *macroProgram) reset() {
	clear(p.ops)
	clear(p.images)
	clear(p.paths)
	p.ops, p.transforms, p.clips = p.ops[:0], p.transforms[:0], p.clips[:0]
	p.gradients, p.images, p.paths = p.gradients[:0], p.images[:0], p.paths[:0]
}

type macroDependency struct {
	macro   ops.Macro
	version uint32
}

type cachedMacro struct {
	source       ops.Macro
	version      uint32
	used, seen   uint64
	program      macroProgram
	dependencies []macroDependency
	rejected     bool
}

// Waiting for a second frame avoids compiling temporary recordings. Limits
// bound decoded storage; unused entries expire to release source Ops/images.
const (
	maxDecodedMacros     = 64
	maxDecodedOps        = 256
	maxMacroDependencies = 32
)

type drawReader struct {
	reader    ops.Reader
	encoded   ops.EncodedOp
	scratch   decodedOp
	root      *ops.Ops
	frame     uint64
	macros    map[ops.Macro]*cachedMacro
	current   *macroProgram
	pc        int
	saturated bool
}

func (r *drawReader) Reset(root *ops.Ops) {
	r.reader.Reset(root)
	r.encoded = ops.EncodedOp{}
	r.root, r.current, r.pc = root, nil, 0
	r.frame++
	r.saturated = false
	for key, entry := range r.macros {
		if entry.used+120 < r.frame {
			delete(r.macros, key)
		}
	}
}

func (r *drawReader) Decode() (*decodedOp, bool) {
	for {
		if r.current != nil && r.pc < len(r.current.ops) {
			op := &r.current.ops[r.pc]
			r.pc++
			return op, true
		}
		r.current = nil
		enc, ok := r.reader.DecodeMacro()
		if !ok {
			return nil, false
		}
		if ops.OpType(enc.Data[0]) == ops.TypeCall {
			macro := ops.DecodeMacro(enc)
			// Root Ops usually reset each frame. Their temporary layout macros
			// would add cache work and retain stale frame references.
			if macro.Ops != r.root {
				if program := r.program(macro); program != nil {
					r.reader.SkipMacro()
					r.current, r.pc = program, 0
				}
			}
			continue
		}
		// Uncached operations retain the original decoders; no wide scratch
		// payload is initialized or copied for these operations.
		r.encoded = enc
		r.scratch = decodedOp{key: enc.Key, typ: ops.OpType(enc.Data[0])}
		return &r.scratch, true
	}
}

func (r *drawReader) program(macro ops.Macro) *macroProgram {
	entry := r.macros[macro]
	if entry == nil {
		if r.saturated {
			return nil
		}
		if r.macros == nil {
			r.macros = make(map[ops.Macro]*cachedMacro, maxDecodedMacros)
		}
		if len(r.macros) == maxDecodedMacros {
			// Never displace another macro used in this frame, avoiding churn
			// when a scene contains more recordings than the cache can hold.
			var oldest *cachedMacro
			for _, e := range r.macros {
				if e.used < r.frame && (oldest == nil || e.used < oldest.used) {
					oldest = e
				}
			}
			if oldest == nil {
				r.saturated = true
				return nil
			}
			delete(r.macros, oldest.source)
			oldest.program.reset()
			clear(oldest.dependencies)
			entry = oldest
			*entry = cachedMacro{program: entry.program, dependencies: entry.dependencies[:0]}
		} else {
			entry = new(cachedMacro)
		}
		entry.source, entry.version = macro, macro.Version()
		entry.seen, entry.used = r.frame, r.frame
		r.macros[macro] = entry
		return nil
	}
	entry.used = r.frame
	valid := entry.version == macro.Version()
	for _, dep := range entry.dependencies {
		if dep.version != dep.macro.Version() {
			valid = false
			break
		}
	}
	if !valid {
		entry.program.reset()
		clear(entry.dependencies)
		entry.dependencies = entry.dependencies[:0]
		entry.version, entry.seen, entry.rejected = macro.Version(), r.frame, false
	}
	if entry.rejected || entry.seen == r.frame {
		return nil
	}
	if len(entry.program.ops) == 0 {
		if !entry.compile(macro, 0) || len(entry.program.ops) < 2 {
			entry.program.reset()
			entry.rejected = true
			return nil
		}
	}
	return &entry.program
}

func (e *cachedMacro) compile(macro ops.Macro, depth int) bool {
	if depth == maxMacroDependencies {
		return false
	}
	found := false
	for _, dep := range e.dependencies {
		if dep.macro.Ops == macro.Ops {
			found = true
			break
		}
	}
	if !found {
		if len(e.dependencies) == maxMacroDependencies {
			return false
		}
		e.dependencies = append(e.dependencies, macroDependency{macro: macro, version: macro.Version()})
	}
	pc := macro.Start
	p := &e.program
	for enc, ok := macro.Decode(&pc); ok; enc, ok = macro.Decode(&pc) {
		d := decodedOp{key: enc.Key, typ: ops.OpType(enc.Data[0])}
		switch d.typ {
		case ops.TypeCall:
			if !e.compile(ops.DecodeMacro(enc), depth+1) {
				return false
			}
			continue
		case ops.TypeDefer, ops.TypeSave, ops.TypeLoad, ops.TypeEmbed:
			// Preserve Reader's deferred queue and native-view geometry handling.
			return false
		case ops.TypeTransform:
			t, push := ops.DecodeTransform(enc.Data)
			d.arg, d.push = uint32(len(p.transforms)), push
			p.transforms = append(p.transforms, t)
		case ops.TypeClip:
			var clip ops.ClipOp
			clip.Decode(enc.Data)
			d.arg = uint32(len(p.clips))
			p.clips = append(p.clips, decodedClip{f32.FRect(clip.Bounds), clip.Outline})
		case ops.TypeLinearGradient:
			d.arg = uint32(len(p.gradients))
			p.gradients = append(p.gradients, decodeLinearGradientOp(enc.Data))
		case ops.TypeImage:
			d.arg = uint32(len(p.images))
			p.images = append(p.images, decodeImageOp(enc.Data, enc.Refs))
		case ops.TypeAux:
			d.arg = uint32(len(p.paths))
			p.paths = append(p.paths, enc.Data[ops.TypeAuxLen:])
		case ops.TypePushOpacity, ops.TypeStroke, ops.TypeColor:
			d.arg = binary.LittleEndian.Uint32(enc.Data[1:])
		case ops.TypePopTransform, ops.TypePopOpacity, ops.TypePath, ops.TypePopClip, ops.TypePaint:
		default:
			continue // The rendering collector ignores input/semantic operations.
		}
		if len(p.ops) == maxDecodedOps {
			return false
		}
		p.ops = append(p.ops, d)
	}
	return true
}

func (r *drawReader) gradient(d *decodedOp) linearGradientOpData {
	if r.current != nil {
		return r.current.gradients[d.arg]
	}
	return decodeLinearGradientOp(r.encoded.Data)
}

func (r *drawReader) image(d *decodedOp) imageOpData {
	if r.current != nil {
		return r.current.images[d.arg]
	}
	return decodeImageOp(r.encoded.Data, r.encoded.Refs)
}

func (r *drawReader) path(d *decodedOp) []byte {
	if r.current != nil {
		return r.current.paths[d.arg]
	}
	return r.encoded.Data[ops.TypeAuxLen:]
}

func (r *drawReader) color(d *decodedOp) color.NRGBA {
	if r.current != nil {
		c := d.arg
		return color.NRGBA{uint8(c), uint8(c >> 8), uint8(c >> 16), uint8(c >> 24)}
	}
	return decodeColorOp(r.encoded.Data)
}

func (r *drawReader) value(d *decodedOp) float32 {
	if r.current != nil {
		return math.Float32frombits(d.arg)
	}
	return decodeStrokeOp(r.encoded.Data) // Stroke and opacity have the same float layout.
}
