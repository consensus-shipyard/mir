package customevents

import (
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/pingpongpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

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
