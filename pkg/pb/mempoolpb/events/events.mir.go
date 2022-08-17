package mempoolpbevents

import (
	types "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

func RequestBatch(next []*types.Event, destModule types1.ModuleID, origin *types2.RequestBatchOrigin) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_Mempool{
			Mempool: &types2.Event{
				Type: &types2.Event_RequestBatch{
					RequestBatch: &types2.RequestBatch{
						Origin: origin,
					},
				},
			},
		},
	}
}

func NewBatch(next []*types.Event, destModule types1.ModuleID, txIds [][]uint8, txs []*requestpb.Request, origin *types2.RequestBatchOrigin) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_Mempool{
			Mempool: &types2.Event{
				Type: &types2.Event_NewBatch{
					NewBatch: &types2.NewBatch{
						TxIds:  txIds,
						Txs:    txs,
						Origin: origin,
					},
				},
			},
		},
	}
}

func RequestTransactions(next []*types.Event, destModule types1.ModuleID, txIds [][]uint8, origin *types2.RequestTransactionsOrigin) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_Mempool{
			Mempool: &types2.Event{
				Type: &types2.Event_RequestTransactions{
					RequestTransactions: &types2.RequestTransactions{
						TxIds:  txIds,
						Origin: origin,
					},
				},
			},
		},
	}
}

func TransactionsResponse(next []*types.Event, destModule types1.ModuleID, present []bool, txs []*requestpb.Request, origin *types2.RequestTransactionsOrigin) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_Mempool{
			Mempool: &types2.Event{
				Type: &types2.Event_TransactionsResponse{
					TransactionsResponse: &types2.TransactionsResponse{
						Present: present,
						Txs:     txs,
						Origin:  origin,
					},
				},
			},
		},
	}
}

func RequestTransactionIDs(next []*types.Event, destModule types1.ModuleID, txs []*requestpb.Request, origin *types2.RequestTransactionIDsOrigin) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_Mempool{
			Mempool: &types2.Event{
				Type: &types2.Event_RequestTransactionIds{
					RequestTransactionIds: &types2.RequestTransactionIDs{
						Txs:    txs,
						Origin: origin,
					},
				},
			},
		},
	}
}

func TransactionIDsResponse(next []*types.Event, destModule types1.ModuleID, txIds [][]uint8, origin *types2.RequestTransactionIDsOrigin) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_Mempool{
			Mempool: &types2.Event{
				Type: &types2.Event_TransactionIdsResponse{
					TransactionIdsResponse: &types2.TransactionIDsResponse{
						TxIds:  txIds,
						Origin: origin,
					},
				},
			},
		},
	}
}

func RequestBatchID(next []*types.Event, destModule types1.ModuleID, txIds [][]uint8, origin *types2.RequestBatchIDOrigin) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_Mempool{
			Mempool: &types2.Event{
				Type: &types2.Event_RequestBatchId{
					RequestBatchId: &types2.RequestBatchID{
						TxIds:  txIds,
						Origin: origin,
					},
				},
			},
		},
	}
}

func BatchIDResponse(next []*types.Event, destModule types1.ModuleID, batchId []uint8, origin *types2.RequestBatchIDOrigin) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_Mempool{
			Mempool: &types2.Event{
				Type: &types2.Event_BatchIdResponse{
					BatchIdResponse: &types2.BatchIDResponse{
						BatchId: batchId,
						Origin:  origin,
					},
				},
			},
		},
	}
}
