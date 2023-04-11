package commonpbevents

import (
	types2 "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	commonpb "github.com/filecoin-project/mir/pkg/pb/commonpb"
	types3 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func ClientProgress(destModule types.ModuleID, progress map[string]*commonpb.DeliveredReqs) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_BatchFetcher{
			BatchFetcher: &types2.Event{
				Type: &types2.Event_ClientProgress{
					ClientProgress: &types3.ClientProgress{
						Progress: progress,
					},
				},
			},
		},
	}
}

func EpochConfig(destModule types.ModuleID, epochNr uint64, firstSn uint64, length uint64, memberships []*types3.Membership) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Checkpoint{
			Checkpoint: &types4.Event{
				Type: &types4.Event_EpochConfig{
					EpochConfig: &types3.EpochConfig{
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
