package eventpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types5 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types6 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	events "github.com/filecoin-project/mir/pkg/pb/eventpb/events"
	types3 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func Init(m dsl.Module, destModule types.ModuleID) {
	dsl.EmitMirEvent(m, events.Init(destModule))
}

func NewRequests(m dsl.Module, destModule types.ModuleID, requests []*types1.Request) {
	dsl.EmitMirEvent(m, events.NewRequests(destModule, requests))
}

func HashRequest[C any](m dsl.Module, destModule types.ModuleID, data []*types2.HashData, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types3.HashOrigin{
		Module: m.ModuleID(),
		Type:   &types3.HashOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.HashRequest(destModule, data, origin))
}

func HashResult(m dsl.Module, destModule types.ModuleID, digests [][]uint8, origin *types3.HashOrigin) {
	dsl.EmitMirEvent(m, events.HashResult(destModule, digests, origin))
}

func SignRequest[C any](m dsl.Module, destModule types.ModuleID, data [][]uint8, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types3.SignOrigin{
		Module: m.ModuleID(),
		Type:   &types3.SignOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.SignRequest(destModule, data, origin))
}

func SignResult(m dsl.Module, destModule types.ModuleID, signature []uint8, origin *types3.SignOrigin) {
	dsl.EmitMirEvent(m, events.SignResult(destModule, signature, origin))
}

func VerifyNodeSigs[C any](m dsl.Module, destModule types.ModuleID, data []*types3.SigVerData, signatures [][]uint8, nodeIds []types.NodeID, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types3.SigVerOrigin{
		Module: m.ModuleID(),
		Type:   &types3.SigVerOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.VerifyNodeSigs(destModule, data, signatures, origin, nodeIds))
}

func NodeSigsVerified(m dsl.Module, destModule types.ModuleID, origin *types3.SigVerOrigin, nodeIds []types.NodeID, valid []bool, errors []error, allOk bool) {
	dsl.EmitMirEvent(m, events.NodeSigsVerified(destModule, origin, nodeIds, valid, errors, allOk))
}

func SendMessage(m dsl.Module, destModule types.ModuleID, msg *types4.Message, destinations []types.NodeID) {
	dsl.EmitMirEvent(m, events.SendMessage(destModule, msg, destinations))
}

func MessageReceived(m dsl.Module, destModule types.ModuleID, from types.NodeID, msg *types4.Message) {
	dsl.EmitMirEvent(m, events.MessageReceived(destModule, from, msg))
}

func DeliverCert(m dsl.Module, destModule types.ModuleID, sn types.SeqNr, cert *types5.Cert) {
	dsl.EmitMirEvent(m, events.DeliverCert(destModule, sn, cert))
}

func AppSnapshotRequest(m dsl.Module, destModule types.ModuleID, replyTo types.ModuleID) {
	dsl.EmitMirEvent(m, events.AppSnapshotRequest(destModule, replyTo))
}

func AppRestoreState(m dsl.Module, destModule types.ModuleID, checkpoint *types6.StableCheckpoint) {
	dsl.EmitMirEvent(m, events.AppRestoreState(destModule, checkpoint))
}

func NewEpoch(m dsl.Module, destModule types.ModuleID, epochNr types.EpochNr) {
	dsl.EmitMirEvent(m, events.NewEpoch(destModule, epochNr))
}
