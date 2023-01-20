package batchdbpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func LookupBatch[C any](m dsl.Module, destModule types.ModuleID, batchId []uint8, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types1.LookupBatchOrigin{
		Module: m.ModuleID(),
		Type:   &types1.LookupBatchOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.LookupBatch(destModule, batchId, origin))
}

func LookupBatchResponse(m dsl.Module, destModule types.ModuleID, found bool, txs []*types2.Request, metadata []uint8, origin *types1.LookupBatchOrigin) {
	dsl.EmitMirEvent(m, events.LookupBatchResponse(destModule, found, txs, metadata, origin))
}

func StoreBatch[C any](m dsl.Module, destModule types.ModuleID, batchId []uint8, txIds [][]uint8, txs []*types2.Request, metadata []uint8, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types1.StoreBatchOrigin{
		Module: m.ModuleID(),
		Type:   &types1.StoreBatchOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.StoreBatch(destModule, batchId, txIds, txs, metadata, origin))
}

func BatchStored(m dsl.Module, destModule types.ModuleID, origin *types1.StoreBatchOrigin) {
	dsl.EmitMirEvent(m, events.BatchStored(destModule, origin))
}
