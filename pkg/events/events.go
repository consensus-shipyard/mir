/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package events

import (
	"google.golang.org/protobuf/types/known/wrapperspb"

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
// This event is the first to be applied to a module after applying all events from the WAL.
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

func HashedRequest(request *requestpb.Request, digest []byte) *requestpb.HashedRequest {
	return &requestpb.HashedRequest{
		Req:    request,
		Digest: digest,
	}
}

// NewClientRequests returns an event representing the reception of new requests from clients.
func NewClientRequests(destModule t.ModuleID, requests []*requestpb.Request) *eventpb.Event {
	return &eventpb.Event{DestModule: destModule.Pb(), Type: &eventpb.Event_NewRequests{
		NewRequests: &eventpb.NewRequests{Requests: requests},
	}}
}

// HashRequest returns an event representing a request to the hashing module for computing hashes of data.
// For each object to be hashed the data argument contains a slice of byte slices representing it (thus [][][]byte).
// The origin is an object used to maintain the context for the requesting module and will be included in the
// HashResult produced by the hashing module.
func HashRequest(destModule t.ModuleID, data [][][]byte, origin *eventpb.HashOrigin) *eventpb.Event {

	// First, construct a slice of HashData objects
	hashData := make([]*commonpb.HashData, len(data))
	for i, d := range data {
		hashData[i] = &commonpb.HashData{Data: d}
	}

	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_HashRequest{HashRequest: &eventpb.HashRequest{
			Data:   hashData,
			Origin: origin,
		}},
	}
}

// HashResult returns an event representing the computation of hashes by the hashing module.
// It contains the computed digests and the HashOrigin,
// an object used to maintain the context for the requesting module,
// i.e., information about what to do with the contained digest.
func HashResult(destModule t.ModuleID, digests [][]byte, origin *eventpb.HashOrigin) *eventpb.Event {
	return &eventpb.Event{DestModule: destModule.Pb(), Type: &eventpb.Event_HashResult{HashResult: &eventpb.HashResult{
		Digests: digests,
		Origin:  origin,
	}}}
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

// RequestReady returns an event signifying that a new request is ready to be inserted into the protocol state machine.
// This normally occurs when the request has been received, persisted, authenticated, and an authenticator is available.
func RequestReady(destModule t.ModuleID, request *requestpb.Request) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_RequestReady{RequestReady: &eventpb.RequestReady{
			Request: request,
		}},
	}
}

// WALAppend returns an event of appending a new entry to the WAL.
// This event is produced by the protocol state machine for persisting its state.
func WALAppend(destModule t.ModuleID, event *eventpb.Event, retentionIndex t.WALRetIndex) *eventpb.Event {
	return &eventpb.Event{DestModule: destModule.Pb(), Type: &eventpb.Event_WalAppend{WalAppend: &eventpb.WALAppend{
		Event:          event,
		RetentionIndex: retentionIndex.Pb(),
	}}}
}

// WALTruncate returns and event on removing all entries from the WAL
// that have been appended with a retentionIndex smaller than the
// specified one.
func WALTruncate(destModule t.ModuleID, retentionIndex t.WALRetIndex) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_WalTruncate{WalTruncate: &eventpb.WALTruncate{
			RetentionIndex: retentionIndex.Pb(),
		}},
	}
}

// WALEntry returns an event of reading an entry from the WAL.
// Those events are used at system initialization.
func WALEntry(persistedEvent *eventpb.Event, retentionIndex t.WALRetIndex) *eventpb.Event {
	return &eventpb.Event{Type: &eventpb.Event_WalEntry{WalEntry: &eventpb.WALEntry{
		Event: persistedEvent,
	}}}
}

// Deliver returns an event of delivering a request batch to the application in sequence number order.
func Deliver(destModule t.ModuleID, sn t.SeqNr, batch *requestpb.Batch) *eventpb.Event {
	return &eventpb.Event{DestModule: destModule.Pb(), Type: &eventpb.Event_Deliver{Deliver: &eventpb.Deliver{
		Sn:    sn.Pb(),
		Batch: batch,
	}}}
}

// StateSnapshotRequest returns an event representing the protocol module asking the application for a state snapshot.
// epoch is the epoch number which initial state is captured in the snapshot.
// The application itself need not be aware of sn, it is only by the protocol
// to be able to identify the response of the application in form of an StateSnapshot event.
func StateSnapshotRequest(destModule t.ModuleID, srcModule t.ModuleID) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_StateSnapshotRequest{StateSnapshotRequest: &eventpb.StateSnapshotRequest{
			Module: srcModule.Pb(),
		}},
	}
}

// StateSnapshotResponse returns an event representing the application making a snapshot of its state.
// epoch is the number of the epoch whose initial state is captured in the snapshot,
// appData is the serialized application state (the snapshot itself)
// and membership is the membership of the epoch.
func StateSnapshotResponse(
	destModule t.ModuleID,
	snapshot *commonpb.StateSnapshot,
) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type:       &eventpb.Event_StateSnapshot{StateSnapshot: snapshot},
	}
}

// StateSnapshot is a constructor for the StateSnapshot protobuf object.
func StateSnapshot(appData []byte, configuration *commonpb.EpochConfig) *commonpb.StateSnapshot {
	return &commonpb.StateSnapshot{
		AppData:       appData,
		Configuration: configuration,
	}
}

// EpochConfig represents the configuration of the system during one epoch
func EpochConfig(epochNr t.EpochNr, memberships []map[t.NodeID]t.NodeAddress) *commonpb.EpochConfig {
	m := make([]*commonpb.Membership, len(memberships))
	for i, membership := range memberships {
		m[i] = t.MembershipPb(membership)
	}

	return &commonpb.EpochConfig{EpochNr: epochNr.Pb(), Memberships: m}
}

// AppRestoreState returns an event representing the protocol module asking the application
// for restoring its state from the snapshot.
func AppRestoreState(destModule t.ModuleID, snapshot *commonpb.StateSnapshot) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_AppRestoreState{AppRestoreState: &eventpb.AppRestoreState{
			Snapshot: snapshot,
		}},
	}
}

func TimerDelay(destModule t.ModuleID, events []*eventpb.Event, delay t.TimeDuration) *eventpb.Event {
	return &eventpb.Event{DestModule: destModule.Pb(), Type: &eventpb.Event_TimerDelay{TimerDelay: &eventpb.TimerDelay{
		Events: events,
		Delay:  delay.Pb(),
	}}}
}

func TimerRepeat(
	destModule t.ModuleID,
	events []*eventpb.Event,
	delay t.TimeDuration,
	retIndex t.TimerRetIndex,
) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_TimerRepeat{TimerRepeat: &eventpb.TimerRepeat{
			Events:         events,
			Delay:          delay.Pb(),
			RetentionIndex: retIndex.Pb(),
		}},
	}
}

func TimerGarbageCollect(destModule t.ModuleID, retIndex t.TimerRetIndex) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_TimerGarbageCollect{TimerGarbageCollect: &eventpb.TimerGarbageCollect{
			RetentionIndex: retIndex.Pb(),
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

func NewConfig(destModule t.ModuleID, nodeIDs []t.NodeID) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_NewConfig{NewConfig: &eventpb.NewConfig{
			NodeIds: t.NodeIDSlicePb(nodeIDs),
		}},
	}
}
