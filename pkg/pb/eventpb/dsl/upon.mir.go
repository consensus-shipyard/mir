package eventpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	eventpb "github.com/filecoin-project/mir/pkg/pb/eventpb"
	types "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing events.

func UponSignRequest(m dsl.Module, handler func(data [][]uint8, origin *types.SignOrigin) error) {
	dsl.UponMirEvent[*types.Event_SignRequest](m, func(ev *types.SignRequest) error {
		return handler(ev.Data, ev.Origin)
	})
}

func UponSignResult(m dsl.Module, handler func(signature []uint8, origin *types.SignOrigin) error) {
	dsl.UponMirEvent[*types.Event_SignResult](m, func(ev *types.SignResult) error {
		return handler(ev.Signature, ev.Origin)
	})
}

func UponNodeSigsVerified(m dsl.Module, handler func(origin *eventpb.SigVerOrigin, nodeIds []types1.NodeID, valid []bool, errors []error, allOk bool) error) {
	dsl.UponMirEvent[*types.Event_NodeSigsVerified](m, func(ev *types.NodeSigsVerified) error {
		return handler(ev.Origin, ev.NodeIds, ev.Valid, ev.Errors, ev.AllOk)
	})
}

func UponSendMessage(m dsl.Module, handler func(msg *types2.Message, destinations []types1.NodeID) error) {
	dsl.UponMirEvent[*types.Event_SendMessage](m, func(ev *types.SendMessage) error {
		return handler(ev.Msg, ev.Destinations)
	})
}

func UponMessageReceived(m dsl.Module, handler func(from types1.NodeID, msg *types2.Message) error) {
	dsl.UponMirEvent[*types.Event_MessageReceived](m, func(ev *types.MessageReceived) error {
		return handler(ev.From, ev.Msg)
	})
}
