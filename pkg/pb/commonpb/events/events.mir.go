package commonpbevents

import (
	types4 "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/types"
	types5 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types1 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func ClientProgress(destModule types.ModuleID, progress map[types1.ClientID]*types2.DeliveredReqs) *types3.Event {
	return &types3.Event{
		DestModule: destModule,
		Type: &types3.Event_BatchFetcher{
			BatchFetcher: &types4.Event{
				Type: &types4.Event_ClientProgress{
					ClientProgress: &types2.ClientProgress{
						Progress: progress,
					},
				},
			},
		},
	}
}

func EpochConfig(destModule types.ModuleID, epochNr types1.EpochNr, firstSn types1.SeqNr, length uint64, memberships []*types2.Membership) *types3.Event {
	return &types3.Event{
		DestModule: destModule,
		Type: &types3.Event_Checkpoint{
			Checkpoint: &types5.Event{
				Type: &types5.Event_EpochConfig{
					EpochConfig: &types2.EpochConfig{
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
