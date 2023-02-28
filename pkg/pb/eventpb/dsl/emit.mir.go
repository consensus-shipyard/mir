package eventpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types3 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	events "github.com/filecoin-project/mir/pkg/pb/eventpb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func Init(m dsl.Module, destModule types.ModuleID) {
	dsl.EmitMirEvent(m, events.Init(destModule))
}

func SignRequest[C any](m dsl.Module, destModule types.ModuleID, data [][]uint8, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types1.SignOrigin{
		Module: m.ModuleID(),
		Type:   &types1.SignOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.SignRequest(destModule, data, origin))
}

func SignResult(m dsl.Module, destModule types.ModuleID, signature []uint8, origin *types1.SignOrigin) {
	dsl.EmitMirEvent(m, events.SignResult(destModule, signature, origin))
}

func VerifyNodeSigs[C any](m dsl.Module, destModule types.ModuleID, data []*types1.SigVerData, signatures [][]uint8, nodeIds []types.NodeID, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types1.SigVerOrigin{
		Module: m.ModuleID(),
		Type:   &types1.SigVerOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.VerifyNodeSigs(destModule, data, signatures, origin, nodeIds))
}

func NodeSigsVerified(m dsl.Module, destModule types.ModuleID, origin *types1.SigVerOrigin, nodeIds []types.NodeID, valid []bool, errors []error, allOk bool) {
	dsl.EmitMirEvent(m, events.NodeSigsVerified(destModule, origin, nodeIds, valid, errors, allOk))
}

func SendMessage(m dsl.Module, destModule types.ModuleID, msg *types2.Message, destinations []types.NodeID) {
	dsl.EmitMirEvent(m, events.SendMessage(destModule, msg, destinations))
}

func MessageReceived(m dsl.Module, destModule types.ModuleID, from types.NodeID, msg *types2.Message) {
	dsl.EmitMirEvent(m, events.MessageReceived(destModule, from, msg))
}

func DeliverCert(m dsl.Module, destModule types.ModuleID, sn types.SeqNr, cert *types3.Cert) {
	dsl.EmitMirEvent(m, events.DeliverCert(destModule, sn, cert))
}

func AppSnapshotRequest(m dsl.Module, destModule types.ModuleID, replyTo types.ModuleID) {
	dsl.EmitMirEvent(m, events.AppSnapshotRequest(destModule, replyTo))
}

func AppRestoreState(m dsl.Module, destModule types.ModuleID, checkpoint *types4.StableCheckpoint) {
	dsl.EmitMirEvent(m, events.AppRestoreState(destModule, checkpoint))
}

func NewEpoch(m dsl.Module, destModule types.ModuleID, epochNr types.EpochNr) {
	dsl.EmitMirEvent(m, events.NewEpoch(destModule, epochNr))
}
