package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/eventmangler"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net/grpc"
	applicationpbevents "github.com/filecoin-project/mir/pkg/pb/blockchainpb/applicationpb/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/timer"
	t "github.com/filecoin-project/mir/pkg/types"
	application "github.com/filecoin-project/mir/samples/blockchain/application"
	"github.com/filecoin-project/mir/samples/blockchain/wsInterceptor"
)

func main() {
	fmt.Println("Starting blockchain")

	logger := logging.ConsoleDebugLogger

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

	mangle := false
	if len(os.Args) >= 4 {
		mangle, err = strconv.ParseBool(os.Args[3])
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
	transport, err := grpc.NewTransport(ownNodeID, membership.Nodes[ownNodeID].Addr, logging.ConsoleInfoLogger)
	if err != nil {
		panic(err)
	}
	if err := transport.Start(); err != nil {
		panic(err)
	}
	transport.Connect(membership)

	timer := timer.New()

	mangler, err := eventmangler.NewModule(
		eventmangler.ModuleConfig{Self: "mangler", Dest: "transport", Timer: "timer"},
		&eventmangler.ModuleParams{MinDelay: time.Second / 1000, MaxDelay: 1 * time.Second, DropRate: 0.05},
	)
	if err != nil {
		panic(err)
	}

	// Instantiate Mir node.
	node, err := mir.NewNode(
		ownNodeID,
		mir.DefaultNodeConfig(),
		map[t.ModuleID]modules.Module{
			"transport":     transport,
			"bcm":           NewBCM(logging.Decorate(logger, "BCM:\t")),
			"miner":         NewMiner(ownNodeID, logging.Decorate(logger, "Miner:\t")),
			"communication": NewCommunication(otherNodes, mangle, logging.Decorate(logger, "Comm:\t")),
			"application":   application.NewApplication(logging.Decorate(logger, "App:\t"), ownNodeID),
			"synchronizer":  NewSynchronizer(ownNodeID, otherNodes, false, logging.Decorate(logger, "Sync:\t")),
			"timer":         timer,
			"mangler":       mangler,
			"devnull":       modules.NullPassive{}, // for messages that are actually destined for the interceptor
		},
		wsInterceptor.NewWsInterceptor(
			func(e *eventpb.Event) bool {
				switch e.Type.(type) {
				case *eventpb.Event_Bcinterceptor:
					return true
				default:
					return false
				}
			},
			8080+ownID,
			logging.Decorate(logger, "WSInter:\t"),
		))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	nodeError := make(chan error)
	go func() {
		nodeError <- node.Run(ctx)
	}()
	fmt.Println("Mir node running.")

	// block until nodeError receives an error
	// fmt.Printf("timer started\n")

	// fmt.Printf("Mir node stopped: %v\n", <-nodeError)

	// ================================================================================
	// Read chat messages from stdin and submit them as transactions.
	// ================================================================================

	scanner := bufio.NewScanner(os.Stdin)

	// Prompt for chat message input.
	fmt.Println("Type in your messages and press 'Enter' to send.")

	// Read chat message from stdin.
	for scanner.Scan() {
		fmt.Println("Gimme more")
		// Submit the chat message as transaction payload to the mempool module.
		if err := node.InjectEvents(ctx, events.ListOf(
			applicationpbevents.MessageInput("application", scanner.Text()).Pb(),
		)); err != nil {
			// Print error if occurred.
			fmt.Println(err)
		}

	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}

	// Stop the node.
	node.Stop()
	transport.Stop()
	fmt.Printf("Mir node stopped")
}
