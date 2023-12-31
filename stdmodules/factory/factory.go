package factory

import (
	"fmt"

	"github.com/filecoin-project/mir/stdevents"
	es "github.com/go-errors/errors"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/messagebuffer"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/transportpb"
)

// TODO: Add support for active modules as well.

// module provides the basic functionality of "submodules".
// It can be used to dynamically create and garbage-collect passive modules,
// and it automatically forwards events to them.
//
// The forwarding mechanism is as follows:
//  1. All events destined for an existing submodule are forwarded to it automatically regardless of the event type.
//  2. Incoming network messages destined for non-existent submodules are buffered within a limit.
//     Once the limit is exceeded, the oldest messages are discarded.
//     If a single message is too large to fit into the buffer, it is discarded.
//  3. Other events destined for non-existent submodules are ignored.
type module struct {
	ownID     stdtypes.ModuleID
	generator ModuleGenerator

	submodules      map[stdtypes.ModuleID]modules.PassiveModule
	moduleRetention map[stdtypes.RetentionIndex][]stdtypes.ModuleID
	retIdx          stdtypes.RetentionIndex
	messageBuffer   *messagebuffer.MessageBuffer // TODO: Split by NodeID (using NewBuffers). Future configurations...?

	eventBuffer map[stdtypes.ModuleID]*stdtypes.EventList
	logger      logging.Logger
}

// New creates a new factory module.
func New(id stdtypes.ModuleID, params ModuleParams, logger logging.Logger) modules.PassiveModule {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}

	return &module{
		ownID:     id,
		generator: params.Generator,

		submodules:      make(map[stdtypes.ModuleID]modules.PassiveModule),
		moduleRetention: make(map[stdtypes.RetentionIndex][]stdtypes.ModuleID),
		retIdx:          0,
		messageBuffer:   messagebuffer.New(params.MsgBufSize, logging.Decorate(logger, "MsgBuf: ", "factory", fmt.Sprintf("%v", id))),

		eventBuffer: make(map[stdtypes.ModuleID]*stdtypes.EventList),
		logger:      logger,
	}
}

func (fm *module) ImplementsModule() {}

func (fm *module) ApplyEvents(evts *stdtypes.EventList) (*stdtypes.EventList, error) {
	// TODO: Perform event processing in parallel (applyEvent will need to be made thread-safe).
	//       The idea is to have one internal thread per submodule, distribute the events to them through channels,
	//       and wait until all are processed.

	eventsOut, err := modules.ApplyEventsSequentially(evts, fm.applyEvent)
	if err != nil {
		return nil, err
	}
	submoduleEventsOut, err := fm.applySubmodulesEvents()
	if err != nil {
		return nil, err
	}
	return eventsOut.PushBackList(submoduleEventsOut), nil
}

func (fm *module) applyEvent(event stdtypes.Event) (*stdtypes.EventList, error) {
	if event.Dest() != fm.ownID {
		// If the module ID of the event does not fully match the factory module ID, the event is for a submodule.
		// Submodule events are not applied directly, but buffered for later concurrent execution.
		// Note that this is different from (and orthogonal to) buffering early messages for non-existent submodules.
		fm.bufferSubmoduleEvent(event)
		return stdtypes.EmptyList(), nil
	}

	// Before applying an event for the factory itself, process all the buffered submodule events
	// (as the factory event might change the submodules themselves).
	eventsOut, err := fm.applySubmodulesEvents()
	if err != nil {
		return nil, err
	}

	var result *stdtypes.EventList
	switch e := event.(type) {
	case *stdevents.Init:
		return stdtypes.EmptyList(), nil // Nothing to do at initialization.
	case *stdevents.NewSubmodule:
		if result, err = fm.NewSubmodule(e.SubmoduleID, e.Params, e.RetentionIndex); err != nil {
			return nil, err
		}
		eventsOut.PushBackList(result)
	case *stdevents.GarbageCollect:
		if result, err = fm.garbageCollect(e.RetentionIndex); err != nil {
			return nil, err
		}
		eventsOut.PushBackList(result)
	default:
		return nil, es.Errorf("unexpected event type: %T", event)
	}

	return eventsOut, nil
}

// bufferSubmoduleEvent buffers event in a map where the keys are the moduleID and the values are lists of events.
func (fm *module) bufferSubmoduleEvent(event stdtypes.Event) {
	smID := event.Dest()
	if _, ok := fm.eventBuffer[smID]; !ok {
		fm.eventBuffer[smID] = stdtypes.EmptyList()
	}

	fm.eventBuffer[smID] = fm.eventBuffer[smID].PushBack(event)
}

// applySubmodulesEvents applies all buffered events to the existing submodules,
// returns the first encountered error if any,
// or the full list of outgoing events after applying all the events to each of the submodules
func (fm *module) applySubmodulesEvents() (*stdtypes.EventList, error) {
	eventsOut := stdtypes.EmptyList()
	errChan := make(chan error)
	evtsChan := make(chan *stdtypes.EventList)

	// Apply submodule events concurrently to their respective submodules.
	existingSubmodules := 0
	for smID, eventList := range fm.eventBuffer {
		if submodule, ok := fm.submodules[smID]; !ok {
			// If the target submodule does not exist (yet), buffer its incoming messages.
			fm.bufferEarlyMsgs(eventList)
		} else {
			// Otherwise, call the submodule's ApplyEvents method in the background.
			existingSubmodules++
			go func(submodule modules.PassiveModule, eventList *stdtypes.EventList) {
				evtsOut, err := submodule.ApplyEvents(eventList)
				errChan <- err
				evtsChan <- evtsOut
			}(submodule, eventList)
		}
	}

	// Wait for all goroutines to complete and collect errors and events.
	for i := 0; i < existingSubmodules; i++ {
		err := <-errChan
		if err != nil {
			return nil, err
		}
		eventsOut.PushBackList(<-evtsChan)
	}

	fm.eventBuffer = make(map[stdtypes.ModuleID]*stdtypes.EventList)
	return eventsOut, nil
}

// bufferEarlyMsgs buffers message events for later application.
// It is used when receiving early messages for submodules that do not exist yet.
func (fm *module) bufferEarlyMsgs(eventList *stdtypes.EventList) {
	iter := eventList.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {
		fm.tryBuffering(event)
	}
}

func (fm *module) tryBuffering(event stdtypes.Event) {

	// We only support proto events.
	pbevent, ok := event.(*eventpb.Event)
	if !ok {
		fm.logger.Log(logging.LevelWarn,
			fmt.Sprintf("Not buffering submodule event (type %T). Only proto events supported", event),
			"moduleID", event.Dest(), "src", event.Src())
		return
	}

	// Check if this is a MessageReceived event.
	isMessageReceivedEvent := false
	var msg *transportpb.Event_MessageReceived
	e, isTransportEvent := pbevent.Type.(*eventpb.Event_Transport)
	if isTransportEvent {
		msg, isMessageReceivedEvent = e.Transport.Type.(*transportpb.Event_MessageReceived)
	}

	if !isMessageReceivedEvent {
		// Events other than MessageReceived are ignored.
		fm.logger.Log(logging.LevelDebug, "Ignoring submodule event. Destination module not found.",
			"moduleID", stdtypes.ModuleID(pbevent.DestModule),
			"eventType", fmt.Sprintf("%T", pbevent.Type),
			"eventValue", fmt.Sprintf("%v", pbevent.Type))
		// TODO: Get rid of Sprintf of the value and just use the value directly. Using Sprintf is just a work-around
		//       for a sloppy implementation of the testing log used in tests that cannot handle pointers yet.
		return
	}

	if !fm.messageBuffer.Store(pbevent) {
		fm.logger.Log(logging.LevelWarn, "Failed buffering incoming submodule message.",
			"moduleID", stdtypes.ModuleID(pbevent.DestModule), "msgType", fmt.Sprintf("%T", msg.MessageReceived.Msg.Type),
			"from", msg.MessageReceived.From)
	}
}

func (fm *module) NewSubmodule(
	id stdtypes.ModuleID,
	params any,
	retIdx stdtypes.RetentionIndex,
) (*stdtypes.EventList, error) {

	// The new module's ID must have the factory's ID as a prefix.
	if id.Top() != fm.ownID {
		return nil, es.Errorf("submodule (%v) must have the factory's ID (%v) as a prefix", id, fm.ownID)
	}

	// Skip creation of submodules that should have been already garbage-collected.
	if retIdx < fm.retIdx {
		fm.logger.Log(logging.LevelWarn, "Ignoring new module instantiation with low retention index.",
			"moduleID", id, "currentRetIdx", fm.retIdx, "moduleRetIdx", retIdx)
		return stdtypes.EmptyList(), nil
	}

	// Create new instance of the submodule.
	if submodule, err := fm.generator(id, params); err == nil {
		fm.submodules[id] = submodule
	} else {
		return nil, err
	}

	// Assign the newly created submodule to its retention index.
	fm.moduleRetention[retIdx] = append(fm.moduleRetention[retIdx], id)

	// Initialize new submodule.
	eventsOut, err := fm.submodules[id].ApplyEvents(stdtypes.ListOf(stdevents.NewInit(id)))
	if err != nil {
		return nil, err
	}

	// Get messages for the new submodule that arrived early and have been buffered.
	bufferedMessages := stdtypes.EmptyList()
	fm.messageBuffer.Iterate(func(msg proto.Message) messagebuffer.Applicable {
		if stdtypes.ModuleID(msg.(*eventpb.Event).DestModule) == id {
			return messagebuffer.Current
		}
		return messagebuffer.Future
	}, func(msg proto.Message) {
		bufferedMessages.PushBack(msg.(*eventpb.Event))
	})

	// Apply buffered messages
	results, err := fm.submodules[id].ApplyEvents(bufferedMessages)
	if err != nil {
		return nil, err
	}
	eventsOut.PushBackList(results)

	// Return all output events.
	return eventsOut, nil
}

func (fm *module) garbageCollect(retIdx stdtypes.RetentionIndex) (*stdtypes.EventList, error) {
	// While the new retention index is larger than the current one
	for retIdx > fm.retIdx {

		// Delete all modules associated with the current retention index.
		for _, mID := range fm.moduleRetention[fm.retIdx] {
			// TODO: Apply a "shutdown" notification event to each garbage-collected module
			//       to give it a chance to clean up.
			delete(fm.submodules, mID)
		}

		// TODO: Allow parametrization of the factory with a custom function that could also garbage-collect
		//   message buffers. In most cases, the destination module of messages also encodes the an epoch number
		//   that is used as a retention index, and thus all messages destined to modules below the retention index
		//   can be garbage-collected. This is, however, not necessarily the case from the perspective of the factory.
		//   But if it is, garbage collection should be easy.
		//   This is not critical, as it makes no difference functionality-wise (the buffers are FIFO anyway),
		//   it just reduces the memory footprint.

		// Increase current retention index.
		delete(fm.moduleRetention, fm.retIdx)
		fm.retIdx++
	}

	return stdtypes.EmptyList(), nil
}
