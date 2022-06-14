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
	"sync"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/filecoin-project/mir"
	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/dummyclient"
	"github.com/filecoin-project/mir/pkg/grpctransport"
	"github.com/filecoin-project/mir/pkg/iss"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/reqstore"
	"github.com/filecoin-project/mir/pkg/requestreceiver"
	t "github.com/filecoin-project/mir/pkg/types"
)

const (

	// Base port number for the nodes to listen to messages from each other.
	// Each node will listen on a port computed by adding its numeric ID to nodeBasePort.
	// Note that protocol messages and requests are following two completely distinct paths to avoid interference
	// of clients with node-to-node communication.
	nodeBasePort = 10000

	// Base port number for the node request receivers to listen to messages from clients.
	// Each request receiver will listen on a port computed by adding its node's numeric ID to reqReceiverBasePort.
	// Note that protocol messages and requests are following two completely distinct paths to avoid interference
	// of clients with node-to-node communication.
	reqReceiverBasePort = 20000
)

// parsedArgs represents parsed command-line parameters passed to the program.
type parsedArgs struct {

	// Numeric ID of this node.
	// The package github.com/hyperledger-labs/mir/pkg/types defines this and other types used by the library.
	OwnID t.NodeID

	// If set, print verbose output to stdout.
	Verbose bool
}

func main() {

	// Parse command-line parameters.
	_ = parseArgs(os.Args)
	args := parseArgs(os.Args)

	// Initialize logger that will be used throughout the code to print log messages.
	var logger logging.Logger
	if args.Verbose {
		logger = logging.ConsoleDebugLogger // Print debug-level info in verbose mode.
	} else {
		logger = logging.ConsoleWarnLogger // Only print errors and warnings by default.
	}

	fmt.Println("Initializing...")

	// ================================================================================
	// Generate system membership info: addresses, ports, etc...
	// ================================================================================

	// IDs of nodes that are part of the system.
	// This example uses a static configuration of 4 nodes.
	nodeIds := []t.NodeID{
		t.NewNodeIDFromInt(0),
		t.NewNodeIDFromInt(1),
		t.NewNodeIDFromInt(2),
		t.NewNodeIDFromInt(3),
	}

	// Generate addresses and ports of participating nodes.
	// All nodes are on the local machine, but listen on different port numbers.
	// Change this or make this configurable do deploy different nodes on different physical machines.
	nodeAddrs := make(map[t.NodeID]string)
	for i := range nodeIds {
		nodeAddrs[t.NewNodeIDFromInt(i)] = fmt.Sprintf("127.0.0.1:%d", nodeBasePort+i)
	}

	// Generate addresses and ports for client request receivers.
	// Each node uses different ports for receiving protocol messages and requests.
	// These addresses will be used by the client code to know where to send its requests
	// (each client sends its requests to all request receivers). Each request receiver,
	// however, will only submit the received requests to its associated Node.
	reqReceiverAddrs := make(map[t.NodeID]string)
	for i := range nodeIds {
		reqReceiverAddrs[t.NewNodeIDFromInt(i)] = fmt.Sprintf("127.0.0.1:%d", reqReceiverBasePort+i)
	}

	// ================================================================================
	// Create and initialize various modules used by mir.
	// ================================================================================

	// Initialize the networking module.
	// Mir will use it for transporting nod-to-node messages.
	net := grpctransport.NewGrpcTransport(nodeAddrs, args.OwnID, nil)
	if err := net.Start(); err != nil {
		panic(err)
	}
	net.Connect(context.Background())

	// Create a new request store. Request payloads will be stored in it.
	// Generally, the request store should be a persistent one,
	// but for this dummy example we use a simple in-memory implementation,
	// as restarts and crash-recovery (where persistence is necessary) are not yet implemented anyway.
	reqStore := reqstore.NewVolatileRequestStore()

	// Instantiate the ISS protocol module with default configuration.
	issConfig := iss.DefaultConfig(nodeIds)
	issProtocol, err := iss.New(args.OwnID, issConfig, logger)
	if err != nil {
		panic(fmt.Errorf("could not instantiate ISS protocol module: %w", err))
	}

	// ================================================================================
	// Create a Mir Node, attaching the ChatApp implementation and other modules.
	// ================================================================================

	// Create a Mir Node, using a default configuration and passing the modules initialized just above.
	node, err := mir.NewNode(args.OwnID, &mir.NodeConfig{Logger: logger}, map[t.ModuleID]modules.Module{
		"net":          net,
		"requestStore": reqStore,
		"iss":          issProtocol,

		// This is the application logic Mir is going to deliver requests to.
		// It requires to have access to the request store, as Mir only passes request references to it.
		// It is the application's responsibility to get the necessary request data from the request store.
		// For the implementation of the application, see app.go.
		"app": NewChatApp(reqStore),

		// Use dummy crypto module that only produces signatures
		// consisting of a single zero byte and treats those signatures as valid.
		// TODO: Remove this line once a default crypto implementation is provided by Mir.
		"crypto": mirCrypto.New(&mirCrypto.DummyCrypto{DummySig: []byte{0}}),
	},
		nil)

	// Exit immediately if Node could not be created.
	if err != nil {
		fmt.Printf("Could not create node: %v\n", err)
		os.Exit(1)
	}

	// ================================================================================
	// Start the Node by establishing network connections and launching necessary processing threads
	// ================================================================================

	// Initialize variables to synchronize Node startup and shutdown.
	ctx := context.Background() // Calling Done() on this context will signal the Node to stop.
	var nodeErr error           // The error returned from running the Node will be stored here.
	var wg sync.WaitGroup       // The Node will call Done() on this WaitGroup when it actually stops.
	wg.Add(1)

	// Start the node in a separate goroutine
	go func() {
		nodeErr = node.Run(ctx)
		wg.Done()
	}()

	// Create a request receiver and start receiving requests.
	// Note that the RequestReceiver is _not_ part of the Node as its module.
	// It is external to the Node and only submits requests it receives to the node.
	reqReceiver := requestreceiver.NewRequestReceiver(node, logger)
	p, err := strconv.Atoi(string(args.OwnID))
	if err != nil {
		panic(fmt.Errorf("could not convert node ID: %w", err))
	}
	if err := reqReceiver.Start(reqReceiverBasePort + p); err != nil {
		panic(err)
	}

	// ================================================================================
	// Create a dummy client for submitting requests (chat messages) to the system.
	// ================================================================================

	// Create a DummyClient. In this example, the client's ID corresponds to the ID of the node it is collocated with,
	// but in general this need not be the case.
	// Also note that the client IDs are in a different namespace than Node IDs.
	// The client also needs to be initialized with a Hasher and Crypto module in order to be able to sign requests.
	// We use a dummy Crypto module set up the same way as the Node's Crypto module,
	// so the client's signatures are accepted.
	client := dummyclient.NewDummyClient(
		t.ClientID(args.OwnID),
		crypto.SHA256,
		&mirCrypto.DummyCrypto{DummySig: []byte{0}},
		logger,
	)

	// Create network connections to all Nodes' request receivers.
	// We use just the background context in this demo app, expecting that the connection will succeed
	// and the Connect() function will return. In a real deployment, the passed context
	// can be used for failure handling, for example to cancel connecting.
	client.Connect(context.Background(), reqReceiverAddrs)

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

	// stop the server,
	if args.Verbose {
		fmt.Println("Stopping server.")
	}
	ctx.Done()
	wg.Wait()

	// and print the error returned by the stopped node.
	fmt.Printf("Node error: %v\n", nodeErr)
}

// Parses the command-line arguments and returns them in a params struct.
func parseArgs(args []string) *parsedArgs {
	app := kingpin.New("chat-demo", "Small chat application to demonstrate the usage of the Mir library.")
	verbose := app.Flag("verbose", "Verbose mode.").Short('v').Bool()
	// Currently the type of the node ID is defined as uint64 by the /pkg/types package.
	// In case that changes, this line will need to be updated.
	ownID := app.Arg("id", "Numeric ID of this node").Required().String()

	if _, err := app.Parse(args[1:]); err != nil { // Skip args[0], which is the name of the program, not an argument.
		app.FatalUsage("could not parse arguments: %v\n", err)
	}

	return &parsedArgs{
		OwnID:   t.NodeID(*ownID),
		Verbose: *verbose,
	}
}
