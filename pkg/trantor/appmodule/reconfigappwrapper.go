package appmodule

import (
	"github.com/filecoin-project/mir/pkg/checkpoint"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
)

// reconfigurableAppLogic is a wrapper around a static app logic that allows a static app logic
// to be used in a reconfigurable SMR system.
// It implements a trivial reconfiguration logic that always returns the same membership.
type reconfigurableAppLogic struct {
	staticAppLogic StaticAppLogic
	membership     *trantorpbtypes.Membership
}

// ApplyTXs only delegates to the static app logic.
func (ra *reconfigurableAppLogic) ApplyTXs(txs []*trantorpbtypes.Transaction) error {
	return ra.staticAppLogic.ApplyTXs(txs)
}

// NewEpoch always returns the same static pre-configured membership.
func (ra *reconfigurableAppLogic) NewEpoch(_ tt.EpochNr) (*trantorpbtypes.Membership, error) {
	return ra.membership, nil
}

// Snapshot only delegates to the static app logic.
func (ra *reconfigurableAppLogic) Snapshot() ([]byte, error) {
	return ra.staticAppLogic.Snapshot()
}

// RestoreState only delegates to the static app logic, ignoring the epoch config.
func (ra *reconfigurableAppLogic) RestoreState(checkpoint *checkpoint.StableCheckpoint) error {
	return ra.staticAppLogic.RestoreState(checkpoint)
}

// Checkpoint only delegates to the static app logic, ignoring the epoch config.
func (ra *reconfigurableAppLogic) Checkpoint(checkpoint *checkpoint.StableCheckpoint) error {
	return ra.staticAppLogic.Checkpoint(checkpoint)
}
