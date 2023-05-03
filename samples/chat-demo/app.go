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
	"os"
	"strings"

	"github.com/multiformats/go-multiaddr"

	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/membership"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// ChatApp and its methods implement the application logic of the small chat demo application
// showcasing the usage of the Mir library.
// An initialized instance of this struct needs to be passed to the trantor.New() method when instantiating an SMR system.
type ChatApp struct {

	// The only state of the application is the chat message history,
	// to which each delivered transaction appends one message.
	messages []string

	// Stores the next membership to be submitted to the Node on the next NewEpoch event.
	newMembership *trantorpbtypes.Membership

	// Name of the directory in which to store the checkpoint files.
	chkpDir string
}

// NewChatApp returns a new instance of the chat demo application.
// The initialMembership argument describes the initial set of nodes that are part of the system.
// The application needs to keep track of the membership, as it will be deciding about when and how it changes.
// If chkpDir is not an empty string, it will be interpreted as a path to the directory to store checkpoint files.
func NewChatApp(initialMembership *trantorpbtypes.Membership, chkpDir string) *ChatApp {
	return &ChatApp{
		messages:      make([]string, 0),
		newMembership: initialMembership,
		chkpDir:       chkpDir,
	}
}

// ApplyTXs applies ordered transactions received from the SMR system to the app state.
// In our case, it simply extends the message history
// by appending the payload of each received transaction as a new chat message.
// Each appended message is also printed to stdout.
// Special messages starting with `Config: ` are recognized, parsed, and treated accordingly.
func (chat *ChatApp) ApplyTXs(txs []*trantorpbtypes.Transaction) error {
	// For each transaction in the batch
	for _, tx := range txs {

		// Convert transaction payload to chat message.
		msgString := string(tx.Data)

		// Print content of chat message.
		chatMessage := fmt.Sprintf("Client %v: %s", tx.ClientId, msgString)

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

		// Parse out the node ID and address.
		// Note that we need to pass the address (inside a temporary membership structure)
		// through the DummyMultiAddrs function to generate a proper (dummy) key.
		nodeID := t.NodeID(tokens[1])
		netAddr, err := multiaddr.NewMultiaddr(tokens[2])
		if err != nil {
			fmt.Printf("Adding node failed. Invalid address: %v\n", err)
		}
		tmpMembership, err := membership.DummyMultiAddrs(&trantorpbtypes.Membership{ // nolint:govet
			map[t.NodeID]*trantorpbtypes.NodeIdentity{nodeID: {nodeID, netAddr.String(), nil, 0}}, // nolint:govet
		})
		if err != nil {
			fmt.Printf("Adding node failed. Could not construct dummy multiaddress: %v\n", err)
		}
		nodeAddr := tmpMembership.Nodes[nodeID].Addr

		// Add the node to the next membership.
		fmt.Printf("Adding node: %v (%v)\n", nodeID, nodeAddr)
		if _, ok := chat.newMembership.Nodes[nodeID]; ok {
			fmt.Printf("Adding node failed. Node already present in membership: %v\n", nodeID)
		} else {
			chat.newMembership.Nodes[nodeID] = &trantorpbtypes.NodeIdentity{nodeID, nodeAddr, nil, 0} // nolint:govet
		}
	case "remove-node":
		// Parse out the node ID and remove it from the next membership.
		nodeID := t.NodeID(tokens[1])
		fmt.Printf("Removing node: %v\n", nodeID)
		if _, ok := chat.newMembership.Nodes[nodeID]; !ok {
			fmt.Printf("Removing node failed. Node not present in membership: %v\n", nodeID)
		} else {
			delete(chat.newMembership.Nodes, nodeID)
		}
	default:
		fmt.Printf("Ignoring config message: %s (tokens: %v)\n", configMsg, tokens)
	}
}

// NewEpoch callback is invoked by the SMR system when it transitions to a new epoch.
// The membership returned from NewEpoch will eventually be used by the system.
func (chat *ChatApp) NewEpoch(_ tt.EpochNr) (*trantorpbtypes.Membership, error) {
	return &trantorpbtypes.Membership{maputil.Copy(chat.newMembership.Nodes)}, nil // nolint:govet
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
func (chat *ChatApp) RestoreState(checkpoint *checkpoint.StableCheckpoint) error {

	// Restore chat messages
	chat.restoreChat(checkpoint.Snapshot.AppData)

	// Restore configuration
	if err := chat.restoreConfiguration(checkpoint.Snapshot.EpochData.EpochConfig); err != nil {
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
func (chat *ChatApp) restoreConfiguration(config *trantorpbtypes.EpochConfig) error {
	chat.newMembership = config.Memberships[len(config.Memberships)-1]
	return nil
}

// Checkpoint does nothing in our case. The sample application does not handle checkpoints.
func (chat *ChatApp) Checkpoint(chkp *checkpoint.StableCheckpoint) (retErr error) {

	// If no checkpoint destination was provided, do nothing.
	if chat.chkpDir == "" {
		return nil
	}

	// Create checkpoint directory if it does not exist.
	if err := os.MkdirAll(chat.chkpDir, 0700); err != nil {
		return fmt.Errorf("failed to create checkpoint directory: %w", err)
	}

	// Name the file according to the epoch number and sequence number.
	fileName := fmt.Sprintf("%s/checkpoint-%04d-%05d", chat.chkpDir, chkp.Epoch(), chkp.SeqNr())

	// Create checkpoint file.
	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create checkpoint file: %w", err)
	}

	// Defer closing file.
	defer func() {
		retErr = f.Close()
	}()

	// Serialize checkpoint data.
	data, err := chkp.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize checkpoint: %w", err)
	}

	// Write checkpoint data to file.
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("failed to write checkpoint data to file: %w", err)
	}

	fmt.Printf("Wrote checkpoint to file: %s\n", fileName)
	return nil
}
