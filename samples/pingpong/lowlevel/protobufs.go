package lowlevel

import (
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	messagepbtypes "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	pingpongpbtypes "github.com/filecoin-project/mir/pkg/pb/pingpongpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

func Message(destModule t.ModuleID, message *pingpongpbtypes.Message) *messagepbtypes.Message {
	return &messagepbtypes.Message{
		DestModule: destModule,
		Type:       &messagepbtypes.Message_Pingpong{Pingpong: message},
	}
}

func PingMessage(destModule t.ModuleID, seqNr uint64) *messagepbtypes.Message {
	return Message(
		destModule,
		&pingpongpbtypes.Message{
			Type: &pingpongpbtypes.Message_Ping{Ping: &pingpongpbtypes.Ping{
				SeqNr: seqNr,
			}}},
	)
}

func PongMessage(destModule t.ModuleID, seqNr uint64) *messagepbtypes.Message {
	return Message(
		destModule,
		&pingpongpbtypes.Message{
			Type: &pingpongpbtypes.Message_Pong{Pong: &pingpongpbtypes.Pong{
				SeqNr: seqNr,
			}}},
	)
}

func Event(destModule t.ModuleID, ppEvent *pingpongpbtypes.Event) *eventpbtypes.Event {
	return &eventpbtypes.Event{
		DestModule: destModule,
		Type: &eventpbtypes.Event_PingPong{
			PingPong: ppEvent,
		},
	}
}

func PingTimeEvent(destModule t.ModuleID) *eventpbtypes.Event {
	return Event(
		destModule,
		&pingpongpbtypes.Event{Type: &pingpongpbtypes.Event_PingTime{
			PingTime: &pingpongpbtypes.PingTime{},
		}},
	)
}
