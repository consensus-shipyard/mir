/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deploytest

import (
	"encoding/binary"
	"fmt"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// FakeApp represents a dummy stub application used for testing only.
type FakeApp struct {
	ProtocolModule t.ModuleID

	Membership map[t.NodeID]t.NodeAddress

	// The state of the FakeApp only consists of a counter of processed requests.
	RequestsProcessed uint64
}

func (fa *FakeApp) ApplyEvents(eventsIn *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsSequentially(eventsIn, fa.ApplyEvent)
}

func (fa *FakeApp) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
		// no actions on init
	case *eventpb.Event_Deliver:
		if err := fa.ApplyBatch(e.Deliver.Batch); err != nil {
			return nil, fmt.Errorf("app batch delivery error: %w", err)
		}
	case *eventpb.Event_AppSnapshotRequest:
		return events.ListOf(events.AppSnapshotResponse(
			t.ModuleID(e.AppSnapshotRequest.Module),
			fa.Snapshot(),
			e.AppSnapshotRequest.Origin,
		)), nil
	case *eventpb.Event_AppRestoreState:
		if err := fa.RestoreState(e.AppRestoreState.Snapshot); err != nil {
			return nil, fmt.Errorf("app restore state error: %w", err)
		}
	case *eventpb.Event_NewEpoch:
		return events.ListOf(events.NewConfig(fa.ProtocolModule, fa.Membership)), nil
	default:
		return nil, fmt.Errorf("unexpected type of App event: %T", event.Type)
	}

	return events.EmptyList(), nil
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (fa *FakeApp) ImplementsModule() {}

// ApplyBatch applies a batch of requests
func (fa *FakeApp) ApplyBatch(batch *requestpb.Batch) error {
	for _, req := range batch.Requests {
		fa.RequestsProcessed++
		fmt.Printf("Received request: \"%s\". Processed requests: %d\n", string(req.Req.Data), fa.RequestsProcessed)
	}
	return nil
}

func (fa *FakeApp) Snapshot() []byte {
	return uint64ToBytes(fa.RequestsProcessed)
}

func (fa *FakeApp) RestoreState(snapshot *commonpb.StateSnapshot) error {
	fa.RequestsProcessed = uint64FromBytes(snapshot.AppData)
	return nil
}

func uint64ToBytes(value uint64) []byte {
	byteValue := make([]byte, 8)
	binary.LittleEndian.PutUint64(byteValue, value)
	return byteValue
}

func uint64FromBytes(bytes []byte) uint64 {
	return binary.LittleEndian.Uint64(bytes)
}
