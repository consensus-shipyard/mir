package customevents

import (
	"fmt"
	"time"

	es "github.com/go-errors/errors"
	"github.com/pkg/errors"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net/grpc"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	eventpbevents "github.com/filecoin-project/mir/pkg/pb/eventpb/events"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	"github.com/filecoin-project/mir/pkg/pb/pingpongpb"
	"github.com/filecoin-project/mir/pkg/timer/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/samples/pingpong/customevents/pingpongevents"
)

type Pingpong struct {
	ownID  t.NodeID
	nextSn uint64
}

func NewPingPong(ownID t.NodeID) *Pingpong {
	return &Pingpong{
		ownID:  ownID,
		nextSn: 0,
	}
}

func (p *Pingpong) ImplementsModule() {}

func (p *Pingpong) ApplyEvents(evts *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsSequentially(evts, p.applyEvent)
}

func (p *Pingpong) applyEvent(event events.Event) (*events.EventList, error) {

	switch evt := event.(type) {
	case *grpc.MessageReceived:
		return p.applyMessageReceived(evt)
	case *eventpb.Event:
		switch e := evt.Type.(type) {
		case *eventpb.Event_Init:
			return p.applyInit()
		case *eventpb.Event_PingPong:
			switch e := e.PingPong.Type.(type) {
			case *pingpongpb.Event_PingTime:
				return p.applyPingTime(e.PingTime)
			default:
				return nil, errors.Errorf("unknown pingpong event type: %T", e)
			}
		default:
			return nil, errors.Errorf("unknown event type: %T", evt.Type)
		}
	default:
		return nil, es.Errorf("The pingpong module only supports proto and pingpong events, received %T", event)
	}
}

func (p *Pingpong) applyInit() (*events.EventList, error) {
	p.nextSn = 0
	timerEvent := eventpbevents.TimerRepeat(
		"timer",
		[]*eventpbtypes.Event{eventpbtypes.EventFromPb(PingTimeEvent("pingpong"))},
		types.Duration(time.Second),
		0,
	)
	return events.ListOf(timerEvent.Pb()), nil
}

func (p *Pingpong) applyPingTime(_ *pingpongpb.PingTime) (*events.EventList, error) {
	return p.sendPing()
}

func (p *Pingpong) sendPing() (*events.EventList, error) {
	var destID t.NodeID
	if p.ownID == "0" {
		destID = "1"
	} else {
		destID = "0"
	}

	ping := grpc.NewOutgoingMessage(
		pingpongevents.Message(pingpongevents.Ping{SeqNr: p.nextSn}),
		"transport",
		"pingpong",
		[]t.NodeID{destID},
	)
	p.nextSn++
	return events.ListOf(ping), nil
}

func (p *Pingpong) applyMessageReceived(msg *grpc.MessageReceived) (*events.EventList, error) {
	m, err := pingpongevents.ParseMessage(msg.Data())

	if err != nil {
		return nil, es.Errorf("failed parsing message: %w", err)
	}

	switch message := m.(type) {
	case *pingpongevents.Ping:
		return p.applyPing(message, msg.SrcNode())
	case *pingpongevents.Pong:
		return p.applyPong(message, msg.SrcNode())
	default:
		return nil, errors.Errorf("unknown message type: %T", message)
	}
}

func (p *Pingpong) applyPing(ping *pingpongevents.Ping, from t.NodeID) (*events.EventList, error) {
	fmt.Printf("Received ping from %s: %d\n", from, ping.SeqNr)

	pong := grpc.NewOutgoingMessage(
		pingpongevents.Message(&pingpongevents.Pong{SeqNr: ping.SeqNr}),
		"transport",
		"pingpong",
		[]t.NodeID{from},
	)
	return events.ListOf(pong), nil
}

func (p *Pingpong) applyPong(pong *pingpongevents.Pong, from t.NodeID) (*events.EventList, error) {
	fmt.Printf("Received pong from %s: %d\n", from, pong.SeqNr)
	return events.EmptyList(), nil
}
