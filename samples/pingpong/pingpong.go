package main

import (
	"fmt"
	"time"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/modules"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	ppdsl "github.com/filecoin-project/mir/pkg/pb/pingpongpb/dsl"
	ppevents "github.com/filecoin-project/mir/pkg/pb/pingpongpb/events"
	ppmsgs "github.com/filecoin-project/mir/pkg/pb/pingpongpb/msgs"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	"github.com/filecoin-project/mir/pkg/timer/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

func NewPingPong(ownNodeID t.NodeID) modules.PassiveModule {

	m := dsl.NewModule("pingpong")
	nextSN := uint64(0)

	dsl.UponInit(m, func() error {
		eventpbdsl.TimerRepeat(
			m,
			"timer",
			[]*eventpbtypes.Event{ppevents.PingTime("pingpong")},
			types.Duration(time.Second),
			0,
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
