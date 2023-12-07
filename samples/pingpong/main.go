package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/filecoin-project/mir/pkg/eventlog"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/debugger"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net/grpc"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/timer"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/samples/pingpong/lowlevel"
)

func main() {
	fmt.Println("Starting ping-pong.")
	// Manually create system membership with just 2 nodes.
	membership := &trantorpbtypes.Membership{map[t.NodeID]*trantorpbtypes.NodeIdentity{ // nolint:govet
		"0": {"0", "/ip4/127.0.0.1/tcp/10000", nil, "1"}, // nolint:govet
		"1": {"1", "/ip4/127.0.0.1/tcp/10001", nil, "1"}, // nolint:govet
	}}

	// Get own ID from command line.
	ownID := t.NodeID(os.Args[1])
	// Get 0 for no debugger, 1 for debugger
	debugg := os.Args[2]

	var interceptor *eventlog.Recorder
	var err error
	interceptor, _ = debugger.InterceptorInit(debugg == "1", ownID) //debugger.InterceptorInit(debugger == "1", ownID)

	// Instantiate network trnasport module and establish connections.
	transport, err := grpc.NewTransport(ownID, membership.Nodes[ownID].Addr, logging.ConsoleWarnLogger)
	if err != nil {
		panic(err)
	}
	if err := transport.Start(); err != nil {
		panic(err)
	}
	transport.Connect(membership)

	// Instantiate Mir node.
	node, err := mir.NewNode(
		ownID,
		mir.DefaultNodeConfig(),
		map[t.ModuleID]modules.Module{
			"transport": transport,
			//"pingpong":  NewPingPong(ownID),
			"pingpong": lowlevel.NewPingPong(ownID),
			"timer":    timer.New(),
		},
		interceptor,
	)
	if err != nil {
		panic(err)
	}

	// Run the node for 5 seconds.
	nodeError := make(chan error)
	go func() {
		nodeError <- node.Run(context.Background())
	}()
	fmt.Println("Mir node running.")
	time.Sleep(5 * time.Second)

	// Stop the node.
	node.Stop()
	transport.Stop()
	fmt.Printf("Mir node stopped: %v\n", <-nodeError)
}
