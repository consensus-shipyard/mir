package leaderselectionpolicy

import (
	"github.com/fxamacker/cbor/v2"

	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// The SimpleLeaderPolicy is a trivial leader selection policy.
// It must be initialized with a set of node IDs and always returns that full set as leaders,
// regardless of which nodes have been suspected. In other words, each node is leader each epoch with this policy.
type SimpleLeaderPolicy struct {
	Membership []t.NodeID
}

func NewSimpleLeaderPolicy(membership *trantorpbtypes.Membership) *SimpleLeaderPolicy {
	return &SimpleLeaderPolicy{
		maputil.GetSortedKeys(membership.Nodes),
	}
}

// Leaders always returns the whole membership for the SimpleLeaderPolicy. All nodes are always leaders.
func (simple *SimpleLeaderPolicy) Leaders() []t.NodeID {
	// All nodes are always leaders.
	return simple.Membership
}

// Suspect does nothing for the SimpleLeaderPolicy.
func (simple *SimpleLeaderPolicy) Suspect(_ tt.EpochNr, _ t.NodeID) {
	// Do nothing.
}

// Reconfigure informs the leader selection policy about a change in the membership.
func (simple *SimpleLeaderPolicy) Reconfigure(membership *trantorpbtypes.Membership) LeaderSelectionPolicy {
	newPolicy := SimpleLeaderPolicy{Membership: make([]t.NodeID, len(membership.Nodes))}
	copy(newPolicy.Membership, maputil.GetSortedKeys(membership.Nodes))
	return &newPolicy
}

func (simple *SimpleLeaderPolicy) Bytes() ([]byte, error) {
	ser, err := getEncMode().Marshal(simple)
	if err != nil {
		return nil, err
	}
	out := []byte(Simple)
	out = append(out, ser...)
	return out, nil
}

func SimpleLeaderPolicyFromBytes(data []byte) (*SimpleLeaderPolicy, error) {
	b := &SimpleLeaderPolicy{}
	err := cbor.Unmarshal(data, b)
	return b, err
}
