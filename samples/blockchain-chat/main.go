package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/blockchain"
	"github.com/filecoin-project/mir/pkg/blockchain/wsInterceptor"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	applicationpbevents "github.com/filecoin-project/mir/pkg/pb/blockchainpb/applicationpb/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/samples/blockchain-chat/application"
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

	logger := logging.ConsoleInfoLogger

	system := blockchain.New(
		ownNodeID,
		application.NewApplication(ownNodeID, logging.Decorate(logger, "Application:\t")),
		*disableMangle,
		*dropRate,
		*minDelay,
		*maxDelay,
		*numberOfNodes,
		logger,
		logging.Decorate(logging.ConsoleInfoLogger, "Transport:\t"))

	// Instantiate Mir node.
	node, err := mir.NewNode(
		ownNodeID,
		mir.DefaultNodeConfig(),
		system.Modules(),
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

	// ==========================================================
	// Read chat messages from stdin and submit them as payloads.
	// ==========================================================

	scanner := bufio.NewScanner(os.Stdin)

	// Prompt for chat message input.
	fmt.Println("Type in your messages and press 'Enter' to send.")

	// Read chat message from stdin.
	for scanner.Scan() {
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
	fmt.Printf("Mir node stopped")
}
