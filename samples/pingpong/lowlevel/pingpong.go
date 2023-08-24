package lowlevel

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	eventpbevents "github.com/filecoin-project/mir/pkg/pb/eventpb/events"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	messagepbtypes "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	pingpongpbtypes "github.com/filecoin-project/mir/pkg/pb/pingpongpb/types"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	transportpbtypes "github.com/filecoin-project/mir/pkg/pb/transportpb/types"
	"github.com/filecoin-project/mir/pkg/timer/types"
	t "github.com/filecoin-project/mir/pkg/types"
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

func (p *Pingpong) applyEvent(event *eventpbtypes.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpbtypes.Event_Init:
		return p.applyInit()
	case *eventpbtypes.Event_Transport:
		switch e := e.Transport.Type.(type) {
		case *transportpbtypes.Event_MessageReceived:
			return p.applyMessageReceived(e.MessageReceived)
		default:
			return nil, errors.Errorf("unknown transport event type: %T", e)
		}
	case *eventpbtypes.Event_PingPong:
		switch e := e.PingPong.Type.(type) {
		case *pingpongpbtypes.Event_PingTime:
			return p.applyPingTime(e.PingTime)
		default:
			return nil, errors.Errorf("unknown pingpong event type: %T", e)
		}
	default:
		return nil, errors.Errorf("unknown event type: %T", event.Type)
	}
}

func (p *Pingpong) applyInit() (*events.EventList, error) {
	p.nextSn = 0
	timerEvent := eventpbevents.TimerRepeat(
		"timer",
		[]*eventpbtypes.Event{PingTimeEvent("pingpong")},
		types.Duration(time.Second),
		0,
	)
	return events.ListOf(timerEvent), nil
}

func (p *Pingpong) applyPingTime(_ *pingpongpbtypes.PingTime) (*events.EventList, error) {
	return p.sendPing()
}

func (p *Pingpong) sendPing() (*events.EventList, error) {
	var destID t.NodeID
	if p.ownID == "0" {
		destID = "1"
	} else {
		destID = "0"
	}

	ping := PingMessage("pingpong", p.nextSn)
	p.nextSn++
	sendMsgEvent := transportpbevents.SendMessage("transport", ping, []t.NodeID{destID})
	return events.ListOf(sendMsgEvent), nil
}

func (p *Pingpong) applyMessageReceived(msg *transportpbtypes.MessageReceived) (*events.EventList, error) {
	switch message := msg.Msg.Type.(type) {
	case *messagepbtypes.Message_Pingpong:
		switch m := message.Pingpong.Type.(type) {
		case *pingpongpbtypes.Message_Ping:
			return p.applyPing(m.Ping, msg.From)
		case *pingpongpbtypes.Message_Pong:
			return p.applyPong(m.Pong, msg.From)
		default:
			return nil, errors.Errorf("unknown pingpong message type: %T", m)
		}
	default:
		return nil, errors.Errorf("unknown message type: %T", message)
	}
}

func (p *Pingpong) applyPing(ping *pingpongpbtypes.Ping, from t.NodeID) (*events.EventList, error) {
	fmt.Printf("Received ping from %s: %d\n", from, ping.SeqNr)
	pong := PongMessage("pingpong", ping.SeqNr)
	sendMsgEvent := transportpbevents.SendMessage("transport", pong, []t.NodeID{from})
	return events.ListOf(sendMsgEvent), nil
}

func (p *Pingpong) applyPong(pong *pingpongpbtypes.Pong, from t.NodeID) (*events.EventList, error) {
	fmt.Printf("Received pong from %s: %d\n", from, pong.SeqNr)
	return events.EmptyList(), nil
}
