package commonpbevents

import (
	types3 "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/types"
	types5 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types4 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func ClientProgress(destModule types.ModuleID, progress map[string]*types1.DeliveredReqs) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_BatchFetcher{
			BatchFetcher: &types3.Event{
				Type: &types3.Event_ClientProgress{
					ClientProgress: &types1.ClientProgress{
						Progress: progress,
					},
				},
			},
		},
	}
}

func EpochConfig(destModule types.ModuleID, epochNr types4.EpochNr, firstSn types4.SeqNr, length uint64, memberships []*types1.Membership) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Checkpoint{
			Checkpoint: &types5.Event{
				Type: &types5.Event_EpochConfig{
					EpochConfig: &types1.EpochConfig{
						EpochNr:     epochNr,
						FirstSn:     firstSn,
						Length:      length,
						Memberships: memberships,
					},
				},
			},
		},
	}
}
