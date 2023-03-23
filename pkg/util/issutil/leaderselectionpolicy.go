/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package issutil

import (
	"fmt"
	"sort"
	"sync"

	"github.com/fxamacker/cbor/v2"

	t "github.com/filecoin-project/mir/pkg/types"
)

type LeaderPolicyType uint64

const (
	Simple LeaderPolicyType = iota
	Blacklist
)

// Initiate cbor.EncMode to core deterministic once only
var encMode cbor.EncMode
var once sync.Once

func getEncMode() cbor.EncMode {
	once.Do(func() {
		encMode, _ = cbor.CoreDetEncOptions().EncMode()
	})
	return encMode
}

// A LeaderSelectionPolicy implements the algorithm for selecting a set of leaders in each ISS epoch.
// In a nutshell, it gathers information about suspected leaders in the past epochs
// and uses it to calculate the set of leaders for future epochs.
// Its state can be updated using Suspect() and the leader set for an epoch is queried using Leaders().
// A leader set policy must be deterministic, i.e., calling Leaders() after the same sequence of Suspect() invocations
// always returns the same set of leaders at every Node.
type LeaderSelectionPolicy interface {

	// Leaders returns the (ordered) list of leaders based on the given epoch e and on the state of this policy object.
	Leaders() []t.NodeID

	// Suspect updates the state of the policy object by announcing it that node `node` has been suspected in epoch `e`.
	Suspect(e t.EpochNr, node t.NodeID)

	// Reconfigure returns a new LeaderSelectionPolicy based on the state of the current one,
	// but using a new configuration.
	// TODO: Use the whole configuration, not just the node IDs.
	Reconfigure(nodeIDs []t.NodeID) LeaderSelectionPolicy

	//// TODO: Define, implement, and use state serialization and restoration for leader selection policies.
	//Pb() commonpb.LeaderSelectionPolicy
	Bytes() ([]byte, error)
}

func LeaderPolicyFromBytes(bytes []byte) (LeaderSelectionPolicy, error) {
	leaderPolicyType := t.Uint64FromBytes(bytes[0:8])

	switch LeaderPolicyType(leaderPolicyType) {
	case Simple:
		return SimpleLeaderPolicyFromBytes(bytes[8:])
	case Blacklist:
		return BlacklistLeaderPolicyFromBytes(bytes[8:])
	default:
		return nil, fmt.Errorf("invalid LeaderSelectionPolicy type: %v", leaderPolicyType)
	}

}

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

type BlacklistLeaderPolicy struct {
	Membership map[t.NodeID]struct{}
	Suspected  map[t.NodeID]t.EpochNr
	MinLeaders int
}

func NewBlackListLeaderPolicy(members []t.NodeID, MinLeaders int) *BlacklistLeaderPolicy {
	membership := make(map[t.NodeID]struct{}, len(members))
	for _, node := range members {
		membership[node] = struct{}{}
	}
	return &BlacklistLeaderPolicy{
		membership,
		make(map[t.NodeID]t.EpochNr, len(members)),
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
func (l *BlacklistLeaderPolicy) Suspect(e t.EpochNr, node t.NodeID) {
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
	suspected := make(map[t.NodeID]t.EpochNr, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		if sus, ok := l.Membership[nodeID]; ok {
			membership[nodeID] = sus // keep former suspect as suspected
			if epoch, ok := l.Suspected[nodeID]; ok {
				suspected[nodeID] = epoch
			}
		} else {
			membership[nodeID] = struct{}{}
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
	out := t.Uint64ToBytes(uint64(Blacklist))
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
