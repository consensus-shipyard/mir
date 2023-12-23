package customevents

import (
	"fmt"
	"time"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net/grpc"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	ppdsl "github.com/filecoin-project/mir/pkg/pb/pingpongpb/dsl"
	"github.com/filecoin-project/mir/samples/pingpong/customevents/pingpongevents"
	stddsl "github.com/filecoin-project/mir/stdevents/dsl"
	t "github.com/filecoin-project/mir/stdtypes"
)

func NewPingPong(ownNodeID t.NodeID) modules.PassiveModule {

	m := dsl.NewModule("pingpong")
	nextSN := uint64(0)

	eventpbdsl.UponInit(m, func() error {
		stddsl.TimerRepeat(m, "timer", time.Second, 0, PingTimeEvent("pingpong"))
		return nil
	})

	ppdsl.UponPingTime(m, func() error {

		// Get ID of other node.
		var destNodeID t.NodeID
		if ownNodeID == "0" {
			destNodeID = "1"
		} else {
			destNodeID = "0"
		}

		// Send PING message.
		dsl.EmitEvent(m, grpc.NewOutgoingMessage(
			pingpongevents.Message(pingpongevents.Ping{SeqNr: nextSN}),
			"transport",
			"pingpong",
			[]t.NodeID{destNodeID},
		))
		nextSN++
		return nil
	})

	pingpongevents.UponGRPCMessage(m, func(ping *pingpongevents.Ping, from t.NodeID) error {
		fmt.Printf("Received ping from %s: %d\n", from, ping.SeqNr)

		dsl.EmitEvent(m, grpc.NewOutgoingMessage(
			pingpongevents.Message(&pingpongevents.Pong{SeqNr: ping.SeqNr}),
			"transport",
			"pingpong",
			[]t.NodeID{from},
		))
		return nil
	})

	pingpongevents.UponGRPCMessage(m, func(pong *pingpongevents.Pong, from t.NodeID) error {
		fmt.Printf("Received pong from %s: %d\n", from, pong.SeqNr)
		return nil
	})

	return m
}
