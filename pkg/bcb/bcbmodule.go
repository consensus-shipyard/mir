package bcb

import (
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/modules"
	bcbpbdsl "github.com/filecoin-project/mir/pkg/pb/bcbpb/dsl"
	bcbpbmsgs "github.com/filecoin-project/mir/pkg/pb/bcbpb/msgs"
	cryptopbdsl "github.com/filecoin-project/mir/pkg/pb/cryptopb/dsl"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

//TODO Sanitize messages received by this module (e.g. check that the sender is the expected one, make sure no crashing if data=nil, etc.)

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig struct {
	Self     stdtypes.ModuleID // id of this module
	Consumer stdtypes.ModuleID // id of the module to send the "Deliver" event to
	Net      stdtypes.ModuleID
	Crypto   stdtypes.ModuleID
}

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams struct {
	InstanceUID []byte            // unique identifier for this instance of BCB, used to prevent cross-instance replay attacks
	AllNodes    []stdtypes.NodeID // the list of participating nodes
	Leader      stdtypes.NodeID   // the id of the leader of the instance
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
	receivedEcho map[stdtypes.NodeID]bool
	echoSigs     map[stdtypes.NodeID][]byte
}

// NewModule returns a passive module for the Signed Echo Broadcast from the textbook "Introduction to reliable and
// secure distributed programming". It serves as a motivating example for the DSL module interface.
// The pseudocode can also be found in https://dcl.epfl.ch/site/_media/education/sdc_byzconsensus.pdf (Algorithm 4
// (Echo broadcast [Rei94]))
func NewModule(mc ModuleConfig, params *ModuleParams, nodeID stdtypes.NodeID) modules.PassiveModule {
	m := dsl.NewModule(mc.Self)

	state := bcbModuleState{
		request: nil,

		sentEcho:     false,
		sentFinal:    false,
		delivered:    false,
		receivedEcho: make(map[stdtypes.NodeID]bool),
		echoSigs:     make(map[stdtypes.NodeID][]byte),
	}

	// upon event <bcb, Broadcast | m> do    // only process s
	bcbpbdsl.UponBroadcastRequest(m, func(data []byte) error {
		if nodeID != params.Leader {
			return es.Errorf("only the leader node can receive requests")
		}
		state.request = data
		transportpbdsl.SendMessage(m, mc.Net, bcbpbmsgs.StartMessage(mc.Self, data), params.AllNodes)
		return nil
	})

	// upon event <al, Deliver | p, [Send, m]> ...
	bcbpbdsl.UponStartMessageReceived(m, func(from stdtypes.NodeID, data []byte) error {
		// ... such that p = s and sentecho = false do
		if from == params.Leader && !state.sentEcho {
			// σ := sign(self, bcb||self||ECHO||m);
			sigMsg := &cryptopbtypes.SignedData{Data: [][]byte{params.InstanceUID, []byte("ECHO"), data}}
			cryptopbdsl.SignRequest(m, mc.Crypto, sigMsg, &signStartMessageContext{})
		}
		return nil
	})

	cryptopbdsl.UponSignResult(m, func(signature []byte, context *signStartMessageContext) error {
		if !state.sentEcho {
			state.sentEcho = true
			transportpbdsl.SendMessage(m, mc.Net, bcbpbmsgs.EchoMessage(mc.Self, signature), []stdtypes.NodeID{params.Leader})
		}
		return nil
	})

	// upon event <al, Deliver | p, [ECHO, m, σ]> do    // only process s
	bcbpbdsl.UponEchoMessageReceived(m, func(from stdtypes.NodeID, signature []byte) error {
		// if echos[p] = ⊥ ∧ verifysig(p, bcb||p||ECHO||m, σ) then
		if nodeID == params.Leader && !state.receivedEcho[from] && state.request != nil {
			state.receivedEcho[from] = true
			sigMsg := &cryptopbtypes.SignedData{Data: [][]byte{params.InstanceUID, []byte("ECHO"), state.request}}
			cryptopbdsl.VerifySig(m, mc.Crypto, sigMsg, signature, from, &verifyEchoContext{signature})
		}
		return nil
	})

	cryptopbdsl.UponSigVerified(m, func(nodeID stdtypes.NodeID, err error, context *verifyEchoContext) error {
		if err == nil {
			state.echoSigs[nodeID] = context.signature
		}
		return nil
	})

	// upon exists m != ⊥ such that #({p ∈ Π | echos[p] = m}) > (N+f)/2 and sentfinal = FALSE do
	dsl.UponStateUpdates(m, func() error {
		if len(state.echoSigs) > (params.GetN()+params.GetF())/2 && !state.sentFinal {
			state.sentFinal = true
			certSigners, certSignatures := maputil.GetKeysAndValues(state.echoSigs)
			transportpbdsl.SendMessage(m, mc.Net,
				bcbpbmsgs.FinalMessage(mc.Self, state.request, certSigners, certSignatures),
				params.AllNodes)
		}
		return nil
	})

	// upon event <al, Deliver | p, [FINAL, m, Σ]> do
	bcbpbdsl.UponFinalMessageReceived(m, func(from stdtypes.NodeID, data []byte, signers []stdtypes.NodeID, signatures [][]byte) error {
		// if #({p ∈ Π | Σ[p] != ⊥ ∧ verifysig(p, bcb||p||ECHO||m, Σ[p])}) > (N+f)/2 and delivered = FALSE do
		if len(signers) == len(signatures) && len(signers) > (params.GetN()+params.GetF())/2 && !state.delivered {
			signedMessage := [][]byte{params.InstanceUID, []byte("ECHO"), data}
			sigMsgs := sliceutil.Repeat(&cryptopbtypes.SignedData{Data: signedMessage}, len(signers))
			cryptopbdsl.VerifySigs(m, mc.Crypto, sigMsgs, signatures, signers, &verifyFinalContext{data})
		}
		return nil
	})

	cryptopbdsl.UponSigsVerified(m, func(_ []stdtypes.NodeID, _ []error, allOK bool, context *verifyFinalContext) error {
		if allOK && !state.delivered {
			state.delivered = true
			bcbpbdsl.Deliver(m, mc.Consumer, context.data)
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
