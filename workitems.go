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
	net *events.EventList

	generic map[t.ModuleID]*events.EventList
}

// NewWorkItems allocates and returns a pointer to a new WorkItems object.
func newWorkItems(modules *modules.Modules) *workItems {

	generic := make(map[t.ModuleID]*events.EventList)

	for moduleID := range modules.GenericModules {
		generic[moduleID] = &events.EventList{}
	}

	return &workItems{
		net: &events.EventList{},

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
		case *eventpb.Event_SendMessage:
			wi.net.PushBack(event)
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

// Methods for clearing the buffers.
// Each of them returns the list of events that have been removed from workItems.

func (wi *workItems) ClearNet() *events.EventList {
	return clearEventList(&wi.net)
}

func clearEventList(listPtr **events.EventList) *events.EventList {
	oldList := *listPtr
	*listPtr = &events.EventList{}
	return oldList
}
