package main

import (
	"fmt"
	"time"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/modules"
	ppdsl "github.com/filecoin-project/mir/pkg/pb/pingpongpb/dsl"
	ppevents "github.com/filecoin-project/mir/pkg/pb/pingpongpb/events"
	ppmsgs "github.com/filecoin-project/mir/pkg/pb/pingpongpb/msgs"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	"github.com/filecoin-project/mir/stdevents"
	stddsl "github.com/filecoin-project/mir/stdevents/dsl"
	t "github.com/filecoin-project/mir/stdtypes"
)

func NewPingPong(ownNodeID t.NodeID) modules.PassiveModule {

	m := dsl.NewModule("pingpong")
	nextSN := uint64(0)

	dsl.UponEvent(m, func(_ *stdevents.Init) error {
		stddsl.TimerRepeat(
			m,
			"timer",
			time.Second,
			0,
			ppevents.PingTime("pingpong").Pb(),
		)
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
		nextSN++
		transportpbdsl.SendMessage(m, "transport", ppmsgs.Ping("pingpong", nextSN), []t.NodeID{destNodeID})
		return nil
	})

	ppdsl.UponPingReceived(m, func(from t.NodeID, seqNr uint64) error {
		fmt.Printf("Received ping from %s: %d\n", from, seqNr)
		transportpbdsl.SendMessage(m, "transport", ppmsgs.Pong("pingpong", seqNr), []t.NodeID{from})
		return nil
	})

	ppdsl.UponPongReceived(m, func(from t.NodeID, seqNr uint64) error {
		fmt.Printf("Received pong from %s: %d\n", from, seqNr)
		return nil
	})

	return m
}
