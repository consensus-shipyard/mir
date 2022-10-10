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
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	"github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/isspb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
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

func HashOrigin(ownModuleID t.ModuleID, origin *isspb.ISSHashOrigin) *eventpb.HashOrigin {
	return &eventpb.HashOrigin{Module: ownModuleID.Pb(), Type: &eventpb.HashOrigin_Iss{Iss: origin}}
}

func SignOrigin(ownModuleID t.ModuleID, origin *isspb.ISSSignOrigin) *eventpb.SignOrigin {
	return &eventpb.SignOrigin{Module: ownModuleID.Pb(), Type: &eventpb.SignOrigin_Iss{Iss: origin}}
}

func SigVerOrigin(ownModuleID t.ModuleID, origin *isspb.ISSSigVerOrigin) *eventpb.SigVerOrigin {
	return &eventpb.SigVerOrigin{Module: ownModuleID.Pb(), Type: &eventpb.SigVerOrigin_Iss{Iss: origin}}
}

func PersistCheckpointEvent(
	ownModuleID t.ModuleID,
	sn t.SeqNr,
	stateSnapshot *commonpb.StateSnapshot,
	appSnapshotHash,
	signature []byte,
) *eventpb.Event {
	return Event(
		ownModuleID,
		&isspb.ISSEvent{Type: &isspb.ISSEvent_PersistCheckpoint{PersistCheckpoint: &isspb.PersistCheckpoint{
			Sn:                sn.Pb(),
			StateSnapshot:     stateSnapshot,
			StateSnapshotHash: appSnapshotHash,
			Signature:         signature,
		}}},
	)
}

func PersistStableCheckpointEvent(ownModuleID t.ModuleID, stableCheckpoint *checkpointpb.StableCheckpoint) *eventpb.Event {
	return Event(
		ownModuleID,
		&isspb.ISSEvent{Type: &isspb.ISSEvent_PersistStableCheckpoint{
			PersistStableCheckpoint: &isspb.PersistStableCheckpoint{
				StableCheckpoint: stableCheckpoint,
			},
		}},
	)
}

func PushCheckpoint(ownModuleID t.ModuleID) *eventpb.Event {
	return Event(ownModuleID, &isspb.ISSEvent{Type: &isspb.ISSEvent_PushCheckpoint{
		PushCheckpoint: &isspb.PushCheckpoint{},
	}})
}

func SBEvent(
	ownModuleID t.ModuleID,
	epoch t.EpochNr,
	instance t.SBInstanceNr,
	event *isspb.SBInstanceEvent,
) *eventpb.Event {
	return Event(
		ownModuleID,
		&isspb.ISSEvent{Type: &isspb.ISSEvent_Sb{Sb: &isspb.SBEvent{
			Epoch:    epoch.Pb(),
			Instance: instance.Pb(),
			Event:    event,
		}}},
	)
}

func LogEntryHashOrigin(ownModuleID t.ModuleID, logEntrySN t.SeqNr) *eventpb.HashOrigin {
	return HashOrigin(ownModuleID, &isspb.ISSHashOrigin{Type: &isspb.ISSHashOrigin_LogEntrySn{
		LogEntrySn: logEntrySN.Pb(),
	}})
}

func SBHashOrigin(ownModuleID t.ModuleID,
	epoch t.EpochNr,
	instance t.SBInstanceNr,
	origin *isspb.SBInstanceHashOrigin,
) *eventpb.HashOrigin {
	return HashOrigin(ownModuleID, &isspb.ISSHashOrigin{Type: &isspb.ISSHashOrigin_Sb{Sb: &isspb.SBHashOrigin{
		Epoch:    epoch.Pb(),
		Instance: instance.Pb(),
		Origin:   origin,
	}}})
}

func StableCheckpointSigVerOrigin(
	ownModuleID t.ModuleID,
	stableCheckpoint *checkpointpb.StableCheckpoint,
) *eventpb.SigVerOrigin {
	return SigVerOrigin(ownModuleID, &isspb.ISSSigVerOrigin{Type: &isspb.ISSSigVerOrigin_StableCheckpoint{
		StableCheckpoint: stableCheckpoint,
	}})
}

func SBSignOrigin(
	ownModuleID t.ModuleID,
	epoch t.EpochNr,
	instance t.SBInstanceNr,
	origin *isspb.SBInstanceSignOrigin,
) *eventpb.SignOrigin {
	return SignOrigin(ownModuleID, &isspb.ISSSignOrigin{Type: &isspb.ISSSignOrigin_Sb{Sb: &isspb.SBSignOrigin{
		Epoch:    epoch.Pb(),
		Instance: instance.Pb(),
		Origin:   origin,
	}}})
}

func SBSigVerOrigin(
	ownModuleID t.ModuleID,
	epoch t.EpochNr,
	instance t.SBInstanceNr,
	origin *isspb.SBInstanceSigVerOrigin,
) *eventpb.SigVerOrigin {
	return SigVerOrigin(ownModuleID, &isspb.ISSSigVerOrigin{Type: &isspb.ISSSigVerOrigin_Sb{Sb: &isspb.SBSigVerOrigin{
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

func SBDeliverEvent(sn t.SeqNr, certData []byte, aborted bool) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_Deliver{
		Deliver: &isspb.SBDeliver{
			Sn:       sn.Pb(),
			CertData: certData,
			Aborted:  aborted,
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

func SBCertRequestEvent() *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_CertRequest{CertRequest: &isspb.SBCertRequest{}}}
}

func SBCertReadyEvent(cert *availabilitypb.Cert) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_CertReady{CertReady: &isspb.SBCertReady{Cert: cert}}}
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
	return &messagepb.Message{DestModule: "iss", Type: &messagepb.Message_Iss{Iss: msg}}
}

func SBMessage(epoch t.EpochNr, instance t.SBInstanceNr, msg *isspb.SBInstanceMessage) *messagepb.Message {
	return Message(&isspb.ISSMessage{Type: &isspb.ISSMessage_Sb{Sb: &isspb.SBMessage{
		Epoch:    epoch.Pb(),
		Instance: instance.Pb(),
		Msg:      msg,
	}}})
}

func StableCheckpointMessage(stableCheckpoint *checkpointpb.StableCheckpoint) *messagepb.Message {
	return Message(&isspb.ISSMessage{Type: &isspb.ISSMessage_StableCheckpoint{StableCheckpoint: stableCheckpoint}})
}
