package main

import (
	"bufio"
	"context"
	"crypto"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/checkpoint"
	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/iss"
	issconfig "github.com/filecoin-project/mir/pkg/iss/config"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/recordingpb"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/trantor/appmodule"
	"github.com/filecoin-project/mir/pkg/util/libp2p"
	"github.com/filecoin-project/mir/stdmodules/timer"
	"github.com/filecoin-project/mir/stdtypes"
)

// debug extracts events from the event log entries and submits them to a node instance
// that it creates for this purpose.
func debug(args *arguments) error {

	// TODO: Empty addresses might not work here, as they will probably produce wrong snapshots.
	membership := &trantorpbtypes.Membership{make(map[stdtypes.NodeID]*trantorpbtypes.NodeIdentity)} // nolint:govet
	for _, nID := range args.membership {
		membership.Nodes[nID] = &trantorpbtypes.NodeIdentity{
			Id:     nID,
			Addr:   libp2p.NewDummyHostAddr(0, 0).String(),
			Key:    nil,
			Weight: "1",
		}
	}

	// Create a debugger node and a new Context.
	node, err := debuggerNode(args.ownID, membership)
	if err != nil {
		return err
	}
	ctx, stopNode := context.WithCancel(context.Background())
	defer stopNode()

	// Create channel for node output events and start printing its contents.
	var nodeOutput chan *stdtypes.EventList
	if args.showNodeEvents {
		nodeOutput = make(chan *stdtypes.EventList)
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

	filenames := strings.Split(*args.src, " ")

	// Create a reader for the input event log file.
	for _, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			kingpin.Errorf("Error opening src file", filename, ": ", err)
			continue
		}

		defer func(file *os.File, offset int64, whence int) {
			_, _ = file.Seek(offset, whence) // resets the file offset for successive reading
			if err = file.Close(); err != nil {
				kingpin.Errorf("Error closing src file", filename, ": ", err)
			}
		}(file, 0, 0)

		reader, err := eventlog.NewReader(file)
		if err != nil {
			kingpin.Errorf("Error opening new reader for src file", filename, ": ", err)
			fmt.Printf("\n\n!!!\nContinuing after error with next file\n!!!\n\n")
			continue
		}
		// Process event log.
		// Each entry can contain multiple events.
		var entry *recordingpb.Entry
		for entry, err = reader.ReadEntry(); err == nil; entry, err = reader.ReadEntry() {

			// Create event metadata structure and fill in fields that are common for all events in the log entry.
			// This structure is modified several times and used as a value parameter.
			metadata := eventMetadata{
				nodeID: stdtypes.NodeID(entry.NodeId),
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
				if err := node.InjectEvents(ctx, stdtypes.ListOf(event)); err != nil {
					kingpin.Errorf("Error injecting events of src file", filename, ": ", err)
					fmt.Printf("\n\n!!!\nContinuing after error with next file\n!!!\n\n")
					continue
				}

				// Increment position of the event in the log (to be used with the next event).
				index++
				if !errors.Is(err, io.EOF) {
					kingpin.Errorf("Error reading event log for src file", filename, ": ", err)
					fmt.Printf("\n\n!!!\nContinuing after error with next file\n!!!\n\n")
					continue
				}
			}
		}
	}

	fmt.Println("End of trace, done debugging. Press Enter to exit.")
	bufio.NewScanner(os.Stdin).Scan()

	return nil //TODO Wrap non-blocking errors and return here
}

// debuggerNode creates a new Mir node instance to be used for debugging.
func debuggerNode(id stdtypes.NodeID, membership *trantorpbtypes.Membership) (*mir.Node, error) {

	// Logger used by the node.
	logger := logging.ConsoleDebugLogger

	cryptoImpl := &mirCrypto.DummyCrypto{DummySig: []byte{0}}

	// Instantiate an ISS protocol module with the default configuration.
	// TODO: The initial app state must be involved here. Otherwise checkpoint hashes might not match.
	issConfig := issconfig.DefaultParams(membership)
	stateSnapshotpb, err := iss.InitialStateSnapshot([]byte{}, issConfig)
	if err != nil {
		return nil, err
	}
	protocol, err := iss.New(
		id,
		iss.ModuleConfig{
			Self:         "iss",
			App:          "batchfetcher",
			Availability: "availability",
			BatchDB:      "batchdb",
			Checkpoint:   "checkpointing",
			Net:          "net",
			Ordering:     "ordering",
			Timer:        "timer",
		},
		issConfig,
		checkpoint.Genesis(stateSnapshotpb),
		crypto.SHA256,
		cryptoImpl,
		logging.Decorate(logger, "ISS: "),
	)
	if err != nil {
		return nil, fmt.Errorf("could not instantiate protocol module: %w", err)
	}

	nullTransport := &NullTransport{}

	// Instantiate and return a minimal Mir Node.
	nodeModules := map[stdtypes.ModuleID]modules.Module{
		"net":    nullTransport,
		"crypto": mirCrypto.New(cryptoImpl),
		"app": appmodule.NewAppModule(
			appmodule.AppLogicFromStatic(
				deploytest.NewFakeApp(),
				&trantorpbtypes.Membership{make(map[stdtypes.NodeID]*trantorpbtypes.NodeIdentity)}, // nolint:govet
			),
			nullTransport,
			"iss",
		),
		"iss":   protocol,
		"timer": timer.New(),
	}
	if err != nil {
		panic(fmt.Errorf("error initializing the Mir modules: %w", err))
	}

	node, err := mir.NewNode(id, mir.DefaultNodeConfig().WithLogger(logger), nodeModules, nil)
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
func printNodeOutput(eventChan chan *stdtypes.EventList) {
	for receivedEvents, ok := <-eventChan; ok; receivedEvents, ok = <-eventChan {
		fmt.Printf("========================================\n")
		fmt.Printf("Node produced the following events:\n\n")
		for _, event := range receivedEvents.Slice() {
			fmt.Println(protojson.Format(event.(*eventpb.Event)))
		}
	}
}
