package modules

import (
	"context"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/statuspb"
)

type ActiveModule interface {
	Module

	// Run performs the module's processing.
	// It usually reads events from eventsIn, processes them, and writes new events to eventsOut.
	// Note, however, that the implementation may produce output events even without receiving any input.
	// Before Run is called, ActiveModule guarantees not to read or write,
	// respectively, from eventsIn and to eventsOut.
	//
	// In general, Run is expected to block until explicitly asked to stop
	// (through a module-specific event or by canceling ctx)
	// and thus should most likely be run in its own goroutine.
	// Run must return after ctx is canceled.
	//
	// Note that the node does not guarantee to always read events from eventsOut.
	// The node might decide at any moment to stop reading from eventsOut for an arbitrary amount of time
	// (e.g. if the Node's internal event buffers become full and the Node needs to wait until they free up).
	//
	// The implementation of Run is required to always read from eventsIn.
	// I.e., when an event is written to eventsIn, Run must eventually read it, regardless of the module's state
	// or any external factors
	// (e.g., the processing of events must not depend on some other goroutine reading from eventsOut).
	Run(
		ctx context.Context,
		eventsIn <-chan *events.EventList,
		eventsOut chan<- *events.EventList,
		interceptor eventlog.Interceptor,
	) error

	// Status returns the current state of the module.
	// If Run has been called and has not returned when calling Status,
	// there is no guarantee about which input events are taken into account
	// when creating the snapshot of the returned state.
	// Mostly for debugging purposes.
	Status() (s *statuspb.ProtocolStatus, err error)
}
