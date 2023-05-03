package orderers

import (
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	factorypbtypes "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	"github.com/filecoin-project/mir/pkg/pb/ordererpb"
	ordererpbtypes "github.com/filecoin-project/mir/pkg/pb/ordererpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

func OrdererEvent(
	destModule t.ModuleID,
	event *ordererpb.Event,
) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_Orderer{
			Orderer: event,
		},
	}
}

func InstanceParams(
	segment *Segment,
	availabilityID t.ModuleID,
	epoch tt.EpochNr,
	validityCheckerType ValidityCheckerType,
) *factorypbtypes.GeneratorParams {
	return &factorypbtypes.GeneratorParams{Type: &factorypbtypes.GeneratorParams_PbftModule{
		PbftModule: &ordererpbtypes.PBFTModule{
			Segment:         segment.PbType(),
			AvailabilityId:  availabilityID.Pb(),
			Epoch:           epoch.Pb(),
			ValidityChecker: uint64(validityCheckerType),
		},
	}}
}
