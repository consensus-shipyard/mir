package main

import (
	"fmt"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/isspb"
	"github.com/filecoin-project/mir/pkg/pb/recordingpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
)

type eventMetadata struct {
	time   int64
	nodeID t.NodeID
	index  uint64
}

// Returns the list of event names present in the given eventlog file,
// along with the total number of events present in the file.
func getEventList(file *os.File) (map[string]struct{}, map[string]struct{}, int, error) {
	events := make(map[string]struct{})
	issEvents := make(map[string]struct{})

	defer func(file *os.File, offset int64, whence int) {
		_, _ = file.Seek(offset, whence) // resets the file offset for successive reading
	}(file, 0, 0)

	reader, err := eventlog.NewReader(file)
	if err != nil {
		return nil, nil, 0, err
	}

	cnt := 0 // Counts the total number of events in the event log.
	var entry *recordingpb.Entry
	for entry, err = reader.ReadEntry(); err == nil; entry, err = reader.ReadEntry() {
		// For each entry of the event log

		for _, event := range entry.Events {
			// For each Event in the entry
			cnt++

			// Add the Event name to the set of known Events.
			events[eventName(event)] = struct{}{}
			switch e := event.Type.(type) {
			case *eventpb.Event_Iss:
				// For ISS Events, also add the type of the ISS event to a set of known ISS events.
				issEvents[issEventName(e.Iss)] = struct{}{}
			}
		}
	}
	if err != io.EOF {
		return events, issEvents, cnt, fmt.Errorf("failed reading event log: %w", err)
	}

	return events, issEvents, cnt, nil
}

// eventName returns a string name of an Event.
func eventName(event *eventpb.Event) string {
	return strings.ReplaceAll(
		reflect.TypeOf(event.Type).Elem().Name(), //gets the type's name i.e. Event_Tick , Event_Iss,etc
		"Event_", "")
}

// issEventName returns a string name of an ISS event.
func issEventName(issEvent *isspb.ISSEvent) string {
	return strings.ReplaceAll(
		reflect.TypeOf(issEvent.Type).Elem().Name(), //gets the type's name i.e. ISSEvent_sb , ISSEvent_PersistCheckpoint,etc
		"ISSEvent_", "") //replaces the given substring from the name
}

// selected returns true if the given event has been selected by the user according to the given criteria.
func selected(event *eventpb.Event, selectedEvents map[string]struct{}, selectedIssEvents map[string]struct{}) bool {
	if _, ok := selectedEvents[eventName(event)]; !ok {
		// If the basic type of the event has not been selected, return false.
		return false
	} else {
		// If the basic type of the event has been selected,
		// check whether the sub-type has been selected as well for ISS events.
		switch e := event.Type.(type) {
		case *eventpb.Event_Iss:
			_, ok := selectedIssEvents[issEventName(e.Iss)]
			return ok
		default:
			return true
		}
	}
}

// Converts a set of strings (represented as a map) to a list.
// Returns a slice containing all the keys present in the given set.
// toList is used to convert sets to a format used by the survey library.
func toList(set map[string]struct{}) []string {
	list := make([]string, 0, len(set))
	for item := range set {
		list = append(list, item)
	}
	sort.Strings(list)
	return list
}

// Converts a list of strings to a set (represented as a map).
// Returns a map of empty structs with one entry for each unique item of the given list (the item being the map key).
// toSet is used to convert lists produced by the survey library to sets for easier lookup.
func toSet(list []string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, item := range list {
		set[item] = struct{}{}
	}
	return set
}
