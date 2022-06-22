package modules

import (
	"context"
	"github.com/filecoin-project/mir/pkg/events"
)

// The NullPassive module is a PassiveModule that ignores all incoming events.
type NullPassive struct {
}

func (n NullPassive) ApplyEvents(_ *events.EventList) (*events.EventList, error) {
	return events.EmptyList(), nil
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (n NullPassive) ImplementsModule() {}

// The NullActive module is an ActiveModule that ignores all incoming events and never produces any events.
type NullActive struct {
	outChan <-chan *events.EventList
}

func (n NullActive) ApplyEvents(_ context.Context, _ *events.EventList) error {
	return nil
}

func (n NullActive) EventsOut() <-chan *events.EventList {
	return n.outChan
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (n NullActive) ImplementsModule() {}
