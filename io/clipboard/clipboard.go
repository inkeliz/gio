package clipboard

import (
	"gioui.org/internal/opconst"
	"gioui.org/io/event"
	"gioui.org/op"
)

type ReadClipboardOp struct {
	Tag event.Tag
}

type WriteClipboardOp struct {
	Text string
}

// Event is generated when a handler request
// the ReadClipboardOp
type Event struct {
	Text string
}

func (h ReadClipboardOp) Add(o *op.Ops) {
	data := o.Write1(opconst.TypeReadClipboardLen, h.Tag)
	data[0] = byte(opconst.TypeReadClipboard)
}

func (h WriteClipboardOp) Add(o *op.Ops) {
	data := o.Write1(opconst.TypeWriteClipboardLen, &h.Text)
	data[0] = byte(opconst.TypeWriteClipboard)
}


func (Event) ImplementsEvent()  {}