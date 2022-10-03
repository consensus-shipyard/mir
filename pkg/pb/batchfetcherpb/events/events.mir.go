package batchfetcherpbevents

import (
	types2 "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
	types "github.com/filecoin-project/mir/pkg/types"
)

func NewOrderedBatch(destModule types.ModuleID, txs []*requestpb.Request) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_BatchFetcher{
			BatchFetcher: &types2.Event{
				Type: &types2.Event_NewOrderedBatch{
					NewOrderedBatch: &types2.NewOrderedBatch{
						Txs: txs,
					},
				},
			},
		},
	}
}
