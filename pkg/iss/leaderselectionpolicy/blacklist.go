package leaderselectionpolicy

import (
	"sort"

	"github.com/fxamacker/cbor/v2"

	"github.com/filecoin-project/mir/pkg/serializing"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

type BlacklistLeaderPolicy struct {
	Membership map[t.NodeID]struct{}
	Suspected  map[t.NodeID]tt.EpochNr
	MinLeaders int
}

func NewBlackListLeaderPolicy(members []t.NodeID, MinLeaders int) *BlacklistLeaderPolicy {
	membership := make(map[t.NodeID]struct{}, len(members))
	for _, node := range members {
		membership[node] = struct{}{}
	}
	return &BlacklistLeaderPolicy{
		membership,
		make(map[t.NodeID]tt.EpochNr, len(members)),
		MinLeaders,
	}
}

// Leaders always returns the whole membership for the SimpleLeaderPolicy. All nodes are always leaders.
func (l *BlacklistLeaderPolicy) Leaders() []t.NodeID {

	curLeaders := make([]t.NodeID, 0, len(l.Membership))
	for nodeID := range l.Membership {
		curLeaders = append(curLeaders, nodeID)
	}

	// Sort the values in ascending order of latest epoch where they each were suspected
	sort.Slice(curLeaders, func(i, j int) bool {
		_, oki := l.Suspected[curLeaders[i]]
		_, okj := l.Suspected[curLeaders[j]]
		// If both nodeIDs have never been suspected, then lexicographical order (both will be leaders)
		if !oki && !okj {
			return string(curLeaders[i]) < string(curLeaders[j])
		}

		// If one node is suspected and the other is not, the suspected one comes last
		if !oki && okj {
			return true
		}
		if oki && !okj {
			return false
		}

		// If both keys are in the suspected map, sort by value in the suspected map
		return l.Suspected[curLeaders[i]] < l.Suspected[curLeaders[j]]
	})

	// Calculate number of leaders
	leadersSize := len(l.Membership) - len(l.Suspected)
	if leadersSize < l.MinLeaders {
		leadersSize = l.MinLeaders // must at least return l.MinLeaders
	}

	// return the leadersSize least recently suspected nodes as currentLeaders
	return curLeaders[:leadersSize]
}

// Suspect adds a new suspect to the list of suspects, or updates its epoch where it was suspected to the given epoch
// if this one is more recent than the one it already has
func (l *BlacklistLeaderPolicy) Suspect(e tt.EpochNr, node t.NodeID) {
	if _, ok := l.Membership[node]; !ok { //node is a not a member
		//TODO error but cannot be passed through
		return
	}

	if epochNr, ok := l.Suspected[node]; !ok || epochNr < e {
		l.Suspected[node] = e // update with latest epoch that node was suspected
	}
}

// Reconfigure informs the leader selection policy about a change in the membership.
func (l *BlacklistLeaderPolicy) Reconfigure(nodeIDs []t.NodeID) LeaderSelectionPolicy {
	membership := make(map[t.NodeID]struct{}, len(nodeIDs))
	suspected := make(map[t.NodeID]tt.EpochNr, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		membership[nodeID] = struct{}{}
		if epoch, ok := l.Suspected[nodeID]; ok {
			suspected[nodeID] = epoch
		}
	}

	newPolicy := BlacklistLeaderPolicy{
		membership,
		suspected,
		l.MinLeaders,
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
