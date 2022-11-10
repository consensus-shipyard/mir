/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mir

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/pkg/errors"

	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/wal"
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

	// Write-ahead log.
	// If not nil, it will be used at initialization to load all events stored in the WAL.
	wal wal.WAL

	// Interceptor of processed events.
	// If not nil, every event is passed to the interceptor (by calling its Intercept method)
	// just before being processed.
	interceptor eventlog.Interceptor

	// A buffer for storing outstanding events that need to be processed by the node.
	// It contains a separate sub-buffer for each type of event.
	workItems eventBuffer

	// Channels for routing work items between modules.
	// Whenever workItems contains events, those events will be written (by the process() method)
	// to the corresponding channel in workChans. When processed by the corresponding module,
	// the result of that processing (also a list of events) will also be written
	// to the appropriate channel in workChans, from which the process() method moves them to the workItems buffer.
	workChans workChans

	// Used to synchronize the exit of the node's worker go routines.
	workErrNotifier *workErrNotifier

	// When true, importing events from ActiveModules is disabled.
	// This value is set to true when the internal event buffer exceeds a certain threshold
	// and reset to false when the buffer is drained under a certain threshold.
	inputPaused bool

	inputPausedCond *sync.Cond

	// If set to true, the node is in debug mode.
	// Only events received through the Step method are applied.
	// Events produced by the modules are, instead of being applied,
	debugMode bool

	// This channel is closed by the Run and Debug methods on returning.
	// Closing of the channel indicates that the node has stopped and the Stop method can also return.
	stopped chan struct{}
}

// NewNode creates a new node with ID id.
// The config parameter specifies Node-level (protocol-independent) configuration, like buffer sizes, logging, ...
// The modules parameter must contain initialized, ready-to-use modules that the new Node will use.
func NewNode(
	id t.NodeID,
	config *NodeConfig,
	m modules.Modules,
	wal wal.WAL,
	interceptor eventlog.Interceptor,
) (*Node, error) {

	// Check that a valid configuration has been provided.
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid node configuration")
	}

	// Make sure that the logger can be accessed concurrently (if there is any logger).
	if config.Logger == nil {
		config.Logger = logging.NilLogger
	} else {
		config.Logger = logging.Synchronize(config.Logger)
	}

	// Return a new Node.
	return &Node{
		ID:     id,
		Config: config,

		eventsIn: make(chan *events.EventList),
		debugOut: make(chan *events.EventList),

		workChans:   newWorkChans(m),
		modules:     m,
		wal:         wal,
		interceptor: interceptor,

		workItems:       newEventBuffer(m),
		workErrNotifier: newWorkErrNotifier(),

		inputPaused:     false,
		inputPausedCond: sync.NewCond(&sync.Mutex{}),

		stopped: make(chan struct{}),
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

	// When done, indicate to the Stop method that it can return.
	defer close(n.stopped)

	// Enable debug mode
	n.debugMode = true

	// If a WAL implementation is available,
	// load the contents of the WAL and enqueue it for processing.
	if n.wal != nil {
		if err := n.processWAL(ctx); err != nil {
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

	// When done, indicate to the Stop method that it can return.
	defer close(n.stopped)

	// If a WAL implementation is available,
	// load the contents of the WAL and enqueue it for processing.
	if n.wal != nil {
		if err := n.processWAL(ctx); err != nil {
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

// Stop stops the Node and returns only after the node has stopped, i.e., after a call to Run or Debug returns.
// If neither Run nor Debug has been called,
// Stop blocks until either Run or Debug is called by another goroutine and returns.
func (n *Node) Stop() {
	// Indicate to the event processing loop that it should stop.
	n.workErrNotifier.Fail(ErrStopped)

	// Wait until event processing stops.
	<-n.stopped
}

// Loads all events stored in the WAL and enqueues them in the node's processing queues.
func (n *Node) processWAL(ctx context.Context) error {

	var storedEvents *events.EventList
	var err error

	// Add all events from the WAL to the new EventList.
	if storedEvents, err = n.wal.LoadAll(ctx); err != nil {
		return fmt.Errorf("could not load WAL events: %w", err)
	}

	// Enqueue all events to the eventBuffer buffers.
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

	// Wait for all worker threads (started by n.StartModules()) to finish when processing is done.
	// Watch out! If process() terminates unexpectedly (e.g. by panicking), this might get stuck!
	defer wg.Wait()

	// Make sure that the event importing goroutines do not get stuck if input is paused due to full buffers.
	// This defer statement must go after waiting for the worker goroutines to finish (wg.Wait() above).
	// so it is executed before wg.Wait() (since defers are stacked). Otherwise, we get into a deadlock.
	defer n.resumeInput()

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
			// TODO: Use a different error here to distinguish this case from calling Node.Stop()
			n.workErrNotifier.Fail(ErrStopped)
		})

		// Add events produced by modules and debugger to the eventBuffer buffers and handle logical time.

		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(n.eventsIn),
		})
		selectReactions = append(selectReactions, func(newEventsVal reflect.Value) {
			newEvents := newEventsVal.Interface().(*events.EventList)
			if err := n.workItems.AddEvents(newEvents); err != nil {
				n.workErrNotifier.Fail(err)
			}

			// Keep track of the size of the input buffer.
			// When it exceeds the PauseInputThreshold, pause the input from active modules.
			if n.workItems.totalEvents > n.Config.PauseInputThreshold {
				n.pauseInput()
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

		// For each generic event buffer in eventBuffer that contains events to be submitted to its corresponding module,
		// create a selectCase for writing those events to the module's work channel.

		for moduleID, buffer := range n.workItems.buffers {
			if buffer.Len() > 0 {

				eventBatch := buffer.Head(n.Config.MaxEventBatchSize)
				numEvents := eventBatch.Len()

				// Create case for writing in the work channel.
				selectCases = append(selectCases, reflect.SelectCase{
					Dir:  reflect.SelectSend,
					Chan: reflect.ValueOf(n.workChans[moduleID]),
					Send: reflect.ValueOf(eventBatch),
				})

				// Create a copy of moduleID to use in the reaction function.
				// If we used moduleID directly in the function definition, it would correspond to the loop variable
				// and have the same value for all cases after the loop finishes iterating.
				var mID = moduleID

				// React to writing to a work channel by emptying the corresponding event buffer
				// (i.e., removing events just written to the channel from the buffer).
				selectReactions = append(selectReactions, func(_ reflect.Value) {
					n.workItems.buffers[mID].RemoveFront(numEvents)

					// Keep track of the size of the event buffer.
					// Whenever it drops below the ResumeInputThreshold, resume input.
					n.workItems.totalEvents -= numEvents
					if n.workItems.totalEvents <= n.Config.ResumeInputThreshold {
						n.resumeInput()
					}
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

			// Create a context that is passed to the module event application function
			// and canceled when module processing is stopped (i.e. this function returns).
			// This is to make sure that the event processing is properly aborted if needed.
			processingCtx, cancelProcessing := context.WithCancel(ctx)
			defer cancelProcessing()

			var continueProcessing = true
			var err error

			for continueProcessing {
				if n.debugMode {
					// In debug mode, all produced events are routed to the debug output.
					continueProcessing, err = n.processModuleEvents(processingCtx, m, workChan, n.debugOut)
				} else {
					// During normal operation, feed all produced events back into the event loop.
					continueProcessing, err = n.processModuleEvents(processingCtx, m, workChan, n.eventsIn)
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

		n.waitForInputEnabled()

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

func (n *Node) pauseInput() {
	n.inputPausedCond.L.Lock()
	n.inputPaused = true
	n.inputPausedCond.L.Unlock()
}

func (n *Node) resumeInput() {
	n.inputPausedCond.L.Lock()
	n.inputPaused = false
	n.inputPausedCond.Broadcast()
	n.inputPausedCond.L.Unlock()
}

func (n *Node) waitForInputEnabled() {
	n.inputPausedCond.L.Lock()
	for n.inputPaused {
		n.inputPausedCond.Wait()
	}
	n.inputPausedCond.L.Unlock()
}

func createInitEvents(m modules.Modules) *events.EventList {
	initEvents := events.EmptyList()
	for moduleID := range m {
		initEvents.PushBack(events.Init(moduleID))
	}
	return initEvents
}
