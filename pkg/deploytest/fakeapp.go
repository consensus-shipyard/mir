/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deploytest

import (
	"encoding/binary"
	"fmt"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"

	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
)

// FakeApp represents a dummy stub application used for testing only.
type FakeApp struct {

	// The state of the FakeApp only consists of a counter of processed requests.
	RequestsProcessed uint64
}

func (fa *FakeApp) ApplyEvents(eventsIn *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsSequentially(eventsIn, fa.ApplyEvent)
}

func (fa *FakeApp) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpb.Event_Deliver:
		if err := fa.ApplyBatch(e.Deliver.Batch); err != nil {
			return nil, fmt.Errorf("app batch delivery error: %w", err)
		}
	case *eventpb.Event_AppSnapshotRequest:
		data, err := fa.Snapshot()
		if err != nil {
			return nil, fmt.Errorf("app snapshot error: %w", err)
		}
		return (&events.EventList{}).PushBack(events.AppSnapshot(
			t.ModuleID(e.AppSnapshotRequest.Module),
			t.EpochNr(e.AppSnapshotRequest.Epoch),
			data,
		)), nil
	case *eventpb.Event_AppRestoreState:
		if err := fa.RestoreState(e.AppRestoreState.Data); err != nil {
			return nil, fmt.Errorf("app restore state error: %w", err)
		}
	default:
		return nil, fmt.Errorf("unexpected type of App event: %T", event.Type)
	}

	return &events.EventList{}, nil
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

func (fa *FakeApp) Snapshot() ([]byte, error) {
	return uint64ToBytes(fa.RequestsProcessed), nil
}

func (fa *FakeApp) RestoreState(snapshot []byte) error {
	return fmt.Errorf("we don't support state transfer in this test (yet)")
}

func uint64ToBytes(value uint64) []byte {
	byteValue := make([]byte, 8)
	binary.LittleEndian.PutUint64(byteValue, value)
	return byteValue
}
