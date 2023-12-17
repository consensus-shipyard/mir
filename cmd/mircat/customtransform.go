package main

import (
	"github.com/filecoin-project/mir/stdtypes"
)

// This function is applied to every event loaded from the event log
// and its return value is used instead of the original event.
// If customEventFilter returns nil, the event is ignored (this can be used for additional event filtering).
// It is meant for ad-hoc editing while debugging, to be able to select events in a fine-grained way.
func customTransform(e stdtypes.Event) stdtypes.Event {

	//moduleID := ""
	//if e.Dest() == moduleID {
	//	return e
	//}
	//return nil

	return e
}
