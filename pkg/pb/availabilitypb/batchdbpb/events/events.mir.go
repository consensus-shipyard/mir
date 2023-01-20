package batchdbpbevents

import (
	types1 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func LookupBatch(destModule types.ModuleID, batchId []uint8, origin *types1.LookupBatchOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_BatchDb{
			BatchDb: &types1.Event{
				Type: &types1.Event_Lookup{
					Lookup: &types1.LookupBatch{
						BatchId: batchId,
						Origin:  origin,
					},
				},
			},
		},
	}
}

func LookupBatchResponse(destModule types.ModuleID, found bool, txs []*types3.Request, origin *types1.LookupBatchOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_BatchDb{
			BatchDb: &types1.Event{
				Type: &types1.Event_LookupResponse{
					LookupResponse: &types1.LookupBatchResponse{
						Found:  found,
						Txs:    txs,
						Origin: origin,
					},
				},
			},
		},
	}
}

func StoreBatch(destModule types.ModuleID, batchId []uint8, txIds [][]uint8, txs []*types3.Request, metadata []uint8, origin *types1.StoreBatchOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_BatchDb{
			BatchDb: &types1.Event{
				Type: &types1.Event_Store{
					Store: &types1.StoreBatch{
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

func BatchStored(destModule types.ModuleID, origin *types1.StoreBatchOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_BatchDb{
			BatchDb: &types1.Event{
				Type: &types1.Event_Stored{
					Stored: &types1.BatchStored{
						Origin: origin,
					},
				},
			},
		},
	}
}
