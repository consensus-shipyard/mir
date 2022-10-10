/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package events

import (
	"container/list"

	"github.com/filecoin-project/mir/pkg/pb/eventpb"
)

// EventList represents a list of Events, e.g. as produced by a module.
type EventList struct {

	// The internal list is intentionally left uninitialized until its actual use.
	// This probably speeds up appending empty lists to other lists.
	list *list.List
}

// EmptyList returns an empty EventList.
// TODO: consider passing EventList by value here and everywhere else.
func EmptyList() *EventList {
	return &EventList{}
}

// ListOf returns EventList containing the given elements.
func ListOf(events ...*eventpb.Event) *EventList {
	res := &EventList{}
	for _, ev := range events {
		res.PushBack(ev)
	}
	return res
}

// Len returns the number of events in the EventList.
func (el *EventList) Len() int {
	if el.list == nil {
		return 0
	}
	return el.list.Len()
}

// PushBack appends an event to the end of the list.
// Returns the EventList itself, for the convenience of chaining multiple calls to PushBack.
func (el *EventList) PushBack(event *eventpb.Event) *EventList {
	if el.list == nil {
		el.list = list.New()
	}

	el.list.PushBack(event)
	return el
}

// PushBackSlice appends all events in newEvents to the end of the current EventList.
func (el *EventList) PushBackSlice(events []*eventpb.Event) *EventList {
	if el.list == nil {
		el.list = list.New()
	}

	for _, event := range events {
		el.list.PushBack(event)
	}

	return el
}

// PushBackList appends all events in newEvents to the end of the current EventList.
func (el *EventList) PushBackList(newEvents *EventList) *EventList {
	if newEvents.list != nil {
		if el.list == nil {
			el.list = list.New()
		}
		// TODO: Check out possible inefficiency. This implementation actually (shallowly) copies the elements from
		//       newEvents.list to el.list. Most of the time it would be enough to simply wire together the two lists.
		//       This would probably require a custom linked list implementation, since list.List does not seem to
		//       support it.
		el.list.PushBackList(newEvents.list)
	}

	return el
}

// Head returns the first up to n events in the list as a new list.
// The original list is not modified.
func (el *EventList) Head(n int) *EventList {
	if el.list == nil {
		return EmptyList()
	}

	result := EmptyList()
	iter := el.Iterator()
	for i := 0; i < n; i++ {
		event := iter.Next()
		if event == nil {
			break
		}
		result.PushBack(event)
	}
	return result
}

// RemoveFront removes the first up to n events from the list.
// Returns the number of events actually removed.
func (el *EventList) RemoveFront(n int) int {
	if el.list == nil {
		return 0
	}

	for i := 0; i < n; i++ {
		if first := el.list.Front(); first != nil {
			el.list.Remove(first)
		} else {
			return i
		}
	}

	return n
}

// Slice returns a slice representation of the current state of the list.
// The returned slice only contains pointers to the events in this list, no deep copying is performed.
// Any modifications performed on the events will affect the contents of both the EventList and the returned slice.
func (el *EventList) Slice() []*eventpb.Event {
	if el.list == nil {
		return nil
	}

	// Create empty result slice.
	events := make([]*eventpb.Event, 0, el.Len())

	// Populate result slice by appending events one by one.
	iter := el.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {
		events = append(events, event)
	}

	// Return populated result slice.
	return events
}

// StripFollowUps collects all follow-up Events of the Events in the list.
// It returns two lists:
// 1. An EventList containing the same events as this list, but with all follow-up events removed.
// 2. An EventList containing only those follow-up events.
func (el *EventList) StripFollowUps() (*EventList, *EventList) {
	// Create list of follow-up Events.
	followUps := EventList{}

	// Create a new EventList for events with follow-ups removed.
	plainEvents := EventList{}

	// Populate list by follow-up events
	iter := el.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {
		plainEvent, strippedEvents := Strip(event)
		plainEvents.PushBack(plainEvent)
		followUps.PushBackList(strippedEvents)
	}

	// Return populated list of follow-up events.
	return &plainEvents, &followUps
}

// Iterator returns a pointer to an EventListIterator object used to iterate over the events in this list,
// starting from the beginning of the list.
func (el *EventList) Iterator() *EventListIterator {
	if el.list == nil {
		return &EventListIterator{}
	}

	return &EventListIterator{
		currentElement: el.list.Front(),
	}
}

// EventListIterator is an object returned from EventList.Iterator
// used to iterate over the elements (Events) of an EventList using the iterator's Next method.
type EventListIterator struct {
	currentElement *list.Element
}

// Next will return the next Event until the end of the associated EventList is encountered.
// Thereafter, it will return nil.
func (eli *EventListIterator) Next() *eventpb.Event {

	// Return nil if list has been exhausted.
	if eli.currentElement == nil {
		return nil
	}

	// Obtain current element and move on to the next one.
	result := eli.currentElement.Value.(*eventpb.Event)
	eli.currentElement = eli.currentElement.Next()

	// Return current element.
	return result
}
