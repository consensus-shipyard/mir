package mir

import (
	"context"
	"fmt"
	"github.com/filecoin-project/mir/pkg/modules"
	"runtime/debug"

	"github.com/filecoin-project/mir/pkg/events"
	t "github.com/filecoin-project/mir/pkg/types"
)

// workChans represents input channels for the modules within the Node.
// the Node.process() method writes events to these channels to route them between the Node's modules.
type workChans map[t.ModuleID]chan *events.EventList

// Allocate and return a new workChans structure.
func newWorkChans(modules modules.Modules) workChans {
	wc := make(map[t.ModuleID]chan *events.EventList)

	for moduleID := range modules {
		wc[moduleID] = make(chan *events.EventList)
	}

	return wc
}

// processEventsPassive reads a single list of input Events from a work channel,
// strips off all associated follow-up Events,
// and processes the bare content of the list using the passed PassiveModule.
// processEventsPassive writes all the stripped off follow-up events along with any Events generated by the processing
// to the workItemInput channel, so they will be added to the workItems buffer for further processing.
//
// If the Node is configured to use an Interceptor, after having removed all follow-up Events,
// processEventsPassive passes the list of input Events to the Interceptor.
//
// If any error occurs, it is returned as the first parameter.
// If context is canceled, processEventsPassive might return a nil error with or without performing event processing.
// The second return value being true indicates that processing can continue
// and processEventsPassive should be called again.
// If the second return is false, processing should be terminated and processEventsPassive should not be called again.
func (n *Node) processModuleEvents(
	ctx context.Context,
	module modules.Module,
	eventSource <-chan *events.EventList,
) (error, bool) {
	var eventsIn *events.EventList
	var inputOpen bool

	// Read input.
	select {
	case eventsIn, inputOpen = <-eventSource:
		if !inputOpen {
			return nil, false
		}
	case <-ctx.Done():
		return nil, false
	case <-n.workErrNotifier.ExitC():
		return nil, false
	}

	// Remove follow-up Events from the input EventList,
	// in order to re-insert them in the processing loop after the input events have been processed.
	plainEvents, followUps := eventsIn.StripFollowUps()
	eventsOut := followUps // Follow-up events go directly to the output after the plainEvents are processed.

	// Intercept the (stripped of all follow-ups) events that are about to be processed.
	// This is only for debugging / diagnostic purposes.
	n.interceptEvents(plainEvents)

	// Process events.
	switch m := module.(type) {

	case modules.PassiveModule:
		// For a passive module, synchronously apply all events and
		// add potential resulting events to the output EventList.

		if newEvents, err := safelyApplyEvents(m, plainEvents); err != nil {
			return err, false
		} else {
			// Add newly generated Events to the output.
			eventsOut.PushBackList(newEvents)
		}

	case modules.ActiveModule:
		// For an active module, only submit the events to the module and let it output the result asynchronously.
		// Note that, unlike with a PassiveModule, an ActiveModule's ApplyEvents method is not invoked "safely",
		// i.e., a potential panic is not caught.
		// This is because an ActiveModule is expected to run its own goroutines.

		if err := m.ApplyEvents(ctx, plainEvents); err != nil {
			return err, false
		}

	default:
		return fmt.Errorf("unknown module type: %T", m), false
	}

	// Return if no output was generated.
	// This is only an optimization to prevent the processor loop from handling empty EventLists.
	if eventsOut.Len() == 0 {
		return nil, true
	}

	// Write output.
	select {
	case n.eventsIn <- eventsOut:
	case <-ctx.Done():
		return nil, false
	case <-n.workErrNotifier.ExitC():
		return nil, false
	}

	return nil, true
}

func safelyApplyEvents(
	module modules.PassiveModule,
	events *events.EventList,
) (result *events.EventList, err error) {
	defer func() {
		if r := recover(); r != nil {
			if rErr, ok := r.(error); ok {
				err = fmt.Errorf("module panicked: %w\nStack trace:\n%s", rErr, string(debug.Stack()))
			} else {
				err = fmt.Errorf("module panicked: %v\nStack trace:\n%s", r, string(debug.Stack()))
			}
		}
	}()

	return module.ApplyEvents(events)
}
