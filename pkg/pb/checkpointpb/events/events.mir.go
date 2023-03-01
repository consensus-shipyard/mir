package checkpointpbevents

import (
	types2 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	commonpb "github.com/filecoin-project/mir/pkg/pb/commonpb"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func StableCheckpoint(destModule types.ModuleID, sn types.SeqNr, snapshot *commonpb.StateSnapshot, cert map[string][]uint8) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Checkpoint{
			Checkpoint: &types2.Event{
				Type: &types2.Event_StableCheckpoint{
					StableCheckpoint: &types2.StableCheckpoint{
						Sn:       sn,
						Snapshot: snapshot,
						Cert:     cert,
					},
				},
			},
		},
	}
}
