package membutil

import (
	"math/big"

	"github.com/fxamacker/cbor/v2"
	es "github.com/go-errors/errors"

	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

func WeightOf(membership *trantorpbtypes.Membership, nodeIDs []t.NodeID) tt.VoteWeight {
	sum := big.NewInt(0)
	for _, node := range maputil.GetValuesOf(membership.Nodes, nodeIDs) {
		sum.Add(sum, node.Weight.BigInt())
	}
	return tt.VoteWeight(sum.String())
}

func TotalWeight(membership *trantorpbtypes.Membership) tt.VoteWeight {
	return WeightOf(membership, maputil.GetKeys(membership.Nodes))
}

func StrongQuorum(membership *trantorpbtypes.Membership) tt.VoteWeight {
	// assuming n > 3f:
	//   return min q: 2q > n+f
	n := TotalWeight(membership).BigInt()
	f := maxFaulty(n)

	result := new(big.Int).Add(n, f)
	result.Div(result, big.NewInt(2))
	result.Add(result, big.NewInt(1))
	return tt.VoteWeight(result.String())

	// If a native integer type was used, the above would be equivalent to
	// return (n+f)/2 + 1
}

func WeakQuorum(membership *trantorpbtypes.Membership) tt.VoteWeight {
	// assuming n > 3f:
	//   return min q: q > f
	n := TotalWeight(membership).BigInt()
	f := maxFaulty(n)

	result := new(big.Int).Add(f, big.NewInt(1))
	return tt.VoteWeight(result.String())

	// If a native integer type was used, the above would be equivalent to
	// return f + 1
}

func HaveStrongQuorum(membership *trantorpbtypes.Membership, nodeIDs []t.NodeID) bool {
	w := WeightOf(membership, nodeIDs).BigInt()
	strongQuorum := StrongQuorum(membership).BigInt()

	// The following replaces WeightOf(membership, nodeIDs) >= StrongQuorum(membership)
	// (if Cmp returned -1, then w would be smaller than strongQuorum)
	return w.Cmp(strongQuorum) != -1
}

func HaveWeakQuorum(membership *trantorpbtypes.Membership, nodeIDs []t.NodeID) bool {
	w := WeightOf(membership, nodeIDs).BigInt()
	weakQuorum := WeakQuorum(membership).BigInt()

	// The following replaces WeightOf(membership, nodeIDs) >= WeakQuorum(membership)
	// (if Cmp returned -1, then w would be smaller than weakQuorum)
	return w.Cmp(weakQuorum) != -1
}

func maxFaulty(n *big.Int) *big.Int {
	// assuming n > 3f:
	//   return max f

	result := new(big.Int).Sub(n, big.NewInt(1))
	result.Div(result, big.NewInt(3))
	return result

	// If a native integer type was used, the above would be equivalent to
	// return (n - 1) / 3
}

func Serialize(membership *trantorpbtypes.Membership) ([]byte, error) {
	em, err := cbor.CoreDetEncOptions().EncMode()
	if err != nil {
		return nil, err
	}
	return em.Marshal(membership)
}

func Deserialize(data []byte) (*trantorpbtypes.Membership, error) {
	var membership trantorpbtypes.Membership
	if err := cbor.Unmarshal(data, &membership); err != nil {
		return nil, es.Errorf("failed to CBOR unmarshal membership: %w", err)
	}
	return &membership, nil
}

func Valid(membership *trantorpbtypes.Membership) error {
	if len(membership.Nodes) == 0 {
		return es.Errorf("membership is empty")
	}

	for nodeID, node := range membership.Nodes {
		if node.Id != nodeID {
			return es.Errorf("node ID %v does not match key under which node is stored %v", node.Id, nodeID)
		}
		if !node.Weight.IsValid() {
			return es.Errorf("invalid weight of node %v (%v): %v", node.Id, node.Addr, node.Weight)
		}
	}

	if TotalWeight(membership).BigInt().Cmp(big.NewInt(0)) == 0 {
		return es.Errorf("zero total weight")
	}

	return nil
}
