package router

import (
	"gioui.org/io/clipboard"
	"gioui.org/io/event"
	"gioui.org/op"
	"testing"
)

func TestClipboardDuplicateEvent(t *testing.T) {
	ops, router, handler := new(op.Ops), new(Router), make([]int, 2)

	// Both must receive the event once
	clipboard.ReadOp{Tag: &handler[0]}.Add(ops)
	clipboard.ReadOp{Tag: &handler[1]}.Add(ops)

	router.Frame(ops)
	event := clipboard.Event{Text: "Test"}
	router.Add(event)
	assertSystemClipboardReadClipboard(t, router, 0)
	assertSystemClipboardEvent(t, router.Events(&handler[0]), true)
	assertSystemClipboardEvent(t, router.Events(&handler[1]), true)
	ops.Reset()

	// No ReadClipboardOp

	router.Frame(ops)
	assertSystemClipboardReadClipboard(t, router, 0)
	assertSystemClipboardEvent(t, router.Events(&handler[0]), false)
	assertSystemClipboardEvent(t, router.Events(&handler[1]), false)
	ops.Reset()

	clipboard.ReadOp{Tag: &handler[0]}.Add(ops)

	router.Frame(ops)
	// No ClipboardEvent sent
	assertSystemClipboardReadClipboard(t, router, 1)
	assertSystemClipboardEvent(t, router.Events(&handler[0]), false)
	assertSystemClipboardEvent(t, router.Events(&handler[1]), false)
	ops.Reset()
}

func TestQueueProcessReadClipboard(t *testing.T) {
	ops, router, handler := new(op.Ops), new(Router), make([]int, 2)
	ops.Reset()

	// Request read
	clipboard.ReadOp{Tag: &handler[0]}.Add(ops)

	router.Frame(ops)
	assertSystemClipboardReadClipboard(t, router, 1)
	ops.Reset()

	// No ReadClipboardOp
	// One receiver must still wait for response

	router.Frame(ops)
	assertSystemClipboardReadClipboardDuplicated(t, router, 1)
	ops.Reset()

	// No ReadClipboardOp
	// One receiver must still wait for response

	router.Frame(ops)
	assertSystemClipboardReadClipboardDuplicated(t, router, 1)
	ops.Reset()

	// No ReadClipboardOp
	// One receiver must still wait for response

	router.Frame(ops)
	// Send the clipboard event
	event := clipboard.Event{Text: "Text 2"}
	router.Add(event)
	assertSystemClipboardReadClipboard(t, router, 0)
	assertSystemClipboardEvent(t, router.Events(&handler[0]), true)
	ops.Reset()

	// No ReadClipboardOp
	// There's no receiver waiting

	router.Frame(ops)
	assertSystemClipboardReadClipboard(t, router, 0)
	assertSystemClipboardEvent(t, router.Events(&handler[0]), false)
	ops.Reset()
}

func TestQueueProcessWriteClipboard(t *testing.T) {
	ops, router := new(op.Ops), new(Router)
	ops.Reset()

	clipboard.WriteOp{Text: "Write 1"}.Add(ops)

	router.Frame(ops)
	assertSystemClipboardWaitingWrite(t, router, "Write 1")
	ops.Reset()

	// No WriteClipboardOp

	router.Frame(ops)
	assertSystemClipboardWaitingWrite(t, router, "")
	ops.Reset()

	clipboard.WriteOp{Text: "Write 2"}.Add(ops)

	router.Frame(ops)
	assertSystemClipboardReadClipboard(t, router, 0)
	assertSystemClipboardWaitingWrite(t, router, "Write 2")
	ops.Reset()

	// No WriteClipboardOp

	router.Frame(ops)
	assertSystemClipboardWaitingWrite(t, router, "")
	ops.Reset()
}

func assertSystemClipboardEvent(t *testing.T, events []event.Event, expected bool) {
	t.Helper()
	var evtClipboard int
	for _, e := range events {
		switch e.(type) {
		case clipboard.Event:
			evtClipboard++
		}
	}
	if evtClipboard <= 0 && expected {
		t.Errorf("expect to receive some event")
	}
	if evtClipboard > 0 && !expected {
		t.Errorf("unexpect event received")
	}
}

func assertSystemClipboardReadClipboard(t *testing.T, router *Router, expected int) {
	t.Helper()
	if len(router.cqueue.receivers) != expected {
		t.Error("unexpect amount of receivers")
	}
	if router.cqueue.ReadClipboard() != (expected > 0) {
		t.Error("missing requests")
	}
}

func assertSystemClipboardReadClipboardDuplicated(t *testing.T, router *Router, expected int) {
	t.Helper()
	if len(router.cqueue.receivers) != expected {
		t.Error("receivers removed")
	}
	if router.cqueue.ReadClipboard() != false {
		t.Error("duplicated requests")
	}
}

func assertSystemClipboardWaitingWrite(t *testing.T, router *Router, expected string) {
	t.Helper()
	if (router.cqueue.text != nil) != (expected != "") {
		t.Error("text not defined")
	}
	text, ok := router.cqueue.WriteClipboard()
	if ok != (expected != "") {
		t.Error("duplicated requests")
	}
	if text != expected {
		t.Error("text didn't match")
	}
}
