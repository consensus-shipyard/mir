package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/checkpoint"
	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/iss"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/recordingpb"
	"github.com/filecoin-project/mir/pkg/systems/trantor"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/issutil"
	"github.com/filecoin-project/mir/pkg/util/libp2p"
)

// debug extracts events from the event log entries and submits them to a node instance
// that it creates for this purpose.
func debug(args *arguments) error {

	// Create a reader for the input event log file.
	reader, err := eventlog.NewReader(args.srcFile)
	if err != nil {
		return err
	}

	// TODO: Empty addresses might not work here, as they will probably produce wrong snapshots.
	membership := make(map[t.NodeID]t.NodeAddress)
	for _, nID := range args.membership {
		membership[nID] = libp2p.NewDummyHostAddr(0, 0)
	}

	// Create a debugger node and a new Context.
	node, err := debuggerNode(args.ownID, membership)
	if err != nil {
		return err
	}
	ctx, stopNode := context.WithCancel(context.Background())
	defer stopNode()

	// Create channel for node output events and start printing its contents.
	var nodeOutput chan *events.EventList
	if args.showNodeEvents {
		nodeOutput = make(chan *events.EventList)
		go printNodeOutput(nodeOutput)
	} else {
		nodeOutput = nil
	}

	// Start the debugger node.
	go func() {
		if err := node.Debug(ctx, nodeOutput); err != nil {
			fmt.Printf("Debugged node stopped with error: %v\n", err)
		}
	}()

	// Keep track of the position of events in the recorded log.
	index := 0

	// Process event log.
	// Each entry can contain multiple events.
	var entry *recordingpb.Entry
	for entry, err = reader.ReadEntry(); err == nil; entry, err = reader.ReadEntry() {

		// Create event metadata structure and fill in fields that are common for all events in the log entry.
		// This structure is modified several times and used as a value parameter.
		metadata := eventMetadata{
			nodeID: t.NodeID(entry.NodeId),
			time:   entry.Time,
		}

		// Process each event in the entry.
		for _, event := range entry.Events {

			// Set the index of the event in the event log.
			metadata.index = uint64(index)

			// If the event was selected by the user for inspection, pause before submitting it to the node.
			// The processing continues after the user's interactive confirmation.
			if selected(event, args.selectedEventNames, args.selectedIssEventNames) &&
				index >= args.offset &&
				(args.limit == 0 || index < args.offset+args.limit) {

				stopBeforeNext(event, metadata)
			}

			// Submit the event to the debugger node.
			if err := node.InjectEvents(ctx, events.ListOf(event)); err != nil {
				return fmt.Errorf("node step failed: %w", err)
			}

			// Increment position of the event in the log (to be used with the next event).
			index++
		}
	}

	if errors.Is(err, io.EOF) {
		return fmt.Errorf("error reading event log: %w", err)
	}

	fmt.Println("End of trace, done debugging. Press Enter to exit.")
	bufio.NewScanner(os.Stdin).Scan()

	return nil
}

// debuggerNode creates a new Mir node instance to be used for debugging.
func debuggerNode(id t.NodeID, membership map[t.NodeID]t.NodeAddress) (*mir.Node, error) {

	// Logger used by the node.
	logger := logging.ConsoleDebugLogger

	// Instantiate an ISS protocol module with the default configuration.
	// TODO: The initial app state must be involved here. Otherwise checkpoint hashes might not match.
	issConfig := issutil.DefaultParams(membership)
	protocol, err := iss.New(
		id,
		iss.DefaultModuleConfig(),
		issConfig,
		checkpoint.Genesis(iss.InitialStateSnapshot([]byte{}, issConfig)),
		logging.Decorate(logger, "ISS: "),
	)
	if err != nil {
		return nil, fmt.Errorf("could not instantiate protocol module: %w", err)
	}

	nullTransport := &NullTransport{}

	// Instantiate and return a minimal Mir Node.
	modulesWithDefaults, err := iss.DefaultModules(map[t.ModuleID]modules.Module{
		"net":    nullTransport,
		"crypto": mirCrypto.New(&mirCrypto.DummyCrypto{DummySig: []byte{0}}),
		"app": trantor.NewAppModule(
			trantor.AppLogicFromStatic(deploytest.NewFakeApp(), map[t.NodeID]t.NodeAddress{}),
			nullTransport,
			"iss",
		),
		"iss": protocol,
	}, iss.DefaultModuleConfig())
	if err != nil {
		panic(fmt.Errorf("error initializing the Mir modules: %w", err))
	}

	node, err := mir.NewNode(id, mir.DefaultNodeConfig().WithLogger(logger), modulesWithDefaults, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("could not instantiate mir node: %w", err)
	}

	return node, nil
}

// stopBeforeNext waits for two confirmations of the user, a confirmation being a new line on the standard input.
// After the first one, event is displayed and after the second one, the function returns.
func stopBeforeNext(event *eventpb.Event, metadata eventMetadata) {
	bufio.NewScanner(os.Stdin).Scan()
	fmt.Printf("========================================\n")
	fmt.Printf("Next step (%d):\n", metadata.index)
	displayEvent(event, metadata)
	bufio.NewScanner(os.Stdin).Scan()
	fmt.Printf("========================================\n")
}

// printNodeOutput reads all events output by a node from the given eventChan channel
// and prints them to standard output.
func printNodeOutput(eventChan chan *events.EventList) {
	for receivedEvents, ok := <-eventChan; ok; receivedEvents, ok = <-eventChan {
		fmt.Printf("========================================\n")
		fmt.Printf("Node produced the following events:\n\n")
		for _, event := range receivedEvents.Slice() {
			fmt.Println(protojson.Format(event))
		}
	}
}
