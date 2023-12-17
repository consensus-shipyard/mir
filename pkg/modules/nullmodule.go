package modules

import (
	"context"

	"github.com/filecoin-project/mir/stdtypes"
)

// The NullPassive module is a PassiveModule that ignores all incoming events.
type NullPassive struct {
}

func (n NullPassive) ApplyEvents(_ *stdtypes.EventList) (*stdtypes.EventList, error) {
	return stdtypes.EmptyList(), nil
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (n NullPassive) ImplementsModule() {}

// The NullActive module is an ActiveModule that ignores all incoming events and never produces any events.
type NullActive struct {
	outChan <-chan *stdtypes.EventList
}

func (n NullActive) ApplyEvents(_ context.Context, _ *stdtypes.EventList) error {
	return nil
}

func (n NullActive) EventsOut() <-chan *stdtypes.EventList {
	return n.outChan
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (n NullActive) ImplementsModule() {}
