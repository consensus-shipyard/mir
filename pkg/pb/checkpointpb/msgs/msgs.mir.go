package checkpointpbmsgs

import (
	types5 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/isspb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types1 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func StableCheckpoint(destModule types.ModuleID, sn types1.SeqNr, snapshot *types2.StateSnapshot, cert map[types.NodeID][]uint8) *types3.Message {
	return &types3.Message{
		DestModule: destModule,
		Type: &types3.Message_Iss{
			Iss: &types4.ISSMessage{
				Type: &types4.ISSMessage_StableCheckpoint{
					StableCheckpoint: &types5.StableCheckpoint{
						Sn:       sn,
						Snapshot: snapshot,
						Cert:     cert,
					},
				},
			},
		},
	}
}

func Checkpoint(destModule types.ModuleID, epoch types1.EpochNr, sn types1.SeqNr, snapshotHash []uint8, signature []uint8) *types3.Message {
	return &types3.Message{
		DestModule: destModule,
		Type: &types3.Message_Checkpoint{
			Checkpoint: &types5.Message{
				Type: &types5.Message_Checkpoint{
					Checkpoint: &types5.Checkpoint{
						Epoch:        epoch,
						Sn:           sn,
						SnapshotHash: snapshotHash,
						Signature:    signature,
					},
				},
			},
		},
	}
}
