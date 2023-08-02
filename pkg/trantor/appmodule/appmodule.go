package appmodule

import (
	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net"
	apppbdsl "github.com/filecoin-project/mir/pkg/pb/apppb/dsl"
	bfpbdsl "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/dsl"
	checkpointpbdsl "github.com/filecoin-project/mir/pkg/pb/checkpointpb/dsl"
	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	isspbdsl "github.com/filecoin-project/mir/pkg/pb/isspb/dsl"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/membutil"
)

// AppModule is the module within the SMR system that handles the application logic.
type AppModule struct {
	// appLogic is the user-provided application logic.
	appLogic AppLogic

	// transport is the network transport.
	// It is required to keep a reference to it in order to connect to new members when a new epoch starts.
	transport net.Transport
}

// NewAppModule creates a new AppModule.
func NewAppModule(appLogic AppLogic, transport net.Transport, moduleID t.ModuleID) modules.PassiveModule {

	appModule := &AppModule{
		appLogic:  appLogic,
		transport: transport,
	}

	m := dsl.NewModule(moduleID)

	// UponNewOrderedBatch sequentially apply a batch of ordered transactions to the application logic
	// and returns an empty event list.
	bfpbdsl.UponNewOrderedBatch(m, func(txs []*trantorpbtypes.Transaction) error {
		return appModule.appLogic.ApplyTXs(txs)
	})

	// UponSnapshotRequest return an event that contains the snapshot back to the originator of the request.
	apppbdsl.UponSnapshotRequest(m, func(replyTo t.ModuleID) error {
		snapshot, err := appModule.appLogic.Snapshot()
		if err != nil {
			return err
		}

		apppbdsl.Snapshot(m, replyTo, snapshot)
		return nil
	})

	// UponRestoreState restores the application state from a snapshot.
	// The snapshot contains both the application state and the configuration corresponding to that version of the state.
	// Emit no event
	apppbdsl.UponRestoreState(m, func(chkp *checkpointpbtypes.StableCheckpoint) error {
		if err := appModule.appLogic.RestoreState((*checkpoint.StableCheckpoint)(chkp)); err != nil {
			return es.Errorf("app restore state error: %w", err)
		}
		return nil
	})

	// UponNewEpoch inform the application logic of the new epoch and emit an event (to the protocol module)
	// containing the configuration for the new epoch.
	apppbdsl.UponNewEpoch(m, func(epochNr types.EpochNr, protocolModule t.ModuleID) error {
		membership, err := appModule.appLogic.NewEpoch(epochNr)
		if err != nil {
			return es.Errorf("error handling NewEpoch event: %w", err)
		}

		if err = membutil.Valid(membership); err != nil {
			return es.Errorf("app logic provided invalid membership: %w", err)
		}

		appModule.transport.Connect(membership)
		isspbdsl.NewConfig(m, protocolModule, epochNr, membership)
		return nil
	})

	checkpointpbdsl.UponStableCheckpoint(m, func(sn types.SeqNr, snapshot *trantorpbtypes.StateSnapshot, cert map[t.NodeID][]uint8) error {
		chkp := &checkpoint.StableCheckpoint{
			Sn:       sn,
			Snapshot: snapshot,
			Cert:     cert,
		}
		if err := appModule.appLogic.Checkpoint(chkp); err != nil {
			return es.Errorf("error handling StableCheckpoint event: %w", err)
		}

		return nil
	})

	return m
}
