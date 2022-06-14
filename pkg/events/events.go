/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package events

import (
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
		Type: event.Type,
		Next: nil,
	}

	// Return new EventList.
	return &newEvent, nextList
}

// ============================================================
// Event Constructors
// ============================================================

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
func ClientRequest(
	destModule t.ModuleID,
	clientID t.ClientID,
	reqNo t.ReqNo,
	data []byte,
	authenticator []byte,
) *eventpb.Event {
	return &eventpb.Event{DestModule: destModule.Pb(), Type: &eventpb.Event_Request{Request: &requestpb.Request{
		ClientId:      clientID.Pb(),
		ReqNo:         reqNo.Pb(),
		Data:          data,
		Authenticator: authenticator,
	}}}
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
func RequestReady(destModule t.ModuleID, requestRef *requestpb.RequestRef) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_RequestReady{RequestReady: &eventpb.RequestReady{
			RequestRef: requestRef,
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

// WALLoadAll returns and event instructing the WAL to load and emit all stored events.
func WALLoadAll(destModule t.ModuleID) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type:       &eventpb.Event_WalLoadAll{WalLoadAll: &eventpb.WALLoadAll{}},
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

// VerifyRequestSig returns an event of a client tracker requesting the verification of a client request signature.
// This event is routed to the Crypto module that issues a RequestSigVerified event in response.
func VerifyRequestSig(destModule t.ModuleID, reqRef *requestpb.RequestRef, signature []byte) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_VerifyRequestSig{VerifyRequestSig: &eventpb.VerifyRequestSig{
			RequestRef: reqRef,
			Signature:  signature,
		}},
	}
}

// RequestSigVerified represents the result of a client signature verification by the Crypto module.
// It is routed to the client tracker that initially issued a VerifyRequestSig event.
func RequestSigVerified(destModule t.ModuleID, reqRef *requestpb.RequestRef, valid bool, error string) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_RequestSigVerified{RequestSigVerified: &eventpb.RequestSigVerified{
			RequestRef: reqRef,
			Valid:      valid,
			Error:      error,
		}},
	}
}

// StoreVerifiedRequest returns an event representing an event the ClientTracker emits
// to request storing a request, including its payload and authenticator, in the request store.
func StoreVerifiedRequest(
	destModule t.ModuleID,
	reqRef *requestpb.RequestRef,
	data []byte,
	authenticator []byte,
) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_StoreVerifiedRequest{StoreVerifiedRequest: &eventpb.StoreVerifiedRequest{
			RequestRef:    reqRef,
			Data:          data,
			Authenticator: authenticator,
		}},
	}
}

// AppSnapshotRequest returns an event representing the protocol module asking the application for a state snapshot.
// epoch is the epoch number which initial state is captured in the snapshot.
// The application itself need not be aware of sn, it is only by the protocol
// to be able to identify the response of the application in form of an AppSnapshot event.
func AppSnapshotRequest(destModule t.ModuleID, srcModule t.ModuleID, epoch t.EpochNr) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_AppSnapshotRequest{AppSnapshotRequest: &eventpb.AppSnapshotRequest{
			Module: srcModule.Pb(),
			Epoch:  epoch.Pb(),
		}},
	}
}

// AppSnapshot returns an event representing the application making a snapshot of its state.
// epoch is the epoch number which initial state is captured in the snapshot,
// and data is the serialized application state (the snapshot itself).
func AppSnapshot(destModule t.ModuleID, epoch t.EpochNr, data []byte) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_AppSnapshot{AppSnapshot: &eventpb.AppSnapshot{
			Epoch: epoch.Pb(),
			Data:  data,
		}},
	}
}

// AppRestoreState returns an event representing the protocol module asking the application for restoring its state from the snapshot.
func AppRestoreState(destModule t.ModuleID, snapshot []byte) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_AppRestoreState{AppRestoreState: &eventpb.AppRestoreState{
			Data: snapshot,
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
