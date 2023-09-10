package common

import (
	"github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/granite"
	"github.com/filecoin-project/mir/pkg/granite/internal/parts/consensustask"
	granitepbtypes "github.com/filecoin-project/mir/pkg/pb/granitepb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/membutil"
	"reflect"
	"time"
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

	ConvergeDelay time.Duration
}

type Step = granite.MsgType
type State struct {
	Round granite.RoundNr
	Step  Step

	ValidatedMsgs   *granite.MsgStore
	UnvalidatedMsgs *granite.MsgStore

	//Decision messages do not need to be stored per round
	DecisionMessages map[t.NodeID]map[string]*granitepbtypes.Decision

	Proposal []byte
	Value    []byte

	Finished bool
}

func (st *State) IsValid(msg *granitepbtypes.ConsensusMsg, params *ModuleParams) bool {
	switch msg.MsgType {
	case granite.CONVERGE:
		//TODO verify ticket and return false if not verified

		if msg.Round == 1 {
			return true
		} else if value, ok := consensustask.GetValueWithStrongQuorum(params.Membership, st.ValidatedMsgs.Msgs[granite.COMMIT][granite.RoundNr(uint64(msg.Round)-1)]); ok && value == nil {
			return true
		} else {
			return maputil.FindAny(st.ValidatedMsgs.Msgs[granite.COMMIT][granite.RoundNr(uint64(msg.Round)-1)], func(_msg *granitepbtypes.ConsensusMsg) bool {
				return reflect.DeepEqual(_msg.Data, msg.Data)
			})
		}

	case granite.PROPOSE:
		if msg.Round == 1 {
			return true
		} else {
			return maputil.FindAny(st.ValidatedMsgs.Msgs[granite.CONVERGE][msg.Round], func(_msg *granitepbtypes.ConsensusMsg) bool {
				return reflect.DeepEqual(_msg.Data, msg.Data)
			})
		}

	case granite.PREPARE:
		keysWithVal := maputil.FindKeysWithFunc(st.ValidatedMsgs.Msgs[granite.PROPOSE][msg.Round], func(_msg *granitepbtypes.ConsensusMsg) bool {
			return reflect.DeepEqual(_msg.Data, msg.Data)
		})
		return membutil.HaveWeakQuorum(params.Membership, keysWithVal)

	case granite.COMMIT:
		if msg.Data != nil {
			keysWithVal := maputil.FindKeysWithFunc(st.ValidatedMsgs.Msgs[granite.PREPARE][msg.Round], func(_msg *granitepbtypes.ConsensusMsg) bool {
				return reflect.DeepEqual(_msg.Data, msg.Data)
			})
			return membutil.HaveStrongQuorum(params.Membership, keysWithVal)
		} else {
			return maputil.HasTwoDistinctValues(st.ValidatedMsgs.Msgs[granite.PREPARE][msg.Round], func(msg, _msg *granitepbtypes.ConsensusMsg) bool {
				return reflect.DeepEqual(_msg.Data, msg.Data)
			})
		}
	}

	return false
}

// TODO this (and also other functions in message validation etc.) can be optimized
func (st *State) FindNewValid(params *ModuleParams) (*granitepbtypes.ConsensusMsg, t.NodeID, bool) {
	for _, store := range st.UnvalidatedMsgs.Msgs {
		for _, nodeMap := range store {
			for source, value := range nodeMap {
				if st.IsValid(value, params) {
					return value, source, true
				}
			}
		}
	}

	var source t.NodeID
	return nil, source, false
}
