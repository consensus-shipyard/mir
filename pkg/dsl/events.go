package dsl

import (
	"github.com/filecoin-project/mir/pkg/events"
	dslpbtypes "github.com/filecoin-project/mir/pkg/pb/dslpb/types"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// Origin creates a dslpb.Origin protobuf.
func Origin(contextID ContextID) *dslpbtypes.Origin {
	return &dslpbtypes.Origin{
		ContextID: contextID.Pb(),
	}
}

// MirOrigin creates a dslpb.Origin protobuf.
func MirOrigin(contextID ContextID) *dslpbtypes.Origin {
	return &dslpbtypes.Origin{
		ContextID: contextID.Pb(),
	}
}

// Dsl functions for emitting events.
// TODO: add missing event types.
// TODO: consider generating this code automatically using a protoc plugin.

// SendMessage emits a request event to send a message over the network.
// The message should be processed on the receiving end using UponMessageReceived.
func SendMessage(m Module, destModule t.ModuleID, msg *messagepb.Message, dest []t.NodeID) {
	EmitEvent(m, events.SendMessage(destModule, msg, dest))
}

// Dsl functions for processing events
// TODO: consider generating this code automatically using a protoc plugin.

// UponInit invokes handler when the module is initialized.
func UponInit(m Module, handler func() error) {
	UponEvent[*eventpb.Event_Init](m, func(ev *eventpb.Init) error {
		return handler()
	})
}

// UponMessageReceived invokes handler when the module receives a message over the network.
func UponMessageReceived(m Module, handler func(from t.NodeID, msg *messagepb.Message) error) {
	UponEvent[*eventpb.Event_MessageReceived](m, func(ev *eventpb.MessageReceived) error {
		return handler(t.NodeID(ev.From), ev.Msg)
	})
}

// UponNewRequests invokes handler when the module receives a NewRequests event.
func UponNewRequests(m Module, handler func(requests []*requestpb.Request) error) {
	UponEvent[*eventpb.Event_NewRequests](m, func(ev *eventpb.NewRequests) error {
		return handler(ev.Requests)
	})
}
