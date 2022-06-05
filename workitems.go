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
)

// WorkItems is a buffer for storing outstanding events that need to be processed by the node.
// It contains a separate list for each destination module.
type workItems map[t.ModuleID]*events.EventList

// NewWorkItems allocates and returns a pointer to a new workItems object.
// The returned workItems contain an empty buffer (EventList) for each module in modules.
func newWorkItems(modules modules.Modules) workItems {
	wi := make(map[t.ModuleID]*events.EventList)

	for moduleID := range modules {
		wi[moduleID] = &events.EventList{}
	}

	return wi
}

// AddEvents adds events produced by modules to the workItems buffer.
// According to their destination modules, the events are distributed to the appropriate internal sub-buffers.
// When AddEvents returns a non-nil error, any subset of the events may have been added.
func (wi workItems) AddEvents(evts *events.EventList) error {
	iter := evts.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {

		buffer, ok := wi[t.ModuleID(event.Destination)]
		if !ok {
			return fmt.Errorf("unknown event destination: %v", event.Destination)
		}
		buffer.PushBack(event)
	}
	return nil
}
