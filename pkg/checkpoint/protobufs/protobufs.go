package protobufs

import (
	"github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	t "github.com/filecoin-project/mir/pkg/types"
)

func Event(destModule t.ModuleID, event *checkpointpb.Event) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_Checkpoint{
			Checkpoint: event,
		},
	}
}

func StableCheckpointEvent(ownModuleID t.ModuleID, stableCheckpoint *checkpointpb.StableCheckpoint) *eventpb.Event {
	return Event(
		ownModuleID,
		&checkpointpb.Event{Type: &checkpointpb.Event_StableCheckpoint{
			StableCheckpoint: stableCheckpoint,
		}},
	)
}

func Message(destModule t.ModuleID, message *checkpointpb.Message) *messagepb.Message {
	return &messagepb.Message{
		DestModule: destModule.Pb(),
		Type:       &messagepb.Message_Checkpoint{Checkpoint: message},
	}
}

func CheckpointMessage(
	destModule t.ModuleID,
	epoch t.EpochNr,
	sn t.SeqNr,
	snapshotHash,
	signature []byte,
) *messagepb.Message {
	return Message(
		destModule,
		&checkpointpb.Message{Type: &checkpointpb.Message_Checkpoint{
			Checkpoint: &checkpointpb.Checkpoint{
				Epoch:        epoch.Pb(),
				Sn:           sn.Pb(),
				SnapshotHash: snapshotHash,
				Signature:    signature,
			},
		}},
	)
}
