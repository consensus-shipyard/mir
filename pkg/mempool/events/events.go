package events

import (
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	mppb "github.com/filecoin-project/mir/pkg/pb/mempoolpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// Event creates an eventpb.Event out of an mempoolpb.Event.
func Event(dest t.ModuleID, ev *mppb.Event) *eventpb.Event {
	return &eventpb.Event{
		DestModule: dest.Pb(),
		Type: &eventpb.Event_Mempool{
			Mempool: ev,
		},
	}
}

// RequestBatch is used by the availability layer to request a new batch of transactions from the mempool.
func RequestBatch(dest t.ModuleID, origin *mppb.RequestBatchOrigin) *eventpb.Event {
	return Event(dest, &mppb.Event{
		Type: &mppb.Event_RequestBatch{
			RequestBatch: &mppb.RequestBatch{
				Origin: origin,
			},
		},
	})
}

// NewBatch is a response to a RequestBatch event.
func NewBatch(dest t.ModuleID, txIDs []t.TxID, txs []*requestpb.Request, origin *mppb.RequestBatchOrigin) *eventpb.Event {
	return Event(dest, &mppb.Event{
		Type: &mppb.Event_NewBatch{
			NewBatch: &mppb.NewBatch{
				TxIds:  t.TxIDSlicePb(txIDs),
				Txs:    txs,
				Origin: origin,
			},
		},
	})
}

// RequestTransactions allows the availability layer to request transactions from the mempool by their IDs.
// It is possible that some of these transactions are not present in the mempool.
func RequestTransactions(dest t.ModuleID, txIDs []t.TxID, origin *mppb.RequestTransactionsOrigin) *eventpb.Event {
	return Event(dest, &mppb.Event{
		Type: &mppb.Event_RequestTransactions{
			RequestTransactions: &mppb.RequestTransactions{
				TxIds:  t.TxIDSlicePb(txIDs),
				Origin: origin,
			},
		},
	})
}

// TransactionsResponse is a response to a RequestTransactions event.
func TransactionsResponse(dest t.ModuleID, present []bool, txs []*requestpb.Request, origin *mppb.RequestTransactionsOrigin) *eventpb.Event {
	return Event(dest, &mppb.Event{
		Type: &mppb.Event_TransactionsResponse{
			TransactionsResponse: &mppb.TransactionsResponse{
				Present: present,
				Txs:     txs,
				Origin:  origin,
			},
		},
	})
}

// RequestTransactionIDs allows other modules to request the mempool module to compute IDs for the given transactions.
// It is possible that some of these transactions are not present in the mempool.
func RequestTransactionIDs(dest t.ModuleID, txs []*requestpb.Request, origin *mppb.RequestTransactionIDsOrigin) *eventpb.Event {
	return Event(dest, &mppb.Event{
		Type: &mppb.Event_RequestTransactionIds{
			RequestTransactionIds: &mppb.RequestTransactionIDs{
				Txs:    txs,
				Origin: origin,
			},
		},
	})
}

// TransactionIDsResponse is a response to a RequestTransactionIDs event.
func TransactionIDsResponse(dest t.ModuleID, txIDs []t.TxID, origin *mppb.RequestTransactionIDsOrigin) *eventpb.Event {
	return Event(dest, &mppb.Event{
		Type: &mppb.Event_TransactionIdsResponse{
			TransactionIdsResponse: &mppb.TransactionIDsResponse{
				TxIds:  t.TxIDSlicePb(txIDs),
				Origin: origin,
			},
		},
	})
}

// RequestBatchID allows other modules to request the mempool module to compute the ID of a batch.
// It is possible that some transactions in the batch are not present in the mempool.
func RequestBatchID(dest t.ModuleID, txIDs []t.TxID, origin *mppb.RequestBatchIDOrigin) *eventpb.Event {
	return Event(dest, &mppb.Event{
		Type: &mppb.Event_RequestBatchId{
			RequestBatchId: &mppb.RequestBatchID{
				TxIds:  t.TxIDSlicePb(txIDs),
				Origin: origin,
			},
		},
	})
}

// BatchIDResponse is a response to a RequestBatchID event.
func BatchIDResponse(dest t.ModuleID, batchID t.BatchID, origin *mppb.RequestBatchIDOrigin) *eventpb.Event {
	return Event(dest, &mppb.Event{
		Type: &mppb.Event_BatchIdResponse{
			BatchIdResponse: &mppb.BatchIDResponse{
				BatchId: batchID.Pb(),
				Origin:  origin,
			},
		},
	})
}
