/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// This file provides constructors for protobuf messages (also used to represent events) used by ISS.
// The primary purpose is convenience and improved readability of the ISS code,
// As creating protobuf objects is rather verbose in Go.
// Moreover, in case the definitions of some protocol buffers change,
// this file should be the only one that will potentially need to change.

// TODO: Write documentation comments for the functions in this file.
//       Part of the text can probably be copy-pasted from the documentation of the functions handling those events.

package iss

import (
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/isspb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ============================================================
// Events
// ============================================================

// ------------------------------------------------------------
// ISS Events

func Event(destModule t.ModuleID, event *isspb.ISSEvent) *eventpb.Event {
	return &eventpb.Event{DestModule: destModule.Pb(), Type: &eventpb.Event_Iss{Iss: event}}
}

func HashOrigin(origin *isspb.ISSHashOrigin) *eventpb.HashOrigin {
	return &eventpb.HashOrigin{Module: OwnModuleName.Pb(), Type: &eventpb.HashOrigin_Iss{Iss: origin}}
}

func SignOrigin(origin *isspb.ISSSignOrigin) *eventpb.SignOrigin {
	return &eventpb.SignOrigin{Module: OwnModuleName.Pb(), Type: &eventpb.SignOrigin_Iss{Iss: origin}}
}

func SigVerOrigin(origin *isspb.ISSSigVerOrigin) *eventpb.SigVerOrigin {
	return &eventpb.SigVerOrigin{Module: OwnModuleName.Pb(), Type: &eventpb.SigVerOrigin_Iss{Iss: origin}}
}

func PersistCheckpointEvent(sn t.SeqNr, appSnapshot, appSnapshotHash, signature []byte) *eventpb.Event {
	return Event(
		OwnModuleName,
		&isspb.ISSEvent{Type: &isspb.ISSEvent_PersistCheckpoint{PersistCheckpoint: &isspb.PersistCheckpoint{
			Sn:              sn.Pb(),
			AppSnapshot:     appSnapshot,
			AppSnapshotHash: appSnapshotHash,
			Signature:       signature,
		}}},
	)
}

func PersistStableCheckpointEvent(stableCheckpoint *isspb.StableCheckpoint) *eventpb.Event {
	return Event(
		OwnModuleName,
		&isspb.ISSEvent{Type: &isspb.ISSEvent_PersistStableCheckpoint{
			PersistStableCheckpoint: &isspb.PersistStableCheckpoint{
				StableCheckpoint: stableCheckpoint,
			},
		}},
	)
}

func StableCheckpointEvent(stableCheckpoint *isspb.StableCheckpoint) *eventpb.Event {
	return Event(
		OwnModuleName,
		&isspb.ISSEvent{Type: &isspb.ISSEvent_StableCheckpoint{
			StableCheckpoint: stableCheckpoint,
		}},
	)
}

func SBEvent(epoch t.EpochNr, instance t.SBInstanceNr, event *isspb.SBInstanceEvent) *eventpb.Event {
	return Event(
		OwnModuleName,
		&isspb.ISSEvent{Type: &isspb.ISSEvent_Sb{Sb: &isspb.SBEvent{
			Epoch:    epoch.Pb(),
			Instance: instance.Pb(),
			Event:    event,
		}}},
	)
}

func LogEntryHashOrigin(logEntrySN t.SeqNr) *eventpb.HashOrigin {
	return HashOrigin(&isspb.ISSHashOrigin{Type: &isspb.ISSHashOrigin_LogEntrySn{LogEntrySn: logEntrySN.Pb()}})
}

func SBHashOrigin(epoch t.EpochNr, instance t.SBInstanceNr, origin *isspb.SBInstanceHashOrigin) *eventpb.HashOrigin {
	return HashOrigin(&isspb.ISSHashOrigin{Type: &isspb.ISSHashOrigin_Sb{Sb: &isspb.SBHashOrigin{
		Epoch:    epoch.Pb(),
		Instance: instance.Pb(),
		Origin:   origin,
	}}})
}

func AppSnapshotHashOrigin(epoch t.EpochNr) *eventpb.HashOrigin {
	return HashOrigin(&isspb.ISSHashOrigin{Type: &isspb.ISSHashOrigin_AppSnapshotEpoch{AppSnapshotEpoch: epoch.Pb()}})
}

func CheckpointSignOrigin(epoch t.EpochNr) *eventpb.SignOrigin {
	return SignOrigin(&isspb.ISSSignOrigin{Type: &isspb.ISSSignOrigin_CheckpointEpoch{CheckpointEpoch: epoch.Pb()}})
}

func CheckpointSigVerOrigin(epoch t.EpochNr) *eventpb.SigVerOrigin {
	return SigVerOrigin(&isspb.ISSSigVerOrigin{Type: &isspb.ISSSigVerOrigin_CheckpointEpoch{CheckpointEpoch: epoch.Pb()}})
}

func SBSignOrigin(epoch t.EpochNr, instance t.SBInstanceNr, origin *isspb.SBInstanceSignOrigin) *eventpb.SignOrigin {
	return SignOrigin(&isspb.ISSSignOrigin{Type: &isspb.ISSSignOrigin_Sb{Sb: &isspb.SBSignOrigin{
		Epoch:    epoch.Pb(),
		Instance: instance.Pb(),
		Origin:   origin,
	}}})
}

func SBSigVerOrigin(
	epoch t.EpochNr,
	instance t.SBInstanceNr,
	origin *isspb.SBInstanceSigVerOrigin,
) *eventpb.SigVerOrigin {
	return SigVerOrigin(&isspb.ISSSigVerOrigin{Type: &isspb.ISSSigVerOrigin_Sb{Sb: &isspb.SBSigVerOrigin{
		Epoch:    epoch.Pb(),
		Instance: instance.Pb(),
		Origin:   origin,
	}}})
}

// ------------------------------------------------------------
// SB Instance Events

func SBInitEvent() *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_Init{Init: &isspb.SBInit{}}}
}

func SBTickEvent() *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_Tick{Tick: &isspb.SBTick{}}}
}

func SBDeliverEvent(sn t.SeqNr, batch *requestpb.Batch, aborted bool) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_Deliver{
		Deliver: &isspb.SBDeliver{
			Sn:      sn.Pb(),
			Batch:   batch,
			Aborted: aborted,
		},
	}}
}

func SBMessageReceivedEvent(message *isspb.SBInstanceMessage, from t.NodeID) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_MessageReceived{
		MessageReceived: &isspb.SBMessageReceived{
			From: from.Pb(),
			Msg:  message,
		},
	}}
}

func SBPendingRequestsEvent(numRequests t.NumRequests) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_PendingRequests{
		PendingRequests: &isspb.SBPendingRequests{
			NumRequests: numRequests.Pb(),
		},
	}}
}

func SBCutBatchEvent(maxSize t.NumRequests) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_CutBatch{CutBatch: &isspb.SBCutBatch{
		MaxSize: maxSize.Pb(),
	}}}
}

func SBBatchReadyEvent(batch *requestpb.Batch, pendingReqsLeft t.NumRequests) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_BatchReady{BatchReady: &isspb.SBBatchReady{
		Batch:               batch,
		PendingRequestsLeft: pendingReqsLeft.Pb(),
	}}}
}

func SBResurrectBatchEvent(batch *requestpb.Batch) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_ResurrectBatch{ResurrectBatch: batch}}
}

func SBWaitForRequestsEvent(reference *isspb.SBReqWaitReference, requests []*requestpb.RequestRef) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_WaitForRequests{
		WaitForRequests: &isspb.SBWaitForRequests{
			Reference: reference,
			Requests:  requests,
		},
	}}
}

func SBRequestsReady(ref *isspb.SBReqWaitReference) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_RequestsReady{RequestsReady: &isspb.SBRequestsReady{
		Ref: ref,
	}}}
}

func SBHashResultEvent(digests [][]byte, origin *isspb.SBInstanceHashOrigin) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_HashResult{HashResult: &isspb.SBHashResult{
		Digests: digests,
		Origin:  origin,
	}}}
}

func SBSignResultEvent(signature []byte, origin *isspb.SBInstanceSignOrigin) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_SignResult{SignResult: &isspb.SBSignResult{
		Signature: signature,
		Origin:    origin,
	}}}
}

func SBNodeSigsVerifiedEvent(
	valid []bool,
	errors []string,
	nodeIDs []t.NodeID,
	origin *isspb.SBInstanceSigVerOrigin,
	allOK bool,
) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_NodeSigsVerified{
		NodeSigsVerified: &isspb.SBNodeSigsVerified{
			NodeIds: t.NodeIDSlicePb(nodeIDs),
			Valid:   valid,
			Errors:  errors,
			Origin:  origin,
			AllOk:   allOK,
		},
	}}
}

// ============================================================
// Messages
// ============================================================

func Message(msg *isspb.ISSMessage) *messagepb.Message {
	return &messagepb.Message{Type: &messagepb.Message_Iss{Iss: msg}}
}

func SBMessage(epoch t.EpochNr, instance t.SBInstanceNr, msg *isspb.SBInstanceMessage) *messagepb.Message {
	return Message(&isspb.ISSMessage{Type: &isspb.ISSMessage_Sb{Sb: &isspb.SBMessage{
		Epoch:    epoch.Pb(),
		Instance: instance.Pb(),
		Msg:      msg,
	}}})
}

func CheckpointMessage(epoch t.EpochNr, sn t.SeqNr, appSnapshotHash, signature []byte) *messagepb.Message {
	return Message(&isspb.ISSMessage{Type: &isspb.ISSMessage_Checkpoint{Checkpoint: &isspb.Checkpoint{
		Epoch:           epoch.Pb(),
		Sn:              sn.Pb(),
		AppSnapshotHash: appSnapshotHash,
		Signature:       signature,
	}}})
}

func StableCheckpointMessage(stableCheckpoint *isspb.StableCheckpoint) *messagepb.Message {
	return Message(&isspb.ISSMessage{Type: &isspb.ISSMessage_StableCheckpoint{StableCheckpoint: stableCheckpoint}})
}

func RetransmitRequestsMessage(requests []*requestpb.RequestRef) *messagepb.Message {
	return Message(&isspb.ISSMessage{Type: &isspb.ISSMessage_RetransmitRequests{
		RetransmitRequests: &isspb.RetransmitRequests{
			Requests: requests,
		},
	}})
}
