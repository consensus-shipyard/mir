package orderers

import (
	"github.com/filecoin-project/mir/pkg/events"
	apbevents "github.com/filecoin-project/mir/pkg/pb/availabilitypb/events"
	apbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	contextstorepbtypes "github.com/filecoin-project/mir/pkg/pb/contextstorepb/types"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	factorypbtypes "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	hasherpbtypes "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
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

func (orderer *Orderer) requestCertOrigin() *events.EventList {
	return events.ListOf(apbevents.RequestCert(
		orderer.moduleConfig.Ava,
		&apbtypes.RequestCertOrigin{
			Module: orderer.moduleConfig.Self,
			Type: &apbtypes.RequestCertOrigin_ContextStore{ContextStore: &contextstorepbtypes.Origin{
				ItemID: 0, // TODO remove this parameter. It is deprecated as now ModuleID is a particular PBFT orderer.
			}},
		},
	).Pb())
}

func HashOrigin(module t.ModuleID, origin *ordererpb.HashOrigin) *hasherpbtypes.HashOrigin {
	return &hasherpbtypes.HashOrigin{
		Module: module,
		Type:   &hasherpbtypes.HashOrigin_Sb{Sb: origin},
	}
}

func SignOrigin(module t.ModuleID, origin *ordererpb.SignOrigin) *cryptopbtypes.SignOrigin {
	return &cryptopbtypes.SignOrigin{
		Module: module,
		Type:   &cryptopbtypes.SignOrigin_Sb{Sb: origin},
	}
}

func SigVerOrigin(module t.ModuleID, origin *ordererpb.SigVerOrigin) *cryptopbtypes.SigVerOrigin {
	return &cryptopbtypes.SigVerOrigin{
		Module: module,
		Type:   &cryptopbtypes.SigVerOrigin_Sb{Sb: origin},
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
			Segment:         ordererpbtypes.PBFTSegmentFromPb(segment.Pb()),
			AvailabilityId:  availabilityID.Pb(),
			Epoch:           epoch.Pb(),
			ValidityChecker: uint64(validityCheckerType),
		},
	}}
}
