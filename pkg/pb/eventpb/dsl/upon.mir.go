package eventpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	eventpb "github.com/filecoin-project/mir/pkg/pb/eventpb"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing events.

func UponNodeSigsVerified(m dsl.Module, handler func(origin *eventpb.SigVerOrigin, nodeIds []types.NodeID, valid []bool, errors []error, allOk bool) error) {
	dsl.UponMirEvent[*types1.Event_NodeSigsVerified](m, func(ev *types1.NodeSigsVerified) error {
		return handler(ev.Origin, ev.NodeIds, ev.Valid, ev.Errors, ev.AllOk)
	})
}

func UponSendMessage(m dsl.Module, handler func(msg *types2.Message, destinations []types.NodeID) error) {
	dsl.UponMirEvent[*types1.Event_SendMessage](m, func(ev *types1.SendMessage) error {
		return handler(ev.Msg, ev.Destinations)
	})
}

func UponMessageReceived(m dsl.Module, handler func(from types.NodeID, msg *types2.Message) error) {
	dsl.UponMirEvent[*types1.Event_MessageReceived](m, func(ev *types1.MessageReceived) error {
		return handler(ev.From, ev.Msg)
	})
}
