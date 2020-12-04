// SPDX-License-Identifier: Unlicense OR MIT

package router

import (
	"gioui.org/internal/opconst"
	"gioui.org/internal/ops"
	"gioui.org/io/event"
)

type systemQueue struct {
	receivers map[event.Tag]struct{}
	// request avoid read clipboard every frame while waiting.
	requested bool
	text      *string
	reader    ops.Reader
}

// WriteClipboard returns the most recent text to be copied
// to the clipboard, if any.
func (q *systemQueue) WriteClipboard() (string, bool) {
	if q.text == nil {
		return "", false
	}
	text := *q.text
	q.text = nil
	return text, true
}

// ReadClipboard reports if any new handler is waiting 
// to read the clipboard.
func (q *systemQueue) ReadClipboard() bool {
	if len(q.receivers) <= 0 || q.requested {
		return false
	}
	q.requested = true
	return true
}

func (q *systemQueue) Push(e event.Event, events *handlerEvents) {
	for r := range q.receivers {
		events.Add(r, e)
		delete(q.receivers, r)
	}
}

func (q *systemQueue) ProcessWriteClipboard(d []byte, refs []interface{}) {
	if opconst.OpType(d[0]) != opconst.TypeSystemClipboardWrite {
		panic("invalid op")
	}
	q.text = refs[0].(*string)
}

func (q *systemQueue) ProcessReadClipboard(d []byte, refs []interface{}) {
	if opconst.OpType(d[0]) != opconst.TypeSystemClipboardRead {
		panic("invalid op")
	}
	if q.receivers == nil {
		q.receivers = make(map[event.Tag]struct{})
	}
	tag := refs[0].(event.Tag)
	if _, ok := q.receivers[tag]; !ok {
		q.receivers[tag] = struct{}{}
		q.requested = false
	}
}
