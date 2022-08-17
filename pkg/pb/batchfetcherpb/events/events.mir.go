package batchfetcherpbevents

import (
	types2 "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

func NewOrderedBatch(next []*types.Event, destModule types1.ModuleID, txs []*requestpb.Request) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_BatchFetcher{
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
