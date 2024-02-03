package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math"
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

// Flags
var disableMangle = flag.Bool("disableMangle", false, "Disable mangling of messages")
var dropRate = flag.Float64("dropRate", 0.05, "The rate at which to drop messages")
var minDelay = flag.Float64("minDelay", 0.001, "The minimum delay to add to messages [seconds]")
var maxDelay = flag.Float64("maxDelay", 1, "The minimum delay to add to messages [seconds]")
var numberOfNodes = flag.Int("numberOfNodes", -1, "The number of nodes in the network [1, inf] REQUIRED")
var nodeID = flag.Int("nodeID", -1, "The ID of the node [0, numberOfNodes-1] REQUIRED")

func main() {
	// Parse command line flags
	flag.Parse()

	requiredFlags := []string{"numberOfNodes", "nodeID"}
	flagsSet := make(map[string]bool)
	missedRequiredFlags := false

	// Check which flags are set
	flag.Visit(func(f *flag.Flag) {
		flagsSet[f.Name] = true
	})

	// Check for missed required flags
	for _, f := range requiredFlags {
		if !flagsSet[f] {
			fmt.Printf("Flag %s is required\n", f)
			missedRequiredFlags = true
		}
	}

	// Exit if missing any required flag
	if missedRequiredFlags {
		os.Exit(2)
	}

	// Check for valid flag values
	if *numberOfNodes < 1 {
		fmt.Printf("Number of nodes must be greater than 0.\n")
		os.Exit(2)
	} else if *nodeID < 0 || *nodeID >= *numberOfNodes {
		fmt.Printf("Node ID must be between 0 and numberOfNodes-1.\n")
		os.Exit(2)
	} else if *minDelay < 0 {
		fmt.Printf("Minimum delay must be greater than or equal to 0.\n")
		os.Exit(2)
	} else if *maxDelay < 0 {
		fmt.Printf("Maximum delay must be greater than or equal to 0.\n")
		os.Exit(2)
	} else if *minDelay > *maxDelay {
		fmt.Printf("Minimum delay must be less than or equal to maximum delay.\n")
		os.Exit(2)
	} else if *dropRate < 0 || *dropRate > 1 {
		fmt.Printf("Drop rate must be between 0 and 1.\n")
		os.Exit(2)
	} else if *disableMangle && (flagsSet["dropRate"] || flagsSet["minDelay"] || flagsSet["maxDelay"]) {
		fmt.Printf("WARNING: Settings for drop rate, minimum delay, and maximum delay are ignored when mangling is disabled.\n")
	}

	ownNodeID := t.NodeID(strconv.Itoa(*nodeID))

	// Setting up node
	fmt.Println("Starting longest chain consensus protocol...")

	logger := logging.ConsoleDebugLogger

	// determine "other" nodes for this node
	nodes := make(map[t.NodeID]*trantorpbtypes.NodeIdentity, *numberOfNodes)
	allNodeIds := make([]t.NodeID, *numberOfNodes)
	otherNodes := make([]t.NodeID, *numberOfNodes-1)
	otherNodesIndex := 0
	for i := 0; i < *numberOfNodes; i++ {
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

	// Instantiate network transport module and establish connections.
	transport, err := grpc.NewTransport(ownNodeID, membership.Nodes[ownNodeID].Addr, logging.ConsoleInfoLogger)
	if err != nil {
		panic(err)
	}
	if err := transport.Start(); err != nil {
		panic(err)
	}
	transport.Connect(membership)

	modules := map[t.ModuleID]modules.Module{
		"transport":    transport,
		"bcm":          NewBCM(logging.Decorate(logger, "BCM:\t")),
		"miner":        NewMiner(ownNodeID, 0.2, logging.Decorate(logger, "Miner:\t")),
		"broadcast":    NewBroadcast(otherNodes, !*disableMangle, logging.Decorate(logger, "Comm:\t")),
		"application":  application.NewApplication(logging.Decorate(logger, "App:\t"), ownNodeID),
		"synchronizer": NewSynchronizer(ownNodeID, otherNodes, logging.Decorate(logger, "Sync:\t")),
		"devnull":      modules.NullPassive{}, // for messages that are actually destined for the interceptor
	}

	if !*disableMangle {
		minDelayDuration := time.Duration(int64(math.Round(*minDelay * float64(time.Second))))
		maxDelayDuration := time.Duration(int64(math.Round(*maxDelay * float64(time.Second))))
		manglerModule, err := eventmangler.NewModule(
			eventmangler.ModuleConfig{Self: "mangler", Dest: "transport", Timer: "timer"},
			&eventmangler.ModuleParams{MinDelay: minDelayDuration, MaxDelay: maxDelayDuration, DropRate: float32(*dropRate)},
		)
		if err != nil {
			panic(err)
		}

		modules["timer"] = timer.New()
		modules["mangler"] = manglerModule
	}

	// Instantiate Mir node.
	node, err := mir.NewNode(
		ownNodeID,
		mir.DefaultNodeConfig(),
		modules,
		wsInterceptor.NewWsInterceptor(
			func(e *eventpb.Event) bool {
				switch e.Type.(type) {
				case *eventpb.Event_Bcinterceptor:
					return true
				default:
					return false
				}
			},
			8080+*nodeID,
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
