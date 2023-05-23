package membutil

import (
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

func WeightOf(membership *trantorpbtypes.Membership, nodeIDs []t.NodeID) tt.VoteWeight {
	w := tt.VoteWeight(0)
	for _, node := range maputil.GetValuesOf(membership.Nodes, nodeIDs) {
		w += node.Weight
	}
	return w
}

func TotalWeight(membership *trantorpbtypes.Membership) tt.VoteWeight {
	return WeightOf(membership, maputil.GetKeys(membership.Nodes))
}

func StrongQuorum(membership *trantorpbtypes.Membership) tt.VoteWeight {
	// assuming n > 3f:
	//   return min q: 2q > n+f
	n := TotalWeight(membership)
	f := maxFaulty(n)
	return (n+f)/2 + 1
}

func WeakQuorum(membership *trantorpbtypes.Membership) tt.VoteWeight {
	// assuming n > 3f:
	//   return min q: q > f
	n := TotalWeight(membership)
	f := maxFaulty(n)
	return f + 1
}

func HaveStrongQuorum(membership *trantorpbtypes.Membership, nodeIDs []t.NodeID) bool {
	return WeightOf(membership, nodeIDs) >= StrongQuorum(membership)
}

func HaveWeakQuorum(membership *trantorpbtypes.Membership, nodeIDs []t.NodeID) bool {
	return WeightOf(membership, nodeIDs) >= WeakQuorum(membership)
}

func maxFaulty(n tt.VoteWeight) tt.VoteWeight {
	// assuming n > 3f:
	//   return max f
	return (n - 1) / 3
}
