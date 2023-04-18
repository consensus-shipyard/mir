package mir

import (
	"github.com/pkg/errors"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/transportpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// eventBuffer is a buffer for storing outstanding events that need to be processed by the node.
// It contains a separate list for each module.
type eventBuffer struct {
	buffers     map[t.ModuleID]*events.EventList
	totalEvents int
}

// newEventBuffer allocates and returns a pointer to a new eventBuffer object.
func newEventBuffer(modules modules.Modules) eventBuffer {

	wi := eventBuffer{
		buffers:     make(map[t.ModuleID]*events.EventList),
		totalEvents: 0,
	}

	for moduleID := range modules {
		wi.buffers[moduleID] = events.EmptyList()
	}

	return wi
}

// AddEvents adds events produced by modules to the eventBuffer buffer.
// According to their DestModule fields, the events are distributed to the appropriate internal sub-buffers.
// When AddEvents returns a non-nil error, any subset of the events may have been added.
func (wi *eventBuffer) AddEvents(events *events.EventList) error {
	// Note that this MUST be a pointer receiver.
	// Otherwise, we'd increment a copy of the event counter rather than the counter itself.
	iter := events.Iterator()

	// For each incoming event
	for event := iter.Next(); event != nil; event = iter.Next() {

		// Look up the corresponding module's buffer and add the event to it.
		if buffer, ok := wi.buffers[t.ModuleID(event.DestModule).Top()]; ok {
			buffer.PushBack(event)
			wi.totalEvents++
		} else {

			// If there is no buffer for the event, check if it is a MessgeReceived event.
			isMessageReceivedEvent := false
			e, isTransportEvent := event.Type.(*eventpb.Event_Transport)
			if isTransportEvent {
				_, isMessageReceivedEvent = e.Transport.Type.(*transportpb.Event_MessageReceived)
			}

			// If there is no buffer but the event is a MessageReceived event,
			// we don't need to return an error,
			// because the message may have been sent by a faulty node.
			// MessageReceived events with non-existent module destinations are thus dropped,
			// without it being considered an error at the local node.
			if !isMessageReceivedEvent {
				return errors.Errorf("no buffer for module %v (adding event of type %T)", event.DestModule, event.Type)
			}
		}
	}
	return nil
}
