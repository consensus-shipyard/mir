/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package events

import (
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	"github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
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

// SignRequest returns an event representing a request to the crypto module for computing the signature over data.
// The origin is an object used to maintain the context for the requesting module and will be included in the
// SignResult produced by the crypto module.
func SignRequest(destModule t.ModuleID, data [][]byte, origin *eventpb.SignOrigin) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_SignRequest{SignRequest: &eventpb.SignRequest{
			Data:   data,
			Origin: origin,
		}},
	}
}

// SignResult returns an event representing the computation of a signature by the crypto module.
// It contains the computed signature and the SignOrigin,
// an object used to maintain the context for the requesting module,
// i.e., information about what to do with the contained signature.
func SignResult(destModule t.ModuleID, signature []byte, origin *eventpb.SignOrigin) *eventpb.Event {
	return &eventpb.Event{DestModule: destModule.Pb(), Type: &eventpb.Event_SignResult{SignResult: &eventpb.SignResult{
		Signature: signature,
		Origin:    origin,
	}}}
}

// VerifyNodeSigs returns an event representing a request to the crypto module
// for verifying a batch of node signatures.
// The origin is an object used to maintain the context for the requesting module and will be included in the
// NodeSigsVerified event produced by the crypto module.
// For each signed message, the data argument contains a slice of byte slices (thus [][][]byte)
func VerifyNodeSigs(
	destModule t.ModuleID,
	data [][][]byte,
	signatures [][]byte,
	nodeIDs []t.NodeID,
	origin *eventpb.SigVerOrigin,
) *eventpb.Event {

	// First, construct a slice of SigVerData objects
	sigVerData := make([]*eventpb.SigVerData, len(data))
	for i, d := range data {
		sigVerData[i] = &eventpb.SigVerData{Data: d}
	}

	// Then return the actual Event.
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_VerifyNodeSigs{VerifyNodeSigs: &eventpb.VerifyNodeSigs{
			Data:       sigVerData,
			Signatures: signatures,
			NodeIds:    t.NodeIDSlicePb(nodeIDs),
			Origin:     origin,
		}},
	}
}

// NodeSigsVerified returns an event representing the result of the verification
// of multiple nodes' signatures by the crypto module.
// It contains the results of the verifications
// (as boolean values and errors produced by the Crypto module if the verification failed)
// and the SigVerOrigin, an object used to maintain the context for the requesting module,
// i.e., information about what to do with the results of the signature verification.
// The allOK argument must be set to true if the valid argument only contains true values.
func NodeSigsVerified(
	destModule t.ModuleID,
	valid []bool,
	errors []string,
	nodeIDs []t.NodeID,
	origin *eventpb.SigVerOrigin,
	allOk bool,
) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_NodeSigsVerified{NodeSigsVerified: &eventpb.NodeSigsVerified{
			Valid:   valid,
			Errors:  errors,
			NodeIds: t.NodeIDSlicePb(nodeIDs),
			Origin:  origin,
			AllOk:   allOk,
		}},
	}
}

// DeliverCert returns an event of delivering a request batch to the application in sequence number order.
func DeliverCert(destModule t.ModuleID, sn t.SeqNr, cert *availabilitypb.Cert) *eventpb.Event {
	return &eventpb.Event{DestModule: destModule.Pb(), Type: &eventpb.Event_DeliverCert{
		DeliverCert: &eventpb.DeliverCert{
			Sn:   sn.Pb(),
			Cert: cert,
		},
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

// AppRestoreState returns an event representing the protocol module asking the application
// for restoring its state from the snapshot.
func AppRestoreState(destModule t.ModuleID, checkpoint *checkpointpb.StableCheckpoint) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_AppRestoreState{AppRestoreState: &eventpb.AppRestoreState{
			Checkpoint: checkpoint,
		}},
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
