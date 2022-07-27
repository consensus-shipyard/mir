/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// ********************************************************************************
//         Chat demo application for demonstrating the usage of Mir              //
//                            (application logic)                                //
// ********************************************************************************

package main

import (
	"fmt"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/multiformats/go-multiaddr"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ChatApp and its methods implement the application logic of the small chat demo application
// showcasing the usage of the Mir library.
// An initialized instance of this struct needs to be passed to the mir.NewNode() method of all nodes
// for the system to run the chat demo app.
type ChatApp struct {

	// The only state of the application is the chat message history,
	// to which each delivered request appends one message.
	messages []string

	// Number of batches applied to the state.
	batchesApplied int

	// Length of an ISS segment.
	segmentLength int

	// Length of the epoch of the consensus protocol.
	// After having delivered this many batches, the app signals the end of an epoch.
	epochLength int

	// The application writes an empty struct after having applied all batches of an epoch.
	epochFinishedC chan struct{}

	// For each epoch number, stores the corresponding membership.
	memberships map[t.EpochNr]map[t.NodeID]t.NodeAddress
}

func (chat *ChatApp) CurrentEpoch() t.EpochNr {
	return t.EpochNr(chat.batchesApplied / chat.epochLength)
}

// NewChatApp returns a new instance of the chat demo application.
// The reqStore must be the same request store that is passed to the mir.NewNode() function as a module.
func NewChatApp(initialMembership map[t.NodeID]t.NodeAddress, segmentLength int) *ChatApp {

	// Initialize the membership for the first epochs.
	memberships := make(map[t.EpochNr]map[t.NodeID]t.NodeAddress)
	for i := 0; i <= configOffset+1; i++ {
		memberships[t.EpochNr(i)] = initialMembership
	}

	return &ChatApp{
		messages:       make([]string, 0),
		batchesApplied: 0,
		segmentLength:  segmentLength,
		epochLength:    len(initialMembership) * segmentLength,
		epochFinishedC: make(chan struct{}),
		memberships:    memberships,
	}
}

func (chat *ChatApp) ApplyEvents(eventsIn *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsSequentially(eventsIn, chat.ApplyEvent)
}

func (chat *ChatApp) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
		// no actions on init
	case *eventpb.Event_Deliver:
		if err := chat.ApplyBatch(e.Deliver.Batch); err != nil {
			return nil, fmt.Errorf("app batch delivery error: %w", err)
		}
	case *eventpb.Event_AppSnapshotRequest:
		data, err := chat.Snapshot()
		if err != nil {
			return nil, fmt.Errorf("app snapshot error: %w", err)
		}
		return events.ListOf(events.AppSnapshot(
			t.ModuleID(e.AppSnapshotRequest.Module),
			t.EpochNr(e.AppSnapshotRequest.Epoch),
			data,
		)), nil
	case *eventpb.Event_AppRestoreState:
		if err := chat.RestoreState(e.AppRestoreState.Data); err != nil {
			return nil, fmt.Errorf("app restore state error: %w", err)
		}
	default:
		return nil, fmt.Errorf("unexpected type of App event: %T", event.Type)
	}

	return events.EmptyList(), nil
}

func (chat *ChatApp) WaitForEpochEnd() {
	<-chat.epochFinishedC
}

// ApplyBatch applies a batch of requests to the state of the application.
// In our case, it simply extends the message history
// by appending the payload of each received request as a new chat message.
// Each appended message is also printed to stdout.
func (chat *ChatApp) ApplyBatch(batch *requestpb.Batch) error {

	// For each request in the batch
	for _, req := range batch.Requests {

		// Convert request payload to chat message.
		msgString := string(req.Req.Data)

		// Print content of chat message.
		chatMessage := fmt.Sprintf("Client %v: %s", req.Req.ClientId, msgString)

		// Append the received chat message to the chat history.
		chat.messages = append(chat.messages, chatMessage)

		// Print received chat message.
		fmt.Println(chatMessage)

		// If this is a config message, treat it correspondingly.
		if len(msgString) > len("Config: ") && msgString[:len("Config: ")] == "Config: " {
			configMsg := msgString[len("Config: "):]
			fmt.Printf("Detected config message: %s\n", configMsg)
			chat.applyConfigMsg(configMsg)
		}
	}

	// Update counter of applied batches.
	chat.batchesApplied++

	if chat.batchesApplied%chat.epochLength == 0 {
		// Initialize new membership to be modified throughput the new epoch
		chat.memberships[chat.CurrentEpoch()+configOffset+1] = maputil.Copy(chat.memberships[chat.CurrentEpoch()+configOffset])

		fmt.Printf("Starting epoch %v.\nMembership:\n%v\n", chat.CurrentEpoch(), chat.memberships[chat.CurrentEpoch()])

		chat.epochLength = len(chat.memberships[chat.CurrentEpoch()]) * chat.segmentLength

		// Notify the manager about a new epoch.
		//chat.epochFinishedC <- struct{}{}
	}

	return nil
}

func (chat *ChatApp) applyConfigMsg(configMsg string) {
	tokens := strings.Fields(configMsg)
	switch tokens[0] {
	case "add-node":

		// Parse out the node ID and address
		nodeID := t.NodeID(tokens[1])
		nodeAddr, err := multiaddr.NewMultiaddr(tokens[2])
		if err != nil {
			fmt.Printf("Adding node failed. Invalid address: %v\n", err)
		}

		fmt.Printf("Adding node: %v (%v)\n", nodeID, nodeAddr)
		if _, ok := chat.memberships[chat.CurrentEpoch()+configOffset+1][nodeID]; ok {
			fmt.Printf("Adding node failed. Node already present in membership: %v\n", nodeID)
		} else {
			chat.memberships[chat.CurrentEpoch()+configOffset+1][nodeID] = nodeAddr
		}
	case "remove-node":
		nodeID := t.NodeID(tokens[1])
		fmt.Printf("Removing node: %v\n", nodeID)
		if _, ok := chat.memberships[chat.CurrentEpoch()+configOffset+1][nodeID]; !ok {
			fmt.Printf("Removing node failed. Node not present in membership: %v\n", nodeID)
		} else {
			delete(chat.memberships[chat.CurrentEpoch()+configOffset+1], nodeID)
		}
	default:
		fmt.Printf("Ignoring config message: %s (tokens: %v)\n", configMsg, tokens)
	}
}

// Snapshot returns a binary representation of the application state.
// The returned value can be passed to RestoreState().
// At the time of writing this comment, the Mir library does not support state transfer
// and Snapshot is never actually called.
// We include its implementation for completeness.
func (chat *ChatApp) Snapshot() ([]byte, error) {

	// We use protocol buffers to serialize the application state.
	state := &AppState{
		Messages: chat.messages,
	}
	return proto.Marshal(state)
}

// RestoreState restores the application's state to the one represented by the passed argument.
// The argument is a binary representation of the application state returned from Snapshot().
// After the chat history is restored, RestoreState prints the whole chat history to stdout.
func (chat *ChatApp) RestoreState(snapshot []byte) error {

	// Unmarshal the protobuf message from its binary form.
	state := &AppState{}
	if err := proto.Unmarshal(snapshot, state); err != nil {
		return err
	}

	// Restore internal state
	chat.messages = state.Messages

	// Print new state
	fmt.Printf("\n CHAT STATE RESTORED. SHOWING ALL CHAT HISTORY FROM THE BEGINNING.\n")
	for _, message := range chat.messages {
		fmt.Println(message)
	}

	return nil
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (chat *ChatApp) ImplementsModule() {}
