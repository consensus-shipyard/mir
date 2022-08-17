package mempoolpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	events "github.com/filecoin-project/mir/pkg/pb/mempoolpb/events"
	types2 "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func RequestBatch(m dsl.Module, next []*types.Event, destModule types1.ModuleID, origin *types2.RequestBatchOrigin) {
	dsl.EmitMirEvent(m, events.RequestBatch(next, destModule, origin))
}

func NewBatch(m dsl.Module, next []*types.Event, destModule types1.ModuleID, txIds [][]uint8, txs []*requestpb.Request, origin *types2.RequestBatchOrigin) {
	dsl.EmitMirEvent(m, events.NewBatch(next, destModule, txIds, txs, origin))
}

func RequestTransactions(m dsl.Module, next []*types.Event, destModule types1.ModuleID, txIds [][]uint8, origin *types2.RequestTransactionsOrigin) {
	dsl.EmitMirEvent(m, events.RequestTransactions(next, destModule, txIds, origin))
}

func TransactionsResponse(m dsl.Module, next []*types.Event, destModule types1.ModuleID, present []bool, txs []*requestpb.Request, origin *types2.RequestTransactionsOrigin) {
	dsl.EmitMirEvent(m, events.TransactionsResponse(next, destModule, present, txs, origin))
}

func RequestTransactionIDs(m dsl.Module, next []*types.Event, destModule types1.ModuleID, txs []*requestpb.Request, origin *types2.RequestTransactionIDsOrigin) {
	dsl.EmitMirEvent(m, events.RequestTransactionIDs(next, destModule, txs, origin))
}

func TransactionIDsResponse(m dsl.Module, next []*types.Event, destModule types1.ModuleID, txIds [][]uint8, origin *types2.RequestTransactionIDsOrigin) {
	dsl.EmitMirEvent(m, events.TransactionIDsResponse(next, destModule, txIds, origin))
}

func RequestBatchID(m dsl.Module, next []*types.Event, destModule types1.ModuleID, txIds [][]uint8, origin *types2.RequestBatchIDOrigin) {
	dsl.EmitMirEvent(m, events.RequestBatchID(next, destModule, txIds, origin))
}

func BatchIDResponse(m dsl.Module, next []*types.Event, destModule types1.ModuleID, batchId []uint8, origin *types2.RequestBatchIDOrigin) {
	dsl.EmitMirEvent(m, events.BatchIDResponse(next, destModule, batchId, origin))
}
