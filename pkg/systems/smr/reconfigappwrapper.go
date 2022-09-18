package smr

import (
	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// reconfigurableAppLogic is a wrapper around a static app logic that allows a static app logic
// to be used in a reconfigurable SMR system.
// It implements a trivial reconfiguration logic that always returns the same membership.
type reconfigurableAppLogic struct {
	staticAppLogic StaticAppLogic
	membership     map[t.NodeID]t.NodeAddress
}

// ApplyTXs only delegates to the static app logic.
func (ra *reconfigurableAppLogic) ApplyTXs(txs []*requestpb.Request) error {
	return ra.staticAppLogic.ApplyTXs(txs)
}

// NewEpoch always returns the same static pre-configured membership.
func (ra *reconfigurableAppLogic) NewEpoch(_ t.EpochNr) (map[t.NodeID]t.NodeAddress, error) {
	return ra.membership, nil
}

// Snapshot only delegates to the static app logic.
func (ra *reconfigurableAppLogic) Snapshot() ([]byte, error) {
	return ra.staticAppLogic.Snapshot()
}

// RestoreState only delegates to the static app logic, ignoring the epoch config.
func (ra *reconfigurableAppLogic) RestoreState(snapshot []byte, _ *commonpb.EpochConfig) error {
	return ra.staticAppLogic.RestoreState(snapshot)
}
