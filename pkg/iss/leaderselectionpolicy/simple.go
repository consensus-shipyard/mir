package leaderselectionpolicy

import (
	"github.com/fxamacker/cbor/v2"

	t "github.com/filecoin-project/mir/pkg/types"
)

// The SimpleLeaderPolicy is a trivial leader selection policy.
// It must be initialized with a set of node IDs and always returns that full set as leaders,
// regardless of which nodes have been suspected. In other words, each node is leader each epoch with this policy.
type SimpleLeaderPolicy struct {
	Membership []t.NodeID
}

func NewSimpleLeaderPolicy(membership []t.NodeID) *SimpleLeaderPolicy {
	return &SimpleLeaderPolicy{
		membership,
	}
}

// Leaders always returns the whole membership for the SimpleLeaderPolicy. All nodes are always leaders.
func (simple *SimpleLeaderPolicy) Leaders() []t.NodeID {
	// All nodes are always leaders.
	return simple.Membership
}

// Suspect does nothing for the SimpleLeaderPolicy.
func (simple *SimpleLeaderPolicy) Suspect(_ t.EpochNr, _ t.NodeID) {
	// Do nothing.
}

// Reconfigure informs the leader selection policy about a change in the membership.
func (simple *SimpleLeaderPolicy) Reconfigure(nodeIDs []t.NodeID) LeaderSelectionPolicy {
	newPolicy := SimpleLeaderPolicy{Membership: make([]t.NodeID, len(nodeIDs))}
	copy(newPolicy.Membership, nodeIDs)
	return &newPolicy
}

func (simple *SimpleLeaderPolicy) Bytes() ([]byte, error) {
	ser, err := getEncMode().Marshal(simple)
	if err != nil {
		return nil, err
	}
	out := t.Uint64ToBytes(uint64(Simple))
	out = append(out, ser...)
	return out, nil
}

func SimpleLeaderPolicyFromBytes(data []byte) (*SimpleLeaderPolicy, error) {
	b := &SimpleLeaderPolicy{}
	err := cbor.Unmarshal(data, b)
	return b, err
}
