package mempoolpbevents

import (
	types2 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
	types "github.com/filecoin-project/mir/pkg/types"
)

func RequestBatch(destModule types.ModuleID, origin *types1.RequestBatchOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Mempool{
			Mempool: &types1.Event{
				Type: &types1.Event_RequestBatch{
					RequestBatch: &types1.RequestBatch{
						Origin: origin,
					},
				},
			},
		},
	}
}

func NewBatch(destModule types.ModuleID, txIds [][]uint8, txs []*requestpb.Request, origin *types1.RequestBatchOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Mempool{
			Mempool: &types1.Event{
				Type: &types1.Event_NewBatch{
					NewBatch: &types1.NewBatch{
						TxIds:  txIds,
						Txs:    txs,
						Origin: origin,
					},
				},
			},
		},
	}
}

func RequestTransactions(destModule types.ModuleID, txIds [][]uint8, origin *types1.RequestTransactionsOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Mempool{
			Mempool: &types1.Event{
				Type: &types1.Event_RequestTransactions{
					RequestTransactions: &types1.RequestTransactions{
						TxIds:  txIds,
						Origin: origin,
					},
				},
			},
		},
	}
}

func TransactionsResponse(destModule types.ModuleID, present []bool, txs []*requestpb.Request, origin *types1.RequestTransactionsOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Mempool{
			Mempool: &types1.Event{
				Type: &types1.Event_TransactionsResponse{
					TransactionsResponse: &types1.TransactionsResponse{
						Present: present,
						Txs:     txs,
						Origin:  origin,
					},
				},
			},
		},
	}
}

func RequestTransactionIDs(destModule types.ModuleID, txs []*requestpb.Request, origin *types1.RequestTransactionIDsOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Mempool{
			Mempool: &types1.Event{
				Type: &types1.Event_RequestTransactionIds{
					RequestTransactionIds: &types1.RequestTransactionIDs{
						Txs:    txs,
						Origin: origin,
					},
				},
			},
		},
	}
}

func TransactionIDsResponse(destModule types.ModuleID, txIds [][]uint8, origin *types1.RequestTransactionIDsOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Mempool{
			Mempool: &types1.Event{
				Type: &types1.Event_TransactionIdsResponse{
					TransactionIdsResponse: &types1.TransactionIDsResponse{
						TxIds:  txIds,
						Origin: origin,
					},
				},
			},
		},
	}
}

func RequestBatchID(destModule types.ModuleID, txIds [][]uint8, origin *types1.RequestBatchIDOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Mempool{
			Mempool: &types1.Event{
				Type: &types1.Event_RequestBatchId{
					RequestBatchId: &types1.RequestBatchID{
						TxIds:  txIds,
						Origin: origin,
					},
				},
			},
		},
	}
}

func BatchIDResponse(destModule types.ModuleID, batchId []uint8, origin *types1.RequestBatchIDOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Mempool{
			Mempool: &types1.Event{
				Type: &types1.Event_BatchIdResponse{
					BatchIdResponse: &types1.BatchIDResponse{
						BatchId: batchId,
						Origin:  origin,
					},
				},
			},
		},
	}
}
