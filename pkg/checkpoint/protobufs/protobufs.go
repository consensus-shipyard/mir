package protobufs

import (
	"github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/factorymodulepb"
	hasherpbtypes "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	messagepbtypes "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
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

func Message(destModule t.ModuleID, message *checkpointpb.Message) *messagepbtypes.Message {
	return &messagepbtypes.Message{
		DestModule: destModule,
		Type:       messagepbtypes.Message_TypeFromPb(&messagepb.Message_Checkpoint{Checkpoint: message}),
	}
}

func CheckpointMessage(
	destModule t.ModuleID,
	epoch t.EpochNr,
	sn t.SeqNr,
	snapshotHash,
	signature []byte,
) *messagepbtypes.Message {
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

func HashOrigin(module t.ModuleID) *hasherpbtypes.HashOrigin {
	return &hasherpbtypes.HashOrigin{
		Module: module,
		Type:   &hasherpbtypes.HashOrigin_Checkpoint{Checkpoint: &checkpointpbtypes.HashOrigin{}},
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
