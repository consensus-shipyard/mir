// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	bfpb "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type App struct {
	logging.Logger

	ProtocolModule t.ModuleID
	Membership     map[t.NodeID]t.NodeAddress
}

func (App) ImplementsModule() {}

func (a *App) ApplyEvents(eventsIn *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsSequentially(eventsIn, a.ApplyEvent)
}

func (a *App) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
	case *eventpb.Event_BatchFetcher:
		switch e := e.BatchFetcher.Type.(type) {
		case *bfpb.Event_NewOrderedBatch:
			for _, req := range e.NewOrderedBatch.Txs {
				a.Log(logging.LevelDebug, fmt.Sprintf("Delivered request %v from client %v", req.ReqNo, req.ClientId))
			}
		default:
			return nil, fmt.Errorf("unexpected availability event type: %T", e)
		}
	case *eventpb.Event_AppSnapshotRequest:
		return events.ListOf(events.AppSnapshotResponse(
			t.ModuleID(e.AppSnapshotRequest.Module),
			[]byte{},
			e.AppSnapshotRequest.Origin,
		)), nil
	case *eventpb.Event_AppRestoreState:
	case *eventpb.Event_NewEpoch:
		return events.ListOf(events.NewConfig(a.ProtocolModule, a.Membership)), nil
	default:
		return nil, fmt.Errorf("unexpected type of App event: %T", event.Type)
	}

	return events.EmptyList(), nil
}
