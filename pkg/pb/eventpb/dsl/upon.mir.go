package eventpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types3 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing events.

func UponInit(m dsl.Module, handler func() error) {
	dsl.UponMirEvent[*types.Event_Init](m, func(ev *types.Init) error {
		return handler()
	})
}

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

func UponDeliverCert(m dsl.Module, handler func(sn types1.SeqNr, cert *types3.Cert) error) {
	dsl.UponMirEvent[*types.Event_DeliverCert](m, func(ev *types.DeliverCert) error {
		return handler(ev.Sn, ev.Cert)
	})
}

func UponAppSnapshotRequest(m dsl.Module, handler func(replyTo types1.ModuleID) error) {
	dsl.UponMirEvent[*types.Event_AppSnapshotRequest](m, func(ev *types.AppSnapshotRequest) error {
		return handler(ev.ReplyTo)
	})
}

func UponAppRestoreState(m dsl.Module, handler func(checkpoint *types4.StableCheckpoint) error) {
	dsl.UponMirEvent[*types.Event_AppRestoreState](m, func(ev *types.AppRestoreState) error {
		return handler(ev.Checkpoint)
	})
}

func UponNewEpoch(m dsl.Module, handler func(epochNr types1.EpochNr) error) {
	dsl.UponMirEvent[*types.Event_NewEpoch](m, func(ev *types.NewEpoch) error {
		return handler(ev.EpochNr)
	})
}
