package mempoolpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
)

// Module-specific dsl functions for processing events.

func UponEvent[W types.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponMirEvent[*types1.Event_Mempool](m, func(ev *types.Event) error {
		w, ok := ev.Type.(W)
		if !ok {
			return nil
		}

		return handler(w.Unwrap())
	})
}

func UponRequestBatch(m dsl.Module, handler func(origin *types.RequestBatchOrigin) error) {
	UponEvent[*types.Event_RequestBatch](m, func(ev *types.RequestBatch) error {
		return handler(ev.Origin)
	})
}

func UponNewBatch[C any](m dsl.Module, handler func(txIds [][]uint8, txs []*requestpb.Request, context *C) error) {
	UponEvent[*types.Event_NewBatch](m, func(ev *types.NewBatch) error {
		originWrapper, ok := ev.Origin.Type.(*types.RequestBatchOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.TxIds, ev.Txs, context)
	})
}

func UponRequestTransactions(m dsl.Module, handler func(txIds [][]uint8, origin *types.RequestTransactionsOrigin) error) {
	UponEvent[*types.Event_RequestTransactions](m, func(ev *types.RequestTransactions) error {
		return handler(ev.TxIds, ev.Origin)
	})
}

func UponTransactionsResponse[C any](m dsl.Module, handler func(present []bool, txs []*requestpb.Request, context *C) error) {
	UponEvent[*types.Event_TransactionsResponse](m, func(ev *types.TransactionsResponse) error {
		originWrapper, ok := ev.Origin.Type.(*types.RequestTransactionsOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.Present, ev.Txs, context)
	})
}

func UponRequestTransactionIDs(m dsl.Module, handler func(txs []*requestpb.Request, origin *types.RequestTransactionIDsOrigin) error) {
	UponEvent[*types.Event_RequestTransactionIds](m, func(ev *types.RequestTransactionIDs) error {
		return handler(ev.Txs, ev.Origin)
	})
}

func UponTransactionIDsResponse[C any](m dsl.Module, handler func(txIds [][]uint8, context *C) error) {
	UponEvent[*types.Event_TransactionIdsResponse](m, func(ev *types.TransactionIDsResponse) error {
		originWrapper, ok := ev.Origin.Type.(*types.RequestTransactionIDsOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.TxIds, context)
	})
}

func UponRequestBatchID(m dsl.Module, handler func(txIds [][]uint8, origin *types.RequestBatchIDOrigin) error) {
	UponEvent[*types.Event_RequestBatchId](m, func(ev *types.RequestBatchID) error {
		return handler(ev.TxIds, ev.Origin)
	})
}

func UponBatchIDResponse[C any](m dsl.Module, handler func(batchId []uint8, context *C) error) {
	UponEvent[*types.Event_BatchIdResponse](m, func(ev *types.BatchIDResponse) error {
		originWrapper, ok := ev.Origin.Type.(*types.RequestBatchIDOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.BatchId, context)
	})
}
