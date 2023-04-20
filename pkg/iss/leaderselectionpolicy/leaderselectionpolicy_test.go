package leaderselectionpolicy

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/mir/pkg/serializing"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/types"
)

func TestSimpleLeaderPolicy(t *testing.T) {
	// Create a SimpleLeaderPolicy with 3 nodes
	nodes := []types.NodeID{"node1", "node2", "node3"}
	policy := SimpleLeaderPolicy{Membership: nodes}

	// Check that all nodes are returned as leaders
	expected := nodes
	actual := policy.Leaders()
	require.Equal(t, expected, actual)

	// Suspect a node and check that it does not affect the leader set
	policy.Suspect(1, "node1")
	expected = nodes
	actual = policy.Leaders()
	require.Equal(t, expected, actual)

	// Reconfigure the policy with a new set of nodes and check that it is returned as the leader set
	newNodes := []types.NodeID{"node4", "node5", "node6"}
	pol, ok := policy.Reconfigure(newNodes).(*SimpleLeaderPolicy)
	policy = *pol
	require.Equal(t, ok, true)
	expected = newNodes
	actual = policy.Leaders()
	require.Equal(t, expected, actual)
}

func TestBlackListLeaderPolicy(t *testing.T) {
	// Create a BlacklistLeaderPolicy with 3 nodes and a minimum of 2 leaders
	nodes := []types.NodeID{"node1", "node2", "node3"}
	policy := NewBlackListLeaderPolicy(nodes, 2)

	// Check that all nodes are returned as leaders
	expected := nodes
	actual := policy.Leaders()
	require.Equal(t, expected, actual)

	// Suspect a node and check that it is not returned as a leader
	policy.Suspect(1, "node1")
	expected = []types.NodeID{"node2", "node3"}
	actual = policy.Leaders()
	sort.Slice(expected, func(i, j int) bool {
		return string(expected[i]) < string(expected[j])
	})
	sort.Slice(actual, func(i, j int) bool {
		return string(actual[i]) < string(actual[j])
	})
	require.Equal(t, expected, actual)

	// Suspect another node and check that it is not returned as a leader
	policy.Suspect(2, "node2")
	expected = []types.NodeID{"node3", "node1"}
	actual = policy.Leaders()
	require.Equal(t, expected, actual)

	// Reconfigure the policy with a new set of nodes and check that it is returned as the leader set
	newNodes := []types.NodeID{"node4", "node5", "node6"}
	policy, ok := policy.Reconfigure(newNodes).(*BlacklistLeaderPolicy)
	require.Equal(t, ok, true)
	expected = newNodes
	actual = policy.Leaders()
	require.Equal(t, expected, actual)
}

func TestBlacklistLeaderPolicy_Conversion(t *testing.T) {
	policy := &BlacklistLeaderPolicy{
		Membership: map[types.NodeID]struct{}{
			"node1": {},
			"node2": {},
			"node3": {},
		},
		Suspected: map[types.NodeID]tt.EpochNr{
			"node1": 1,
			"node2": 2,
		},
		MinLeaders: 2,
	}
	bytes, _ := policy.Bytes()
	// check the first 8 bytes
	policyType := serializing.Uint64FromBytes(bytes[0:8])
	assert.Equal(t, uint64(Blacklist), policyType)

	reconstitutedPolicy, err := LeaderPolicyFromBytes(bytes)
	assert.NoError(t, err)
	assert.IsType(t, &BlacklistLeaderPolicy{}, reconstitutedPolicy)
	reconstitutedBlacklistPolicy := reconstitutedPolicy.(*BlacklistLeaderPolicy)
	require.Equal(t, policy.Membership, reconstitutedBlacklistPolicy.Membership)
	require.Equal(t, policy.Suspected, reconstitutedBlacklistPolicy.Suspected)
	require.Equal(t, policy.MinLeaders, reconstitutedBlacklistPolicy.MinLeaders)
	require.Equal(t, policy.Leaders(), reconstitutedBlacklistPolicy.Leaders())
}

func TestSimpleLeaderPolicy_OneNodeMembership(t *testing.T) {
	simplePolicy := NewSimpleLeaderPolicy([]types.NodeID{"node1"})
	// Verify that bytes() returns an error when membership is empty
	bytes, _ := simplePolicy.Bytes()
	// Verify that LeaderPolicyFromBytes() returns an error when passed an empty byte slice
	reconstitutedPolicy, err := LeaderPolicyFromBytes(bytes)
	reconstitutedSimpleLeaderPolicy := reconstitutedPolicy.(*SimpleLeaderPolicy)
	assert.Equal(t, err, nil)
	require.Equal(t, simplePolicy.Membership, reconstitutedSimpleLeaderPolicy.Membership)

	blacklistPolicy := NewBlackListLeaderPolicy([]types.NodeID{"node1"}, 1)
	bytes, _ = blacklistPolicy.Bytes()
	// Verify that LeaderPolicyFromBytes() returns an error when passed an empty byte slice
	reconstitutedPolicy, err = LeaderPolicyFromBytes(bytes)
	reconstitutedBlacklistPolicy := reconstitutedPolicy.(*BlacklistLeaderPolicy)
	assert.Equal(t, err, nil)
	require.Equal(t, blacklistPolicy.Membership, reconstitutedBlacklistPolicy.Membership)
	require.Equal(t, blacklistPolicy.Leaders(), reconstitutedBlacklistPolicy.Leaders())
}

func TestBlacklistLeaderPolicy_EmptySuspected(t *testing.T) {
	policy := &BlacklistLeaderPolicy{
		Membership: map[types.NodeID]struct{}{
			"node1": {},
			"node2": {},
			"node3": {},
		},
		Suspected:  map[types.NodeID]tt.EpochNr{},
		MinLeaders: 2,
	}
	bytes, _ := policy.Bytes()
	// check the first 8 bytes
	policyType := serializing.Uint64FromBytes(bytes[0:8])
	assert.Equal(t, uint64(Blacklist), policyType)

	reconstitutedPolicy, err := LeaderPolicyFromBytes(bytes)
	assert.NoError(t, err)
	assert.IsType(t, &BlacklistLeaderPolicy{}, reconstitutedPolicy)
	reconstitutedBlacklistPolicy := reconstitutedPolicy.(*BlacklistLeaderPolicy)
	assert.Equal(t, policy.Membership, reconstitutedBlacklistPolicy.Membership)
	assert.Equal(t, policy.MinLeaders, reconstitutedBlacklistPolicy.MinLeaders)
	assert.Equal(t, policy.Suspected, reconstitutedBlacklistPolicy.Suspected)
	policyLeaders := policy.Leaders()
	reconstitutedPolicyLeaders := policy.Leaders()
	assert.Equal(t, policyLeaders, reconstitutedPolicyLeaders)
}
