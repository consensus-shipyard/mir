/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// ********************************************************************************
//                                                                               //
//         Chat demo application for demonstrating the usage of Mir              //
//                            (application logic)                                //
//                                                                               //
// ********************************************************************************

package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/multiformats/go-multiaddr"

	"github.com/filecoin-project/mir/pkg/pb/checkpointpb"

	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// ChatApp and its methods implement the application logic of the small chat demo application
// showcasing the usage of the Mir library.
// An initialized instance of this struct needs to be passed to the smr.New() method when instantiating an SMR system.
type ChatApp struct {

	// The only state of the application is the chat message history,
	// to which each delivered transaction appends one message.
	messages []string

	// Stores the next membership to be submitted to the Node on the next NewEpoch event.
	newMembership map[t.NodeID]t.NodeAddress
}

// NewChatApp returns a new instance of the chat demo application.
// The initialMembership argument describes the initial set of nodes that are part of the system.
// The application needs to keep track of the membership, as it will be deciding about when and how it changes.
func NewChatApp(initialMembership map[t.NodeID]t.NodeAddress) *ChatApp {
	return &ChatApp{
		messages:      make([]string, 0),
		newMembership: initialMembership,
	}
}

// ApplyTXs applies ordered transactions received from the SMR system to the app state.
// In our case, it simply extends the message history
// by appending the payload of each received transaction as a new chat message.
// Each appended message is also printed to stdout.
// Special messages starting with `Config: ` are recognized, parsed, and treated accordingly.
func (chat *ChatApp) ApplyTXs(txs []*requestpb.Request) error {
	// For each transaction in the batch
	for _, req := range txs {

		// Convert transaction payload to chat message.
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
			chat.applyConfigTX(configMsg)
		}
	}

	return nil
}

// applyConfigMsg parses and applies a special configuration transaction.
// The supported configuration transactions are:
// - Config: add-node <nodeID> <nodeAddress>
// - Config: remove-node <nodeID>
// The membership change is communicated to the SMR system the next time NewEpoch() is called
// and takes effect after a pre-configured number of epochs (using the system's default for simplicity).
func (chat *ChatApp) applyConfigTX(configMsg string) {

	// Split config message into whitespace-separated tokens.
	tokens := strings.Fields(configMsg)
	if len(tokens) == 0 {
		fmt.Printf("Ignoring empty config message.\n")
	}

	// Parse config message and execute the appropriate action.
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

		// Add the node to the next membership.
		fmt.Printf("Adding node: %v (%v)\n", nodeID, nodeAddr)
		if _, ok := chat.newMembership[nodeID]; ok {
			fmt.Printf("Adding node failed. Node already present in membership: %v\n", nodeID)
		} else {
			chat.newMembership[nodeID] = nodeAddr
		}
	case "remove-node":
		// Parse out the node ID and remove it from the next membership.
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

// NewEpoch callback is invoked by the SMR system when it transitions to a new epoch.
// The membership returned from NewEpoch will eventually be used by the system.
func (chat *ChatApp) NewEpoch(_ t.EpochNr) (map[t.NodeID]t.NodeAddress, error) {
	return maputil.Copy(chat.newMembership), nil
}

// Snapshot produces a StateSnapshotResponse event containing the current snapshot of the chat app state.
// The snapshot is a binary representation of the application state that can be passed to applyRestoreState().
// In our case, it only serializes the chat message history.
func (chat *ChatApp) Snapshot() ([]byte, error) {
	data := make([]byte, 0)
	for _, msg := range chat.messages {
		data = append(data, []byte(msg)...) // message data
		data = append(data, 0)              // zero byte as separator
	}

	return data, nil
}

// RestoreState restores the application's state to the one represented by the passed argument.
// The argument is a binary representation of the application state returned from Snapshot().
// After the chat history is restored, RestoreState prints the whole chat history to stdout.
func (chat *ChatApp) RestoreState(appData []byte, epochConfig *commonpb.EpochConfig) error {

	// Restore chat messages
	chat.restoreChat(appData)

	// Restore configuration
	if err := chat.restoreConfiguration(epochConfig); err != nil {
		return err
	}

	// Print new state
	fmt.Printf("\nCHAT STATE RESTORED. SHOWING ALL CHAT HISTORY FROM THE BEGINNING.\n\n")
	for _, message := range chat.messages {
		fmt.Println(message)
	}

	return nil
}

// restoreChat deserializes the chat history from the passed byte slice.
func (chat *ChatApp) restoreChat(data []byte) {
	chat.messages = make([]string, 0)
	if len(data) > 0 {
		for _, msg := range bytes.Split(data[:len(data)-1], []byte{0}) { // len(data)-1 to strip off the last zero byte
			chat.messages = append(chat.messages, string(msg))
		}
	}
}

// restoreConfiguration saves the system membership, so it can be further updated by subsequent config transactions.
func (chat *ChatApp) restoreConfiguration(config *commonpb.EpochConfig) error {
	chat.newMembership = t.Membership(config.Memberships[len(config.Memberships)-1])
	return nil
}

// Checkpoint does nothing in our case. The sample application does not handle checkpoints.
func (chat *ChatApp) Checkpoint(_ *checkpointpb.StableCheckpoint) error {
	return nil
}
