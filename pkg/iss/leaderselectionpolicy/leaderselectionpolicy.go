package leaderselectionpolicy

import (
	"sync"

	"github.com/fxamacker/cbor/v2"
	es "github.com/go-errors/errors"

	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/serializing"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

type LeaderPolicyType uint64

const (
	Simple LeaderPolicyType = iota
	Blacklist
)

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
	Suspect(e tt.EpochNr, node t.NodeID)

	// Reconfigure returns a new LeaderSelectionPolicy based on the state of the current one,
	// but using a new membership.
	Reconfigure(membership *trantorpbtypes.Membership) LeaderSelectionPolicy

	Bytes() ([]byte, error)
}

func LeaderPolicyFromBytes(bytes []byte) (LeaderSelectionPolicy, error) {
	leaderPolicyType := serializing.Uint64FromBytes(bytes[0:8])

	switch LeaderPolicyType(leaderPolicyType) {
	case Simple:
		return SimpleLeaderPolicyFromBytes(bytes[8:])
	case Blacklist:
		return BlacklistLeaderPolicyFromBytes(bytes[8:])
	default:
		return nil, es.Errorf("invalid LeaderSelectionPolicy type: %v", leaderPolicyType)
	}

}
