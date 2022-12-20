package orderers

import (
	availabilityevents "github.com/filecoin-project/mir/pkg/availability/events"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	"github.com/filecoin-project/mir/pkg/pb/contextstorepb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/factorymodulepb"
	"github.com/filecoin-project/mir/pkg/pb/isspb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/pb/ordererspb"
	t "github.com/filecoin-project/mir/pkg/types"
)

func OrdererEvent(
	destModule t.ModuleID,
	event *ordererspb.SBInstanceEvent,
) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_SbEvent{
			SbEvent: event,
		},
	}
}

func (orderer *Orderer) requestCertOrigin() *events.EventList {
	return events.ListOf(availabilityevents.RequestCert(
		orderer.moduleConfig.Ava,
		&availabilitypb.RequestCertOrigin{
			Module: orderer.moduleConfig.Self.Pb(),
			Type: &availabilitypb.RequestCertOrigin_ContextStore{ContextStore: &contextstorepb.Origin{
				ItemID: 0, // TODO remove this parameter. It is deprecated as now ModuleID is a particular PBFT orderer.
			}},
		},
	))
}

func HashOrigin(module t.ModuleID, origin *ordererspb.SBInstanceHashOrigin) *eventpb.HashOrigin {
	return &eventpb.HashOrigin{
		Module: module.Pb(),
		Type:   &eventpb.HashOrigin_Sb{Sb: origin},
	}
}

func SignOrigin(module t.ModuleID, origin *ordererspb.SBInstanceSignOrigin) *eventpb.SignOrigin {
	return &eventpb.SignOrigin{
		Module: module.Pb(),
		Type:   &eventpb.SignOrigin_Sb{Sb: origin},
	}
}

func SigVerOrigin(module t.ModuleID, origin *ordererspb.SBInstanceSigVerOrigin) *eventpb.SigVerOrigin {
	return &eventpb.SigVerOrigin{
		Module: module.Pb(),
		Type:   &eventpb.SigVerOrigin_Sb{Sb: origin},
	}
}

func InstanceParams(
	segment *Segment,
	availabilityID t.ModuleID,
	epoch t.EpochNr,
) *factorymodulepb.GeneratorParams {
	return &factorymodulepb.GeneratorParams{Type: &factorymodulepb.GeneratorParams_PbftModule{
		PbftModule: &ordererspb.PBFTModule{
			Segment: &ordererspb.PBFTSegment{
				Leader:     segment.Leader.Pb(),
				Membership: t.NodeIDSlicePb(segment.Membership),
				SeqNrs:     t.SeqNrSlicePb(segment.SeqNrs),
			},
			AvailabilityId: availabilityID.Pb(),
			Epoch:          epoch.Pb(),
		},
	}}
}

func OrdererMessage(msg *ordererspb.SBInstanceMessage, destModule t.ModuleID) *messagepb.Message {
	return &messagepb.Message{DestModule: string(destModule), Type: &messagepb.Message_SbMessage{SbMessage: msg}}
}

func SBDeliverEvent(sn t.SeqNr, certData []byte, aborted bool, leader t.NodeID) *isspb.ISSEvent {
	return &isspb.ISSEvent{Type: &isspb.ISSEvent_SbDeliver{SbDeliver: &isspb.SBDeliver{
		Sn:       sn.Pb(),
		CertData: certData,
		Aborted:  aborted,
		Leader:   leader.Pb(),
	}}}
}
