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

func PushCheckpoint(ownModuleID t.ModuleID) *eventpb.Event {
	return Event(ownModuleID, &isspb.ISSEvent{Type: &isspb.ISSEvent_PushCheckpoint{
		PushCheckpoint: &isspb.PushCheckpoint{},
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
