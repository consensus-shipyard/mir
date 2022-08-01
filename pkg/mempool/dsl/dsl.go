package dsl

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	mpevents "github.com/filecoin-project/mir/pkg/mempool/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	mppb "github.com/filecoin-project/mir/pkg/pb/mempoolpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

// RequestBatch is used by the availability layer to request a new batch of transactions from the mempool.
func RequestBatch[C any](m dsl.Module, dest t.ModuleID, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &mppb.RequestBatchOrigin{
		Module: m.ModuleID().Pb(),
		Type: &mppb.RequestBatchOrigin_Dsl{
			Dsl: dsl.Origin(contextID),
		},
	}

	dsl.EmitEvent(m, mpevents.RequestBatch(dest, origin))
}

// NewBatch is a response to a RequestBatch event.
func NewBatch(m dsl.Module, dest t.ModuleID, txIDs []t.TxID, txs []*requestpb.Request, origin *mppb.RequestBatchOrigin) {
	dsl.EmitEvent(m, mpevents.NewBatch(dest, txIDs, txs, origin))
}

// RequestTransactions allows the availability layer to request transactions from the mempool by their IDs.
// It is possible that some of these transactions are not present in the mempool.
func RequestTransactions[C any](m dsl.Module, dest t.ModuleID, txIDs []t.TxID, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &mppb.RequestTransactionsOrigin{
		Module: m.ModuleID().Pb(),
		Type: &mppb.RequestTransactionsOrigin_Dsl{
			Dsl: dsl.Origin(contextID),
		},
	}

	dsl.EmitEvent(m, mpevents.RequestTransactions(dest, txIDs, origin))
}

// TransactionsResponse is a response to a RequestTransactions event.
func TransactionsResponse(m dsl.Module, dest t.ModuleID, present []bool, txs []*requestpb.Request, origin *mppb.RequestTransactionsOrigin) {
	dsl.EmitEvent(m, mpevents.TransactionsResponse(dest, present, txs, origin))
}

// RequestTransactionIDs allows other modules to request the mempool module to compute IDs for the given transactions.
// It is possible that some of these transactions are not present in the mempool.
func RequestTransactionIDs[C any](m dsl.Module, dest t.ModuleID, txs []*requestpb.Request, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &mppb.RequestTransactionIDsOrigin{
		Module: m.ModuleID().Pb(),
		Type: &mppb.RequestTransactionIDsOrigin_Dsl{
			Dsl: dsl.Origin(contextID),
		},
	}

	dsl.EmitEvent(m, mpevents.RequestTransactionIDs(dest, txs, origin))
}

// TransactionIDsResponse is a response to a RequestTransactionIDs event.
func TransactionIDsResponse(m dsl.Module, dest t.ModuleID, txIDs []t.TxID, origin *mppb.RequestTransactionIDsOrigin) {
	dsl.EmitEvent(m, mpevents.TransactionIDsResponse(dest, txIDs, origin))
}

// RequestBatchID allows other modules to request the mempool module to compute the ID of a batch.
// It is possible that some transactions in the batch are not present in the mempool.
func RequestBatchID[C any](m dsl.Module, dest t.ModuleID, txIDs []t.TxID, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &mppb.RequestBatchIDOrigin{
		Module: m.ModuleID().Pb(),
		Type: &mppb.RequestBatchIDOrigin_Dsl{
			Dsl: dsl.Origin(contextID),
		},
	}

	dsl.EmitEvent(m, mpevents.RequestBatchID(dest, txIDs, origin))
}

// BatchIDResponse is a response to a RequestBatchID event.
func BatchIDResponse(m dsl.Module, dest t.ModuleID, batchID t.BatchID, origin *mppb.RequestBatchIDOrigin) {
	dsl.EmitEvent(m, mpevents.BatchIDResponse(dest, batchID, origin))
}

// Module-specific dsl functions for processing events.

// UponEvent registers a handler for the given mempool event type.
func UponEvent[EvWrapper mppb.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponEvent[*eventpb.Event_Mempool](m, func(ev *mppb.Event) error {
		evWrapper, ok := ev.Type.(EvWrapper)
		if !ok {
			return nil
		}
		return handler(evWrapper.Unwrap())
	})
}

// UponRequestBatch registers a handler for the RequestBatch events.
func UponRequestBatch(m dsl.Module, handler func(origin *mppb.RequestBatchOrigin) error) {
	UponEvent[*mppb.Event_RequestBatch](m, func(ev *mppb.RequestBatch) error {
		return handler(ev.Origin)
	})
}

// UponNewBatch registers a handler for the NewBatch events.
func UponNewBatch[C any](m dsl.Module, handler func(txIDs []t.TxID, txs []*requestpb.Request, context *C) error) {
	UponEvent[*mppb.Event_NewBatch](m, func(ev *mppb.NewBatch) error {
		originWrapper, ok := ev.Origin.Type.(*mppb.RequestBatchOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(t.TxIDSlice(ev.TxIds), ev.Txs, context)
	})
}

// UponRequestTransactions registers a handler for the RequestTransactions events.
func UponRequestTransactions(m dsl.Module, handler func(txIDs []t.TxID, origin *mppb.RequestTransactionsOrigin) error) {
	UponEvent[*mppb.Event_RequestTransactions](m, func(ev *mppb.RequestTransactions) error {
		return handler(t.TxIDSlice(ev.TxIds), ev.Origin)
	})
}

// UponTransactionsResponse registers a handler for the TransactionsResponse events.
func UponTransactionsResponse[C any](m dsl.Module, handler func(present []bool, txs []*requestpb.Request, context *C) error) {
	UponEvent[*mppb.Event_TransactionsResponse](m, func(ev *mppb.TransactionsResponse) error {
		originWrapper, ok := ev.Origin.Type.(*mppb.RequestTransactionsOrigin_Dsl)
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

// UponRequestTransactionIDs registers a handler for the RequestTransactionIDs events.
func UponRequestTransactionIDs(m dsl.Module, handler func(txs []*requestpb.Request, origin *mppb.RequestTransactionIDsOrigin) error) {
	UponEvent[*mppb.Event_RequestTransactionIds](m, func(ev *mppb.RequestTransactionIDs) error {
		return handler(ev.Txs, ev.Origin)
	})
}

// UponTransactionIDsResponse registers a handler for the TransactionIDsResponse events.
func UponTransactionIDsResponse[C any](m dsl.Module, handler func(txIDs []t.TxID, context *C) error) {
	UponEvent[*mppb.Event_TransactionIdsResponse](m, func(ev *mppb.TransactionIDsResponse) error {
		originWrapper, ok := ev.Origin.Type.(*mppb.RequestTransactionIDsOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(t.TxIDSlice(ev.TxIds), context)
	})
}

// UponRequestBatchID registers a handler for the RequestBatchID events.
func UponRequestBatchID(m dsl.Module, handler func(txIDs []t.TxID, origin *mppb.RequestBatchIDOrigin) error) {
	UponEvent[*mppb.Event_RequestBatchId](m, func(ev *mppb.RequestBatchID) error {
		return handler(t.TxIDSlice(ev.TxIds), ev.Origin)
	})
}

// UponBatchIDResponse registers a handler for the BatchIDResponse events.
func UponBatchIDResponse[C any](m dsl.Module, handler func(batchID t.BatchID, context *C) error) {
	UponEvent[*mppb.Event_BatchIdResponse](m, func(ev *mppb.BatchIDResponse) error {
		originWrapper, ok := ev.Origin.Type.(*mppb.RequestBatchIDOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(t.BatchID(ev.BatchId), context)
	})
}
