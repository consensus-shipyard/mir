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
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
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

// EpochConfig represents the configuration of the system during one epoch
func EpochConfig(
	epochNr tt.EpochNr,
	firstSn tt.SeqNr,
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
