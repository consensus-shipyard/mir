// Package activemodule provides a implementation of the ActiveModule's event handling
package activemodule

import (
	"context"
	"fmt"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/events"
)

// EventProcessor can express the module-specific logic of an ActiveModule.
// When making use the package-provided default implementation of an ActiveModule (RunActiveModule),
// a function of this type must be passed as an argument to RunActiveModule.
// ActiveModuleEventProcessor takes the following arguments:
// - ctx: A Context the canceling of which will abort the processing of the logic and releases all associated resources.
//        In particular, if the processing spawned any goroutines, all those goroutines must terminate,
//        even if blocked on channel reads/writes or I/O.
// - eventsIn: A list of events to process. The default implementation will call this function repeatedly,
//             each time the Node submits new events to the ActiveModule for processing.
// - eventsOut: A channel to which events produced by the ActiveModule's implementation must be written.
//
// If an error occurs during event processing, the function ActiveModuleEventProcessor returns it.
// Otherwise, it returns nil.
//
// Each invocation of ActiveModuleEventProcessor must be non-blocking.
// Note that while it is expected that ActiveModuleEventProcessor causes writes to the eventsOut channel,
// there is no guarantee of the channel being also read from, potentially resulting in writes to block.
// Nevertheless, ActiveModuleEventProcessor must never block.
// This can be achieved, for example, by having the writes to eventsOut happen in a separate goroutine
// spawned by ActiveModuleEventProcessor.
//
// The default implementation never invokes an ActiveModuleEventProcessor concurrently.
type EventProcessor func(
	ctx context.Context,
	eventsIn *events.EventList,
	eventsOut chan<- *events.EventList,
) error

// RunActiveModule is the default implementation of the ActiveModule.Run method.
// Compared to ActiveModule.Run, it receives one more argument (eventProcessor) specifying,
// which is a function implementing the module-specific logic
// (see the documentation of the ActiveModuleEventProcessor type for details).
func RunActiveModule(
	ctx context.Context,
	eventsIn <-chan *events.EventList,
	eventsOut chan<- *events.EventList,
	interceptor eventlog.Interceptor,
	eventProcessor EventProcessor,
) error {
	for {
		var inputList *events.EventList

		// Read input.
		select {
		case inputList = <-eventsIn:
		case <-ctx.Done():
			return nil
		}

		// Process obtained events.
		if err := processEvents(ctx, inputList, eventsOut, interceptor, eventProcessor); err != nil {
			return err
		}
	}
}

// processEvents performs the default handling of events incoming to an ActiveModule.
// It strips the follow-up events, performs the processing as defined by the given eventProcessor,
// and writes follow-up events in the output channel.
// If the given interceptor is not nil, it passes the input events (stripped of follow-ups) to the interceptor as well.
func processEvents(
	ctx context.Context,
	inputList *events.EventList,
	eventsOut chan<- *events.EventList,
	interceptor eventlog.Interceptor,
	eventProcessor EventProcessor,
) error {

	// Remove follow-up Events from the input EventList,
	// in order to re-insert them in the processing loop after the input events have been processed.
	plainEvents, followUps := inputList.StripFollowUps()

	// Intercept the (stripped of all follow-ups) events that are about to be processed.
	// This is only for debugging / diagnostic purposes.
	if interceptor != nil {
		if err := interceptor.Intercept(plainEvents); err != nil {
			return fmt.Errorf("error intercepting events: %w", err)
		}
	}

	// Process events.
	err := eventProcessor(ctx, plainEvents, eventsOut)
	if err != nil {
		return fmt.Errorf("could not process events: %w", err)
	}

	// Return if there are no follow-up events
	// This is an optimization to prevent the processor loop from handling empty EventLists.
	if followUps.Len() == 0 {
		return nil
	}

	// Write follow-up events to the output channel.
	// This must happen in a separate goroutine in order to prevent the event processing loop from blocking.
	go func() {
		select {
		case eventsOut <- followUps:
		case <-ctx.Done():
		}
	}()

	return nil
}
