/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package issutil

import (
	t "github.com/filecoin-project/mir/pkg/types"
	"sort"
)

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

	// TODO: Define, implement, and use state serialization and restoration for leader selection policies.
}

// The SimpleLeaderPolicy is a trivial leader selection policy.
// It must be initialized with a set of node IDs and always returns that full set as leaders,
// regardless of which nodes have been suspected. In other words, each node is leader each epoch with this policy.
type SimpleLeaderPolicy struct {
	Membership []t.NodeID
}

// Leaders always returns the whole membership for the SimpleLeaderPolicy. All nodes are always leaders.
func (simple *SimpleLeaderPolicy) Leaders() []t.NodeID {
	// All nodes are always leaders.
	return simple.Membership
}

// Suspect does nothing for the SimpleLeaderPolicy.
func (simple *SimpleLeaderPolicy) Suspect(e t.EpochNr, node t.NodeID) {
	// Do nothing.
}

// Reconfigure informs the leader selection policy about a change in the membership.
func (simple *SimpleLeaderPolicy) Reconfigure(nodeIDs []t.NodeID) LeaderSelectionPolicy {
	newPolicy := SimpleLeaderPolicy{Membership: make([]t.NodeID, len(nodeIDs))}
	copy(newPolicy.Membership, nodeIDs)
	return &newPolicy
}

type keyVal struct {
}
type BlackListLeaderPolicy struct {
	Membership     map[t.NodeID]bool // this bool is not to check membership but if it was ever suspected
	Suspected      map[t.NodeID]t.EpochNr
	minLeaders     int
	currentLeaders []t.NodeID
	updated        bool
}

func NewBlackListLeaderPolicy(members []t.NodeID, minLeaders int) *BlackListLeaderPolicy {
	membership := make(map[t.NodeID]bool, len(members))
	for _, node := range members {
		membership[node] = false // no one is suspected at first
	}
	return &BlackListLeaderPolicy{
		membership,
		make(map[t.NodeID]t.EpochNr, len(members)),
		minLeaders,
		members,
		true,
	}
}

// Leaders always returns the whole membership for the SimpleLeaderPolicy. All nodes are always leaders.
func (l *BlackListLeaderPolicy) Leaders() []t.NodeID {
	// All nodes are always leaders.
	if l.updated { //no need to recalculate, no changes since last time
		return l.currentLeaders
	}
	// defer marking l.currentLeaders as updated
	defer func() {
		l.updated = true
	}()

	curLeaders := make([]t.NodeID, len(l.Membership), len(l.Membership))
	for nodeID, _ := range l.Membership {
		curLeaders = append(curLeaders, nodeID)
	}

	// Sort the values in ascending order of latest epoch where they each were suspected
	sort.Slice(curLeaders, func(i, j int) bool {
		_, oki := l.Suspected[curLeaders[i]]
		_, okj := l.Suspected[curLeaders[j]]
		// If both nodeIDs have never been suspected, then order does not matter (both will be leaders)
		if !oki && !okj {
			return i < j
		}

		// If one node is suspected and the other is not, the suspected one comes last
		if !oki && okj {
			return false
		}
		if oki && !okj {
			return true
		}

		// If both keys are in the suspected map, sort by value in the suspected map
		return l.Suspected[curLeaders[i]] < l.Suspected[curLeaders[j]]
	})

	// Calculate number of leaders
	leadersSize := len(l.Membership) - len(l.Suspected)
	if leadersSize < l.minLeaders {
		leadersSize = l.minLeaders // must at least return l.minLeaders
	}

	// assign the leadersSize least recently suspected nodes as currentLeaders
	l.currentLeaders = curLeaders[:leadersSize]

	return l.currentLeaders
}

// Suspect adds a new suspect to the list of suspects, or updates its epoch where it was suspected to the given epoch
// if this one is more recent than the one it already has
func (l *BlackListLeaderPolicy) Suspect(e t.EpochNr, node t.NodeID) {
	if _, ok := l.Membership[node]; !ok { //node is a not a member
		//TODO error but cannot be returned
	}

	if epochNr, ok := l.Suspected[node]; !ok || epochNr < e {
		l.Membership[node] = true //mark suspected
		l.updated = false         // mark that next call to Leaders() needs to recalculate the leaders
		l.Suspected[node] = e     // update with latest epoch that node was suspected
	}
}

// Reconfigure informs the leader selection policy about a change in the membership.
func (l *BlackListLeaderPolicy) Reconfigure(nodeIDs []t.NodeID) LeaderSelectionPolicy {
	membership := make(map[t.NodeID]bool, len(nodeIDs))
	suspected := make(map[t.NodeID]t.EpochNr, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		if sus, ok := l.Membership[nodeID]; ok {
			membership[nodeID] = sus // keep former suspect as suspected
			if epoch, ok := l.Suspected[nodeID]; ok {
				suspected[nodeID] = epoch
			}
		} else {
			membership[nodeID] = false // new nodes are not suspected initially
		}
	}

	newPolicy := BlackListLeaderPolicy{
		membership,
		suspected,
		l.minLeaders,
		make([]t.NodeID, len(nodeIDs), len(nodeIDs)),
		false,
	}

	return &newPolicy
}
