package protobufs

import (
	"github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/factorymodulepb"
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

func EpochProgressEvent(
	destModule t.ModuleID,
	nodeID t.NodeID,
	epochNr t.EpochNr,
) *eventpb.Event {
	return Event(
		destModule,
		&checkpointpb.Event{Type: &checkpointpb.Event_EpochProgress{EpochProgress: &checkpointpb.EpochProgress{
			NodeId: nodeID.Pb(),
			Epoch:  epochNr.Pb(),
		}}},
	)
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

func HashOrigin(module t.ModuleID) *eventpb.HashOrigin {
	return &eventpb.HashOrigin{
		Module: module.Pb(),
		Type:   &eventpb.HashOrigin_Checkpoint{Checkpoint: &checkpointpb.HashOrigin{}},
	}
}

func SignOrigin(module t.ModuleID) *eventpb.SignOrigin {
	return &eventpb.SignOrigin{
		Module: module.Pb(),
		Type:   &eventpb.SignOrigin_Checkpoint{Checkpoint: &checkpointpb.SignOrigin{}},
	}
}

func SigVerOrigin(module t.ModuleID) *eventpb.SigVerOrigin {
	return &eventpb.SigVerOrigin{
		Module: module.Pb(),
		Type:   &eventpb.SigVerOrigin_Checkpoint{Checkpoint: &checkpointpb.SigVerOrigin{}},
	}
}

func InstanceParams(
	membership map[t.NodeID]t.NodeAddress,
	resendPeriod t.TimeDuration,
	leaderPolicyData []byte,
	epochConfig *commonpb.EpochConfig,
) *factorymodulepb.GeneratorParams {
	return &factorymodulepb.GeneratorParams{Type: &factorymodulepb.GeneratorParams_Checkpoint{
		Checkpoint: &checkpointpb.InstanceParams{
			Membership:       t.MembershipPb(membership),
			ResendPeriod:     resendPeriod.Pb(),
			LeaderPolicyData: leaderPolicyData,
			EpochConfig:      epochConfig,
		},
	}}
}
