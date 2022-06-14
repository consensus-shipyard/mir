/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

Refactored: 1
*/

package mir

import (
	"fmt"
	"github.com/filecoin-project/mir/pkg/modules"
	t "github.com/filecoin-project/mir/pkg/types"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
)

// WorkItems is a buffer for storing outstanding events that need to be processed by the node.
// It contains a separate list for each type of event.
type workItems struct {
	net      *events.EventList
	protocol *events.EventList
	timer    *events.EventList

	generic map[t.ModuleID]*events.EventList
}

// NewWorkItems allocates and returns a pointer to a new WorkItems object.
func newWorkItems(modules *modules.Modules) *workItems {

	generic := make(map[t.ModuleID]*events.EventList)

	for moduleID := range modules.GenericModules {
		generic[moduleID] = &events.EventList{}
	}

	return &workItems{
		net:      &events.EventList{},
		protocol: &events.EventList{},
		timer:    &events.EventList{},

		generic: generic,
	}
}

// AddEvents adds events produced by modules to the workItems buffer.
// According to their types, the events are distributed to the appropriate internal sub-buffers.
// When AddEvents returns a non-nil error, any subset of the events may have been added.
func (wi *workItems) AddEvents(events *events.EventList) error {
	iter := events.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {

		// If the event has a target generic module, add it to the corresponding buffer and go to next event.
		if buffer, ok := wi.generic[t.ModuleID(event.DestModule)]; ok {
			buffer.PushBack(event)
			continue
		}

		switch t := event.Type.(type) {
		case *eventpb.Event_Init:
			wi.protocol.PushBack(event)
			// TODO: Should the Init event also go elsewhere? Clients? All the modules?
		case *eventpb.Event_SendMessage:
			wi.net.PushBack(event)
		case *eventpb.Event_MessageReceived, *eventpb.Event_Iss, *eventpb.Event_RequestReady,
			*eventpb.Event_AppSnapshot:
			wi.protocol.PushBack(event)
		case *eventpb.Event_HashResult:
			// For hash results, their origin determines the destination.
			switch t.HashResult.Origin.Type.(type) {
			case *eventpb.HashOrigin_Iss:
				// The ISS origin goes to the Protocol module (ISS is a protocol implementation).
				wi.protocol.PushBack(event)
			}
		case *eventpb.Event_SignResult:
			// For sign results, their origin determines the destination.
			switch t.SignResult.Origin.Type.(type) {
			case *eventpb.SignOrigin_Iss:
				// The ISS origin goes to the Protocol module (ISS is a protocol implementation).
				wi.protocol.PushBack(event)
			}
		case *eventpb.Event_NodeSigsVerified:
			// For signature verification results, their origin determines the destination.
			switch t.NodeSigsVerified.Origin.Type.(type) {
			case *eventpb.SigVerOrigin_Iss:
				// The ISS origin goes to the Protocol module (ISS is a protocol implementation).
				wi.protocol.PushBack(event)
			}
		case *eventpb.Event_WalEntry:
			switch walEntry := t.WalEntry.Event.Type.(type) {
			case *eventpb.Event_Iss:
				// TODO: WAL loading is now disabled for ISS by commenting out the next line.
				//       Implement recovery and re-enable (un-comment).
				// wi.protocol.PushBack(t.WalEntry.Event)
			case *eventpb.Event_PersistDummyBatch:
				wi.protocol.PushBack(t.WalEntry.Event)
			default:
				return fmt.Errorf("unsupported WAL entry event type %T", walEntry)
			}
		case *eventpb.Event_TimerDelay, *eventpb.Event_TimerRepeat, *eventpb.Event_TimerGarbageCollect:
			wi.timer.PushBack(event)
		default:
			return fmt.Errorf("cannot add event of unknown type %T", t)
		}
	}
	return nil
}

// Getters.

func (wi *workItems) Net() *events.EventList {
	return wi.net
}

func (wi *workItems) Protocol() *events.EventList {
	return wi.protocol
}

func (wi *workItems) Timer() *events.EventList {
	return wi.timer
}

// Methods for clearing the buffers.
// Each of them returns the list of events that have been removed from workItems.

func (wi *workItems) ClearNet() *events.EventList {
	return clearEventList(&wi.net)
}

func (wi *workItems) ClearProtocol() *events.EventList {
	return clearEventList(&wi.protocol)
}

func (wi *workItems) ClearTimer() *events.EventList {
	return clearEventList(&wi.timer)
}

func clearEventList(listPtr **events.EventList) *events.EventList {
	oldList := *listPtr
	*listPtr = &events.EventList{}
	return oldList
}
