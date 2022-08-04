package bcb

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/bcb/bcbdsl"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/modules"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig struct {
	Self     t.ModuleID // id of this module
	Consumer t.ModuleID // id of the module to send the "Deliver" event to
	Net      t.ModuleID
	Crypto   t.ModuleID
}

// DefaultModuleConfig returns a valid module config with default names for all modules.
func DefaultModuleConfig(consumer t.ModuleID) *ModuleConfig {
	return &ModuleConfig{
		Self:     "bcb",
		Consumer: consumer,
		Net:      "net",
		Crypto:   "crypto",
	}
}

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams struct {
	InstanceUID []byte     // unique identifier for this instance of BCB, used to prevent cross-instance replay attacks
	AllNodes    []t.NodeID // the list of participating nodes
	Leader      t.NodeID   // the id of the leader of the instance
}

// GetN returns the total number of nodes.
func (params *ModuleParams) GetN() int {
	return len(params.AllNodes)
}

// GetF returns the maximum tolerated number of faulty nodes.
func (params *ModuleParams) GetF() int {
	return (params.GetN() - 1) / 3
}

// bcbModuleState represents the state of the bcb module.
type bcbModuleState struct {
	// this variable is not part of the original protocol description, but it greatly simplifies the code
	request []byte

	sentEcho     bool
	sentFinal    bool
	delivered    bool
	receivedEcho map[t.NodeID]bool
	echoSigs     map[t.NodeID][]byte
}

// NewModule returns a passive module for the Signed Echo Broadcast from the textbook "Introduction to reliable and
// secure distributed programming". It serves as a motivating example for the DSL module interface.
// The pseudocode can also be found in https://dcl.epfl.ch/site/_media/education/sdc_byzconsensus.pdf (Algorithm 4
// (Echo broadcast [Rei94]))
func NewModule(mc *ModuleConfig, params *ModuleParams, nodeID t.NodeID) modules.PassiveModule {
	m := dsl.NewModule(mc.Self)

	state := bcbModuleState{
		request: nil,

		sentEcho:     false,
		sentFinal:    false,
		delivered:    false,
		receivedEcho: make(map[t.NodeID]bool),
		echoSigs:     make(map[t.NodeID][]byte),
	}

	dsl.UponInit(m, func() error {
		// no initialization required
		return nil
	})

	// upon event <bcb, Broadcast | m> do    // only process s
	bcbdsl.UponRequest(m, func(data []byte) error {
		if nodeID != params.Leader {
			return fmt.Errorf("only the leader node can receive requests")
		}
		state.request = data
		dsl.SendMessage(m, mc.Net, StartMessage(mc.Self, data), params.AllNodes)
		return nil
	})

	// upon event <al, Deliver | p, [Send, m]> ...
	bcbdsl.UponStartMessageReceived(m, func(from t.NodeID, data []byte) error {
		// ... such that p = s and sentecho = false do
		if from == params.Leader && !state.sentEcho {
			// σ := sign(self, bcb||self||ECHO||m);
			sigMsg := [][]byte{params.InstanceUID, []byte("ECHO"), data}
			dsl.SignRequest(m, mc.Crypto, sigMsg, &signStartMessageContext{})
		}
		return nil
	})

	dsl.UponSignResult(m, func(signature []byte, context *signStartMessageContext) error {
		if !state.sentEcho {
			state.sentEcho = true
			dsl.SendMessage(m, mc.Net, EchoMessage(mc.Self, signature), []t.NodeID{params.Leader})
		}
		return nil
	})

	// upon event <al, Deliver | p, [ECHO, m, σ]> do    // only process s
	bcbdsl.UponEchoMessageReceived(m, func(from t.NodeID, signature []byte) error {
		// if echos[p] = ⊥ ∧ verifysig(p, bcb||p||ECHO||m, σ) then
		if nodeID == params.Leader && !state.receivedEcho[from] && state.request != nil {
			state.receivedEcho[from] = true
			sigMsg := [][]byte{params.InstanceUID, []byte("ECHO"), state.request}
			dsl.VerifyOneNodeSig(m, mc.Crypto, sigMsg, signature, from, &verifyEchoContext{signature})
		}
		return nil
	})

	dsl.UponOneNodeSigVerified(m, func(nodeID t.NodeID, err error, context *verifyEchoContext) error {
		if err == nil {
			state.echoSigs[nodeID] = context.signature
		}
		return nil
	})

	// upon exists m != ⊥ such that #({p ∈ Π | echos[p] = m}) > (N+f)/2 and sentfinal = FALSE do
	dsl.UponCondition(m, func() error {
		if len(state.echoSigs) > (params.GetN()+params.GetF())/2 && !state.sentFinal {
			state.sentFinal = true
			certSigners, certSignatures := maputil.GetKeysAndValues(state.echoSigs)
			dsl.SendMessage(m, mc.Net,
				FinalMessage(mc.Self, state.request, certSigners, certSignatures),
				params.AllNodes)
		}
		return nil
	})

	// upon event <al, Deliver | p, [FINAL, m, Σ]> do
	bcbdsl.UponFinalMessageReceived(m, func(from t.NodeID, data []byte, signers []t.NodeID, signatures [][]byte) error {
		// if #({p ∈ Π | Σ[p] != ⊥ ∧ verifysig(p, bcb||p||ECHO||m, Σ[p])}) > (N+f)/2 and delivered = FALSE do
		if len(signers) == len(signatures) && len(signers) > (params.GetN()+params.GetF())/2 && !state.delivered {
			sigMsgs := sliceutil.Repeat([][]byte{params.InstanceUID, []byte("ECHO"), data}, len(signers))
			dsl.VerifyNodeSigs(m, mc.Crypto, sigMsgs, signatures, signers, &verifyFinalContext{data})
		}
		return nil
	})

	dsl.UponNodeSigsVerified(m, func(nodeIDs []t.NodeID, errs []error, allOK bool, context *verifyFinalContext) error {
		if allOK && !state.delivered {
			state.delivered = true
			bcbdsl.Deliver(m, mc.Consumer, context.data)
		}
		return nil
	})

	return m
}

// Context data structures

type signStartMessageContext struct{}

type verifyEchoContext struct {
	signature []byte
}

type verifyFinalContext struct {
	data []byte
}
