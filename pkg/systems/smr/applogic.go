package smr

import (
	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// AppLogic represents the application logic of an SMR system.
// It holds the state of the replicated state machine and defines the semantics of transactions applied to it.
// It also defines the membership of the system through the return value of NewEpoch.
type AppLogic interface {

	// ApplyTXs applies a batch of transactions to the state machine.
	ApplyTXs(txs []*requestpb.Request) error

	// NewEpoch is called by the SMR system when a new epoch is started.
	// It returns the membership of a new epoch.
	// Note that, due to pipelining, the membership NewEpoch returns is not necessarily used immediately
	// in the epoch that is just starting.
	// It might define the membership of a future epoch.
	NewEpoch(nr t.EpochNr) (map[t.NodeID]t.NodeAddress, error)

	// Snapshot returns a snapshot of the application state.
	Snapshot() ([]byte, error)

	// RestoreState restores the application state from stable checkpoint.
	RestoreState(checkpoint *checkpoint.StableCheckpoint) error

	// Checkpoint is called by the SMR system when it produces a checkpoint.
	// A checkpoint contains the state of the application at a particular point in time
	// from which the system can recover, including the configuration of the system at that point in time
	// and a certificate of validity of the checkpoint.
	Checkpoint(checkpoint *checkpoint.StableCheckpoint) error
}

// StaticAppLogic represents the logic of an application that is not reconfigurable.
// It is simpler than AppLogic, as it does not need to define the membership of the system.
type StaticAppLogic interface {

	// ApplyTXs applies a batch of transactions to the state machine.
	ApplyTXs(txs []*requestpb.Request) error

	// Snapshot returns a snapshot of the application state.
	Snapshot() ([]byte, error)

	// RestoreState restores the application state from a snapshot.
	RestoreState(checkpoint *checkpoint.StableCheckpoint) error

	// Checkpoint is called by the SMR system when it produces a checkpoint.
	// A checkpoint contains the state of the application at a particular point in time
	// from which the system can recover, including the configuration of the system at that point in time
	// and a certificate of validity of the checkpoint.
	Checkpoint(checkpoint *checkpoint.StableCheckpoint) error
}

// AppLogicFromStatic augments the static application logic with a default implementation of the reconfiguration logic
// that simply always uses the same membership.
func AppLogicFromStatic(staticAppLogic StaticAppLogic, membership map[t.NodeID]t.NodeAddress) AppLogic {
	return &reconfigurableAppLogic{
		staticAppLogic: staticAppLogic,
		membership:     membership,
	}
}
