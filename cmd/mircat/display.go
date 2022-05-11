package main

//handles the processing, display and retrieval of events from a given eventlog file

import (
	"fmt"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/ttacon/chalk"
	"google.golang.org/protobuf/encoding/protojson"
	"os"
	"strconv"
)

// extracts events from eventlog entries and
// forwards them for display
func displayEvents(srcFile *os.File, events map[string]struct{}, issEvents map[string]struct{}, offset int) error {

	//new reader
	reader, err := eventlog.NewReader(srcFile)

	if err != nil {
		return err
	}

	index := uint64(0) // a counter set to track the log indices

	for entry, err := reader.ReadEntry(); err == nil; entry, err = reader.ReadEntry() {
		metadata := eventMetadata{
			nodeID: t.NodeID(entry.NodeId),
			time:   entry.Time,
		}
		//getting events from entry
		for _, event := range entry.Events {
			metadata.index = index

			if _, ok := events[eventName(event)]; ok && index >= uint64(offset) {
				// If event type has been selected for displaying

				switch e := event.Type.(type) {
				case *eventpb.Event_Iss:
					// Only display selected sub-types of the ISS Event
					if _, ok := issEvents[issEventName(e.Iss)]; ok {
						displayEvent(event, metadata)
					}
				default:
					displayEvent(event, metadata)
				}
			}

			index++
		}
	}
	return nil
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

//Creates and returns a prefix tag for event display using event metadata
func getMetaTag(eventType string, metadata eventMetadata) string {
	boldGreen := chalk.Green.NewStyle().WithTextStyle(chalk.Bold) //setting font color and style
	boldCyan := chalk.Cyan.NewStyle().WithTextStyle(chalk.Bold)
	return fmt.Sprintf("%s %s",
		boldGreen.Style(fmt.Sprintf("[ Event_%s ]", eventType)),
		boldCyan.Style(fmt.Sprintf("[ Node #%v ] [ Time _%s ] [ Index #%s ]",
			metadata.nodeID,
			strconv.FormatInt(metadata.time, 10),
			strconv.FormatUint(metadata.index, 10))),
	)
}

//displays the event
func display(eventType string, event string, metadata eventMetadata) {
	whiteText := chalk.White.NewStyle().WithTextStyle(chalk.Bold)
	metaTag := getMetaTag(eventType, metadata)
	fmt.Printf("%s\n%s \n", metaTag, whiteText.Style(event))
}
