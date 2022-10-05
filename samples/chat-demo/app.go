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
	"fmt"
	"strings"

	"github.com/multiformats/go-multiaddr"

	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
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
}

// NewChatApp returns a new instance of the chat demo application.
// The reqStore must be the same request store that is passed to the mir.NewNode() function as a module.
func NewChatApp(initialMembership map[t.NodeID]t.NodeAddress) *ChatApp {

	return &ChatApp{
		messages:      make([]string, 0),
		newMembership: initialMembership,
	}
}

// ApplyTXs applies transactions received from the availability layer to the app state.
// In our case, it simply extends the message history
// by appending the payload of each received request as a new chat message.
// Each appended message is also printed to stdout.
// Special messages starting with `Config: ` are recognized, parsed, and treated accordingly.
func (chat *ChatApp) ApplyTXs(txs []*requestpb.Request) error {
	// For each request in the batch
	for _, req := range txs {

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

	return nil
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

func (chat *ChatApp) NewEpoch(_ t.EpochNr) (map[t.NodeID]t.NodeAddress, error) {
	return maputil.Copy(chat.newMembership), nil
}

// Snapshot produces a StateSnapshotResponse event containing the current snapshot of the chat app state.
// The snapshot is a binary representation of the application state that can be passed to applyRestoreState().
func (chat *ChatApp) Snapshot() ([]byte, error) {
	return chat.serializeMessages(), nil
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
