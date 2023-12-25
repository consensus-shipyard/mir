package lowlevel

import (
	"fmt"
	"time"

	"github.com/filecoin-project/mir/stdevents"
	es "github.com/go-errors/errors"
	"github.com/pkg/errors"

	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	messagepbtypes "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	"github.com/filecoin-project/mir/pkg/pb/pingpongpb"
	"github.com/filecoin-project/mir/pkg/pb/transportpb"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	"github.com/filecoin-project/mir/stdtypes"
)

type Pingpong struct {
	ownID  stdtypes.NodeID
	nextSn uint64
}

func NewPingPong(ownID stdtypes.NodeID) *Pingpong {
	return &Pingpong{
		ownID:  ownID,
		nextSn: 0,
	}
}

func (p *Pingpong) ImplementsModule() {}

func (p *Pingpong) ApplyEvents(evts *stdtypes.EventList) (*stdtypes.EventList, error) {
	return modules.ApplyEventsSequentially(evts, p.applyEvent)
}

func (p *Pingpong) applyEvent(event stdtypes.Event) (*stdtypes.EventList, error) {

	// Ignore Init event.
	_, ok := event.(*stdevents.Init)
	if ok {
		return stdtypes.EmptyList(), nil
	}
	// We only support proto events.
	pbevent, ok := event.(*eventpb.Event)
	if !ok {
		return nil, es.Errorf("The pingpong module only supports proto events, received %T", event)
	}

	switch e := pbevent.Type.(type) {
	case *eventpb.Event_Transport:
		switch e := e.Transport.Type.(type) {
		case *transportpb.Event_MessageReceived:
			return p.applyMessageReceived(e.MessageReceived)
		default:
			return nil, errors.Errorf("unknown transport event type: %T", e)
		}
	case *eventpb.Event_PingPong:
		switch e := e.PingPong.Type.(type) {
		case *pingpongpb.Event_PingTime:
			return p.applyPingTime(e.PingTime)
		default:
			return nil, errors.Errorf("unknown pingpong event type: %T", e)
		}
	default:
		return nil, errors.Errorf("unknown event type: %T", pbevent.Type)
	}
}

func (p *Pingpong) applyInit() (*stdtypes.EventList, error) {
	p.nextSn = 0
	timerEvent := stdevents.NewTimerRepeat(
		"timer",
		time.Second,
		0,
		PingTimeEvent("pingpong"),
	)
	return stdtypes.ListOf(timerEvent), nil
}

func (p *Pingpong) applyPingTime(_ *pingpongpb.PingTime) (*stdtypes.EventList, error) {
	return p.sendPing()
}

func (p *Pingpong) sendPing() (*stdtypes.EventList, error) {
	var destID stdtypes.NodeID
	if p.ownID == "0" {
		destID = "1"
	} else {
		destID = "0"
	}

	ping := PingMessage("pingpong", p.nextSn)
	p.nextSn++
	sendMsgEvent := transportpbevents.SendMessage("transport", messagepbtypes.MessageFromPb(ping), []stdtypes.NodeID{destID})
	return stdtypes.ListOf(sendMsgEvent.Pb()), nil
}

func (p *Pingpong) applyMessageReceived(msg *transportpb.MessageReceived) (*stdtypes.EventList, error) {
	switch message := msg.Msg.Type.(type) {
	case *messagepb.Message_Pingpong:
		switch m := message.Pingpong.Type.(type) {
		case *pingpongpb.Message_Ping:
			return p.applyPing(m.Ping, stdtypes.NodeID(msg.From))
		case *pingpongpb.Message_Pong:
			return p.applyPong(m.Pong, stdtypes.NodeID(msg.From))
		default:
			return nil, errors.Errorf("unknown pingpong message type: %T", m)
		}
	default:
		return nil, errors.Errorf("unknown message type: %T", message)
	}
}

func (p *Pingpong) applyPing(ping *pingpongpb.Ping, from stdtypes.NodeID) (*stdtypes.EventList, error) {
	fmt.Printf("Received ping from %s: %d\n", from, ping.SeqNr)
	pong := PongMessage("pingpong", ping.SeqNr)
	sendMsgEvent := transportpbevents.SendMessage("transport", messagepbtypes.MessageFromPb(pong), []stdtypes.NodeID{from})
	return stdtypes.ListOf(sendMsgEvent.Pb()), nil
}

func (p *Pingpong) applyPong(pong *pingpongpb.Pong, from stdtypes.NodeID) (*stdtypes.EventList, error) {
	fmt.Printf("Received pong from %s: %d\n", from, pong.SeqNr)
	return stdtypes.EmptyList(), nil
}
