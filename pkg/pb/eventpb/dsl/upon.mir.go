package eventpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
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

func UponSignResult[C any](m dsl.Module, handler func(signature []uint8, context *C) error) {
	dsl.UponMirEvent[*types.Event_SignResult](m, func(ev *types.SignResult) error {
		originWrapper, ok := ev.Origin.Type.(*types.SignOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.Signature, context)
	})
}

func UponVerifyNodeSigs(m dsl.Module, handler func(data []*types.SigVerData, signatures [][]uint8, origin *types.SigVerOrigin, nodeIds []types1.NodeID) error) {
	dsl.UponMirEvent[*types.Event_VerifyNodeSigs](m, func(ev *types.VerifyNodeSigs) error {
		return handler(ev.Data, ev.Signatures, ev.Origin, ev.NodeIds)
	})
}

func UponNodeSigsVerified[C any](m dsl.Module, handler func(nodeIds []types1.NodeID, valid []bool, errors []error, allOk bool, context *C) error) {
	dsl.UponMirEvent[*types.Event_NodeSigsVerified](m, func(ev *types.NodeSigsVerified) error {
		originWrapper, ok := ev.Origin.Type.(*types.SigVerOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.NodeIds, ev.Valid, ev.Errors, ev.AllOk, context)
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
