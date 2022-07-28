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
	"crypto"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/filecoin-project/mir"
	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/dummyclient"
	"github.com/filecoin-project/mir/pkg/iss"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/net/grpc"
	"github.com/filecoin-project/mir/pkg/net/libp2p"
	"github.com/filecoin-project/mir/pkg/requestreceiver"
	t "github.com/filecoin-project/mir/pkg/types"
	grpctools "github.com/filecoin-project/mir/pkg/util/grpc"
	libp2ptools "github.com/filecoin-project/mir/pkg/util/libp2p"
)

const (

	// Base port number for the nodes to listen to messages from each other.
	// The nodes will listen on ports starting from nodeBasePort through nodeBasePort+3.
	// Note that protocol messages and requests are following two completely distinct paths to avoid interference
	// of clients with node-to-node communication.
	nodeBasePort = 10000

	// Base port number for the node request receivers to listen to messages from clients.
	// Request receivers will listen on port reqReceiverBasePort through reqReceiverBasePort+3.
	// Note that protocol messages and requests are following two completely distinct paths to avoid interference
	// of clients with node-to-node communication.
	reqReceiverBasePort = 20000

	// The number of nodes participating in the chat.
	nodeNumber = 4
)

// parsedArgs represents parsed command-line parameters passed to the program.
type parsedArgs struct {

	// ID of this node.
	// The package github.com/hyperledger-labs/mir/pkg/types defines this and other types used by the library.
	OwnID t.NodeID

	// If set, print verbose output to stdout.
	Verbose bool

	// Network transport.
	Net string
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
	if args.Verbose {
		logger = logging.ConsoleDebugLogger // Print debug-level info in verbose mode.
	} else {
		logger = logging.ConsoleWarnLogger // Only print errors and warnings by default.
	}

	ctx := context.Background()

	fmt.Println("Initializing...")

	ownID, err := strconv.Atoi(string(args.OwnID))
	if err != nil {
		return fmt.Errorf("unable to convert node ID: %w", err)
	}
	if ownID < 0 || ownID > nodeNumber-1 {
		return fmt.Errorf("ID must be in [0, %d]", nodeNumber-1)
	}

	// ================================================================================
	// Generate system membership info: addresses, ports, etc...
	// ================================================================================

	// IDs of nodes that are part of the system.
	// This example uses a static configuration of nodeNumber nodes.
	nodeIDs := make([]t.NodeID, nodeNumber)
	for i := 0; i < nodeNumber; i++ {
		nodeIDs[i] = t.NewNodeIDFromInt(i)
	}

	// Generate addresses and ports for client request receivers.
	// Each node uses different ports for receiving protocol messages and requests.
	// These addresses will be used by the client code to know where to send its requests
	// (each client sends its requests to all request receivers). Each request receiver,
	// however, will only submit the received requests to its associated Node.
	reqReceiverAddrs := make(map[t.NodeID]string)
	for i := range nodeIDs {
		reqReceiverAddrs[t.NewNodeIDFromInt(i)] = fmt.Sprintf("127.0.0.1:%d", reqReceiverBasePort+i)
	}

	// ================================================================================
	// Create and initialize various modules used by mir.
	// ================================================================================

	// Initialize the networking module.
	// Mir will use it for transporting nod-to-node messages.

	// In the current implementation, all nodes are on the local machine, but listen on different port numbers.
	// Change this or make this configurable to deploy different nodes on different physical machines.
	var transport net.Transport
	nodeAddrs := make(map[t.NodeID]t.NodeAddress)
	switch strings.ToLower(args.Net) {
	case "grpc":
		for i := range nodeIDs {
			nodeAddrs[t.NewNodeIDFromInt(i)] = t.NodeAddress(grpctools.NewDummyMultiaddr(i + nodeBasePort))
		}
		transport, err = grpc.NewTransport(args.OwnID, nodeAddrs[args.OwnID], logger)
	case "libp2p":
		h := libp2ptools.NewDummyHost(ownID, nodeBasePort)
		for i := range nodeIDs {
			nodeAddrs[t.NewNodeIDFromInt(i)] = t.NodeAddress(libp2ptools.NewDummyMultiaddr(i, nodeBasePort))
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

	// Instantiate the ISS protocol module with default configuration.
	issConfig := iss.DefaultConfig(nodeIDs)
	issProtocol, err := iss.New(args.OwnID, issConfig, logger)
	if err != nil {
		return fmt.Errorf("could not instantiate ISS protocol module: %w", err)
	}

	// ================================================================================
	// Create a Mir Node, attaching the ChatApp implementation and other modules.
	// ================================================================================

	// Create a Mir Node, using a default configuration and passing the modules initialized just above.
	modulesWithDefaults, err := iss.DefaultModules(map[t.ModuleID]modules.Module{
		"net": transport,
		"iss": issProtocol,

		// This is the application logic Mir is going to deliver requests to.
		// For the implementation of the application, see app.go.
		"app": NewChatApp(),

		// Use dummy crypto module that only produces signatures
		// consisting of a single zero byte and treats those signatures as valid.
		// TODO: Remove this line once a default crypto implementation is provided by Mir.
		"crypto": mirCrypto.New(&mirCrypto.DummyCrypto{DummySig: []byte{0}}),
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

	// Create a request receiver and start receiving requests.
	// Note that the RequestReceiver is _not_ part of the Node as its module.
	// It is external to the Node and only submits requests it receives to the node.
	reqReceiver := requestreceiver.NewRequestReceiver(node, "iss", logger)
	if err := reqReceiver.Start(reqReceiverBasePort + ownID); err != nil {
		return fmt.Errorf("could not start request receiver: %w", err)
	}

	// ================================================================================
	// Create a dummy client for submitting requests (chat messages) to the system.
	// ================================================================================

	// Create a DummyClient. In this example, the client's ID corresponds to the ID of the node it is collocated with,
	// but in general this need not be the case.
	// Also note that the client IDs are in a different namespace than Node IDs.
	// The client also needs to be initialized with a Hasher and MirModule module in order to be able to sign requests.
	// We use a dummy MirModule module set up the same way as the Node's MirModule module,
	// so the client's signatures are accepted.
	client := dummyclient.NewDummyClient(
		t.ClientID(args.OwnID),
		crypto.SHA256,
		logger,
	)

	// Create network connections to all Nodes' request receivers.
	// We use just the background context in this demo app, expecting that the connection will succeed
	// and the Connect() function will return. In a real deployment, the passed context
	// can be used for failure handling, for example to cancel connecting.
	client.Connect(ctx, reqReceiverAddrs)

	// ================================================================================
	// Read chat messages from stdin and submit them as requests.
	// ================================================================================

	scanner := bufio.NewScanner(os.Stdin)

	// Prompt for chat message input.
	fmt.Println("Type in your messages and press 'Enter' to send.")

	// Read chat message from stdin.
	for scanner.Scan() {
		// Submit the chat message as request payload.
		if err := client.SubmitRequest(
			scanner.Bytes(),
		); err != nil {
			fmt.Println(err)
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
		fmt.Println("Done sending messages.")
	}
	client.Disconnect()

	// stop the request receiver,
	reqReceiver.Stop()

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
	// Currently, the type of the node ID is defined as uint64 by the /pkg/types package.
	// In case that changes, this line will need to be updated.
	n := app.Flag("net", "Network transport.").Short('n').Default("libp2p").String()
	ownID := app.Arg("id", "ID of this node").Required().String()

	if _, err := app.Parse(args[1:]); err != nil { // Skip args[0], which is the name of the program, not an argument.
		app.FatalUsage("could not parse arguments: %v\n", err)
	}

	return &parsedArgs{
		OwnID:   t.NodeID(*ownID),
		Verbose: *verbose,
		Net:     *n,
	}
}
