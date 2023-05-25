package leaderselectionpolicy

import (
	"sort"

	"github.com/fxamacker/cbor/v2"

	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/serializing"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/membutil"
)

// BlacklistLeaderPolicy represents the state of the blacklist leader selection policy.
type BlacklistLeaderPolicy struct {
	Membership *trantorpbtypes.Membership
	Suspected  map[t.NodeID]tt.EpochNr
	// All fields must be exported, otherwise they would be ignored when serializing.
}

func NewBlackListLeaderPolicy(membership *trantorpbtypes.Membership) *BlacklistLeaderPolicy {
	return &BlacklistLeaderPolicy{
		membership,
		make(map[t.NodeID]tt.EpochNr, len(membership.Nodes)),
	}
}

// Leaders returns a list of least recently suspected leaders
// that cumulatively have a strong quorum of the voting power.
func (l *BlacklistLeaderPolicy) Leaders() []t.NodeID {

	nodeIDs := maputil.GetKeys(l.Membership.Nodes)

	// Sort the values in ascending order of the latest epoch where they each were suspected.
	sort.Slice(nodeIDs, func(i, j int) bool {
		_, oki := l.Suspected[nodeIDs[i]]
		_, okj := l.Suspected[nodeIDs[j]]

		// If both nodeIDs have never been suspected, then lexicographical order (both will be leaders)
		if !oki && !okj {
			return string(nodeIDs[i]) < string(nodeIDs[j])
		}

		// If one node is suspected and the other is not, the suspected one comes last
		if !oki && okj {
			return true
		}
		if oki && !okj {
			return false
		}

		// If both keys are in the suspected map, sort by value in the suspected map
		return l.Suspected[nodeIDs[i]] < l.Suspected[nodeIDs[j]]
	})

	// Calculate the minimal number of least recently suspected nodes for a strong quorum.
	minLeaders := 1 // We always need at least one leader.
	for !membutil.HaveStrongQuorum(l.Membership, nodeIDs[:minLeaders]) {
		minLeaders++
	}

	// Calculate number of leaders
	numLeaders := len(l.Membership.Nodes) - len(l.Suspected)
	if numLeaders < minLeaders {
		numLeaders = minLeaders // must at least return l.minLeaders
	}

	// return the leadersSize least recently suspected nodes as currentLeaders
	return nodeIDs[:numLeaders]
}

// Suspect adds a new suspect to the list of suspects, or updates its epoch where it was suspected to the given epoch
// if this one is more recent than the one it already has.
func (l *BlacklistLeaderPolicy) Suspect(e tt.EpochNr, node t.NodeID) {
	if _, ok := l.Membership.Nodes[node]; !ok { // node is a not a member
		//TODO error but cannot be passed through
		return
	}

	if epochNr, ok := l.Suspected[node]; !ok || epochNr < e {
		l.Suspected[node] = e // update with latest epoch that node was suspected
	}
}

// Reconfigure informs the leader selection policy about a change in the membership.
func (l *BlacklistLeaderPolicy) Reconfigure(membership *trantorpbtypes.Membership) LeaderSelectionPolicy {
	suspected := make(map[t.NodeID]tt.EpochNr, len(membership.Nodes))
	for _, nodeID := range maputil.GetKeys(membership.Nodes) {
		if epoch, ok := l.Suspected[nodeID]; ok {
			suspected[nodeID] = epoch
		}
	}

	newPolicy := BlacklistLeaderPolicy{
		membership,
		suspected,
	}

	return &newPolicy
}

func (l *BlacklistLeaderPolicy) Bytes() ([]byte, error) {
	ser, err := getEncMode().Marshal(*l)
	if err != nil {
		return nil, err
	}
	out := serializing.Uint64ToBytes(uint64(Blacklist))
	out = append(out, ser...)
	return out, nil
}

func BlacklistLeaderPolicyFromBytes(data []byte) (*BlacklistLeaderPolicy, error) {
	b := &BlacklistLeaderPolicy{}
	if err := cbor.Unmarshal(data, b); err != nil {
		return nil, err
	}
	return b, nil
}
