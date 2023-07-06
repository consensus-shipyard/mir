package lowlevel

import (
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/pb/pingpongpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

func Message(destModule t.ModuleID, message *pingpongpb.Message) *messagepb.Message {
	return &messagepb.Message{
		DestModule: destModule.Pb(),
		Type:       &messagepb.Message_Pingpong{Pingpong: message},
	}
}

func PingMessage(destModule t.ModuleID, seqNr uint64) *messagepb.Message {
	return Message(
		destModule,
		&pingpongpb.Message{
			Type: &pingpongpb.Message_Ping{Ping: &pingpongpb.Ping{
				SeqNr: seqNr,
			}}},
	)
}

func PongMessage(destModule t.ModuleID, seqNr uint64) *messagepb.Message {
	return Message(
		destModule,
		&pingpongpb.Message{
			Type: &pingpongpb.Message_Pong{Pong: &pingpongpb.Pong{
				SeqNr: seqNr,
			}}},
	)
}

func Event(destModule t.ModuleID, ppEvent *pingpongpb.Event) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_PingPong{
			PingPong: ppEvent,
		},
	}
}

func PingTimeEvent(destModule t.ModuleID) *eventpb.Event {
	return Event(
		destModule,
		&pingpongpb.Event{Type: &pingpongpb.Event_PingTime{
			PingTime: &pingpongpb.PingTime{},
		}},
	)
}
