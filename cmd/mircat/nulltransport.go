package main

import (
	"context"

	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/stdtypes"
)

type NullTransport struct{}

func (n *NullTransport) ImplementsModule() {}

func (n *NullTransport) ApplyEvents(_ context.Context, _ *stdtypes.EventList) error {
	return nil
}

func (n *NullTransport) EventsOut() <-chan *stdtypes.EventList {
	return nil
}

func (n *NullTransport) Start() error {
	return nil
}

func (n *NullTransport) Stop() {}

func (n *NullTransport) Send(_ stdtypes.NodeID, _ *messagepb.Message) error {
	return nil
}

func (n *NullTransport) Connect(_ *trantorpbtypes.Membership) {
}

func (n *NullTransport) WaitFor(_ int) error {
	return nil
}

func (n *NullTransport) CloseOldConnections(_ *trantorpbtypes.Membership) {
}
