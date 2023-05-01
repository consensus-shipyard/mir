package trantor

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/dsl"
	checkpointpbdsl "github.com/filecoin-project/mir/pkg/pb/checkpointpb/dsl"
	commonpbtypes "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	isspbdsl "github.com/filecoin-project/mir/pkg/pb/isspb/dsl"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	requestpbtypes "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	"github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"

	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net"
	apppbdsl "github.com/filecoin-project/mir/pkg/pb/apppb/dsl"
	bfpbdsl "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/dsl"
	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
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
func NewAppModule(appLogic AppLogic, transport net.Transport, protocolModule t.ModuleID) modules.PassiveModule {

	appModule := &AppModule{
		appLogic:  appLogic,
		transport: transport,
	}

	m := dsl.NewModule("appModule")

	// UponNewOrderedBatch sequentially apply a batch of ordered transactions to the application logic
	// and returns an empty event list.
	bfpbdsl.UponNewOrderedBatch(m, func(txs []*requestpbtypes.Request) error {

		if err := appModule.appLogic.ApplyTXs(
			sliceutil.Transform(
				txs,
				func(i int, tx *requestpbtypes.Request) *requestpb.Request {
					return tx.Pb()
				})); err != nil {
			return err
		}

		return nil
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

	// UponRestoreState restore the application state from a snapshot.
	// The snapshot contains both the application state and the configuration corresponding to that version of the state.
	// Emit no event
	apppbdsl.UponRestoreState(m, func(chkp *checkpointpbtypes.StableCheckpoint) error {
		if err := appModule.appLogic.RestoreState((*checkpoint.StableCheckpoint)(chkp)); err != nil {
			return fmt.Errorf("app restore state error: %w", err)
		}
		return nil
	})

	// UponNewEpoch inform the application logic of the new epoch and emit an event (to the protocol module)
	// containing the configuration for the new epoch.
	apppbdsl.UponNewEpoch(m, func(epochNr types.EpochNr, protocolModule t.ModuleID) error {
		membership, err := appModule.appLogic.NewEpoch(epochNr)
		if err != nil {
			return fmt.Errorf("error handling NewEpoch event: %w", err)
		}
		appModule.transport.Connect(membership)
		// TODO: Save the origin module ID in the event and use it here, instead of saving the m.protocolModule.
		isspbdsl.NewConfig(m, protocolModule, epochNr, membership)
		return nil
	})

	checkpointpbdsl.UponStableCheckpoint(m, func(sn types.SeqNr, snapshot *commonpbtypes.StateSnapshot, cert map[t.NodeID][]uint8) error {
		chkp := &checkpoint.StableCheckpoint{
			Sn:       sn,
			Snapshot: snapshot,
			Cert:     cert,
		}
		if err := appModule.appLogic.Checkpoint(chkp); err != nil {
			return fmt.Errorf("error handling StableCheckpoint event: %w", err)
		}

		return nil
	})

	return m
}
