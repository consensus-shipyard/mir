package modules

import (
	"github.com/filecoin-project/mir/stdtypes"
)

type PassiveModule interface {
	Module

	// ApplyEvents applies a list of input events to the module, making it advance its state
	// and returns a (potentially empty) list of output events that the application of the input events results in.
	ApplyEvents(events *stdtypes.EventList) (*stdtypes.EventList, error)
}
