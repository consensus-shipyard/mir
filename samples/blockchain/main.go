package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net/grpc"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

func main() {
	fmt.Println("Starting blockchain")

	numberOfNodes, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	idInput := os.Args[2]
	ownID, err := strconv.Atoi(idInput)
	if err != nil {
		panic(err)
	}
	ownNodeID := t.NodeID(idInput)

	msgLossProb := 0.0
	if len(os.Args) <= 3 {
		fmt.Printf("No message loss probability provided, defaulting to 0.0\n")
	} else {
		msgLossProb, err = strconv.ParseFloat(os.Args[3], 32)
		if err != nil {
			panic(err)
		}
	}

	nodes := make(map[t.NodeID]*trantorpbtypes.NodeIdentity, numberOfNodes)
	allNodeIds := make([]t.NodeID, numberOfNodes)
	otherNodes := make([]t.NodeID, numberOfNodes-1)
	otherNodesIndex := 0
	for i := 0; i < numberOfNodes; i++ {
		nodeIdStr := strconv.Itoa(i)
		nodeId := t.NodeID(nodeIdStr)
		allNodeIds[i] = nodeId
		if nodeId != ownNodeID {
			otherNodes[otherNodesIndex] = nodeId
			otherNodesIndex++
		}
		nodes[nodeId] = &trantorpbtypes.NodeIdentity{Id: nodeId, Addr: "/ip4/127.0.0.1/tcp/1000" + nodeIdStr, Key: nil, Weight: "1"}
	}
	membership := &trantorpbtypes.Membership{Nodes: nodes}

	// Instantiate network trnasport module and establish connections.
	transport, err := grpc.NewTransport(ownNodeID, membership.Nodes[ownNodeID].Addr, logging.ConsoleWarnLogger)
	if err != nil {
		panic(err)
	}
	if err := transport.Start(); err != nil {
		panic(err)
	}
	transport.Connect(membership)

	// Instantiate Mir node.
	node, err := mir.NewNode(
		ownNodeID,
		mir.DefaultNodeConfig(),
		map[t.ModuleID]modules.Module{
			"transport":     transport,
			"bcm":           NewBCM(8080 + ownID),
			"miner":         NewMiner(),
			"communication": NewCommunication(otherNodes, float32(msgLossProb)),
			"tpm":           NewTPM(),
		},
		nil,
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

	// block until nodeError receives an error
	// fmt.Printf("timer started\n")

	switch <-nodeError {
	default:
		fmt.Printf("Mir node stopped: %v\n", <-nodeError)
	}

	// time.Sleep(5 * time.Second)
	// fmt.Printf("timer up")

	// Stop the node.
	node.Stop()
	transport.Stop()
	fmt.Printf("Mir node stopped: %v\n", <-nodeError)
}
