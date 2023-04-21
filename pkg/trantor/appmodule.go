package trantor

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net"
	apppbevents "github.com/filecoin-project/mir/pkg/pb/apppb/events"
	apppbtypes "github.com/filecoin-project/mir/pkg/pb/apppb/types"
	bfpb "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	isspbevents "github.com/filecoin-project/mir/pkg/pb/isspb/events"
	t "github.com/filecoin-project/mir/pkg/types"
)

// AppModule is the module within the SMR system that handles the application logic.
type AppModule struct {
	// appLogic is the user-provided application logic.
	appLogic AppLogic

	// transport is the network transport.
	// It is required to keep a reference to it in order to connect to new members when a new epoch starts.
	transport net.Transport

	// protocolModule is the ID of the protocol module.
	// It is required to send events (new configurations) to the protocol module.
	// TODO: Remove this. Instead, save the origin module ID in the NewEpoch event and use that.
	protocolModule t.ModuleID
}

// NewAppModule creates a new AppModule.
func NewAppModule(appLogic AppLogic, transport net.Transport, protocolModule t.ModuleID) *AppModule {
	return &AppModule{
		appLogic:       appLogic,
		transport:      transport,
		protocolModule: protocolModule,
	}
}

// ApplyEvents applies a list of events to the AppModule.
func (m *AppModule) ApplyEvents(eventsIn *events.EventList) (*events.EventList, error) {
	// Events must be applied sequentially, since the application logic is not expected to be thread-safe.
	return modules.ApplyEventsSequentially(eventsIn, m.ApplyEvent)
}

// ApplyEvent applies a single event to the AppModule.
func (m *AppModule) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
		// no actions on init
	case *eventpb.Event_BatchFetcher:
		switch e := e.BatchFetcher.Type.(type) {
		case *bfpb.Event_NewOrderedBatch:
			return m.applyNewOrderedBatch(e.NewOrderedBatch)
		default:
			return nil, fmt.Errorf("unexpected availability event type: %T", e)
		}
	case *eventpb.Event_App:
		switch e := apppbtypes.EventFromPb(e.App).Type.(type) {
		case *apppbtypes.Event_SnapshotRequest:
			return m.applyAppSnapshotRequest(e.SnapshotRequest)
		case *apppbtypes.Event_RestoreState:
			return m.applyAppRestoreState(e.RestoreState)
		case *apppbtypes.Event_NewEpoch:
			return m.applyNewEpoch(e.NewEpoch)
		default:
			return nil, fmt.Errorf("unexpected app event type: %T", e)
		}
	case *eventpb.Event_Checkpoint:
		switch e := checkpointpbtypes.EventFromPb(e.Checkpoint).Type.(type) {
		case *checkpointpbtypes.Event_StableCheckpoint:
			return m.applyStableCheckpoint((*checkpoint.StableCheckpoint)(e.StableCheckpoint))
		default:
			return nil, fmt.Errorf("unexpected checkpoint event type: %T", e)
		}
	default:
		return nil, fmt.Errorf("unexpected type of App event: %T", event.Type)
	}

	return events.EmptyList(), nil
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (m *AppModule) ImplementsModule() {}

// applyNewOrderedBatch sequentially applies a batch of ordered transactions to the application logic
// and returns an empty event list.
func (m *AppModule) applyNewOrderedBatch(batch *bfpb.NewOrderedBatch) (*events.EventList, error) {

	if err := m.appLogic.ApplyTXs(batch.Txs); err != nil {
		return nil, err
	}
	return events.EmptyList(), nil
}

// applyAppSnapshotRequest takes a snapshot of the application state
// and returns an event that contains the snapshot back to the originator of the request.
func (m *AppModule) applyAppSnapshotRequest(snapshotRequest *apppbtypes.SnapshotRequest) (*events.EventList, error) {
	snapshot, err := m.appLogic.Snapshot()
	if err != nil {
		return nil, err
	}
	return events.ListOf(apppbevents.Snapshot(
		snapshotRequest.ReplyTo,
		snapshot,
	).Pb()), nil
}

// applyRestoreState restores the application state from a snapshot.
// The snapshot contains both the application state and the configuration corresponding to that version of the state.
// applyRestoreState returns an empty event list.
func (m *AppModule) applyAppRestoreState(restoreState *apppbtypes.RestoreState) (*events.EventList, error) {
	if err := m.appLogic.RestoreState((*checkpoint.StableCheckpoint)(restoreState.Checkpoint)); err != nil {
		return nil, fmt.Errorf("app restore state error: %w", err)
	}
	return events.EmptyList(), nil
}

// applyNewEpoch applies a new epoch event.
// It informs the application logic of the new epoch and returns an event (to the protocol module)
// containing the configuration for the new epoch.
func (m *AppModule) applyNewEpoch(newEpoch *apppbtypes.NewEpoch) (*events.EventList, error) {
	membership, err := m.appLogic.NewEpoch(newEpoch.EpochNr)
	if err != nil {
		return nil, fmt.Errorf("error handling NewEpoch event: %w", err)
	}
	m.transport.Connect(membership)
	// TODO: Save the origin module ID in the event and use it here, instead of saving the m.protocolModule.
	return events.ListOf(isspbevents.NewConfig(
		m.protocolModule,
		newEpoch.EpochNr,
		membership,
	).Pb()), nil
}

func (m *AppModule) applyStableCheckpoint(stableCheckpoint *checkpoint.StableCheckpoint) (*events.EventList, error) {
	if err := m.appLogic.Checkpoint(stableCheckpoint); err != nil {
		return nil, fmt.Errorf("error handling StableCheckpoint event: %w", err)
	}
	return events.EmptyList(), nil
}
