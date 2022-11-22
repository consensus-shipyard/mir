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
	bfpb "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
	"github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// FakeApp represents a dummy stub application used for testing only.
type FakeApp struct {
	ProtocolModule t.ModuleID

	Membership map[t.NodeID]t.NodeAddress

	// The state of the FakeApp only consists of a counter of processed requests.
	RequestsProcessed uint64
}

func NewFakeApp(protocolModule t.ModuleID, membership map[t.NodeID]t.NodeAddress) *FakeApp {
	return &FakeApp{
		ProtocolModule:    protocolModule,
		Membership:        membership,
		RequestsProcessed: 0,
	}
}

func (fa *FakeApp) ApplyEvents(eventsIn *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsSequentially(eventsIn, fa.ApplyEvent)
}

func (fa *FakeApp) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
		// no actions on init
	case *eventpb.Event_BatchFetcher:
		switch e := e.BatchFetcher.Type.(type) {
		case *bfpb.Event_NewOrderedBatch:
			return fa.applyNewOrderedBatch(e.NewOrderedBatch)
		default:
			return nil, fmt.Errorf("unexpected availability event type: %T", e)
		}
	case *eventpb.Event_AppSnapshotRequest:
		return events.ListOf(events.AppSnapshotResponse(
			t.ModuleID(e.AppSnapshotRequest.ReplyTo),
			fa.Snapshot(),
		)), nil
	case *eventpb.Event_AppRestoreState:
		if err := fa.RestoreState(e.AppRestoreState.Checkpoint); err != nil {
			return nil, fmt.Errorf("app restore state error: %w", err)
		}
	case *eventpb.Event_NewEpoch:
		return events.ListOf(events.NewConfig(fa.ProtocolModule, t.EpochNr(e.NewEpoch.EpochNr), fa.Membership)), nil
	case *eventpb.Event_Checkpoint:
		switch e := e.Checkpoint.Type.(type) {
		case *checkpointpb.Event_StableCheckpoint:
			return events.EmptyList(), nil // Ignore checkpoints.
		default:
			return nil, fmt.Errorf("unexpected checkpoint event type: %T", e)
		}
	default:
		return nil, fmt.Errorf("unexpected type of App event: %T", event.Type)
	}

	return events.EmptyList(), nil
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (fa *FakeApp) ImplementsModule() {}

// ApplyBatch applies a batch of transactions.
func (fa *FakeApp) applyNewOrderedBatch(batch *bfpb.NewOrderedBatch) (*events.EventList, error) {

	for _, req := range batch.Txs {
		fa.RequestsProcessed++
		fmt.Printf("Received request: %q. Processed requests: %d\n", string(req.Data), fa.RequestsProcessed)
	}
	return events.EmptyList(), nil
}

func (fa *FakeApp) Snapshot() []byte {
	return uint64ToBytes(fa.RequestsProcessed)
}

func (fa *FakeApp) RestoreState(checkpoint *checkpointpb.StableCheckpoint) error {
	fa.RequestsProcessed = uint64FromBytes(checkpoint.Snapshot.AppData)
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
