package leaderselectionpolicy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/serializing"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

func TestSimpleLeaderPolicy(t *testing.T) {
	// Create a SimpleLeaderPolicy with 3 nodes
	nodes := testMembership([]types.NodeID{"node1", "node2", "node3"})
	policy := NewSimpleLeaderPolicy(nodes)

	// Check that all nodes are returned as leaders
	expected := maputil.GetSortedKeys(nodes.Nodes)
	actual := policy.Leaders()
	require.Equal(t, expected, actual)

	// Suspect a node and check that it does not affect the leader set
	policy.Suspect(1, "node1")
	expected = maputil.GetSortedKeys(nodes.Nodes)
	actual = policy.Leaders()
	require.Equal(t, expected, actual)

	// Reconfigure the policy with a new set of nodes and check that it is returned as the leader set
	newNodes := testMembership([]types.NodeID{"node4", "node5", "node6"})
	pol, ok := policy.Reconfigure(newNodes).(*SimpleLeaderPolicy)
	policy = pol
	require.Equal(t, ok, true)
	expected = maputil.GetSortedKeys(newNodes.Nodes)
	actual = policy.Leaders()
	require.Equal(t, expected, actual)
}

func TestBlackListLeaderPolicy(t *testing.T) {
	// Create a BlacklistLeaderPolicy with 3 nodes and a minimum of 2 leaders
	nodes := testMembership([]types.NodeID{"node1", "node2", "node3", "node4"})
	policy := NewBlackListLeaderPolicy(nodes)

	// Check that all nodes are returned as leaders
	expected := maputil.GetSortedKeys(nodes.Nodes)
	actual := policy.Leaders()
	require.Equal(t, expected, actual)

	// Suspect a node and check that it is not returned as a leader
	policy.Suspect(1, "node1")
	expected = []types.NodeID{"node2", "node3", "node4"}
	actual = policy.Leaders()
	require.Equal(t, expected, actual)

	// Suspect another node and check that the previously suspected node is again returned as a leader.
	policy.Suspect(2, "node4")
	expected = []types.NodeID{"node2", "node3", "node1"}
	actual = policy.Leaders()
	require.Equal(t, expected, actual)

	// Reconfigure the policy with a new set of nodes and check that it is returned as the leader set.
	policy, ok := policy.Reconfigure(testMembership([]types.NodeID{"node1", "node5"})).(*BlacklistLeaderPolicy)
	require.Equal(t, ok, true)
	expected = []types.NodeID{"node5", "node1"}
	actual = policy.Leaders()
	require.Equal(t, expected, actual)

	// Add one more node and check that the suspected node is removed again from the leader set.
	policy, ok = policy.Reconfigure(testMembership([]types.NodeID{"node1", "node5", "node6"})).(*BlacklistLeaderPolicy)
	require.Equal(t, ok, true)
	expected = []types.NodeID{"node5", "node6"}
	actual = policy.Leaders()
	require.Equal(t, expected, actual)
}

func TestBlacklistLeaderPolicy_Conversion(t *testing.T) {
	policy := NewBlackListLeaderPolicy(testMembership([]types.NodeID{"node1", "node2", "node3"}))
	policy.Suspected = map[types.NodeID]tt.EpochNr{
		"node1": 1,
		"node2": 2,
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
	require.Equal(t, policy.Leaders(), reconstitutedBlacklistPolicy.Leaders())
}

func TestLeaderPolicy_OneNodeMembership(t *testing.T) {
	simplePolicy := NewSimpleLeaderPolicy(testMembership([]types.NodeID{"node1"}))
	// Verify that bytes() returns an error when membership is empty
	bytes, _ := simplePolicy.Bytes()
	// Verify that LeaderPolicyFromBytes() returns an error when passed an empty byte slice
	reconstitutedPolicy, err := LeaderPolicyFromBytes(bytes)
	reconstitutedSimpleLeaderPolicy := reconstitutedPolicy.(*SimpleLeaderPolicy)
	assert.Equal(t, err, nil)
	require.Equal(t, simplePolicy.Membership, reconstitutedSimpleLeaderPolicy.Membership)

	blacklistPolicy := NewBlackListLeaderPolicy(testMembership([]types.NodeID{"node1"}))
	bytes, _ = blacklistPolicy.Bytes()
	// Verify that LeaderPolicyFromBytes() returns an error when passed an empty byte slice
	reconstitutedPolicy, err = LeaderPolicyFromBytes(bytes)
	reconstitutedBlacklistPolicy := reconstitutedPolicy.(*BlacklistLeaderPolicy)
	assert.Equal(t, err, nil)
	require.Equal(t, blacklistPolicy.Membership, reconstitutedBlacklistPolicy.Membership)
	require.Equal(t, blacklistPolicy.Leaders(), reconstitutedBlacklistPolicy.Leaders())
}

func TestBlacklistLeaderPolicy_EmptySuspected(t *testing.T) {
	policy := NewBlackListLeaderPolicy(testMembership([]types.NodeID{"node1", "node2", "node3"}))
	bytes, _ := policy.Bytes()
	// check the first 8 bytes
	policyType := serializing.Uint64FromBytes(bytes[0:8])
	assert.Equal(t, uint64(Blacklist), policyType)

	reconstitutedPolicy, err := LeaderPolicyFromBytes(bytes)
	assert.NoError(t, err)
	assert.IsType(t, &BlacklistLeaderPolicy{}, reconstitutedPolicy)
	reconstitutedBlacklistPolicy := reconstitutedPolicy.(*BlacklistLeaderPolicy)
	assert.Equal(t, policy.Membership, reconstitutedBlacklistPolicy.Membership)
	assert.Equal(t, policy.Suspected, reconstitutedBlacklistPolicy.Suspected)
	policyLeaders := policy.Leaders()
	reconstitutedPolicyLeaders := policy.Leaders()
	assert.Equal(t, policyLeaders, reconstitutedPolicyLeaders)
}

func testMembership(nodeIDs []types.NodeID) *trantorpbtypes.Membership {
	m := trantorpbtypes.Membership{Nodes: map[types.NodeID]*trantorpbtypes.NodeIdentity{}}
	for _, nodeID := range nodeIDs {
		m.Nodes[nodeID] = &trantorpbtypes.NodeIdentity{
			Id:     nodeID,
			Addr:   "",
			Key:    nil,
			Weight: "1",
		}
	}
	return &m
}
