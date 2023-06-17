package factorymodule

import (
	"fmt"

	es "github.com/go-errors/errors"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/messagebuffer"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	factorypbtypes "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	"github.com/filecoin-project/mir/pkg/pb/transportpb"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// TODO: Add support for active modules as well.

// FactoryModule provides the basic functionality of "submodules".
// It can be used to dynamically create and garbage-collect passive modules,
// and it automatically forwards events to them.
// See: protos/factorypb/factorypb.proto for details on the interface of the factory module itself.
//
// The forwarding mechanism is as follows:
//  1. All events destined for an existing submodule are forwarded to it automatically regardless of the event type.
//  2. Incoming network messages destined for non-existent submodules are buffered within a limit.
//     Once the limit is exceeded, the oldest messages are discarded.
//     If a single message is too large to fit into the buffer, it is discarded.
//  3. Other events destined for non-existent submodules are ignored.
type FactoryModule struct {
	ownID     t.ModuleID
	generator ModuleGenerator

	submodules      map[t.ModuleID]modules.PassiveModule
	moduleRetention map[tt.RetentionIndex][]t.ModuleID
	retIdx          tt.RetentionIndex
	messageBuffer   *messagebuffer.MessageBuffer // TODO: Split by NodeID (using NewBuffers). Future configurations...?

	eventBuffer map[t.ModuleID]*events.EventList
	logger      logging.Logger
}

// New creates a new factory module.
func New(id t.ModuleID, params ModuleParams, logger logging.Logger) *FactoryModule {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}

	// Zero value of the t.NodeID type.
	// This is used as a dummy value, as for now the node ID is ignored by the message buffer.
	// TODO: This is hacky, fix when using separate buffers for different nodes.
	var zeroID t.NodeID

	return &FactoryModule{
		ownID:     id,
		generator: params.Generator,

		submodules:      make(map[t.ModuleID]modules.PassiveModule),
		moduleRetention: make(map[tt.RetentionIndex][]t.ModuleID),
		retIdx:          0,
		messageBuffer:   messagebuffer.New(zeroID, params.MsgBufSize, logging.Decorate(logger, "MsgBuf: ")),

		eventBuffer: make(map[t.ModuleID]*events.EventList),
		logger:      logger,
	}
}

func (fm *FactoryModule) ImplementsModule() {}

func (fm *FactoryModule) ApplyEvents(evts *events.EventList) (*events.EventList, error) {
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

func (fm *FactoryModule) applyEvent(event *eventpb.Event) (*events.EventList, error) {
	if t.ModuleID(event.DestModule) == fm.ownID {
		switch e := event.Type.(type) {
		case *eventpb.Event_Init:
			return events.EmptyList(), nil // Nothing to do at initialization.
		case *eventpb.Event_Factory:

			// Before applying an event for the factory itself, process all the buffered submodule events
			// (as the factory event might change the submodules themselves).
			submoduleOutputEvts, err := fm.applySubmodulesEvents()
			if err != nil {
				return nil, err
			}

			// Apply the factory event itself, appending its output to the result of submodule event processing.
			switch e := factorypbtypes.EventFromPb(e.Factory).Type.(type) {
			case *factorypbtypes.Event_NewModule:
				evOut, err := fm.applyNewModule(e.NewModule)
				if err != nil {
					return nil, err
				}
				return submoduleOutputEvts.PushBackList(evOut), nil
			case *factorypbtypes.Event_GarbageCollect:
				evOut, err := fm.applyGarbageCollect(e.GarbageCollect)
				if err != nil {
					return nil, err
				}
				return submoduleOutputEvts.PushBackList(evOut), nil
			default:
				return nil, es.Errorf("unsupported factory event subtype: %T", e)
			}
		default:
			return nil, es.Errorf("unsupported event type for factory module: %T", e)
		}
	}

	// Submodule events are not applied directly, but buffered for later concurrent execution.
	// Note that this is different from (and orthogonal to) buffering early messages for non-existent submodules.
	fm.bufferSubmoduleEvent(event)

	return events.EmptyList(), nil
}

// bufferSubmoduleEvent buffers event in a map where the keys are the moduleID and the values are lists of events.
func (fm *FactoryModule) bufferSubmoduleEvent(event *eventpb.Event) {
	smID := t.ModuleID(event.DestModule)
	if _, ok := fm.eventBuffer[smID]; !ok {
		fm.eventBuffer[smID] = events.EmptyList()
	}

	fm.eventBuffer[smID] = fm.eventBuffer[smID].PushBack(event)
}

// applySubmodulesEvents applies all buffered events to the existing submodules,
// returns the first encountered error if any,
// or the full list of outgoing events after applying all the events to each of the submodules
func (fm *FactoryModule) applySubmodulesEvents() (*events.EventList, error) {
	eventsOut := events.EmptyList()
	errChan := make(chan error)
	evtsChan := make(chan *events.EventList)

	// Apply submodule events concurrently to their respective submodules.
	existingSubmodules := 0
	for smID, eventList := range fm.eventBuffer {
		if submodule, ok := fm.submodules[smID]; !ok {
			// If the target submodule does not exist (yet), buffer its incoming messages.
			fm.bufferEarlyMsgs(eventList)
		} else {
			// Otherwise, call the submodule's ApplyEvents method in the background.
			existingSubmodules++
			go func(submodule modules.PassiveModule, eventList *events.EventList) {
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

	fm.eventBuffer = make(map[t.ModuleID]*events.EventList)
	return eventsOut, nil
}

// bufferEarlyMsgs buffers message events for later application.
// It is used when receiving early messages for submodules that do not exist yet.
func (fm *FactoryModule) bufferEarlyMsgs(eventList *events.EventList) {
	iter := eventList.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {
		fm.tryBuffering(event)
	}
}

func (fm *FactoryModule) tryBuffering(event *eventpb.Event) {

	// Check if this is a MessageReceived event.
	isMessageReceivedEvent := false
	var msg *transportpb.Event_MessageReceived
	e, isTransportEvent := event.Type.(*eventpb.Event_Transport)
	if isTransportEvent {
		msg, isMessageReceivedEvent = e.Transport.Type.(*transportpb.Event_MessageReceived)
	}

	if !isMessageReceivedEvent {
		// Events other than MessageReceived are ignored.
		fm.logger.Log(logging.LevelDebug, "Ignoring submodule event. Destination module not found.",
			"moduleID", t.ModuleID(event.DestModule),
			"eventType", fmt.Sprintf("%T", event.Type),
			"eventValue", fmt.Sprintf("%v", event.Type))
		// TODO: Get rid of Sprintf of the value and just use the value directly. Using Sprintf is just a work-around
		//       for a sloppy implementation of the testing log used in tests that cannot handle pointers yet.
		return
	}

	if !fm.messageBuffer.Store(event) {
		fm.logger.Log(logging.LevelWarn, "Failed buffering incoming submodule message.",
			"moduleID", t.ModuleID(event.DestModule), "msgType", fmt.Sprintf("%T", msg.MessageReceived.Msg.Type),
			"from", msg.MessageReceived.From)
	}
}

func (fm *FactoryModule) applyNewModule(newModule *factorypbtypes.NewModule) (*events.EventList, error) {

	// Convenience variables
	id := newModule.ModuleId
	retIdx := newModule.RetentionIndex

	// The new module's ID must have the factory's ID as a prefix.
	if id.Top() != fm.ownID {
		return nil, es.Errorf("submodule (%v) must have the factory's ID (%v) as a prefix", id, fm.ownID)
	}

	// Skip creation of submodules that should have been already garbage-collected.
	if retIdx < fm.retIdx {
		fm.logger.Log(logging.LevelWarn, "Ignoring new module instantiation with low retention index.",
			"moduleID", id, "currentRetIdx", fm.retIdx, "moduleRetIdx", retIdx)
		return events.EmptyList(), nil
	}

	// Create new instance of the submodule.
	if submodule, err := fm.generator(id, newModule.Params); err == nil {
		fm.submodules[id] = submodule
	} else {
		return nil, err
	}

	// Assign the newly created submodule to its retention index.
	fm.moduleRetention[retIdx] = append(fm.moduleRetention[retIdx], id)

	// Initialize new submodule.
	eventsOut, err := fm.submodules[id].ApplyEvents(events.ListOf(events.Init(id)))
	if err != nil {
		return nil, err
	}

	// Get messages for the new submodule that arrived early and have been buffered.
	bufferedMessages := events.EmptyList()
	fm.messageBuffer.Iterate(func(_ t.NodeID, msg proto.Message) messagebuffer.Applicable {
		if t.ModuleID(msg.(*eventpb.Event).DestModule) == id {
			return messagebuffer.Current
		}
		return messagebuffer.Future
	}, func(_ t.NodeID, msg proto.Message) {
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

func (fm *FactoryModule) applyGarbageCollect(gc *factorypbtypes.GarbageCollect) (*events.EventList, error) {
	// While the new retention index is larger than the current one
	for gc.RetentionIndex > fm.retIdx {

		// Delete all modules associated with the current retention index.
		for _, mID := range fm.moduleRetention[fm.retIdx] {
			// TODO: Apply a "shutdown" notification event to each garbage-collected module
			//       to give it a chance to clean up.
			delete(fm.submodules, mID)
		}

		// Increase current retention index.
		delete(fm.moduleRetention, fm.retIdx)
		fm.retIdx++
	}

	return events.EmptyList(), nil
}
