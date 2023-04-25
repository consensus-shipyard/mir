package batchdbpbevents

import (
	types1 "github.com/filecoin-project/mir/pkg/availability/multisigcollector/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	types5 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func LookupBatch(destModule types.ModuleID, batchId types1.BatchID, origin *types2.LookupBatchOrigin) *types3.Event {
	return &types3.Event{
		DestModule: destModule,
		Type: &types3.Event_BatchDb{
			BatchDb: &types2.Event{
				Type: &types2.Event_Lookup{
					Lookup: &types2.LookupBatch{
						BatchId: batchId,
						Origin:  origin,
					},
				},
			},
		},
	}
}

func LookupBatchResponse(destModule types.ModuleID, found bool, txs []*types4.Transaction, origin *types2.LookupBatchOrigin) *types3.Event {
	return &types3.Event{
		DestModule: destModule,
		Type: &types3.Event_BatchDb{
			BatchDb: &types2.Event{
				Type: &types2.Event_LookupResponse{
					LookupResponse: &types2.LookupBatchResponse{
						Found:  found,
						Txs:    txs,
						Origin: origin,
					},
				},
			},
		},
	}
}

func StoreBatch(destModule types.ModuleID, batchId types1.BatchID, txIds []types5.TxID, txs []*types4.Transaction, metadata []uint8, origin *types2.StoreBatchOrigin) *types3.Event {
	return &types3.Event{
		DestModule: destModule,
		Type: &types3.Event_BatchDb{
			BatchDb: &types2.Event{
				Type: &types2.Event_Store{
					Store: &types2.StoreBatch{
						BatchId:  batchId,
						TxIds:    txIds,
						Txs:      txs,
						Metadata: metadata,
						Origin:   origin,
					},
				},
			},
		},
	}
}

func BatchStored(destModule types.ModuleID, origin *types2.StoreBatchOrigin) *types3.Event {
	return &types3.Event{
		DestModule: destModule,
		Type: &types3.Event_BatchDb{
			BatchDb: &types2.Event{
				Type: &types2.Event_Stored{
					Stored: &types2.BatchStored{
						Origin: origin,
					},
				},
			},
		},
	}
}
