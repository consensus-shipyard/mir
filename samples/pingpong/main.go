package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/filecoin-project/mir/pkg/debugger"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/eventlog"
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

	debug := flag.Bool("d", false, "Enable debug mode")
	debugPort := flag.String("port", "", "Debug port number")
	flag.Parse()
	var ownID t.NodeID
	var interceptor *eventlog.Recorder
	var err error

	if *debug {
		// In debug mode, expect the next argument to be the node ID
		if flag.NArg() > 0 {
			ownID = t.NodeID(flag.Arg(0))
		} else {
			fmt.Println("Node ID must be provided in debug mode")
			os.Exit(1)
		}
		interceptor, err = debugger.NewWebSocketDebugger(ownID, *debugPort, logging.ConsoleInfoLogger) // replace ownID with port number
		if err != nil {
			panic(err)
		}
	} else {
		// If not in debug mode, use the first argument as the node ID
		if len(os.Args) > 1 {
			ownID = t.NodeID(os.Args[1])
		} else {
			fmt.Println("Node ID must be provided")
			os.Exit(1)
		}
	}

	// Instantiate network transport module and establish connections.
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
	time.Sleep(50 * time.Second)

	// Stop the node.
	node.Stop()
	transport.Stop()
	fmt.Printf("Mir node stopped: %v\n", <-nodeError)
}
