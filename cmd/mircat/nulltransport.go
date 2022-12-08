package main

import (
	"context"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type NullTransport struct{}

func (n *NullTransport) ImplementsModule() {}

func (n *NullTransport) ApplyEvents(_ context.Context, _ *events.EventList) error {
	return nil
}

func (n *NullTransport) EventsOut() <-chan *events.EventList {
	return nil
}

func (n *NullTransport) Start() error {
	return nil
}

func (n *NullTransport) Stop() {}

func (n *NullTransport) Send(_ t.NodeID, _ *messagepb.Message) error {
	return nil
}

func (n *NullTransport) Connect(_ map[t.NodeID]t.NodeAddress) {
}

func (n *NullTransport) WaitFor(_ int) {
}

func (n *NullTransport) CloseOldConnections(_ map[t.NodeID]t.NodeAddress) {
}
