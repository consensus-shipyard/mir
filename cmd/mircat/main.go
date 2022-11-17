package main

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"

	t "github.com/filecoin-project/mir/pkg/types"
)

// mircat is a tool for reviewing Mir state machine recordings.
// It understands the format encoded via github.com/filecoin-project/mir/eventlog
// and is able to parse and filter these log files based on the events.

// arguments represents the parameters passed to mircat.
type arguments struct {

	// File containing the event log to read.
	srcFile *os.File

	// Number of events at the start of the event log to be skipped (including the ones not selected).
	// If debugging, the skipped events will still be passed to the node before the interactive debugging starts.
	offset int

	// The number of events to display.
	// When debugging, this is the maximum number of events to pause before submitting to the node.
	// All remaining events will simply be flushed without user prompt.
	limit int

	// Events selected by the user for displaying.
	selectedEventNames map[string]struct{}

	// If ISS Events have been selected for displaying, this variable contains the types of ISS events to be displayed.
	selectedIssEventNames map[string]struct{}

	// Events with specific destination modules selected by the user for displaying
	selectedEventDests map[string]struct{}

	// If set to true, start a Node in debug mode with the given event log.
	debug bool

	// The rest of the fields are only used in debug mode and are otherwise ignored.

	// The ID of the node being debugged.
	// It must correspond to the ID of the node that produced the event log being read.
	ownID t.NodeID

	// IDs, in order, of all nodes in the deployment that produced the event log.
	membership []t.NodeID

	// If set to true, the events produced by the node are printed to standard output.
	// Otherwise, they are simply dropped.
	showNodeEvents bool
}

func main() {

	// Parse command-line arguments
	kingpin.Version("0.0.1")
	args, err := parseArgs(os.Args[1:])
	if err != nil {
		kingpin.Fatalf("Cannot parse given argument", err)
	}

	// Scan the event log and collect all occurring event types.
	fmt.Println("Scanning input file.")
	allEvents, allISSEvents, allDests, totalEvents, err := getEventList(args.srcFile)
	if err != nil {
		kingpin.Errorf("Error parsing src file", err)
		fmt.Printf("\n\n!!!\nContinuing after error. Event list might be incomplete!\n!!!\n\n")
	}
	fmt.Printf("Total number of events found: %d\n", totalEvents)

	// If no event types have been selected through command-line arguments,
	// have the user interactively select the events to include in the output.
	if len(args.selectedEventNames) == 0 {

		// Select top-level events
		args.selectedEventNames = checkboxes("Please select the events", allEvents)

		// If any ISS events occur in the event log and the user selected the ISS event type,
		// have the user select which of those should be included in the output.
		if _, ok := args.selectedEventNames["Iss"]; ok {
			args.selectedIssEventNames = checkboxes("Please select the ISS events", allISSEvents)
		}
	}

	// // If no event destinations have been selected through command-line arguments,
	// // have the user interactively select the event destinations' to include in the output.
	if len(args.selectedEventDests) == 0 {

		// Select top-level events
		args.selectedEventDests = checkboxes("Please select the event destinations", allDests)

	}

	fmt.Println("Command-line arguments for selecting the chosen filters:\n" +
		selectionArgs(args.selectedEventNames, args.selectedIssEventNames, args.selectedEventDests))

	// Display selected events or enter debug mode.
	if args.debug {
		err = debug(args)
		if err != nil {
			kingpin.Errorf("Error debugging node", err)
		}
	} else {
		err = displayEvents(args)
		if err != nil {
			kingpin.Errorf("Error Processing Events", err)
		}
	}

}

// parse the command line arguments
func parseArgs(args []string) (*arguments, error) {
	if len(args) == 0 {
		return nil, errors.Errorf("required input \" --src <Src_File> \" not found")
	}

	app := kingpin.New("mircat", "Utility for processing Mir state event logs.")
	src := app.Flag("src", "The input file to read.").Required().File()
	events := app.Flag("event", "Event types to be displayed.").Short('e').Strings()
	issEvents := app.Flag("iss-event", "Types of ISS Events to be displayed if ISS events are selected.").Short('s').Strings()
	eventDests := app.Flag("event-dest", "Event destination types to be displayed.").Short('r').Strings()
	offset := app.Flag("offset", "The first offset events will not be displayed.").Default("0").Int()
	limit := app.Flag("limit", "Maximum number of events to consider for display or debug").Default("0").Int()
	dbg := app.Flag("debug", "Start a Node in debug mode with the given event log.").Short('d').Bool()
	id := app.Flag("own-id", "ID of the node to use for debugging.").String()
	membership := app.Flag(
		"node-id",
		"ID of one membership node, specified once for each node (debugging only).",
	).Short('m').Strings()
	showNodeEvents := app.Flag("show-node-output", "Show events generated by the node (debugging only)").Bool()

	_, err := app.Parse(args)
	if err != nil {
		return nil, err
	}

	return &arguments{
		srcFile:               *src,
		debug:                 *dbg,
		ownID:                 t.NodeID(*id),
		membership:            t.NodeIDSlice(*membership),
		showNodeEvents:        *showNodeEvents,
		offset:                *offset,
		limit:                 *limit,
		selectedEventNames:    toSet(*events),
		selectedIssEventNames: toSet(*issEvents),
		selectedEventDests:    toSet(*eventDests),
	}, nil
}

// Prompts users with a list of available Events to select from.
// Returns a set of selected Events.
func checkboxes(label string, opts map[string]struct{}) map[string]struct{} {

	// Use survey library to get a list selected event names.
	selected := make([]string, 0)
	prompt := &survey.MultiSelect{
		Message: label,
		Options: toList(opts),
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		fmt.Printf("Error selecting event types: %v", err)
	}

	return toSet(selected)
}

func selectionArgs(events map[string]struct{}, issEvents map[string]struct{}, dests map[string]struct{}) string {
	argStr := ""

	for _, eventName := range toList(events) {
		argStr += " --event " + eventName
	}

	for _, issEventName := range toList(issEvents) {
		argStr += " --iss-event " + issEventName
	}

	for _, dest := range toList(dests) {
		argStr += " --event-dest " + dest
	}

	return argStr
}
