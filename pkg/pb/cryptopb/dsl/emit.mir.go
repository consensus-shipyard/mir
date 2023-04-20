package cryptopbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/cryptopb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

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

func VerifySig[C any](m dsl.Module, destModule types.ModuleID, data *types1.SignedData, signature []uint8, nodeId types.NodeID, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types1.SigVerOrigin{
		Module: m.ModuleID(),
		Type:   &types1.SigVerOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.VerifySig(destModule, data, signature, origin, nodeId))
}

func SigVerified(m dsl.Module, destModule types.ModuleID, origin *types1.SigVerOrigin, nodeId types.NodeID, error error) {
	dsl.EmitMirEvent(m, events.SigVerified(destModule, origin, nodeId, error))
}

func VerifySigs[C any](m dsl.Module, destModule types.ModuleID, data []*types1.SignedData, signatures [][]uint8, nodeIds []types.NodeID, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types1.SigVerOrigin{
		Module: m.ModuleID(),
		Type:   &types1.SigVerOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.VerifySigs(destModule, data, signatures, origin, nodeIds))
}

func SigsVerified(m dsl.Module, destModule types.ModuleID, origin *types1.SigVerOrigin, nodeIds []types.NodeID, errors []error, allOk bool) {
	dsl.EmitMirEvent(m, events.SigsVerified(destModule, origin, nodeIds, errors, allOk))
}
