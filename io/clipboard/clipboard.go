package clipboard

import (
	"gioui.org/internal/opconst"
	"gioui.org/io/event"
	"gioui.org/op"
)

// ReadOp requests the clipboard content, the
// declared handler will receive the content.
// Multiple reads may be coalesced to a single event.
type ReadOp struct {
	Tag event.Tag
}

// WriteOp writes the Text content to the clipboard.
type WriteOp struct {
	Text string
}

// Event is sent once for each request for the
// clipboard content.
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