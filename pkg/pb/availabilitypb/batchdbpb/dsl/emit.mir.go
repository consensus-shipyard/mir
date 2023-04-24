package batchdbpbdsl

import (
	types1 "github.com/filecoin-project/mir/pkg/availability/multisigcollector/types"
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/events"
	types2 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	types4 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func LookupBatch[C any](m dsl.Module, destModule types.ModuleID, batchId types1.BatchID, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types2.LookupBatchOrigin{
		Module: m.ModuleID(),
		Type:   &types2.LookupBatchOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.LookupBatch(destModule, batchId, origin))
}

func LookupBatchResponse(m dsl.Module, destModule types.ModuleID, found bool, txs []*types3.Request, origin *types2.LookupBatchOrigin) {
	dsl.EmitMirEvent(m, events.LookupBatchResponse(destModule, found, txs, origin))
}

func StoreBatch[C any](m dsl.Module, destModule types.ModuleID, batchId types1.BatchID, txIds []types4.TxID, txs []*types3.Request, metadata []uint8, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types2.StoreBatchOrigin{
		Module: m.ModuleID(),
		Type:   &types2.StoreBatchOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.StoreBatch(destModule, batchId, txIds, txs, metadata, origin))
}

func BatchStored(m dsl.Module, destModule types.ModuleID, origin *types2.StoreBatchOrigin) {
	dsl.EmitMirEvent(m, events.BatchStored(destModule, origin))
}
