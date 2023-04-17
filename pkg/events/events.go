/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package events

import (
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// Strip returns a new identical (shallow copy of the) event,
// but with all follow-up events (stored under event.Next) removed.
// The removed events are stored in a new EventList that Strip returns a pointer to as the second return value.
func Strip(event *eventpb.Event) (*eventpb.Event, *EventList) {

	// Create new EventList.
	nextList := &EventList{}

	// Add all follow-up events to the new EventList.
	for _, e := range event.Next {
		nextList.PushBack(e)
	}

	// Create a new event with follow-ups removed.
	newEvent := eventpb.Event{
		Type:       event.Type,
		DestModule: event.DestModule,
		Next:       nil,
	}

	// Return new EventList.
	return &newEvent, nextList
}

func Redirect(event *eventpbtypes.Event, destination t.ModuleID) *eventpbtypes.Event {
	return &eventpbtypes.Event{
		Type:       event.Type,
		Next:       event.Next,
		DestModule: destination,
	}
}

// ============================================================
// Event Constructors
// ============================================================

func TestingString(dest t.ModuleID, s string) *eventpb.Event {
	return &eventpb.Event{
		DestModule: dest.Pb(),
		Type: &eventpb.Event_TestingString{
			TestingString: wrapperspb.String(s),
		},
	}
}

func TestingUint(dest t.ModuleID, u uint64) *eventpb.Event {
	return &eventpb.Event{
		DestModule: dest.Pb(),
		Type: &eventpb.Event_TestingUint{
			TestingUint: wrapperspb.UInt64(u),
		},
	}
}

// Init returns an event instructing a module to initialize.
// This event is the first to be applied to a module.
func Init(destModule t.ModuleID) *eventpb.Event {
	return &eventpb.Event{DestModule: destModule.Pb(), Type: &eventpb.Event_Init{Init: &eventpb.Init{}}}
}

// SendMessage returns an event of sending the message message to destinations.
// destinations is a slice of replica IDs that will be translated to actual addresses later.
func SendMessage(destModule t.ModuleID, message *messagepb.Message, destinations []t.NodeID) *eventpb.Event {

	// TODO: This conversion can potentially be very inefficient!

	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_SendMessage{SendMessage: &eventpb.SendMessage{
			Destinations: t.NodeIDSlicePb(destinations),
			Msg:          message,
		}},
	}
}

// MessageReceived returns an event representing the reception of a message from another node.
// The from parameter is the ID of the node the message was received from.
func MessageReceived(destModule t.ModuleID, from t.NodeID, message *messagepb.Message) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_MessageReceived{MessageReceived: &eventpb.MessageReceived{
			From: from.Pb(),
			Msg:  message,
		}},
	}
}

// ClientRequest returns an event representing the reception of a request from a client.
func ClientRequest(clientID t.ClientID, reqNo t.ReqNo, data []byte) *requestpb.Request {
	return &requestpb.Request{
		ClientId: clientID.Pb(),
		ReqNo:    reqNo.Pb(),
		Data:     data,
	}
}

// NewClientRequests returns an event representing the reception of new requests from clients.
func NewClientRequests(destModule t.ModuleID, requests []*requestpb.Request) *eventpb.Event {
	return &eventpb.Event{DestModule: destModule.Pb(), Type: &eventpb.Event_NewRequests{
		NewRequests: &eventpb.NewRequests{Requests: requests},
	}}
}

// AppSnapshotRequest returns an event representing the protocol module asking the application for a state snapshot.
func AppSnapshotRequest(destModule t.ModuleID, replyTo t.ModuleID) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_AppSnapshotRequest{AppSnapshotRequest: &eventpb.AppSnapshotRequest{
			ReplyTo: replyTo.Pb(),
		}},
	}
}

// AppSnapshotResponse returns an event representing the application making a snapshot of its state.
// appData is the serialized application state (the snapshot itself)
// and origin is the origin of the corresponding AppSnapshotRequest.
func AppSnapshotResponse(
	destModule t.ModuleID,
	appData []byte,
) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_AppSnapshot{AppSnapshot: &eventpb.AppSnapshot{
			AppData: appData,
		}},
	}
}

// EpochConfig represents the configuration of the system during one epoch
func EpochConfig(
	epochNr t.EpochNr,
	firstSn t.SeqNr,
	length int,
	memberships []map[t.NodeID]t.NodeAddress,
) *commonpb.EpochConfig {

	m := make([]*commonpb.Membership, len(memberships))
	for i, membership := range memberships {
		m[i] = t.MembershipPb(membership)
	}

	return &commonpb.EpochConfig{
		EpochNr:     epochNr.Pb(),
		FirstSn:     firstSn.Pb(),
		Length:      uint64(length),
		Memberships: m,
	}
}

func NewEpoch(destModule t.ModuleID, epochNr t.EpochNr) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_NewEpoch{NewEpoch: &eventpb.NewEpoch{
			EpochNr: epochNr.Pb(),
		}},
	}
}

func NewConfig(destModule t.ModuleID, epochNr t.EpochNr, membership map[t.NodeID]t.NodeAddress) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_NewConfig{NewConfig: &eventpb.NewConfig{
			EpochNr:    epochNr.Pb(),
			Membership: t.MembershipPb(membership),
		}},
	}
}
