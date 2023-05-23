package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net/grpc"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/timer"
	t "github.com/filecoin-project/mir/pkg/types"
)

func main() {
	fmt.Println("Starting ping-pong.")

	membership := &trantorpbtypes.Membership{map[t.NodeID]*trantorpbtypes.NodeIdentity{ // nolint:govet
		"0": {"0", "/ip4/127.0.0.1/tcp/10000", nil, 1}, // nolint:govet
		"1": {"1", "/ip4/127.0.0.1/tcp/10001", nil, 1}, // nolint:govet
	}}

	ownID := t.NodeID(os.Args[1])
	tranport, err := grpc.NewTransport(ownID, membership.Nodes[ownID].Addr, logging.ConsoleWarnLogger)
	if err != nil {
		panic(err)
	}
	if err := tranport.Start(); err != nil {
		panic(err)
	}
	tranport.Connect(membership)

	node, err := mir.NewNode(
		ownID,
		mir.DefaultNodeConfig(),
		map[t.ModuleID]modules.Module{
			"transport": tranport,
			"pingpong":  NewPingPong(ownID),
			"timer":     timer.New(),
		},
		nil,
	)
	if err != nil {
		panic(err)
	}

	nodeError := make(chan error)
	go func() {
		nodeError <- node.Run(context.Background())
	}()

	fmt.Println("Mir node running.")

	time.Sleep(5 * time.Second)

	node.Stop()

	fmt.Printf("Mir node error: %v\n", <-nodeError)
}
