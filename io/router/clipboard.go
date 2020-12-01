package router

import (
	"gioui.org/internal/opconst"
	"gioui.org/internal/ops"
	"gioui.org/io/event"
	"gioui.org/op"
)

type clipboardQueue struct {
	receiver  event.Tag
	requested bool
	text      *string
	reader    ops.Reader
}

// WriteClipboard returns the last text supossed to be
// copied to clipboard as determined in Frame.
func (q *clipboardQueue) WriteClipboard() *string {
	if q.text != nil {
		t := q.text
		q.text = nil
		return t
	}
	return nil
}

// ReadClipboard returns true if there's any request
// to read the clipboard.
func (q *clipboardQueue) ReadClipboard() bool {
	if q.receiver != nil && !q.requested {
		q.requested = true
		return true
	}
	return false
}

func (q *clipboardQueue) Frame(root *op.Ops, events *handlerEvents) {
	q.reader.Reset(root)

	receiver, text := q.resolveClipboard(events)
	if text != nil {
		q.text = text
	}
	if receiver != nil {
		q.receiver = receiver
		q.requested = false
	}
}

func (q *clipboardQueue) Push(e event.Event, events *handlerEvents) {
	if q.receiver != nil {
		events.Add(q.receiver, e)
		q.receiver = nil
	}
}

func (q *clipboardQueue) resolveClipboard(events *handlerEvents) (receiver event.Tag, text *string) {
loop:
	for encOp, ok := q.reader.Decode(); ok; encOp, ok = q.reader.Decode() {
		switch opconst.OpType(encOp.Data[0]) {
		case opconst.TypeWriteClipboard:
			text = decodeWriteClipboard(encOp.Data, encOp.Refs)
		case opconst.TypeReadClipboard:
			receiver = decodeReadClipboard(encOp.Data, encOp.Refs)
		case opconst.TypePush:
			newReceiver, newWrite := q.resolveClipboard(events)
			if newWrite != nil {
				text = newWrite
			}
			if newReceiver != nil {
				receiver = newReceiver
			}
		case opconst.TypePop:
			break loop
		}
	}
	return receiver, text
}

func decodeWriteClipboard(d []byte, refs []interface{}) *string {
	if opconst.OpType(d[0]) != opconst.TypeWriteClipboard {
		panic("invalid op")
	}
	return refs[0].(*string)
}

func decodeReadClipboard(d []byte, refs []interface{}) event.Tag {
	if opconst.OpType(d[0]) != opconst.TypeReadClipboard {
		panic("invalid op")
	}
	return refs[0].(event.Tag)
}
