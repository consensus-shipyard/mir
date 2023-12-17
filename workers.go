package mir

import (
	"context"
	"fmt"
	"runtime/debug"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
)

// workChans represents input channels for the modules within the Node.
// the Node.process() method writes events to these channels to route them between the Node's modules.
type workChans map[stdtypes.ModuleID]chan *stdtypes.EventList

// Allocate and return a new workChans structure.
func newWorkChans(modules modules.Modules) workChans {
	wc := make(map[stdtypes.ModuleID]chan *stdtypes.EventList)

	for moduleID := range modules {
		wc[moduleID] = make(chan *stdtypes.EventList)
	}

	return wc
}

// processModuleEvents reads a single list of input Events from a work channel,
// strips off all associated follow-up Events,
// and processes the bare content of the list using the passed PassiveModule.
// processModuleEvents writes all the stripped off follow-up events along with any Events generated by the processing
// to the eventSink channel if it is not nil.
//
// If the Node is configured to use an Interceptor, after having removed all follow-up Events,
// processModuleEvents passes the list of input Events to the Interceptor.
//
// If the first return value is false,
// processing should be terminated and processModuleEvents should not be called again.
// The first return value being true indicates that processing can continue
// and processModuleEvents should be called again.
//
// If any error occurs, it is returned as the second parameter.
// If context is canceled, processModuleEvents might return a nil error with or without performing event processing.
func (n *Node) processModuleEvents(
	ctx context.Context,
	module modules.Module,
	eventSource <-chan *stdtypes.EventList,
	eventSink chan<- *stdtypes.EventList,
	sw *Stopwatch,
) (bool, error) {
	var eventsIn *stdtypes.EventList
	var inputOpen bool

	// Read input.
	select {
	case eventsIn, inputOpen = <-eventSource:
		if !inputOpen {
			return false, nil
		}
	case <-ctx.Done():
		return false, nil
	case <-n.workErrNotifier.ExitC():
		return false, nil
	}

	eventsOut := stdtypes.EmptyList()

	sw.Start()

	// Intercept the (stripped of all follow-ups) events that are about to be processed.
	// This is only for debugging / diagnostic purposes.
	n.interceptEvents(eventsIn)

	// In Trace mode, log all events.
	if n.Config.Logger.MinLevel() <= logging.LevelTrace {
		iter := eventsIn.Iterator()
		for event := iter.Next(); event != nil; event = iter.Next() {
			n.Config.Logger.Log(logging.LevelTrace,
				fmt.Sprintf("Event for module %v: %v", event.Dest(), event.ToString()))
		}
	}

	// Process events.
	switch m := module.(type) {

	case modules.PassiveModule:
		// For a passive module, synchronously apply all events and
		// add potential resulting events to the output EventList.

		var newEvents *stdtypes.EventList
		var err error
		if newEvents, err = safelyApplyEventsPassive(m, eventsIn); err != nil {
			return false, err
		}

		// Add newly generated Events to the output.
		eventsOut.PushBackList(newEvents)

	case modules.ActiveModule:
		// For an active module, only submit the events to the module and let it output the result asynchronously.

		if err := safelyApplyEventsActive(ctx, m, eventsIn); err != nil {
			return false, err
		}

	default:
		return false, es.Errorf("unknown module type: %T", m)
	}

	sw.Stop()

	// Return if no output was generated.
	// This is only an optimization to prevent the processor loop from handling empty EventLists.
	if eventsOut.Len() == 0 {
		return true, nil
	}

	// Skip writing output if there is no channel to write it to.
	if eventSink == nil {
		return true, nil
	}

	// Write output.
	select {
	case eventSink <- eventsOut:
		return true, nil
	case <-ctx.Done():
		return false, nil
	case <-n.workErrNotifier.ExitC():
		return false, nil
	}
}

func safelyApplyEventsPassive(
	module modules.PassiveModule,
	events *stdtypes.EventList,
) (result *stdtypes.EventList, err error) {
	defer func() {
		if r := recover(); r != nil {
			if rErr, ok := r.(error); ok {
				err = es.Errorf("module panicked: %w\nStack trace:\n%s", rErr, string(debug.Stack()))
			} else {
				err = es.Errorf("module panicked: %v\nStack trace:\n%s", r, string(debug.Stack()))
			}
		}
	}()

	return module.ApplyEvents(events)
}

func safelyApplyEventsActive(ctx context.Context, module modules.ActiveModule, events *stdtypes.EventList) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if rErr, ok := r.(error); ok {
				err = es.Errorf("module panicked: %w\nStack trace:\n%s", rErr, string(debug.Stack()))
			} else {
				err = es.Errorf("module panicked: %v\nStack trace:\n%s", r, string(debug.Stack()))
			}
		}
	}()

	return module.ApplyEvents(ctx, events)
}
