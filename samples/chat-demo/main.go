/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// ********************************************************************************
//                                                                               //
//         Chat demo application for demonstrating the usage of Mir              //
//                             (main executable)                                 //
//                                                                               //
// ********************************************************************************

package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/filecoin-project/mir/pkg/deploytest"

	"github.com/filecoin-project/mir/pkg/systems/smr"
	"github.com/filecoin-project/mir/pkg/util/errstack"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/membership"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/libp2p"
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

	// Network transport type
	Net string

	// Name of the file containing the initial membership for joining nodes.
	InitMembershipFile string
}

// main is just the wrapper for executing the run() and printing a potential error.
func main() {
	if err := run(); err != nil {
		errstack.Println(err)
		os.Exit(1)
	}
}

// run is the actual code of the program.
func run() error {

	// ================================================================================
	// Basic initialization and configuration
	// ================================================================================

	// Convenience variables
	var err error
	ctx := context.Background()

	// Parse command-line parameters.
	args := parseArgs(os.Args)

	// Initialize logger that will be used throughout the code to print log messages.
	var logger logging.Logger
	if args.Trace {
		logger = logging.ConsoleTraceLogger // Print trace-level info.
	} else if args.Verbose {
		logger = logging.ConsoleDebugLogger // Print debug-level info in verbose mode.
	} else {
		logger = logging.ConsoleWarnLogger // Only print errors and warnings by default.
	}

	fmt.Println("Initializing...")

	// ================================================================================
	// Load system membership info: IDs, addresses, ports, etc...
	// ================================================================================

	// For the dummy chat application, we require node IDs to be numeric,
	// as other metadata is derived from node IDs.
	ownNumericID, err := strconv.Atoi(string(args.OwnID))
	if err != nil {
		return errors.Wrap(err, "node IDs must be numeric in the sample app")
	}

	// Load initial system membership from the file indicated through the command line.
	initialAddrs, err := membership.FromFileName(args.InitMembershipFile)
	if err != nil {
		return errors.Wrap(err, "could not load membership")
	}
	initialMembership, err := membership.DummyMultiAddrs(initialAddrs)
	if err != nil {
		return errors.Wrap(err, "could not create dummy multiaddrs")
	}

	// ================================================================================
	// Instantiate the Mir node with the appropriate set of modules.
	// ================================================================================

	// Assemble listening address.
	// In this demo code, we always listen on tha address 0.0.0.0.
	portStr, err := getPortStr(initialMembership[args.OwnID])
	if err != nil {
		return fmt.Errorf("could not parse port from own address: %w", err)
	}
	addrStr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", portStr)
	listenAddr, err := multiaddr.NewMultiaddr(addrStr)
	if err != nil {
		return fmt.Errorf("could not create listen address: %w", err)
	}

	// Create a dummy libp2p host for network communication (this is why we need a numeric ID)
	h, err := libp2p.NewDummyHostWithPrivKey(
		t.NodeAddress(libp2p.NewDummyMultiaddr(ownNumericID, listenAddr)),
		libp2p.NewDummyHostKey(ownNumericID),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create libp2p host")
	}

	// Create a dummy crypto implementation that locally generates all keys in a pseudo-random manner.
	localCrypto := deploytest.NewLocalCryptoSystem("pseudo", membership.GetIDs(initialMembership), logger)

	// Create a Mir SMR system.
	smrSystem, err := smr.New(
		args.OwnID,
		h,
		initialMembership,
		localCrypto.Crypto(args.OwnID),
		NewChatApp(initialMembership),
		smr.DefaultParams(initialMembership),
		logger,
	)
	if err != nil {
		return errors.Wrap(err, "could not create SMR system")
	}

	// Create a Mir node, passing it all the modules of the SMR system.
	node, err := mir.NewNode(args.OwnID, mir.DefaultNodeConfig().WithLogger(logger), smrSystem.Modules(), nil, nil)
	if err != nil {
		return errors.Wrap(err, "could not create node")
	}

	// ================================================================================
	// Start the Node by launching necessary processing threads.
	// ================================================================================

	// Start the SMR system.
	// This will start all the goroutines that need to run within the modules of the SMR system.
	// For example, the network module will start listening for incoming connections and create outgoing ones.
	// The modules will become ready to be used by the node (but the node itself is not yet started).
	if err := smrSystem.Start(ctx); err != nil {
		return errors.Wrap(err, "could not start SMR system")
	}

	// Start the node in a separate goroutine
	var nodeErr error // The error returned from running the Node will be stored here.
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

		// Submit the chat message as request payload to the mempool module.
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

	// Stop the system.
	if args.Verbose {
		fmt.Println("Stopping SMR system.")
	}
	smrSystem.Stop()

	// Stop the node.
	if args.Verbose {
		fmt.Println("Stopping node.")
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
	ownID := app.Arg("id", "ID of this node").Required().String()
	initMembershipFile := app.Flag("init-membership", "File containing the initial system membership.").
		Short('i').Required().String()

	if _, err := app.Parse(args[1:]); err != nil { // Skip args[0], which is the name of the program, not an argument.
		app.FatalUsage("could not parse arguments: %v\n", err)
	}

	return &parsedArgs{
		OwnID:              t.NodeID(*ownID),
		Verbose:            *verbose,
		Trace:              *trace,
		InitMembershipFile: *initMembershipFile,
	}
}

func getPortStr(address t.NodeAddress) (string, error) {
	_, addrStr, err := manet.DialArgs(address)
	if err != nil {
		return "", err
	}

	_, portStr, err := net.SplitHostPort(addrStr)
	if err != nil {
		return "", err
	}

	return portStr, nil
}
