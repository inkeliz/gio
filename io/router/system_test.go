package router

import (
	"gioui.org/io/event"
	"gioui.org/io/system"
	"gioui.org/op"
	"testing"
)

func TestSystemClipboardDuplicateEvent(t *testing.T) {
	ops, router, handler := new(op.Ops), new(Router), make([]int, 2)

	// Both must receive the event once
	system.ReadClipboardOp{Tag: &handler[0]}.Add(ops)
	system.ReadClipboardOp{Tag: &handler[1]}.Add(ops)

	router.Frame(ops)
	event := system.ClipboardEvent{Text: "Test"}
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

	system.ReadClipboardOp{Tag: &handler[0]}.Add(ops)

	router.Frame(ops)
	// No ClipboardEvent sent
	assertSystemClipboardReadClipboard(t, router, 1)
	assertSystemClipboardEvent(t, router.Events(&handler[0]), false)
	assertSystemClipboardEvent(t, router.Events(&handler[1]), false)
	ops.Reset()
}

func TestSystemQueueProcessReadClipboard(t *testing.T) {
	ops, router, handler := new(op.Ops), new(Router), make([]int, 2)
	ops.Reset()

	// Request read
	system.ReadClipboardOp{Tag: &handler[0]}.Add(ops)

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
	event := system.ClipboardEvent{Text: "Text 2"}
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

func TestSystemQueueProcessWriteClipboard(t *testing.T) {
	ops, router := new(op.Ops), new(Router)
	ops.Reset()

	system.WriteClipboardOp{Text: "Write 1"}.Add(ops)

	router.Frame(ops)
	assertSystemClipboardWaitingWrite(t, router, "Write 1")
	ops.Reset()

	// No WriteClipboardOp

	router.Frame(ops)
	assertSystemClipboardWaitingWrite(t, router, "")
	ops.Reset()

	system.WriteClipboardOp{Text: "Write 2"}.Add(ops)

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
		case system.ClipboardEvent:
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
	if len(router.squeue.receivers) != expected {
		t.Error("unexpect amount of receivers")
	}
	if router.squeue.ReadClipboard() != (expected > 0) {
		t.Error("missing requests")
	}
}

func assertSystemClipboardReadClipboardDuplicated(t *testing.T, router *Router, expected int) {
	t.Helper()
	if len(router.squeue.receivers) != expected {
		t.Error("receivers removed")
	}
	if router.squeue.ReadClipboard() != false {
		t.Error("duplicated requests")
	}
}

func assertSystemClipboardWaitingWrite(t *testing.T, router *Router, expected string) {
	t.Helper()
	if (router.squeue.text != nil) != (expected != "") {
		t.Error("text not defined")
	}
	text, ok := router.squeue.WriteClipboard()
	if ok != (expected != "") {
		t.Error("duplicated requests")
	}
	if text != expected {
		t.Error("text didn't match")
	}
}
