// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type App struct {
	logging.Logger
}

func (App) ImplementsModule() {}

func (a *App) ApplyEvents(eventsIn *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsSequentially(eventsIn, a.ApplyEvent)
}

func (a *App) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
	case *eventpb.Event_Deliver:
		for _, req := range e.Deliver.Batch.Requests {
			a.Log(logging.LevelDebug, fmt.Sprintf("Delivered request %v from client %v", req.Req.ReqNo, req.Req.ClientId))
		}
	case *eventpb.Event_AppSnapshotRequest:
		return events.ListOf(events.AppSnapshot(
			t.ModuleID(e.AppSnapshotRequest.Module),
			t.EpochNr(e.AppSnapshotRequest.Epoch),
			nil,
		)), nil
	case *eventpb.Event_AppRestoreState:
	default:
		return nil, fmt.Errorf("unexpected type of App event: %T", event.Type)
	}

	return events.EmptyList(), nil
}
