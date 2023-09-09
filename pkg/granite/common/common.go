package common

import (
	"github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/granite"
	granitepbtypes "github.com/filecoin-project/mir/pkg/pb/granitepb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig struct {
	Self           t.ModuleID // id of this module
	App            t.ModuleID
	Ava            t.ModuleID
	Crypto         t.ModuleID
	Hasher         t.ModuleID
	Net            t.ModuleID
	Ord            t.ModuleID
	PPrepValidator t.ModuleID
	Timer          t.ModuleID
}

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams struct {

	// InstanceUID is a unique identifier for this instance used to prevent replay attacks.
	InstanceUID []byte `json:"-"`

	// Membership defines the set of nodes participating in a particular instance of Granite.
	Membership *trantorpbtypes.Membership `json:"-"`

	Crypto crypto.Crypto
}

type Step = granite.MsgType
type State struct {
	Round granite.RoundNr
	Step  Step

	ValidatedMsgs   *granite.MsgStore
	UnvalidatedMsgs *granite.MsgStore

	//Decision messages do not need to be stored per round
	DecisionMessages map[t.NodeID]map[string]*granitepbtypes.Decision
}

func (st *State) IsValid(msg *granitepbtypes.ConsensusMsg) bool {
	//TODO implement message validation logic (algorithm 5 currently in Overleaf's doc)
	return true
}

// TODO this can be optimized
func (st *State) FindNewValid() (*granitepbtypes.ConsensusMsg, t.NodeID, bool) {
	for _, store := range st.UnvalidatedMsgs.Msgs {
		for _, nodeMap := range store {
			for source, value := range nodeMap {
				if st.IsValid(value) {
					return value, source, true
				}
			}
		}
	}

	var source t.NodeID
	return nil, source, false
}
