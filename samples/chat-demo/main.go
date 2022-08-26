/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// ********************************************************************************
//         Chat demo application for demonstrating the usage of Mir              //
//                             (main executable)                                 //
//                                                                               //
//                     Run with --help flag for usage info.                      //
// ********************************************************************************

package main

import (
	"bufio"
	"context"
	"fmt"
	net2 "net"
	"os"
	"strconv"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/availability/batchdb/fakebatchdb"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector"
	"github.com/filecoin-project/mir/pkg/batchfetcher"
	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/iss"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/membership"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/net/grpc"
	"github.com/filecoin-project/mir/pkg/net/libp2p"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
	libp2ptools "github.com/filecoin-project/mir/pkg/util/libp2p"
)

const (

	// Base port number for the node request receivers to listen to messages from clients.
	// Request receivers will listen on port reqReceiverBasePort through reqReceiverBasePort+3.
	// Note that protocol messages and requests are following two completely distinct paths to avoid interference
	// of clients with node-to-node communication.
	reqReceiverBasePort = 20000

	// Number of epochs by which to delay configuration changes.
	// If a configuration is agreed upon in epoch e, it will take effect in epoch e + 1 + configOffset.
	configOffset = 2
)

// parsedArgs represents parsed command-line parameters passed to the program.
type parsedArgs struct {

	// ID of this node.
	// The package github.com/hyperledger-labs/mir/pkg/types defines this and other types used by the library.
	OwnID t.NodeID

	// If set, print debug output to stdout.
	Verbose bool

	// If set, print trace output to stdout.
	Trace bool

	// Network transport.
	Net string

	// File containing the initial membership for joining nodes.
	InitMembershipFile *os.File
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	// Parse command-line parameters.
	_ = parseArgs(os.Args)
	args := parseArgs(os.Args)

	var err error

	// Initialize logger that will be used throughout the code to print log messages.
	var logger logging.Logger
	if args.Trace {
		logger = logging.ConsoleTraceLogger // Print trace-level info.
	} else if args.Verbose {
		logger = logging.ConsoleDebugLogger // Print debug-level info in verbose mode.
	} else {
		logger = logging.ConsoleWarnLogger // Only print errors and warnings by default.
	}

	ctx := context.Background()

	fmt.Println("Initializing...")

	// ================================================================================
	// Load system membership info: IDs, addresses, ports, etc...
	// ================================================================================

	ownNumericID, err := strconv.Atoi(string(args.OwnID))
	if err != nil {
		return fmt.Errorf("node IDs must be numeric in the sample app: %w", err)
	}

	initialMembership, err := membership.FromFile(args.InitMembershipFile)
	if err != nil {
		return fmt.Errorf("could not load membership: %w", err)
	}
	if err := args.InitMembershipFile.Close(); err != nil {
		return fmt.Errorf("could not close membership file: %w", err)
	}
	addresses, err := membership.GetIPs(initialMembership)
	if err != nil {
		return fmt.Errorf("could not load node IPs: %w", err)
	}

	// Generate addresses and ports for client request receivers.
	// Each node uses different ports for receiving protocol messages and requests.
	// These addresses will be used by the client code to know where to send its requests
	// (each client sends its requests to all request receivers). Each request receiver,
	// however, will only submit the received requests to its associated Node.
	reqReceiverAddrs := make(map[t.NodeID]string)
	for nodeID, nodeIP := range addresses {
		numericID, err := strconv.Atoi(string(nodeID))
		if err != nil {
			return fmt.Errorf("node IDs must be numeric in the sample app: %w", err)
		}
		reqReceiverAddrs[nodeID] = net2.JoinHostPort(nodeIP, fmt.Sprintf("%d", reqReceiverBasePort+numericID))
	}

	// ================================================================================
	// Create and initialize various modules used by mir.
	// ================================================================================

	// Initialize the networking module.
	// Mir will use it for transporting nod-to-node messages.

	// In the current implementation, all nodes are on the local machine, but listen on different port numbers.
	// Change this or make this configurable to deploy different nodes on different physical machines.
	var transport net.Transport
	var nodeAddrs map[t.NodeID]t.NodeAddress
	switch strings.ToLower(args.Net) {
	case "grpc":
		transport, err = grpc.NewTransport(args.OwnID, initialMembership[args.OwnID], logger)
		nodeAddrs = initialMembership
	case "libp2p":
		h := libp2ptools.NewDummyHost(ownNumericID, initialMembership[args.OwnID])
		nodeAddrs, err = membership.DummyMultiAddrs(initialMembership)
		if err != nil {
			return fmt.Errorf("could not generate libp2p multiaddresses: %w", err)
		}
		transport, err = libp2p.NewTransport(h, args.OwnID, logger)
	default:
		return fmt.Errorf("unknown network transport %s", strings.ToLower(args.Net))
	}
	if err != nil {
		return fmt.Errorf("failed to get network transport %w", err)
	}

	if err := transport.Start(); err != nil {
		return fmt.Errorf("could not start network transport: %w", err)
	}
	transport.Connect(ctx, nodeAddrs)

	chatApp := NewChatApp(initialMembership, transport)

	issConfig := iss.DefaultConfig(initialMembership)
	issConfig.ConfigOffset = configOffset
	issProtocol, err := iss.New(
		args.OwnID,
		iss.DefaultModuleConfig(), issConfig, iss.InitialStateSnapshot(chatApp.serializeMessages(), issConfig),
		logger,
	)
	if err != nil {
		return fmt.Errorf("could not instantiate ISS protocol module: %w", err)
	}

	// Use a simple mempool for incoming requests.
	mempool := simplemempool.NewModule(
		&simplemempool.ModuleConfig{
			Self:   "mempool",
			Hasher: "hasher",
		},
		&simplemempool.ModuleParams{
			MaxTransactionsInBatch: 10,
		},
	)

	// Use fake batch database.
	batchdb := fakebatchdb.NewModule(
		&fakebatchdb.ModuleConfig{
			Self: "batchdb",
		},
	)

	availability := multisigcollector.NewReconfigurableModule(
		&multisigcollector.ModuleConfig{
			Self:    "availability",
			Mempool: "mempool",
			BatchDB: "batchdb",
			Net:     "net",
			Crypto:  "crypto",
		},
		args.OwnID,
		logger,
	)

	// The batch fetcher receives availability certificates from the ordering protocol (ISS),
	// retrieves the corresponding transaction batches, and delivers them to the application.
	batchFetcher := batchfetcher.NewModule(batchfetcher.DefaultModuleConfig())

	// ================================================================================
	// Create a Mir Node, attaching the ChatApp implementation and other modules.
	// ================================================================================

	// Create a Mir Node, using a default configuration and passing the modules initialized just above.
	modulesWithDefaults, err := iss.DefaultModules(map[t.ModuleID]modules.Module{
		"net":          transport,
		"iss":          issProtocol,
		"batchfetcher": batchFetcher,

		// This is the application logic Mir is going to deliver requests to.
		// For the implementation of the application, see app.go.
		"app": chatApp,

		// Use dummy crypto module that only produces signatures
		// consisting of a single zero byte and treats those signatures as valid.
		// TODO: Remove this line once a default crypto implementation is provided by Mir.
		"crypto": mirCrypto.New(&mirCrypto.DummyCrypto{DummySig: []byte{0}}),

		"mempool":      mempool,
		"batchdb":      batchdb,
		"availability": availability,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize Mir modules: %w", err)
	}

	node, err := mir.NewNode(args.OwnID, &mir.NodeConfig{Logger: logger}, modulesWithDefaults, nil, nil)
	if err != nil {
		return fmt.Errorf("could not create node: %w", err)
	}

	// ================================================================================
	// Start the Node by establishing network connections and launching necessary processing threads
	// ================================================================================

	// Initialize variables to synchronize Node startup and shutdown.
	var nodeErr error // The error returned from running the Node will be stored here.

	// Start the node in a separate goroutine
	go func() {
		nodeErr = node.Run(ctx)
	}()

	// ================================================================================
	// Read chat messages from stdin and submit them as requests.
	// ================================================================================

	scanner := bufio.NewScanner(os.Stdin)

	// Prompt for chat message input.
	fmt.Println("Type in your messages and press 'Enter' to send.")

	// Read chat message from stdin.
	nextReqNo := t.ReqNo(0)
	for scanner.Scan() {

		// Submit the chat message as request payload.
		err := node.InjectEvents(ctx, events.ListOf(events.NewClientRequests(
			"mempool",
			[]*requestpb.Request{events.ClientRequest(t.ClientID(args.OwnID), nextReqNo, scanner.Bytes())})),
		)

		// Print error if occurred.
		if err != nil {
			fmt.Println(err)
		} else {
			nextReqNo++
		}

	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}

	// ================================================================================
	// Shut down.
	// ================================================================================

	// After sending a few messages, we disconnect the client,
	if args.Verbose {
		fmt.Println("Stopping transport.")
	}

	// stop the gRPC transport,
	transport.Stop()

	// stop the server,
	if args.Verbose {
		fmt.Println("Stopping server.")
	}
	node.Stop()

	return nodeErr
}

// Parses the command-line arguments and returns them in a params struct.
func parseArgs(args []string) *parsedArgs {
	app := kingpin.New("chat-demo", "Small chat application to demonstrate the usage of the Mir library.")
	verbose := app.Flag("verbose", "Verbose mode.").Short('v').Bool()
	trace := app.Flag("trace", "Very verbose mode.").Bool()
	// Currently, the type of the node ID is defined as uint64 by the /pkg/types package.
	// In case that changes, this line will need to be updated.
	n := app.Flag("net", "Network transport.").Short('n').Default("libp2p").String()
	ownID := app.Arg("id", "ID of this node").Required().String()
	initMembershipFile := app.Flag("init-membership", "File containing the initial system membership.").Short('i').Required().File()

	if _, err := app.Parse(args[1:]); err != nil { // Skip args[0], which is the name of the program, not an argument.
		app.FatalUsage("could not parse arguments: %v\n", err)
	}

	return &parsedArgs{
		OwnID:              t.NodeID(*ownID),
		Verbose:            *verbose,
		Trace:              *trace,
		Net:                *n,
		InitMembershipFile: *initMembershipFile,
	}
}
