package mempoolpbdsl

import (
	types4 "github.com/filecoin-project/mir/pkg/availability/multisigcollector/types"
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/mempoolpb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	types2 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func RequestBatch[C any](m dsl.Module, destModule types.ModuleID, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types1.RequestBatchOrigin{
		Module: m.ModuleID(),
		Type:   &types1.RequestBatchOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.RequestBatch(destModule, origin))
}

func NewBatch(m dsl.Module, destModule types.ModuleID, txIds []types2.TxID, txs []*types3.Request, origin *types1.RequestBatchOrigin) {
	dsl.EmitMirEvent(m, events.NewBatch(destModule, txIds, txs, origin))
}

func RequestTransactions[C any](m dsl.Module, destModule types.ModuleID, txIds []types2.TxID, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types1.RequestTransactionsOrigin{
		Module: m.ModuleID(),
		Type:   &types1.RequestTransactionsOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.RequestTransactions(destModule, txIds, origin))
}

func TransactionsResponse(m dsl.Module, destModule types.ModuleID, present []bool, txs []*types3.Request, origin *types1.RequestTransactionsOrigin) {
	dsl.EmitMirEvent(m, events.TransactionsResponse(destModule, present, txs, origin))
}

func RequestTransactionIDs[C any](m dsl.Module, destModule types.ModuleID, txs []*types3.Request, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types1.RequestTransactionIDsOrigin{
		Module: m.ModuleID(),
		Type:   &types1.RequestTransactionIDsOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.RequestTransactionIDs(destModule, txs, origin))
}

func TransactionIDsResponse(m dsl.Module, destModule types.ModuleID, txIds []types2.TxID, origin *types1.RequestTransactionIDsOrigin) {
	dsl.EmitMirEvent(m, events.TransactionIDsResponse(destModule, txIds, origin))
}

func RequestBatchID[C any](m dsl.Module, destModule types.ModuleID, txIds []types2.TxID, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types1.RequestBatchIDOrigin{
		Module: m.ModuleID(),
		Type:   &types1.RequestBatchIDOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.RequestBatchID(destModule, txIds, origin))
}

func BatchIDResponse(m dsl.Module, destModule types.ModuleID, batchId types4.BatchID, origin *types1.RequestBatchIDOrigin) {
	dsl.EmitMirEvent(m, events.BatchIDResponse(destModule, batchId, origin))
}

func NewRequests(m dsl.Module, destModule types.ModuleID, requests []*types3.Request) {
	dsl.EmitMirEvent(m, events.NewRequests(destModule, requests))
}
