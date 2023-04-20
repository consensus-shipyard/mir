package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/isspb"
	"github.com/filecoin-project/mir/pkg/pb/recordingpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type eventMetadata struct {
	time   int64
	nodeID t.NodeID
	index  uint64
}

// Returns the list of event names and destinations present in the given eventlog file,
// along with the total number of events present in the file.
func getEventList(filenames *[]string) (map[string]struct{}, map[string]struct{}, map[string]struct{}, int, error) {
	events := make(map[string]struct{})
	issEvents := make(map[string]struct{})
	eventDests := make(map[string]struct{})

	totalCount := 0
	for _, filename := range *filenames {
		// Open the file
		file, err := os.Open(filename)
		if err != nil {
			kingpin.Errorf("Error opening src file", filename, ": ", err)
			fmt.Printf("\n\n!!!\nContinuing after error with new file. Event list might be incomplete!\n!!!\n\n")
			continue
		}

		defer func(file *os.File, offset int64, whence int) {
			_, _ = file.Seek(offset, whence) // resets the file offset for successive reading
			if err = file.Close(); err != nil {
				kingpin.Errorf("Error closing src file", filename, ": ", err)
			}
		}(file, 0, 0)

		reader, err := eventlog.NewReader(file)
		if err != nil {
			kingpin.Errorf("Error creating new reader of src file", filename, ": ", err)
			fmt.Printf("\n\n!!!\nContinuing after error with new file. Event list might be incomplete!\n!!!\n\n")
			continue
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
				eventDests[event.DestModule] = struct{}{}
				switch e := event.Type.(type) {
				case *eventpb.Event_Iss:
					// For ISS Events, also add the type of the ISS event to a set of known ISS events.
					issEvents[issEventName(e.Iss)] = struct{}{}
				}
			}
		}
		if !errors.Is(err, io.EOF) {
			kingpin.Errorf("Error reading src file", filename, ": ", err)
			fmt.Printf("\n\n!!!\nContinuing after error. Event list might be incomplete!\n!!!\n\n")
		}
		totalCount += cnt
	}

	return events, issEvents, eventDests, totalCount, nil // TODO wrap non-blocking errors and return them here
}

// eventName returns a string name of an Event.
func eventName(event *eventpb.Event) string {
	return strings.ReplaceAll(
		reflect.TypeOf(event.Type).Elem().Name(), // gets the type's name i.e. Event_Tick , Event_Iss,etc
		"Event_", "")
}

// issEventName returns a string name of an ISS event.
func issEventName(issEvent *isspb.Event) string {
	return strings.ReplaceAll(
		reflect.TypeOf(issEvent.Type).Elem().Name(), // gets the type's name i.e. ISSEvent_sb , ISSEvent_PersistCheckpoint,etc
		"ISSEvent_", "") // replaces the given substring from the name
}

// selected returns true if the given event has been selected by the user according to the given criteria.
func selected(event *eventpb.Event, selectedEvents map[string]struct{}, selectedIssEvents map[string]struct{}) bool {
	if _, ok := selectedEvents[eventName(event)]; !ok {
		// If the basic type of the event has not been selected, return false.
		return false
	}

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
