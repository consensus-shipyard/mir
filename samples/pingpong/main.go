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

	ownID := t.NodeID(os.Args[1])
	tranport, err := grpc.NewTransport(ownID, addrs[ownID], logging.ConsoleWarnLogger)
	if err != nil {
		panic(err)
	}
	if err := tranport.Start(); err != nil {
		panic(err)
	}
	tranport.Connect(addrs)

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
