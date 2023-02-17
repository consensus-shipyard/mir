// Handles the processing, display and retrieval of events from a given eventlog file
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ttacon/chalk"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/recordingpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// extracts events from eventlog entries and
// forwards them for display
func displayEvents(args *arguments) error { //nolint:gocognit
	// TODO: linting is disabled, since this code needs to be substantially re-written anyway.

	// new reader

	// Keep track of the position of events in the recorded log.
	index := 0

	filenames := strings.Split(*args.src, " ")

	// If debugging a module, instantiate it.
	var module modules.PassiveModule
	if args.dbgModule {
		module = customModule()
	}

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

		var entry *recordingpb.Entry
		for entry, err = reader.ReadEntry(); err == nil; entry, err = reader.ReadEntry() {
			metadata := eventMetadata{
				nodeID: t.NodeID(entry.NodeId),
				time:   entry.Time,
			}
			// getting events from entry
			for _, event := range entry.Events {
				metadata.index = uint64(index)

				_, validEvent := args.selectedEventNames[eventName(event)]
				_, validDest := args.selectedEventDests[event.DestModule]
				event = customTransform(event) // Apply custom transformation to event.

				if validEvent && validDest && event != nil && index >= args.offset && (args.limit == 0 || index < args.offset+args.limit) {
					// If event type has been selected for displaying

					switch e := event.Type.(type) {
					case *eventpb.Event_Iss:
						// Only display selected sub-types of the ISS Event
						if _, validIssEvent := args.selectedIssEventNames[issEventName(e.Iss)]; validIssEvent {
							displayEvent(event, metadata)
						}
					default:
						displayEvent(event, metadata)
					}

					if args.dbgModule {
						if _, err := module.ApplyEvents(events.ListOf(event)); err != nil {
							return fmt.Errorf("error replaying event to module: %w", err)
						}
					}
				}

				index++
			}
		}

		if !errors.Is(err, io.EOF) {
			kingpin.Errorf("Error reading event log for src file", filename, ": ", err)
			fmt.Printf("\n\n!!!\nContinuing after error with next file\n!!!\n\n")
			continue
		}
	}
	fmt.Println("End of trace.")

	return nil //TODO Wrap non-blocking errors and return here
}

// Displays one event according to its type.
func displayEvent(event *eventpb.Event, metadata eventMetadata) {

	switch e := event.Type.(type) {
	case *eventpb.Event_Iss:
		display(fmt.Sprintf("%s : %s", eventName(event), issEventName(e.Iss)), protojson.Format(event), metadata)
	default:
		display(eventName(event), protojson.Format(event), metadata)
	}
}

// Creates and returns a prefix tag for event display using event metadata
func getMetaTag(eventType string, metadata eventMetadata) string {
	boldGreen := chalk.Green.NewStyle().WithTextStyle(chalk.Bold) // setting font color and style
	boldCyan := chalk.Cyan.NewStyle().WithTextStyle(chalk.Bold)
	return fmt.Sprintf("%s %s",
		boldGreen.Style(fmt.Sprintf("[ Event_%s ]", eventType)),
		boldCyan.Style(fmt.Sprintf("[ Node #%v ] [ Time _%s ] [ Index #%s ]",
			metadata.nodeID,
			strconv.FormatInt(metadata.time, 10),
			strconv.FormatUint(metadata.index, 10))),
	)
}

// displays the event
func display(eventType string, event string, metadata eventMetadata) {
	whiteText := chalk.White.NewStyle().WithTextStyle(chalk.Bold)
	metaTag := getMetaTag(eventType, metadata)
	fmt.Printf("%s\n%s \n", metaTag, whiteText.Style(event))
}
