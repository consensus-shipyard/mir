package main

import (
	"context"

	"github.com/filecoin-project/mir/pkg/events"
	messagepbtypes "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
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

func (n *NullTransport) Send(_ t.NodeID, _ *messagepbtypes.Message) error {
	return nil
}

func (n *NullTransport) Connect(_ *trantorpbtypes.Membership) {
}

func (n *NullTransport) WaitFor(_ int) error {
	return nil
}

func (n *NullTransport) CloseOldConnections(_ *trantorpbtypes.Membership) {
}
