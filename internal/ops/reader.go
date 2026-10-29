// SPDX-License-Identifier: Unlicense OR MIT

package ops

import (
	"encoding/binary"
)

// Reader parses an ops list.
type Reader struct {
	pc        PC
	stack     []macro
	ops       *Ops
	deferOps  Ops
	deferDone bool
}

// EncodedOp represents an encoded op returned by
// Reader.
type EncodedOp struct {
	Key  Key
	Data []byte
	Refs []any
}

// Key is a unique key for a given op.
type Key struct {
	ops     *Ops
	pc      uint32
	version uint32
}

// Shadow of op.MacroOp.
type macroOp struct {
	ops   *Ops
	start PC
	end   PC
}

// PC is an instruction counter for an operation list.
type PC struct {
	data uint32
	refs uint32
}

type macro struct {
	ops   *Ops
	retPC PC
	endPC PC
}

type opMacroDef struct {
	endpc PC
}

func (pc PC) Add(op OpType) PC {
	size, numRefs := op.props()
	return PC{
		data: pc.data + size,
		refs: pc.refs + numRefs,
	}
}

// Reset start reading from the beginning of ops.
func (r *Reader) Reset(ops *Ops) {
	r.ResetAt(ops, PC{})
}

// ResetAt is like Reset, except it starts reading from pc.
func (r *Reader) ResetAt(ops *Ops, pc PC) {
	r.stack = r.stack[:0]
	Reset(&r.deferOps)
	r.deferDone = false
	r.pc = pc
	r.ops = ops
}

// Macro identifies a recorded range. Version changes when its Ops is reset.
// The range is part of the identity because one Ops can hold several macros.
type Macro struct {
	Ops        *Ops
	Start, End PC
}

func DecodeMacro(op EncodedOp) Macro {
	var m macroOp
	m.decode(op.Data, op.Refs)
	return Macro{Ops: m.ops, Start: m.start, End: m.end}
}

func (m Macro) Version() uint32 { return m.Ops.version }

// Decode reads one unexpanded operation in the macro. It skips definitions,
// leaving calls and deferred operations for the caller to handle.
func (m Macro) Decode(pc *PC) (EncodedOp, bool) {
	for pc.data < m.End.data {
		data := m.Ops.data[pc.data:m.End.data]
		typ := OpType(data[0])
		n, nrefs := typ.props()
		if typ == TypeMacro {
			var def opMacroDef
			def.decode(data)
			if def.endpc == (PC{}) {
				return EncodedOp{}, false
			}
			*pc = def.endpc
			continue
		}
		if typ == TypeAux {
			n = m.End.data - pc.data
		}
		enc := EncodedOp{Key: Key{ops: m.Ops, pc: pc.data, version: m.Version()}, Data: data[:n], Refs: m.Ops.refs[pc.refs : pc.refs+nrefs]}
		pc.data += n
		pc.refs += nrefs
		return enc, true
	}
	return EncodedOp{}, false
}

func (r *Reader) Decode() (EncodedOp, bool) { return r.decode(false) }

// DecodeMacro is Decode with calls reported just after entering their body.
// SkipMacro can then bypass a body that the caller has already decoded.
// Normal Decode users, including input routing, never see these calls.
func (r *Reader) DecodeMacro() (EncodedOp, bool) { return r.decode(true) }

func (r *Reader) SkipMacro() {
	n := len(r.stack) - 1
	m := r.stack[n]
	r.ops, r.pc = m.ops, m.retPC
	r.stack = r.stack[:n]
}

func (r *Reader) decode(reportCalls bool) (EncodedOp, bool) {
	if r.ops == nil {
		return EncodedOp{}, false
	}
	deferring := false
	for {
		if len(r.stack) > 0 {
			b := r.stack[len(r.stack)-1]
			if r.pc == b.endPC {
				r.ops = b.ops
				r.pc = b.retPC
				r.stack = r.stack[:len(r.stack)-1]
				continue
			}
		}
		data := r.ops.data
		data = data[r.pc.data:]
		refs := r.ops.refs
		if len(data) == 0 {
			if r.deferDone {
				return EncodedOp{}, false
			}
			r.deferDone = true
			// Execute deferred macros.
			r.ops = &r.deferOps
			r.pc = PC{}
			continue
		}
		key := Key{ops: r.ops, pc: r.pc.data, version: r.ops.version}
		t := OpType(data[0])
		n, nrefs := t.props()
		data = data[:n]
		refs = refs[r.pc.refs:]
		refs = refs[:nrefs]
		switch t {
		case TypeDefer:
			deferring = true
			r.pc.data += n
			r.pc.refs += nrefs
			continue
		case TypeAux:
			// An Aux operations is always wrapped in a macro, and
			// its length is the remaining space.
			block := r.stack[len(r.stack)-1]
			n += block.endPC.data - r.pc.data - TypeAuxLen
			data = data[:n]
		case TypeCall:
			if deferring {
				deferring = false
				// Copy macro for deferred execution.
				if nrefs != 1 {
					panic("internal error: unexpected number of macro refs")
				}
				deferData := Write1(&r.deferOps, int(n), refs[0])
				copy(deferData, data)
				r.pc.data += n
				r.pc.refs += nrefs
				continue
			}
			var op macroOp
			op.decode(data, refs)
			retPC := r.pc
			retPC.data += n
			retPC.refs += nrefs
			r.stack = append(r.stack, macro{
				ops:   r.ops,
				retPC: retPC,
				endPC: op.end,
			})
			r.ops = op.ops
			r.pc = op.start
			if reportCalls {
				return EncodedOp{Key: key, Data: data, Refs: refs}, true
			}
			continue
		case TypeMacro:
			var op opMacroDef
			op.decode(data)
			if op.endpc != (PC{}) {
				r.pc = op.endpc
			} else {
				// Treat an incomplete macro as containing all remaining ops.
				r.pc.data = uint32(len(r.ops.data))
				r.pc.refs = uint32(len(r.ops.refs))
			}
			continue
		}
		r.pc.data += n
		r.pc.refs += nrefs
		return EncodedOp{Key: key, Data: data, Refs: refs}, true
	}
}

func (op *opMacroDef) decode(data []byte) {
	if len(data) < TypeMacroLen || OpType(data[0]) != TypeMacro {
		panic("invalid op")
	}
	bo := binary.LittleEndian
	data = data[:TypeMacroLen]
	op.endpc.data = bo.Uint32(data[1:])
	op.endpc.refs = bo.Uint32(data[5:])
}

func (m *macroOp) decode(data []byte, refs []any) {
	if len(data) < TypeCallLen || len(refs) < 1 || OpType(data[0]) != TypeCall {
		panic("invalid op")
	}
	bo := binary.LittleEndian
	data = data[:TypeCallLen]

	m.ops = refs[0].(*Ops)
	m.start.data = bo.Uint32(data[1:])
	m.start.refs = bo.Uint32(data[5:])
	m.end.data = bo.Uint32(data[9:])
	m.end.refs = bo.Uint32(data[13:])
}
