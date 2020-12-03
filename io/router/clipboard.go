// SPDX-License-Identifier: Unlicense OR MIT

package router

import (
	"gioui.org/internal/opconst"
	"gioui.org/internal/ops"
	"gioui.org/io/event"
)

type clipboardQueue struct {
	receiver map[event.Tag]struct{}
	// requested avoids multiples clipboard reads while wait for response.
	requested bool
	text      *string
	reader    ops.Reader
}

// WriteClipboard returns the last text supossed to be
// copied to clipboard as determined in Frame.
func (q *clipboardQueue) WriteClipboard() (string, bool) {
	if q.text == nil {
		return "", false
	}
	t := q.text
	q.text = nil
	return *t, true
}

// ReadClipboard returns true if there's any request
// to read the clipboard.
func (q *clipboardQueue) ReadClipboard() bool {
	if len(q.receiver) <= 0 || q.requested {
		return false
	}
	q.requested = true
	return true
}

func (q *clipboardQueue) SetWriteClipboard(d []byte, refs []interface{}) {
	if q.receiver == nil {
		q.receiver = make(map[event.Tag]struct{})
	}
	if opconst.OpType(d[0]) != opconst.TypeClipboardWrite {
		panic("invalid op")
	}
	q.text = refs[0].(*string)
}

func (q *clipboardQueue) SetReadClipboard(d []byte, refs []interface{}) {
	if opconst.OpType(d[0]) != opconst.TypeClipboardRead {
		panic("invalid op")
	}
	q.receiver[refs[0].(event.Tag)] = struct{}{}
	q.requested = false
}

func (q *clipboardQueue) Push(e event.Event, events *handlerEvents) {
	for r := range q.receiver {
		events.Add(r, e)
	}
	q.receiver = nil
}