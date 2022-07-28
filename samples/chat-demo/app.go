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
	"bytes"
	"context"
	"fmt"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/multiformats/go-multiaddr"
	"strings"

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

	// The current epoch number.
	currentEpoch t.EpochNr

	// For each epoch number, stores the corresponding membership.
	memberships []map[t.NodeID]t.NodeAddress

	// Network transport module used by Mir.
	// The app needs a reference to it in order to manage connections when the membership changes.
	transport net.Transport
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (chat *ChatApp) ImplementsModule() {}

// NewChatApp returns a new instance of the chat demo application.
// The reqStore must be the same request store that is passed to the mir.NewNode() function as a module.
func NewChatApp(initialMembership map[t.NodeID]t.NodeAddress, transport net.Transport) *ChatApp {

	// Initialize the membership for the first epochs.
	// We use configOffset+2 memberships to account for:
	// - The first epoch (epoch 0)
	// - The configOffset epochs that already have a fixed membership (epochs 1 to configOffset)
	// - The membership of the following epoch (configOffset+1) initialized with the same membership,
	//   but potentially changed during the first epoch (epoch 0) through special configuration requests.
	memberships := make([]map[t.NodeID]t.NodeAddress, configOffset+2)
	for i := 0; i < configOffset+2; i++ {
		memberships[i] = initialMembership
	}

	return &ChatApp{
		messages:     make([]string, 0),
		currentEpoch: 0,
		memberships:  memberships,
		transport:    transport,
	}
}

func (chat *ChatApp) ApplyEvents(eventsIn *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsSequentially(eventsIn, chat.ApplyEvent)
}

func (chat *ChatApp) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
		return events.EmptyList(), nil // no actions on init
	case *eventpb.Event_Deliver:
		return chat.applyBatch(e.Deliver.Batch)
	case *eventpb.Event_NewEpoch:
		return chat.applyNewEpoch(e.NewEpoch)
	case *eventpb.Event_StateSnapshotRequest:
		return chat.applySnapshotRequest(e.StateSnapshotRequest)
	case *eventpb.Event_AppRestoreState:
		return chat.applyRestoreState(e.AppRestoreState.Snapshot)
	default:
		return nil, fmt.Errorf("unexpected type of App event: %T", event.Type)
	}
}

// applyBatch applies a batch of requests to the state of the application.
// In our case, it simply extends the message history
// by appending the payload of each received request as a new chat message.
// Each appended message is also printed to stdout.
func (chat *ChatApp) applyBatch(batch *requestpb.Batch) (*events.EventList, error) {
	eventsOut := events.EmptyList()

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
			chat.applyConfigMsg(configMsg)
		}
	}

	return eventsOut, nil
}

func (chat *ChatApp) applyConfigMsg(configMsg string) {
	tokens := strings.Fields(configMsg)

	if len(tokens) == 0 {
		fmt.Printf("Ignoring empty config message.\n")
	}

	switch tokens[0] {
	case "add-node":

		if len(tokens) < 3 {
			fmt.Printf("Ignoring config message: %s (tokens: %v). Need 3 tokens.\n", configMsg, tokens)
		}

		// Parse out the node ID and address
		nodeID := t.NodeID(tokens[1])
		nodeAddr, err := multiaddr.NewMultiaddr(tokens[2])
		if err != nil {
			fmt.Printf("Adding node failed. Invalid address: %v\n", err)
		}

		fmt.Printf("Adding node: %v (%v)\n", nodeID, nodeAddr)
		if _, ok := chat.memberships[chat.currentEpoch+configOffset+1][nodeID]; ok {
			fmt.Printf("Adding node failed. Node already present in membership: %v\n", nodeID)
		} else {
			chat.memberships[chat.currentEpoch+configOffset+1][nodeID] = nodeAddr
		}
	case "remove-node":
		nodeID := t.NodeID(tokens[1])
		fmt.Printf("Removing node: %v\n", nodeID)
		if _, ok := chat.memberships[chat.currentEpoch+configOffset+1][nodeID]; !ok {
			fmt.Printf("Removing node failed. Node not present in membership: %v\n", nodeID)
		} else {
			delete(chat.memberships[chat.currentEpoch+configOffset+1], nodeID)
		}
	default:
		fmt.Printf("Ignoring config message: %s (tokens: %v)\n", configMsg, tokens)
	}
}

func (chat *ChatApp) applyNewEpoch(newEpoch *eventpb.NewEpoch) (*events.EventList, error) {

	// Sanity check.
	if t.EpochNr(newEpoch.EpochNr) != chat.currentEpoch+1 {
		return nil, fmt.Errorf("expected next epoch to be %d, got %d", chat.currentEpoch+1, newEpoch.EpochNr)
	}

	fmt.Printf("New epoch: %d\n", newEpoch.EpochNr)

	// Convenience variable.
	newMembership := chat.memberships[newEpoch.EpochNr+configOffset]

	// Append a new membership data structure to be modified throughout the new epoch.
	chat.memberships = append(chat.memberships, maputil.Copy(newMembership))

	// Create network connections to all nodes in the new membership.
	if nodeAddrs, err := dummyMultiAddrs(newMembership); err != nil {
		return nil, err
	} else {
		chat.transport.Connect(context.Background(), nodeAddrs)
	}

	// Update current epoch number.
	chat.currentEpoch = t.EpochNr(newEpoch.EpochNr)

	// Notify ISS about the new membership.
	return events.ListOf(events.NewConfig("iss", maputil.GetSortedKeys(newMembership))), nil
}

// applySnapshotRequest produces a StateSnapshotResponse event containing the current snapshot of the chat app state.
// The snapshot is a binary representation of the application state that can be passed to applyRestoreState().
func (chat *ChatApp) applySnapshotRequest(snapshotRequest *eventpb.StateSnapshotRequest) (*events.EventList, error) {
	return events.ListOf(events.StateSnapshotResponse(
		t.ModuleID(snapshotRequest.Module),
		events.StateSnapshot(chat.serializeMessages(), events.EpochConfig(chat.currentEpoch, chat.memberships)),
	)), nil
}

func (chat *ChatApp) serializeMessages() []byte {
	data := make([]byte, 0)
	for _, msg := range chat.messages {
		data = append(data, []byte(msg)...)
		data = append(data, 0)
	}

	return data
}

// RestoreState restores the application's state to the one represented by the passed argument.
// The argument is a binary representation of the application state returned from Snapshot().
// After the chat history is restored, RestoreState prints the whole chat history to stdout.
func (chat *ChatApp) applyRestoreState(snapshot *commonpb.StateSnapshot) (*events.EventList, error) {

	// Restore chat messages
	chat.restoreChat(snapshot.AppData)

	// Restore configuration
	if err := chat.restoreConfiguration(snapshot.Configuration); err != nil {
		return nil, err
	}

	// Print new state
	fmt.Printf("\nCHAT STATE RESTORED. SHOWING ALL CHAT HISTORY FROM THE BEGINNING.\n\n")
	for _, message := range chat.messages {
		fmt.Println(message)
	}

	return events.EmptyList(), nil
}

func (chat *ChatApp) restoreChat(data []byte) {
	chat.messages = make([]string, 0)
	if len(data) > 0 {
		for _, msg := range bytes.Split(data[:len(data)-1], []byte{0}) { // len(data)-1 to strip off the last zero byte
			chat.messages = append(chat.messages, string(msg))
		}
	}
}

func (chat *ChatApp) restoreConfiguration(config *commonpb.EpochConfig) error {
	chat.currentEpoch = t.EpochNr(config.EpochNr)
	chat.memberships = make([]map[t.NodeID]t.NodeAddress, len(config.Memberships))
	for e, mem := range config.Memberships {
		chat.memberships[e] = make(map[t.NodeID]t.NodeAddress)
		for nID, nAddr := range mem.Membership {
			var err error
			chat.memberships[e][t.NodeID(nID)], err = multiaddr.NewMultiaddr(nAddr)
			if err != nil {
				return err
			}
		}
	}

	fmt.Printf("Restored app memberships: %d (epoch: %d)\n", len(chat.memberships), chat.currentEpoch)

	return nil
}
