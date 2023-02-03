package main

import "github.com/filecoin-project/mir/pkg/pb/eventpb"

// This function is applied to every event loaded from the event log
// and its return value is used instead of the original event.
// If customEventFilter returns nil, the event is ignored (this can be used for additional event filtering).
// It is meant for ad-hoc editing while debugging, to be able to select events in a fine-grained way.
func customTransform(e *eventpb.Event) *eventpb.Event {
	return e
}
