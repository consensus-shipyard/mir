package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/multiformats/go-multiaddr"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net/grpc"
	"github.com/filecoin-project/mir/pkg/timer"
	t "github.com/filecoin-project/mir/pkg/types"
)

func main() {
	fmt.Println("Starting ping-pong.")

	addrs := make(map[t.NodeID]t.NodeAddress)
	var err error
	if addrs["0"], err = multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/10000"); err != nil {
		panic(err)
	}
	if addrs["1"], err = multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/10001"); err != nil {
		panic(err)
	}

	fmt.Println("Connecting to peer.")
	ownID := t.NodeID(os.Args[1])
	transport, err := grpc.NewTransport(ownID, addrs[ownID], logging.ConsoleWarnLogger)
	if err != nil {
		panic(err)
	}
	if err := transport.Start(); err != nil {
		panic(err)
	}
	transport.Connect(context.Background(), addrs)

	node, err := mir.NewNode(
		ownID,
		&mir.NodeConfig{Logger: logging.ConsoleWarnLogger},
		map[t.ModuleID]modules.Module{
			"transport": transport,
			"pingpong":  newPingpong(ownID),
			"timer":     timer.New(),
		},
		nil,
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

	time.Sleep(10 * time.Second)

	node.Stop()

	fmt.Printf("Mir node error: %v\n", <-nodeError)
}
