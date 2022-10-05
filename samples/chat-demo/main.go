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
	"os"
	"strconv"

	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/filecoin-project/mir/pkg/systems/smr"
	"github.com/filecoin-project/mir/pkg/util/errstack"

	"github.com/filecoin-project/mir"
	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
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

	// Network transport.
	Net string

	// Name of the file containing the initial membership for joining nodes.
	InitMembershipFile string
}

func main() {
	if err := run(); err != nil {
		errstack.Println(err)
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

	// Create a Mir SMR system.
	h, err := libp2p.NewDummyHostWithPrivKey(
		args.OwnID,
		libp2p.NewDummyHostKey(ownNumericID),
		initialMembership,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create libp2p host: %s")
	}
	smrSystem, err := smr.New(
		args.OwnID,
		h,
		initialMembership,
		&mirCrypto.DummyCrypto{DummySig: []byte{0}},
		NewChatApp(initialMembership),
		logger,
	)
	if err != nil {
		return errors.Wrap(err, "could not create SMR system")
	}

	if err := smrSystem.Start(ctx); err != nil {
		return errors.Wrap(err, "could not start SMR system")
	}

	node, err := mir.NewNode(args.OwnID, &mir.NodeConfig{Logger: logger}, smrSystem.Modules(), nil, nil)
	if err != nil {
		return errors.Wrap(err, "could not create node")
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
	n := app.Flag("net", "Network transport.").Short('n').Default("libp2p").String()
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
		Net:                *n,
		InitMembershipFile: *initMembershipFile,
	}
}
