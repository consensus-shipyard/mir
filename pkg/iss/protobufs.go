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
	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/pb/checkpointpb"
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

func SigVerOrigin(ownModuleID t.ModuleID, origin *isspb.ISSSigVerOrigin) *eventpb.SigVerOrigin {
	return &eventpb.SigVerOrigin{Module: ownModuleID.Pb(), Type: &eventpb.SigVerOrigin_Iss{Iss: origin}}
}

func PushCheckpoint(ownModuleID t.ModuleID) *eventpb.Event {
	return Event(ownModuleID, &isspb.ISSEvent{Type: &isspb.ISSEvent_PushCheckpoint{
		PushCheckpoint: &isspb.PushCheckpoint{},
	}})
}

func LogEntryHashOrigin(ownModuleID t.ModuleID, logEntrySN t.SeqNr) *eventpb.HashOrigin {
	return HashOrigin(ownModuleID, &isspb.ISSHashOrigin{Type: &isspb.ISSHashOrigin_LogEntrySn{
		LogEntrySn: logEntrySN.Pb(),
	}})
}

func StableCheckpointSigVerOrigin(
	ownModuleID t.ModuleID,
	stableCheckpoint *checkpointpb.StableCheckpoint,
) *eventpb.SigVerOrigin {
	return SigVerOrigin(ownModuleID, &isspb.ISSSigVerOrigin{Type: &isspb.ISSSigVerOrigin_StableCheckpoint{
		StableCheckpoint: stableCheckpoint,
	}})
}

// ============================================================
// Messages
// ============================================================

func Message(msg *isspb.ISSMessage) *messagepb.Message {
	return &messagepb.Message{DestModule: "iss", Type: &messagepb.Message_Iss{Iss: msg}}
}

func StableCheckpointMessage(stableCheckpoint *checkpoint.StableCheckpoint) *messagepb.Message {
	return Message(&isspb.ISSMessage{Type: &isspb.ISSMessage_StableCheckpoint{StableCheckpoint: stableCheckpoint.Pb()}})
}
