package mir

import (
	"github.com/filecoin-project/mir/pkg/modules"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/stdtypes"
)

// eventBuffer is a buffer for storing outstanding events that need to be processed by the node.
// It contains a separate list for each module.
type eventBuffer struct {
	buffers     map[t.ModuleID]*stdtypes.EventList
	totalEvents int
}

// newEventBuffer allocates and returns a pointer to a new eventBuffer object.
func newEventBuffer(modules modules.Modules) eventBuffer {

	wi := eventBuffer{
		buffers:     make(map[t.ModuleID]*stdtypes.EventList),
		totalEvents: 0,
	}

	for moduleID := range modules {
		wi.buffers[moduleID] = stdtypes.EmptyList()
	}

	return wi
}

// Add adds events produced by modules to the eventBuffer buffer.
// According to their DestModule fields, the events are distributed to the appropriate internal sub-buffers.
// When Add returns a non-nil error, any subset of the events may have been added.
func (eb *eventBuffer) Add(events *stdtypes.EventList) error {
	// Note that this MUST be a pointer receiver.
	// Otherwise, we'd increment a copy of the event counter rather than the counter itself.
	iter := events.Iterator()

	// For each incoming event
	for event := iter.Next(); event != nil; event = iter.Next() {

		// Look up the corresponding module's buffer and add the event to it.
		if buffer, ok := eb.buffers[event.Dest().Top()]; ok {
			buffer.PushBack(event)
			eb.totalEvents++
		} else { //nolint:revive,staticcheck

			// Silently drop events for which the destination module does not exist.
			// WARNING: While this is the desired behavior when it comes to untrusted events (MessageReceived events)
			// from potentially Faulty nodes, in general this is very dangerous. If a locally generated event has a
			// wrong destination by accident, this error will be harder to spot.
			// TODO: Instead, introduce a concept of a default module (at Mir level) that will receive all such events.
			//   That module (supplied by the Mir user, with a reasonable default implementation) can then figure out
			//   what to do with those events and execute code similar to the one commented out below.
			//   Below is the old code checking for network messages and only silently ignoring those,
			//   otherwise raising an error.

			//// If there is no buffer for the event, check if it is a MessgeReceived event.
			//isMessageReceivedEvent := false
			//e, isTransportEvent := event.Type.(*eventpb.Event_Transport)
			//if isTransportEvent {
			//	_, isMessageReceivedEvent = e.Transport.Type.(*transportpb.Event_MessageReceived)
			//}
			//
			//// If there is no buffer but the event is a MessageReceived event,
			//// we don't need to return an error,
			//// because the message may have been sent by a faulty node.
			//// MessageReceived events with non-existent module destinations are thus dropped,
			//// without it being considered an error at the local node.
			//if !isMessageReceivedEvent {
			//	return errors.Errorf("no buffer for module %v (adding event of type %T)", event.DestModule, event.Type)
			//}
		}
	}
	return nil
}

// Stats returns a map containing the number of buffered events for each module.
func (eb *eventBuffer) Stats() map[t.ModuleID]int {
	stats := make(map[t.ModuleID]int)
	for mID, evts := range eb.buffers {
		stats[mID] = evts.Len()
	}
	return stats
}
