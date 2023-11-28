package eventlog

import (
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/apppb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// EventNewEpochLogger returns a file that splits an event list into multiple lists
// every time an event eventpb.Event_NewLogFile is found
func EventNewEpochLogger(appModuleID t.ModuleID) func(*events.EventList) []*events.EventList {
	eventNewLogFileLogger := func(event *eventpb.Event) bool {
		appEvent, ok := event.Type.(*eventpb.Event_App)
		if !ok {
			return false
		}

		_, ok = appEvent.App.Type.(*apppb.Event_NewEpoch)
		return ok && t.ModuleID(event.DestModule) == appModuleID
	}
	return EventTrackerLogger(eventNewLogFileLogger)
}

// EventTrackerLogger returns a function that tracks every single event of EventList and
// creates a new file for every event such that newFile(event) = True
func EventTrackerLogger(newFile func(event *eventpb.Event) bool) func(*events.EventList) []*events.EventList {
	return func(evts *events.EventList) []*events.EventList {
		var result []*events.EventList
		currentChunk := events.EmptyList()

		for _, event := range evts.Slice() {
			if newFile(event) {
				result = append(result, currentChunk)
				currentChunk = events.ListOf(event)
			} else {
				currentChunk.PushBack(event)
			}
		}

		// If there is a remaining chunk with fewer than the desired number of events, append it to the result
		if currentChunk.Len() > 0 {
			result = append(result, currentChunk)
		}

		return result
	}
}

// EventLimitLogger returns a function for the interceptor that splits the logging file
// every eventLimit number of events
func EventLimitLogger(eventLimit int64) func(*events.EventList) []*events.EventList {
	var eventCount int64
	return func(evts *events.EventList) []*events.EventList {
		var result []*events.EventList
		currentChunk := events.EmptyList()

		// Iterate over the events in the input slice
		for _, event := range evts.Slice() {
			// Add the current element to the current chunk
			currentChunk.PushBack(event)
			eventCount++
			// If the current chunk has the desired number of events, append it to the result and start a new chunk
			if eventCount%eventLimit == 0 {
				result = append(result, currentChunk)
				currentChunk = events.EmptyList()
			}
		}

		// If there is a remaining chunk with fewer than the desired number of events, append it to the result
		if currentChunk.Len() > 0 {
			result = append(result, currentChunk)
		}

		return result
	}
}

func OneFileLogger() func(*events.EventList) []*events.EventList {
	return func(evts *events.EventList) []*events.EventList {
		return []*events.EventList{evts}
	}
}
