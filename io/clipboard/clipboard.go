// SPDX-License-Identifier: Unlicense OR MIT

package clipboard

import (
	"gioui.org/internal/opconst"
	"gioui.org/io/event"
	"gioui.org/op"
)

// ReadOp requests the text of the clipboard, the
// declared handler will receive `clipboard.Event`
// once.
type ReadOp struct {
	Tag event.Tag
}

// WriteOp writes the Text into the clipboard.
type WriteOp struct {
	Text string
}

// Event is generated when a handler request
// the ReadOp
type Event struct {
	Text string
}

func (h ReadOp) Add(o *op.Ops) {
	data := o.Write1(opconst.TypeClipboardReadLen, h.Tag)
	data[0] = byte(opconst.TypeClipboardRead)
}

func (h WriteOp) Add(o *op.Ops) {
	data := o.Write1(opconst.TypeClipboardWriteLen, &h.Text)
	data[0] = byte(opconst.TypeClipboardWrite)
}


func (Event) ImplementsEvent()  {}