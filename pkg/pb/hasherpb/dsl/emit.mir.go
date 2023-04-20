package hasherpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types1 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	events "github.com/filecoin-project/mir/pkg/pb/hasherpb/events"
	types2 "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func Request[C any](m dsl.Module, destModule types.ModuleID, data []*types1.HashData, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types2.HashOrigin{
		Module: m.ModuleID(),
		Type:   &types2.HashOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.Request(destModule, data, origin))
}

func Result(m dsl.Module, destModule types.ModuleID, digests [][]uint8, origin *types2.HashOrigin) {
	dsl.EmitMirEvent(m, events.Result(destModule, digests, origin))
}

func RequestOne[C any](m dsl.Module, destModule types.ModuleID, data *types1.HashData, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types2.HashOrigin{
		Module: m.ModuleID(),
		Type:   &types2.HashOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.RequestOne(destModule, data, origin))
}

func ResultOne(m dsl.Module, destModule types.ModuleID, digest []uint8, origin *types2.HashOrigin) {
	dsl.EmitMirEvent(m, events.ResultOne(destModule, digest, origin))
}
