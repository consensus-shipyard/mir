package modules

import (
	"context"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/statuspb"
)

type ActiveModule interface {
	Module

	// ApplyEvents applies a list of input events to the module, making it advance its state
	// and potentially write a list of events output events to the ActiveModule's output channel.
	//
	// ApplyEvents takes the following arguments:
	// - ctx: A Context the canceling of which will abort the processing of the module's logic
	//        and releases all associated resources.
	//        In particular, if the processing spawned any goroutines, all those goroutines must terminate,
	//        even if blocked on channel reads/writes or I/O.
	// - events: A list of events to process. The Node will call this function repeatedly,
	//           each time it submits new events to the ActiveModule for processing.
	//
	// If an error occurs during event processing, ApplyEvents returns it. Otherwise, it returns nil.
	//
	// Each invocation of ApplyEvents must be non-blocking.
	// Note that while it is expected that ApplyEvents causes writes to the output channel,
	// there is no guarantee of the channel being also read from, potentially resulting in writes to block.
	// Nevertheless, ApplyEvents must never block.
	// This can be achieved, for example, by having the writes to the output channel happen in a separate goroutine
	// spawned by ApplyEvents.
	//
	// The Node never invokes an ApplyEvents concurrently.
	ApplyEvents(ctx context.Context, events *events.EventList) error

	// EventsOut returns a channel to which output events produced by the ActiveModule's implementation will be written.
	//
	// Note that the implementation may produce output events even without receiving any input.
	//
	// Note also that the Node does not guarantee to always read events from the channel returned by EventsOut.
	// The node might decide at any moment to stop reading from eventsOut for an arbitrary amount of time
	// (e.g. if the Node's internal event buffers become full and the Node needs to wait until they free up).
	// Even then, calls to ApplyEvents must be non-blocking.
	EventsOut() <-chan *events.EventList

	// Status returns the current state of the module.
	// If Run has been called and has not returned when calling Status,
	// there is no guarantee about which input events are taken into account
	// when creating the snapshot of the returned state.
	// Mostly for debugging purposes.
	Status() (s *statuspb.ProtocolStatus, err error)
}
