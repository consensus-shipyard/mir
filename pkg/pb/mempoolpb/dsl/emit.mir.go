// Code generated by Mir codegen. DO NOT EDIT.

package mempoolpbdsl

import (
	types4 "github.com/filecoin-project/mir/pkg/availability/multisigcollector/types"
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/mempoolpb/events"
	types2 "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	types1 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func RequestBatch[C any](m dsl.Module, destModule types.ModuleID, epoch types1.EpochNr, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types2.RequestBatchOrigin{
		Module: m.ModuleID(),
		Type:   &types2.RequestBatchOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.RequestBatch(destModule, epoch, origin))
}

func NewBatch(m dsl.Module, destModule types.ModuleID, txIds []types1.TxID, txs []*types3.Transaction, origin *types2.RequestBatchOrigin) {
	dsl.EmitMirEvent(m, events.NewBatch(destModule, txIds, txs, origin))
}

func RequestTransactions[C any](m dsl.Module, destModule types.ModuleID, txIds []types1.TxID, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types2.RequestTransactionsOrigin{
		Module: m.ModuleID(),
		Type:   &types2.RequestTransactionsOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.RequestTransactions(destModule, txIds, origin))
}

func TransactionsResponse(m dsl.Module, destModule types.ModuleID, foundIds []types1.TxID, foundTxs []*types3.Transaction, missingIds []types1.TxID, origin *types2.RequestTransactionsOrigin) {
	dsl.EmitMirEvent(m, events.TransactionsResponse(destModule, foundIds, foundTxs, missingIds, origin))
}

func RequestTransactionIDs[C any](m dsl.Module, destModule types.ModuleID, txs []*types3.Transaction, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types2.RequestTransactionIDsOrigin{
		Module: m.ModuleID(),
		Type:   &types2.RequestTransactionIDsOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.RequestTransactionIDs(destModule, txs, origin))
}

func TransactionIDsResponse(m dsl.Module, destModule types.ModuleID, txIds []types1.TxID, origin *types2.RequestTransactionIDsOrigin) {
	dsl.EmitMirEvent(m, events.TransactionIDsResponse(destModule, txIds, origin))
}

func RequestBatchID[C any](m dsl.Module, destModule types.ModuleID, txIds []types1.TxID, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types2.RequestBatchIDOrigin{
		Module: m.ModuleID(),
		Type:   &types2.RequestBatchIDOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.RequestBatchID(destModule, txIds, origin))
}

func BatchIDResponse(m dsl.Module, destModule types.ModuleID, batchId types4.BatchID, origin *types2.RequestBatchIDOrigin) {
	dsl.EmitMirEvent(m, events.BatchIDResponse(destModule, batchId, origin))
}

func NewTransactions(m dsl.Module, destModule types.ModuleID, transactions []*types3.Transaction) {
	dsl.EmitMirEvent(m, events.NewTransactions(destModule, transactions))
}

func BatchTimeout(m dsl.Module, destModule types.ModuleID, batchReqID uint64) {
	dsl.EmitMirEvent(m, events.BatchTimeout(destModule, batchReqID))
}

func NewEpoch(m dsl.Module, destModule types.ModuleID, epochNr types1.EpochNr, clientProgress *types3.ClientProgress) {
	dsl.EmitMirEvent(m, events.NewEpoch(destModule, epochNr, clientProgress))
}
