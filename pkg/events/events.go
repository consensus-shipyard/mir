/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package events

import (
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/filecoin-project/mir/pkg/contextstore"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/contextstorepb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/factorymodulepb"
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

func Redirect(event *eventpb.Event, destination t.ModuleID) *eventpb.Event {
	return &eventpb.Event{
		Type:       event.Type,
		Next:       event.Next,
		DestModule: destination.Pb(),
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

// WALAppend returns an event of appending a new entry to the WAL.
// This event is produced by the protocol state machine for persisting its state.
func WALAppend(destModule t.ModuleID, event *eventpb.Event, retentionIndex t.RetentionIndex) *eventpb.Event {
	return &eventpb.Event{DestModule: destModule.Pb(), Type: &eventpb.Event_WalAppend{WalAppend: &eventpb.WALAppend{
		Event:          event,
		RetentionIndex: retentionIndex.Pb(),
	}}}
}

// WALTruncate returns and event on removing all entries from the WAL
// that have been appended with a retentionIndex smaller than the
// specified one.
func WALTruncate(destModule t.ModuleID, retentionIndex t.RetentionIndex) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_WalTruncate{WalTruncate: &eventpb.WALTruncate{
			RetentionIndex: retentionIndex.Pb(),
		}},
	}
}

// WALEntry returns an event of reading an entry from the WAL.
// Those events are used at system initialization.
func WALEntry(persistedEvent *eventpb.Event, retentionIndex t.RetentionIndex) *eventpb.Event {
	return &eventpb.Event{Type: &eventpb.Event_WalEntry{WalEntry: &eventpb.WALEntry{
		Event: persistedEvent,
	}}}
}

// Deliver returns an event of delivering a request batch to the application in sequence number order.
func DeliverCert(destModule t.ModuleID, sn t.SeqNr, cert *availabilitypb.Cert) *eventpb.Event {
	return &eventpb.Event{DestModule: destModule.Pb(), Type: &eventpb.Event_DeliverCert{
		DeliverCert: &eventpb.DeliverCert{
			Sn:   sn.Pb(),
			Cert: cert,
		},
	}}
}

// AppSnapshotRequest returns an event representing the protocol module asking the application for a state snapshot.
func AppSnapshotRequest(destModule t.ModuleID, srcModule t.ModuleID, contextID contextstore.ItemID) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_AppSnapshotRequest{AppSnapshotRequest: &eventpb.AppSnapshotRequest{
			Module: srcModule.Pb(),
			Origin: contextstore.Origin(contextID),
		}},
	}
}

// AppSnapshotResponse returns an event representing the application making a snapshot of its state.
// appData is the serialized application state (the snapshot itself)
// and origin is the origin of the corresponding AppSnapshotRequest.
func AppSnapshotResponse(
	destModule t.ModuleID,
	appData []byte,
	origin *contextstorepb.Origin,
) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_AppSnapshot{AppSnapshot: &eventpb.AppSnapshot{
			AppData: appData,
			Origin:  origin,
		}},
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
	retIndex t.RetentionIndex,
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

func TimerGarbageCollect(destModule t.ModuleID, retIndex t.RetentionIndex) *eventpb.Event {
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

func NewConfig(destModule t.ModuleID, membership map[t.NodeID]t.NodeAddress) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_NewConfig{NewConfig: &eventpb.NewConfig{
			Membership: t.MembershipPb(membership),
		}},
	}
}

func Factory(destModule t.ModuleID, evt *factorymodulepb.Factory) *eventpb.Event {
	return &eventpb.Event{
		Type:       &eventpb.Event_Factory{Factory: evt},
		DestModule: destModule.Pb(),
	}
}

// ThreshSign returns an event representing a request to the threshcrypto module for computing the signature share over data.
// The origin is an object used to maintain the context for the requesting module and will be included in the
// ThreshSignResult produced by the crypto module.
func ThreshSign(destModule t.ModuleID, data [][]byte, origin *eventpb.ThreshSignOrigin) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_ThreshSign{ThreshSign: &eventpb.ThreshSign{
			Data:   data,
			Origin: origin,
		}},
	}
}

// ThreshSignResult returns an event representing the computation of a signature share by the threshcrypto module.
// It contains the computed signature share and the ThreshSignOrigin,
// an object used to maintain the context for the requesting module,
// i.e., information about what to do with the contained signature.
func ThreshSignResult(destModule t.ModuleID, signatureShare []byte, origin *eventpb.ThreshSignOrigin) *eventpb.Event {
	return &eventpb.Event{DestModule: destModule.Pb(), Type: &eventpb.Event_ThreshSignResult{ThreshSignResult: &eventpb.ThreshSignResult{
		SignatureShare: signatureShare,
		Origin:         origin,
	}}}
}

// ThreshVerShare returns an event representing a request to the threshcrypto module for verifying
// a signature share over data against the group/module's public key.
// The origin is an object used to maintain the context for the requesting module and will be included in the
// ThreshVerShareResult produced by the crypto module.
func ThreshVerShare(destModule t.ModuleID, data [][]byte, sigShare []byte, origin *eventpb.ThreshVerShareOrigin) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_ThreshVerShare{ThreshVerShare: &eventpb.ThreshVerShare{
			Data:           data,
			SignatureShare: sigShare,
			Origin:         origin,
		}},
	}
}

// ThreshVerShareResult returns an event representing the verification of a signature share by the threshcrypto module.
// It contains the result of the verification (boolean and a string error if applicable) and the ThreshVerShareOrigin,
// an object used to maintain the context for the requesting module,
// i.e., information about what to do with the contained signature.
func ThreshVerShareResult(destModule t.ModuleID, ok bool, err string, origin *eventpb.ThreshVerShareOrigin) *eventpb.Event {
	return &eventpb.Event{DestModule: destModule.Pb(), Type: &eventpb.Event_ThreshVerShareResult{ThreshVerShareResult: &eventpb.ThreshVerShareResult{
		Ok:     ok,
		Error:  err,
		Origin: origin,
	}}}
}

// ThreshVerFull returns an event representing a request to the threshcrypto module for verifying
// a full signature over data against the group/module's public key.
// The origin is an object used to maintain the context for the requesting module and will be included in the
// ThreshVerFullResult produced by the crypto module.
func ThreshVerFull(destModule t.ModuleID, data [][]byte, sigFull []byte, origin *eventpb.ThreshVerFullOrigin) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_ThreshVerFull{ThreshVerFull: &eventpb.ThreshVerFull{
			Data:          data,
			FullSignature: sigFull,
			Origin:        origin,
		}},
	}
}

// ThreshVerFullResult returns an event representing the verification of a full signature by the threshcrypto module.
// It contains the result of the verification (boolean and a string error if applicable) and the ThreshVerFullOrigin,
// an object used to maintain the context for the requesting module,
// i.e., information about what to do with the contained signature.
func ThreshVerFullResult(destModule t.ModuleID, ok bool, err string, origin *eventpb.ThreshVerFullOrigin) *eventpb.Event {
	return &eventpb.Event{DestModule: destModule.Pb(), Type: &eventpb.Event_ThreshVerFullResult{ThreshVerFullResult: &eventpb.ThreshVerFullResult{
		Ok:     ok,
		Error:  err,
		Origin: origin,
	}}}
}

// ThreshRecover returns an event representing a request to the threshcrypto module for recovering
// a full signature share over data from signature shares.
// The full signature can only be recovered if enough shares are provided, and if they were created from the group's
// private key shares, therefore the full signature is always valid for this data in the group.
// The origin is an object used to maintain the context for the requesting module and will be included in the
// ThreshRecoverResult produced by the crypto module.
func ThreshRecover(destModule t.ModuleID, data [][]byte, sigShares [][]byte, origin *eventpb.ThreshRecoverOrigin) *eventpb.Event {
	return &eventpb.Event{
		DestModule: destModule.Pb(),
		Type: &eventpb.Event_ThreshRecover{ThreshRecover: &eventpb.ThreshRecover{
			Data:            data,
			SignatureShares: sigShares,
			Origin:          origin,
		}},
	}
}

// ThreshRecoverResult returns an event representing the recovery of a full signature by the threshcrypto module.
// It contains the result of the recovery (boolean, the recovered signature, and a string error if applicable)
// and the ThreshRecoverOrigin, an object used to maintain the context for the requesting module,
// i.e., information about what to do with the contained signature.
func ThreshRecoverResult(destModule t.ModuleID, fullSig []byte, ok bool, err string, origin *eventpb.ThreshRecoverOrigin) *eventpb.Event {
	return &eventpb.Event{DestModule: destModule.Pb(), Type: &eventpb.Event_ThreshRecoverResult{ThreshRecoverResult: &eventpb.ThreshRecoverResult{
		FullSignature: fullSig,
		Ok:            ok,
		Error:         err,
		Origin:        origin,
	}}}
}
