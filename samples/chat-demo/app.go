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
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/statuspb"
	"github.com/filecoin-project/mir/pkg/reqstore"
	t "github.com/filecoin-project/mir/pkg/types"

	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/pb/requestpb"
)

// ChatApp and its methods implement the application logic of the small chat demo application
// showcasing the usage of the Mir library.
// An initialized instance of this struct needs to be passed to the mir.NewNode() method of all nodes
// for the system to run the chat demo app.
type ChatApp struct {
	modules.Module

	// The only state of the application is the chat message history,
	// to which each delivered request appends one message.
	messages []string

	// The request store module (also passed to the Mir library at startup)
	// is used for accessing the request payloads containing the chat message data.
	reqStore *reqstore.VolatileRequestStore
}

// NewChatApp returns a new instance of the chat demo application.
// The reqStore must be the same request store that is passed to the mir.NewNode() function as a module.
func NewChatApp(reqStore *reqstore.VolatileRequestStore) *ChatApp {
	return &ChatApp{
		messages: make([]string, 0),
		reqStore: reqStore,
	}
}

func (chat *ChatApp) ApplyEvents(eventsIn *events.EventList) (*events.EventList, error) {

	eventsOut := &events.EventList{}

	iter := eventsIn.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {
		evts, err := chat.ApplyEvent(event)
		if err != nil {
			return nil, err
		}
		eventsOut.PushBackList(evts)
	}

	return eventsOut, nil
}

func (chat *ChatApp) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpb.Event_Deliver:
		if err := chat.ApplyBatch(e.Deliver.Batch); err != nil {
			return nil, fmt.Errorf("app batch delivery error: %w", err)
		}
	case *eventpb.Event_AppSnapshotRequest:
		data, err := chat.Snapshot()
		if err != nil {
			return nil, fmt.Errorf("app snapshot error: %w", err)
		}
		return (&events.EventList{}).PushBack(events.AppSnapshot(
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

	return &events.EventList{}, nil
}

// ApplyBatch applies a batch of requests to the state of the application.
// In our case, it simply extends the message history
// by appending the payload of each received request as a new chat message.
// Each appended message is also printed to stdout.
func (chat *ChatApp) ApplyBatch(batch *requestpb.Batch) error {

	// For each request in the batch
	for _, reqRef := range batch.Requests {

		// Extract request data from the request store and construct a printable chat message.
		reqData, err := chat.reqStore.GetRequest(reqRef)
		if err != nil {
			return err
		}
		chatMessage := fmt.Sprintf("Client %v: %s", reqRef.ClientId, string(reqData))

		// Append the received chat message to the chat history.
		chat.messages = append(chat.messages, chatMessage)

		// Print received chat message.
		fmt.Println(chatMessage)
	}
	return nil
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

// Status returns a representation of the ChatApp's internal state for the purpose of debugging.
// Currently it is not used by Mir and thus does not need to be implemented.
func (chat *ChatApp) Status() (s *statuspb.ProtocolStatus, err error) {
	panic("not implemented")
}
