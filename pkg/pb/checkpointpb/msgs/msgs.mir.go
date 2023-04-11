package checkpointpbmsgs

import (
	types4 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/isspb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func StableCheckpoint(destModule types.ModuleID, sn types.SeqNr, snapshot *types1.StateSnapshot, cert map[types.NodeID][]uint8) *types2.Message {
	return &types2.Message{
		DestModule: destModule,
		Type: &types2.Message_Iss{
			Iss: &types3.ISSMessage{
				Type: &types3.ISSMessage_StableCheckpoint{
					StableCheckpoint: &types4.StableCheckpoint{
						Sn:       sn,
						Snapshot: snapshot,
						Cert:     cert,
					},
				},
			},
		},
	}
}

func Checkpoint(destModule types.ModuleID, epoch types.EpochNr, sn types.SeqNr, snapshotHash []uint8, signature []uint8) *types2.Message {
	return &types2.Message{
		DestModule: destModule,
		Type: &types2.Message_Checkpoint{
			Checkpoint: &types4.Message{
				Type: &types4.Message_Checkpoint{
					Checkpoint: &types4.Checkpoint{
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
