/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mir

import (
	"context"
	"fmt"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"reflect"
	"sync"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	t "github.com/filecoin-project/mir/pkg/types"
)

var ErrStopped = fmt.Errorf("stopped at caller request")

// Node is the local instance of Mir and the application's interface to the mir library.
type Node struct {
	ID     t.NodeID    // Protocol-level node ID
	Config *NodeConfig // Node-level (protocol-independent) configuration, like buffer sizes, logging, ...

	// Incoming events to be processed by the node.
	// E.g., all modules' output events are written in this channel,
	// from where the Node processor reads and redistributes the events to their respective workItems buffers.
	// External events are also funneled through this channel towards the workItems buffers.
	eventsIn chan *events.EventList

	// During debugging, Events that would normally be inserted in the workItems event buffer
	// (and thus inserted in the event loop) are written to this channel instead if it is not nil.
	// If this channel is nil, those Events are discarded.
	debugOut chan *events.EventList

	// Implementations of networking, hashing, request store, WAL, etc.
	// The state machine is also a module of the node.
	modules modules.Modules

	// Interceptor of processed events.
	// If not nil, every event is passed to the interceptor (by calling its Intercept method)
	// just before being processed.
	interceptor eventlog.Interceptor

	// A buffer for storing outstanding events that need to be processed by the node.
	// It contains a separate sub-buffer for each type of event.
	workItems workItems

	// Channels for routing work items between modules.
	// Whenever workItems contains events, those events will be written (by the process() method)
	// to the corresponding channel in workChans. When processed by the corresponding module,
	// the result of that processing (also a list of events) will also be written
	// to the appropriate channel in workChans, from which the process() method moves them to the workItems buffer.
	workChans workChans

	// Used to synchronize the exit of the node's worker go routines.
	workErrNotifier *workErrNotifier

	// If set to true, the node is in debug mode.
	// Only events received through the Step method are applied.
	// Events produced by the modules are, instead of being applied,
	debugMode bool
}

// NewNode creates a new node with ID id.
// The config parameter specifies Node-level (protocol-independent) configuration, like buffer sizes, logging, ...
// The modules parameter must contain initialized, ready-to-use modules that the new Node will use.
func NewNode(
	id t.NodeID,
	config *NodeConfig,
	m modules.Modules,
	interceptor eventlog.Interceptor,
) (*Node, error) {
	// Return a new Node.
	return &Node{
		ID:     id,
		Config: config,

		eventsIn: make(chan *events.EventList),
		debugOut: make(chan *events.EventList),

		workChans:   newWorkChans(m),
		modules:     m,
		interceptor: interceptor,

		workItems:       newWorkItems(m),
		workErrNotifier: newWorkErrNotifier(),
	}, nil
}

// Debug runs the Node in debug mode.
// If the node has been instantiated with a WAL, its contents will be loaded.
// Then, the Node will ony process events submitted through the Step method.
// All internally generated events will be ignored
// and, if the eventsOut argument is not nil, written to eventsOut instead.
// Note that if the caller supplies such a channel, the caller is expected to read from it.
// Otherwise, the Node's execution might block while writing to the channel.
func (n *Node) Debug(ctx context.Context, eventsOut chan *events.EventList) error {
	// Enable debug mode
	n.debugMode = true

	// If a WAL implementation is available,
	// load the contents of the WAL and enqueue it for processing.
	// TODO: The WAL module is assumed to be stored under the "wal" key here. Generalize!
	if n.modules["wal"] != nil {
		if err := n.processWAL(); err != nil {
			n.workErrNotifier.Fail(err)
			return fmt.Errorf("could not process WAL: %w", err)
		}
	}

	// Set up channel for outputting internal events
	n.debugOut = eventsOut

	// Start processing of events.
	return n.process(ctx)

}

// InjectEvents inserts a list of Events in the Node.
func (n *Node) InjectEvents(ctx context.Context, events *events.EventList) error {

	// Enqueue event in a work channel to be handled by the processing thread.
	select {
	case n.eventsIn <- events:
		return nil
	case <-n.workErrNotifier.ExitStatusC():
		return n.workErrNotifier.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Run starts the Node.
// First, it loads the contents of the WAL and enqueues all its contents for processing.
// This makes sure that the WAL events end up first in all the modules' processing queues.
// Then it adds an Init event to the work items, giving the modules the possibility
// to perform additional initialization based on the state recovered from the WAL.
// Run then launches the processing of incoming messages, and internal events.
// The node stops when the ctx is canceled.
// The function call is blocking and only returns when the node stops.
func (n *Node) Run(ctx context.Context) error {

	// If a WAL implementation is available,
	// load the contents of the WAL and enqueue it for processing.
	if n.modules["wal"] != nil {
		if err := n.processWAL(); err != nil {
			n.workErrNotifier.Fail(err)
			return fmt.Errorf("could not process WAL: %w", err)
		}
	}

	// Submit the Init event to the modules.
	if err := n.workItems.AddEvents(createInitEvents(n.modules)); err != nil {
		n.workErrNotifier.Fail(err)
		return fmt.Errorf("failed to add init event: %w", err)
	}

	// Start processing of events.
	return n.process(ctx)
}

// Loads all events stored in the WAL and enqueues them in the node's processing queues.
func (n *Node) processWAL() error {

	var storedEvents *events.EventList
	var err error

	// Add all events from the WAL to the new EventList.
	// TODO: The WAL module is assumed to be stored under the "wal" key here. Generalize!
	if storedEvents, err = n.modules["wal"].(modules.PassiveModule).ApplyEvents(
		events.ListOf(events.WALLoadAll("wal")),
	); err != nil {
		return fmt.Errorf("could not load WAL events: %w", err)
	}

	// Enqueue all events to the workItems buffers.
	if err = n.workItems.AddEvents(storedEvents); err != nil {
		return fmt.Errorf("could not enqueue WAL events for processing: %w", err)
	}

	// If we made it all the way here, no error occurred.
	return nil

}

// Performs all internal work of the node,
// which mostly consists of routing events between the node's modules.
// Stops and returns when ctx is canceled.
func (n *Node) process(ctx context.Context) error { //nolint:gocyclo

	var wg sync.WaitGroup // Synchronizes all the worker functions
	defer wg.Wait()       // Watch out! If process() terminates unexpectedly (e.g. by panicking), this might get stuck!

	// Start processing module events.
	n.startModules(ctx, &wg)

	// This loop shovels events between the appropriate channels, until a stopping condition is satisfied.
	var returnErr error
	for returnErr == nil {

		// Initialize slices of select cases and the corresponding reactions to each case being selected.
		selectCases := make([]reflect.SelectCase, 0)
		selectReactions := make([]func(receivedVal reflect.Value), 0)

		// If the context has been canceled, set the corresponding stopping value at the Node's WorkErrorNotifier,
		// making the processing stop when the WorkErrorNotifier's channel is selected the next time.
		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		})
		selectReactions = append(selectReactions, func(_ reflect.Value) {
			n.workErrNotifier.Fail(ErrStopped)
		})

		// Add events produced by modules and debugger to the workItems buffers and handle logical time.

		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(n.eventsIn),
		})
		selectReactions = append(selectReactions, func(newEventsVal reflect.Value) {
			newEvents := newEventsVal.Interface().(*events.EventList)
			if err := n.workItems.AddEvents(newEvents); err != nil {
				n.workErrNotifier.Fail(err)
			}
		})

		// If an error occurred, stop processing.

		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(n.workErrNotifier.ExitC()),
		})
		selectReactions = append(selectReactions, func(_ reflect.Value) {
			returnErr = n.workErrNotifier.Err()
		})

		// For each generic event buffer in workItems that contains events to be submitted to its corresponding module,
		// create a selectCase for writing those events to the module's work channel.

		for moduleID, buffer := range n.workItems {
			if buffer.Len() > 0 {

				// Create case for writing in the work channel.
				selectCases = append(selectCases, reflect.SelectCase{
					Dir:  reflect.SelectSend,
					Chan: reflect.ValueOf(n.workChans[moduleID]),
					Send: reflect.ValueOf(buffer),
				})

				// Create a copy of moduleID to use in the reaction function.
				// If we used moduleID directly in the function definition, it would correspond to the loop variable
				// and have the same value for all cases after the loop finishes iterating.
				var mID = moduleID

				// React to writing to a work channel by emptying the corresponding event buffer
				// (i.e., removing events just written to the channel from the buffer).
				selectReactions = append(selectReactions, func(_ reflect.Value) {
					n.workItems[mID] = events.EmptyList()
				})
			}
		}

		// Choose one case from above and execute the corresponding reaction.

		chosenCase, receivedValue, _ := reflect.Select(selectCases)
		selectReactions[chosenCase](receivedValue)

	}

	n.workErrNotifier.SetExitStatus(nil, nil) // TODO: change this when statuses are implemented.
	return returnErr
}

func (n *Node) startModules(ctx context.Context, wg *sync.WaitGroup) {

	// The modules mostly read events from their respective channels in n.workChans,
	// process them correspondingly, and write the results (also represented as events) in the appropriate channels.
	for moduleID, module := range n.modules {

		// For each module, we start a worker function reads a single work item (EventList) and processes it.
		wg.Add(1)
		go func(mID t.ModuleID, m modules.Module, workChan chan *events.EventList) {
			defer wg.Done()

			var continueProcessing = true
			var err error

			for continueProcessing {
				if n.debugMode {
					// In debug mode, all produced events are routed to the debug output.
					continueProcessing, err = n.processModuleEvents(ctx, m, workChan, n.debugOut)
				} else {
					// During normal operation, feed all produced events back into the event loop.
					continueProcessing, err = n.processModuleEvents(ctx, m, workChan, n.eventsIn)
				}
				if err != nil {
					n.workErrNotifier.Fail(fmt.Errorf("could not process PassiveModule (%v) events: %w", mID, err))
					return
				}
			}

		}(moduleID, module, n.workChans[moduleID])

		// Depending on the module type (and the way output events are communicated back to the node),
		// start a goroutine importing the modules' output events
		switch m := module.(type) {
		case modules.PassiveModule:
			// Nothing else to be done for a PassiveModule
		case modules.ActiveModule:
			// Start a goroutine to import the ActiveModule's output events to workItemInput.
			wg.Add(1)
			go func() {
				defer wg.Done()
				if n.debugMode {
					// In debug mode, all produced events are routed to the debug output.
					n.importEvents(ctx, m.EventsOut(), n.debugOut)
				} else {
					// During normal operation, feed all produced events back into the event loop.
					n.importEvents(ctx, m.EventsOut(), n.eventsIn)
				}
			}()
		default:
			n.workErrNotifier.Fail(fmt.Errorf("unknown module type: %T", m))
		}
	}
}

// importEvents reads events from eventSource and writes them to the eventSink until
// - eventSource is closed or
// - ctx is canceled or
// - an error occurred in the Node and was announced through the Node's workErrorNotifier.
func (n *Node) importEvents(
	ctx context.Context,
	eventSource <-chan *events.EventList,
	eventSink chan<- *events.EventList,
) {
	for {

		// First, try to read events from the input.
		select {
		case newEvents, ok := <-eventSource:

			// Return if input channel has been closed
			if !ok {
				return
			}

			// Skip writing events if there is no event sink.
			if eventSink == nil {
				continue
			}

			// If input events have been read, try to write them to the Node's central input channel.
			select {
			case eventSink <- newEvents:
			case <-ctx.Done():
				return
			case <-n.workErrNotifier.ExitC():
				return
			}

		case <-ctx.Done():
			return
		case <-n.workErrNotifier.ExitC():
			return
		}
	}
}

// If the interceptor module is present, passes events to it. Otherwise, does nothing.
// If an error occurs passing events to the interceptor, notifies the node by means of the workErrorNotifier.
// Note: The passed Events should be free of any follow-up Events,
// as those will be intercepted separately when processed.
// Make sure to call the Strip method of the EventList before passing it to interceptEvents.
func (n *Node) interceptEvents(events *events.EventList) {
	if n.interceptor != nil {
		if err := n.interceptor.Intercept(events); err != nil {
			n.workErrNotifier.Fail(err)
		}
	}
}

func createInitEvents(m modules.Modules) *events.EventList {
	initEvents := events.EmptyList()
	for moduleID := range m {
		initEvents.PushBack(events.Init(moduleID))
	}
	return initEvents
}
