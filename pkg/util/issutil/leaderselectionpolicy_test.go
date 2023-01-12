package issutil

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

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
