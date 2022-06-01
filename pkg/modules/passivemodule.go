package modules

import (
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/statuspb"
)

type PassiveModule interface {

	// ApplyEvent applies a single input event to the module, making it advance its state
	// and return a (potentially empty) list of output events that the application of the input event results in.
	ApplyEvent(event *eventpb.Event) (*events.EventList, error)

	// Status returns the current state of the module.
	// Mostly for debugging purposes.
	Status() (s *statuspb.ProtocolStatus, err error)
}
