package main

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/pb/pingpongpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/samples/pingpong/protobufs"
)

type pingpong struct {
	ownID t.NodeID
	seqNr uint64
}

func newPingpong(ownID t.NodeID) *pingpong {
	return &pingpong{
		ownID: ownID,
		seqNr: 0,
	}
}

func (p *pingpong) ImplementsModule() {}

func (p *pingpong) ApplyEvents(events *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsSequentially(events, p.applyEvent)
}

func (p *pingpong) applyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
		return p.applyInit()
	case *eventpb.Event_Pingpong:
		switch e.Pingpong.Type.(type) {
		case *pingpongpb.Event_PingTime:
			return p.applyPingTime()
		default:
			return nil, errors.Errorf("unknown pingpong event type: %T", e.Pingpong.Type)
		}
	case *eventpb.Event_MessageReceived:
		return p.applyMessageReceived(e.MessageReceived)
	default:
		return nil, errors.Errorf("unexpected event type: %T", e)
	}
}

func (p *pingpong) applyInit() (*events.EventList, error) {
	return events.ListOf(events.TimerRepeat("timer",
		[]*eventpb.Event{protobufs.TimeoutEvent("pingpong")},
		t.TimeDuration(time.Second),
		0,
	)), nil
}

func (p *pingpong) applyPingTime() (*events.EventList, error) {
	ping := protobufs.PingMessage("pingpong", p.seqNr)
	p.seqNr++

	var destNodeID t.NodeID
	if p.ownID == "0" {
		destNodeID = "1"
	} else {
		destNodeID = "0"
	}

	return events.ListOf(events.SendMessage("transport", ping, []t.NodeID{destNodeID})), nil

}

func (p *pingpong) applyMessageReceived(msgEvent *eventpb.MessageReceived) (*events.EventList, error) {
	switch msg := msgEvent.Msg.Type.(type) {
	case *messagepb.Message_Pingpong:
		switch pingpongMsg := msg.Pingpong.Type.(type) {
		case *pingpongpb.Message_Ping:
			return p.applyPing(pingpongMsg.Ping, t.NodeID(msgEvent.From))
		case *pingpongpb.Message_Pong:
			return p.applyPong(pingpongMsg.Pong)
		default:
			return nil, errors.Errorf("unknown pingpong message type %T", pingpongMsg)
		}
	default:
		return nil, errors.Errorf("unexpected message type %T", msg)
	}
}

func (p *pingpong) applyPing(ping *pingpongpb.Ping, from t.NodeID) (*events.EventList, error) {
	fmt.Println("Received ping with seq nr: ", ping.SeqNr)
	pong := protobufs.PongMessage("pingpong", ping.SeqNr)
	return events.ListOf(events.SendMessage("transport", pong, []t.NodeID{from})), nil
}

func (p *pingpong) applyPong(pong *pingpongpb.Pong) (*events.EventList, error) {
	fmt.Println("Received pong with seq nr: ", pong.SeqNr)
	return events.EmptyList(), nil
}
