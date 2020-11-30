// SPDX-License-Identifier: Unlicense OR MIT

package router

import (
	"gioui.org/internal/opconst"
	"gioui.org/internal/ops"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/op"
)

type TextInputState uint8

type keyQueue struct {
	focus    event.Tag
	handlers map[event.Tag]*keyHandler
	reader   ops.Reader
	state    TextInputState
}

type keyHandler struct {
	// visible will be true if the InputOp is present
	// in the current frame.
	visible bool
}

type listenerPriority uint8

const (
	priDefault listenerPriority = iota
	priCurrentFocus
	priNone
	priNewFocus
)

const (
	TextInputKeep TextInputState = iota
	TextInputClose
	TextInputOpen
)

// InputState returns the last text input state as
// determined in Frame.
func (q *keyQueue) InputState() TextInputState {
	return q.state
}

func (q *keyQueue) Frame(root *op.Ops, events *handlerEvents) {
	if q.handlers == nil {
		q.handlers = make(map[event.Tag]*keyHandler)
	}
	for _, h := range q.handlers {
		h.visible = false
	}
	q.reader.Reset(root)

	focus, _, keyboard := q.resolveFocus(events)
	for k, h := range q.handlers {
		if k != focus && h.visible {
			delete(q.handlers, k)
			events.Add(k, key.FocusEvent{Focus: false})
		}
	}

	if focus != q.focus {
		if q.focus != nil {
			events.Add(q.focus, key.FocusEvent{Focus: false})
		}
		q.focus = focus
		if q.focus != nil {
			events.Add(q.focus, key.FocusEvent{Focus: true})
		} else {
			keyboard = TextInputClose
		}
	}

	q.state = keyboard
}

func (q *keyQueue) Push(e event.Event, events *handlerEvents) {
	if q.focus != nil {
		events.Add(q.focus, e)
	}
}

func (q *keyQueue) resolveFocus(events *handlerEvents) (tag event.Tag, pri listenerPriority, keyboard TextInputState) {
loop:
	for encOp, ok := q.reader.Decode(); ok; encOp, ok = q.reader.Decode() {
		switch opconst.OpType(encOp.Data[0]) {
		case opconst.TypeKeyFocus:
			op := decodeFocusOp(encOp.Data, encOp.Refs)
			if op.Focus {
				pri = priNewFocus
			} else {
				pri, keyboard = priNone, TextInputClose
			}
		case opconst.TypeKeySoftKeyboard:
			op := decodeSoftKeyboardOp(encOp.Data, encOp.Refs)
			if op.Show {
				keyboard = TextInputOpen
			} else {
				keyboard = TextInputClose
			}
		case opconst.TypeKeyInput:
			op := decodeKeyInputOp(encOp.Data, encOp.Refs)

			tag = op.Tag
			if tag == q.focus {
				pri = priCurrentFocus
			}

			h, ok := q.handlers[tag]
			if !ok {
				h = new(keyHandler)
				q.handlers[tag] = h
			}
			h.visible = true
		case opconst.TypePush:
			newK, newPri, newKeyboard := q.resolveFocus(events)
			if newKeyboard > keyboard {
				keyboard = newKeyboard
			}
			if newPri.replaces(pri) {
				tag, pri = newK, newPri
			}
		case opconst.TypePop:
			break loop
		}
	}

	return tag, pri, keyboard
}

func (p listenerPriority) replaces(p2 listenerPriority) bool {
	// Favor earliest default focus or latest requested focus.
	return p > p2 || p == p2 && p == priNewFocus
}

func decodeKeyInputOp(d []byte, refs []interface{}) key.InputOp {
	if opconst.OpType(d[0]) != opconst.TypeKeyInput {
		panic("invalid op")
	}
	return key.InputOp{
		Tag: refs[0].(event.Tag),
	}
}

func decodeSoftKeyboardOp(d []byte, refs []interface{}) key.SoftKeyboardOp {
	if opconst.OpType(d[0]) != opconst.TypeKeySoftKeyboard {
		panic("invalid op")
	}
	return key.SoftKeyboardOp{
		Show: d[1] != 0,
	}
}

func decodeFocusOp(d []byte, refs []interface{}) key.FocusOp {
	if opconst.OpType(d[0]) != opconst.TypeKeyFocus {
		panic("invalid op")
	}
	return key.FocusOp{
		Focus: d[1] != 0,
	}
}
