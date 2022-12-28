package eventlog

import (
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
)

// Returns a file that splits an eventTime slice into multiple slices
// every time a an event eventpb.Event_NewLogFile is found
func EventNewLogFileLogger() func(time EventTime) *[]EventTime {
	return EventTrackerLogger(func(event eventpb.Event) bool {
		_, ok := event.Type.(*eventpb.Event_NewLogFile)
		return ok
	})
}

// eventTrackerLogger returns a function that tracks every single event of EventTime and
// creates a new file for every event such that newFile(event) = True
func EventTrackerLogger(newFile func(event eventpb.Event) bool) func(time EventTime) *[]EventTime {
	return func(eventTime EventTime) *[]EventTime {
		var result []EventTime
		// Create a variable to hold the current chunk
		currentChunk := &EventTime{
			Time:   eventTime.Time,
			Events: events.EmptyList(),
		}

		for _, event := range eventTime.Events.Slice() {
			if newFile(*event) {
				result = append(result, *currentChunk)
				currentChunk = &EventTime{
					Time:   eventTime.Time,
					Events: events.EmptyList(),
				}
			} else {
				currentChunk.Events.PushBack(event)
			}
		}

		// If there is a remaining chunk with fewer than the desired number of events, append it to the result
		if currentChunk.Events.Len() > 0 {
			result = append(result, *currentChunk)
		}

		return &result
	}
}

// EventLimitLogger returns a function for the interceptor that splits the logging file
// every eventLimit number of events
func EventLimitLogger(eventLimit int64) func(EventTime) *[]EventTime {
	var eventCount int64 = 0
	return func(eventTime EventTime) *[]EventTime {
		// Create a slice to hold the slices of eventTime elements
		var result []EventTime
		// Create a variable to hold the current chunk
		currentChunk := &EventTime{
			Time:   eventTime.Time,
			Events: events.EmptyList(),
		}

		// Iterate over the events in the input slice
		for _, event := range eventTime.Events.Slice() {
			// Add the current element to the current chunk
			currentChunk.Events.PushBack(event)
			eventCount++
			// If the current chunk has the desired number of events, append it to the result and start a new chunk
			if eventCount%eventLimit == 0 {
				result = append(result, *currentChunk)
				currentChunk = &EventTime{
					Time:   eventTime.Time,
					Events: events.EmptyList(),
				}
			}
		}

		// If there is a remaining chunk with fewer than the desired number of events, append it to the result
		if currentChunk.Events.Len() > 0 {
			result = append(result, *currentChunk)
		}

		return &result
	}
}

func OneFileLogger() func(EventTime) *[]EventTime {
	return func(eventTime EventTime) *[]EventTime {
		return &[]EventTime{eventTime}
	}
}
