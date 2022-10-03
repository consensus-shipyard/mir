package mempoolpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/mempoolpb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func RequestBatch(m dsl.Module, destModule types.ModuleID, origin *types1.RequestBatchOrigin) {
	dsl.EmitMirEvent(m, events.RequestBatch(destModule, origin))
}

func NewBatch(m dsl.Module, destModule types.ModuleID, txIds [][]uint8, txs []*requestpb.Request, origin *types1.RequestBatchOrigin) {
	dsl.EmitMirEvent(m, events.NewBatch(destModule, txIds, txs, origin))
}

func RequestTransactions(m dsl.Module, destModule types.ModuleID, txIds [][]uint8, origin *types1.RequestTransactionsOrigin) {
	dsl.EmitMirEvent(m, events.RequestTransactions(destModule, txIds, origin))
}

func TransactionsResponse(m dsl.Module, destModule types.ModuleID, present []bool, txs []*requestpb.Request, origin *types1.RequestTransactionsOrigin) {
	dsl.EmitMirEvent(m, events.TransactionsResponse(destModule, present, txs, origin))
}

func RequestTransactionIDs(m dsl.Module, destModule types.ModuleID, txs []*requestpb.Request, origin *types1.RequestTransactionIDsOrigin) {
	dsl.EmitMirEvent(m, events.RequestTransactionIDs(destModule, txs, origin))
}

func TransactionIDsResponse(m dsl.Module, destModule types.ModuleID, txIds [][]uint8, origin *types1.RequestTransactionIDsOrigin) {
	dsl.EmitMirEvent(m, events.TransactionIDsResponse(destModule, txIds, origin))
}

func RequestBatchID(m dsl.Module, destModule types.ModuleID, txIds [][]uint8, origin *types1.RequestBatchIDOrigin) {
	dsl.EmitMirEvent(m, events.RequestBatchID(destModule, txIds, origin))
}

func BatchIDResponse(m dsl.Module, destModule types.ModuleID, batchId []uint8, origin *types1.RequestBatchIDOrigin) {
	dsl.EmitMirEvent(m, events.BatchIDResponse(destModule, batchId, origin))
}
