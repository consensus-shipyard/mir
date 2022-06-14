package mir

import (
	"fmt"
	"github.com/filecoin-project/mir/pkg/modules"
	t "github.com/filecoin-project/mir/pkg/types"

	"github.com/filecoin-project/mir/pkg/events"
)

// WorkItems is a buffer for storing outstanding events that need to be processed by the node.
// It contains a separate list for each module.
type workItems map[t.ModuleID]*events.EventList

// NewWorkItems allocates and returns a pointer to a new WorkItems object.
func newWorkItems(modules modules.Modules) workItems {

	wi := make(map[t.ModuleID]*events.EventList)

	for moduleID := range modules {
		wi[moduleID] = &events.EventList{}
	}

	return wi
}

// AddEvents adds events produced by modules to the workItems buffer.
// According to their DestModule fields, the events are distributed to the appropriate internal sub-buffers.
// When AddEvents returns a non-nil error, any subset of the events may have been added.
func (wi workItems) AddEvents(events *events.EventList) error {
	iter := events.Iterator()

	// For each incoming event
	for event := iter.Next(); event != nil; event = iter.Next() {

		// Look up the corresponding module's buffer and add the event to it.
		if buffer, ok := wi[t.ModuleID(event.DestModule)]; ok {
			buffer.PushBack(event)
		} else {
			return fmt.Errorf("no buffer for module %v (adding event of type %T)", event.DestModule, event.Type)
		}
	}
	return nil
}
