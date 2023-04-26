package protobufs

import (
	"github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	factorypbtypes "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	hasherpbtypes "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	timertypes "github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
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
	epochNr tt.EpochNr,
) *eventpb.Event {
	return Event(
		destModule,
		&checkpointpb.Event{Type: &checkpointpb.Event_EpochProgress{EpochProgress: &checkpointpb.EpochProgress{
			NodeId: nodeID.Pb(),
			Epoch:  epochNr.Pb(),
		}}},
	)
}

func HashOrigin(module t.ModuleID) *hasherpbtypes.HashOrigin {
	return &hasherpbtypes.HashOrigin{
		Module: module,
		Type:   &hasherpbtypes.HashOrigin_Checkpoint{Checkpoint: &checkpointpbtypes.HashOrigin{}},
	}
}

func SignOrigin(module t.ModuleID) *cryptopbtypes.SignOrigin {
	return &cryptopbtypes.SignOrigin{
		Module: module,
		Type:   &cryptopbtypes.SignOrigin_Checkpoint{Checkpoint: &checkpointpbtypes.SignOrigin{}},
	}
}

func SigVerOrigin(module t.ModuleID) *cryptopbtypes.SigVerOrigin {
	return &cryptopbtypes.SigVerOrigin{
		Module: module,
		Type:   &cryptopbtypes.SigVerOrigin_Checkpoint{Checkpoint: &checkpointpbtypes.SigVerOrigin{}},
	}
}

func InstanceParams(
	membership *trantorpbtypes.Membership,
	resendPeriod timertypes.Duration,
	leaderPolicyData []byte,
	epochConfig *trantorpbtypes.EpochConfig,
) *factorypbtypes.GeneratorParams {
	return &factorypbtypes.GeneratorParams{Type: &factorypbtypes.GeneratorParams_Checkpoint{Checkpoint: &checkpointpbtypes.InstanceParams{
		Membership:       membership,
		ResendPeriod:     resendPeriod,
		LeaderPolicyData: leaderPolicyData,
		EpochConfig:      epochConfig,
	}}}
}
