// Package transfer contains operations and events for brokering data transfers.
//
// The transfer protocol is as follows:
//
//   - Data sources use [SourceFilter] to receive an [InitiateEvent] when a drag
//     is initiated, and an [RequestEvent] for each initiation of a data transfer.
//     Sources respond to requests with [OfferCmd].
//   - Data targets use [TargetFilter] to receive an [DataEvent] for receiving data.
//     The target must close the data event after use.
//
// When a user initiates a pointer-guided drag and drop transfer, the
// source as well as all potential targets receive an InitiateEvent.
// Potential targets are targets with at least one MIME type in common
// with the source. When a drag gesture completes, a CancelEvent is sent
// to the source and all potential targets.
//
// Note that the RequestEvent is sent to the source upon drop.
package transfer

import (
	"io"
	"net/url"

	"gioui.org/io/event"
)

// OfferCmd is used by data sources as a response to a RequestEvent.
type OfferCmd struct {
	Tag event.Tag
	// Type is the MIME type of Data.
	// It must be the Type from the corresponding RequestEvent.
	Type string
	// Data contains the offered data. It is closed when the
	// transfer is complete or cancelled.
	// Data must be kept valid until closed, and it may be used from
	// a goroutine separate from the one processing the frame.
	Data io.ReadCloser
}

func (OfferCmd) ImplementsCommand() {}

// SourceFilter filters for any [RequestEvent] that match a MIME type
// as well as [InitiateEvent] and [CancelEvent].
// Use multiple filters to offer multiple types.
type SourceFilter struct {
	// Target is a tag included in a previous event.Op.
	Target event.Tag
	// Type is the MIME type supported by this source.
	Type string
}

// TargetFilter filters for any [DataEvent] whose type matches a MIME type
// as well as [CancelEvent]. Use multiple filters to accept multiple types.
type TargetFilter struct {
	// Target is a tag included in a previous event.Op.
	Target event.Tag
	// Type is the MIME type accepted by this target.
	Type string
}

// URLFilter filters for any [URLEvent] which the scheme matches against
// the scheme from the event.
type URLFilter struct {
	// If it's blank/empty, it will handle any scheme.
	Scheme string
}

func (URLFilter) ImplementsFilter() {}

// RequestEvent requests data from a data source. The source must
// respond with an OfferCmd.
type RequestEvent struct {
	// Type is the first matched type between the source and the target.
	Type string
}

func (RequestEvent) ImplementsEvent() {}

// InitiateEvent is sent to a data source when a drag-and-drop
// transfer gesture is initiated.
//
// Potential data targets also receive the event.
type InitiateEvent struct{}

func (InitiateEvent) ImplementsEvent() {}

// CancelEvent is sent to data sources and targets to cancel the
// effects of an InitiateEvent.
type CancelEvent struct{}

func (CancelEvent) ImplementsEvent() {}

// DataEvent is sent to the target receiving the transfer.
type DataEvent struct {
	// Type is the MIME type of Data.
	Type string
	// Open returns the transfer data. It is only valid to call Open in the frame
	// the DataEvent is received. The caller must close the return value after use.
	Open func() io.ReadCloser
}

func (DataEvent) ImplementsEvent() {}

// URLEvent is generated when the app is invoked with a URL. This event is sent to
// all SchemeOps that match the URL scheme.
//
// In order to receive URLEvents, it's necessary to use SchemeOp and the app must
// register a scheme. The scheme can be registered using gogio, with `-schemes <scheme>`
// flag.
//
// Due to differences between OSes, URLEvent might not be generated when the URL is
// larger than 2048 characters. Domain names using non-ASCII characters have
// undefined behavior, it can be either encoded as Punycode, QueryEscape,
// or even as raw UTF-8.
type URLEvent struct {
	URL *url.URL
}

func (URLEvent) ImplementsEvent() {}

func (SourceFilter) ImplementsFilter() {}
func (TargetFilter) ImplementsFilter() {}
