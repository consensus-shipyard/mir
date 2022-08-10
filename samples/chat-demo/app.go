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
	"strings"

	availabilityevents "github.com/filecoin-project/mir/pkg/availability/events"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/contextstorepb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"

	"github.com/multiformats/go-multiaddr"
)

// ChatApp and its methods implement the application logic of the small chat demo application
// showcasing the usage of the Mir library.
// An initialized instance of this struct needs to be passed to the mir.NewNode() method of all nodes
// for the system to run the chat demo app.
type ChatApp struct {

	// The only state of the application is the chat message history,
	// to which each delivered request appends one message.
	messages []string

	// Stores the next membership to be submitted to the Node on the next NewEpoch event.
	newMembership map[t.NodeID]t.NodeAddress

	// Network transport module used by Mir.
	// The app needs a reference to it in order to manage connections when the membership changes.
	transport net.Transport
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (chat *ChatApp) ImplementsModule() {}

// NewChatApp returns a new instance of the chat demo application.
// The reqStore must be the same request store that is passed to the mir.NewNode() function as a module.
func NewChatApp(initialMembership map[t.NodeID]t.NodeAddress, transport net.Transport) *ChatApp {

	return &ChatApp{
		messages:      make([]string, 0),
		newMembership: initialMembership,
		transport:     transport,
	}
}

func (chat *ChatApp) ApplyEvents(eventsIn *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsSequentially(eventsIn, chat.ApplyEvent)
}

func (chat *ChatApp) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
		return events.EmptyList(), nil // no actions on init
	case *eventpb.Event_NewEpoch:
		return chat.applyNewEpoch(e.NewEpoch)
	case *eventpb.Event_Deliver:
		return chat.ApplyDeliver(e.Deliver)
	case *eventpb.Event_Availability:
		switch e := e.Availability.Type.(type) {
		case *availabilitypb.Event_ProvideTransactions:
			return chat.applyProvideTransactions(e.ProvideTransactions)
		default:
			return nil, fmt.Errorf("unexpected availability event type: %T", e)
		}
	case *eventpb.Event_AppSnapshotRequest:
		return chat.applySnapshotRequest(e.AppSnapshotRequest)
	case *eventpb.Event_AppRestoreState:
		return chat.applyRestoreState(e.AppRestoreState.Snapshot)
	default:
		return nil, fmt.Errorf("unexpected type of App event: %T", event.Type)
	}
}

// ApplyDeliver applies a batch of requests to the state of the application.
// In our case, it simply extends the message history
// by appending the payload of each received request as a new chat message.
// Each appended message is also printed to stdout.
func (chat *ChatApp) ApplyDeliver(deliver *eventpb.Deliver) (*events.EventList, error) {

	switch c := deliver.Cert.Type.(type) {
	case *availabilitypb.Cert_Msc:
		if len(c.Msc.BatchId) == 0 {
			fmt.Println("Received empty batch availability certificate.")
			return events.EmptyList(), nil
		} else {
			return events.ListOf(availabilityevents.RequestTransactions(
				"availability",
				deliver.Cert,
				&availabilitypb.RequestTransactionsOrigin{
					Module: "app",
					Type: &availabilitypb.RequestTransactionsOrigin_ContextStore{
						ContextStore: &contextstorepb.Origin{ItemID: 0},
					},
				},
			)), nil
		}
	default:
		return nil, fmt.Errorf("unknown availability certificate type: %T", deliver.Cert.Type)
	}
}

func (chat *ChatApp) applyProvideTransactions(ptx *availabilitypb.ProvideTransactions) (*events.EventList, error) {
	// For each request in the batch
	for _, req := range ptx.Txs {

		// Convert request payload to chat message.
		msgString := string(req.Data)

		// Print content of chat message.
		chatMessage := fmt.Sprintf("Client %v: %s", req.ClientId, msgString)

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

	return events.EmptyList(), nil
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
		if _, ok := chat.newMembership[nodeID]; ok {
			fmt.Printf("Adding node failed. Node already present in membership: %v\n", nodeID)
		} else {
			chat.newMembership[nodeID] = nodeAddr
		}
	case "remove-node":
		nodeID := t.NodeID(tokens[1])
		fmt.Printf("Removing node: %v\n", nodeID)
		if _, ok := chat.newMembership[nodeID]; !ok {
			fmt.Printf("Removing node failed. Node not present in membership: %v\n", nodeID)
		} else {
			delete(chat.newMembership, nodeID)
		}
	default:
		fmt.Printf("Ignoring config message: %s (tokens: %v)\n", configMsg, tokens)
	}
}

func (chat *ChatApp) applyNewEpoch(newEpoch *eventpb.NewEpoch) (*events.EventList, error) {

	// Create network connections to all nodes in the new membership.
	var nodeAddrs map[t.NodeID]t.NodeAddress
	var err error
	if nodeAddrs, err = dummyMultiAddrs(chat.newMembership); err != nil {
		return nil, err
	}

	chat.transport.Connect(context.Background(), nodeAddrs)

	// Notify ISS about the new membership.
	return events.ListOf(events.NewConfig("iss", maputil.Copy(chat.newMembership))), nil
}

// applySnapshotRequest produces a StateSnapshotResponse event containing the current snapshot of the chat app state.
// The snapshot is a binary representation of the application state that can be passed to applyRestoreState().
func (chat *ChatApp) applySnapshotRequest(snapshotRequest *eventpb.AppSnapshotRequest) (*events.EventList, error) {
	return events.ListOf(events.AppSnapshotResponse(
		t.ModuleID(snapshotRequest.Module),
		chat.serializeMessages(),
		snapshotRequest.Origin,
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
	chat.newMembership = t.Membership(config.Memberships[len(config.Memberships)-1])
	return nil
}
