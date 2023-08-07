package simpleacc

import (
	"github.com/filecoin-project/mir/pkg/accountability/simpleacc/common"
	"github.com/filecoin-project/mir/pkg/accountability/simpleacc/internal/certificates/fullcertificates"
	"github.com/filecoin-project/mir/pkg/accountability/simpleacc/internal/certificates/lightcertificates"
	incommon "github.com/filecoin-project/mir/pkg/accountability/simpleacc/internal/common"
	"github.com/filecoin-project/mir/pkg/accountability/simpleacc/internal/poms"
	"github.com/filecoin-project/mir/pkg/accountability/simpleacc/internal/predecisions"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/factorymodule"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	accpbtypes "github.com/filecoin-project/mir/pkg/pb/accountabilitypb/types"
	factorypbtypes "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig = common.ModuleConfig

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams = common.ModuleParams

// NewModule creates a new instance of the (optinal) accountability
// module.
// This module can receive decisions from a module that ensures agreement
// (for example, receive a decision from the ordering module, instead of
// the ordering module delivering them to the application layer directly),
// and treats them as predecisions. It performs two all-to-all broadcasts
// with signatures to ensure accountability. The first broadcast is a
// signed predecision per participant. In the second broadcast, each
// participant broadcasts a certificate containing a strong quorum of
// signed predecisions that they each delivered from the first
// broadcast. Termination occurs once a process receives a strong quorum
// of correctly signed predecisions.
// *Accountability* states that if an adversary (controlling less than a
// strong quorum, but perhaps more or as much as a weak quorum) causes
// a disagreement (two different correct processes delivering different
// decisions) then all correct processes eventually receive Proofs-of-Misbehavior (PoMs)
// for a provably malicious coalition at least the size of a weak quorum.
// In the case of this module, a PoM is a pair of different predecisions signed
// by the same node.
// The module keeps looking for PoMs with newly received messages
// (signed predecisions or certificates) after termination, until
// it is garbage collected.
//

// Intuition of correctness: a process only terminates if it receives a
// strong quorum of signed predecisions from distinct processes, forming
// a certificate.  Once this process forms a certificate, it shares it
// with the rest of participants. If all correct processes terminate,
// then that means all correct processes will (i) deliver a strong quorum
// of signed predecisions and (ii) broadcast them in a certificate. Thus,
// if two correct processes p_1, p_2 disagree (decide d_1, d_2,
// respectively) then that means they must have each delivered a strong
// quorum of signed predecisions for different predecisions. By quorum
// intersection, this means that at least a weak quorum of processes have
// signed respective predecisions for d_1, d_2 and sent each of them to
// the respective correct process p_1, p_2. Once p_1 receives the
// certificate that p_2 broadcasted, p_1 will then generate a weak quorum
// of PoMs (and vice versa) and broadcast it to the rest of processes.
//
// This module effectively implements a variant of the accountability
// module of Civit et al. at https://ieeexplore.ieee.org/document/9820722/
// Except that it does not implement the optimization using threshold
// signatures (as we have members with associated weight)
//

// The optimistic variant of this module is a parametrizable optimization
// in which certificates are optimistically believed to be correct. This way,
// in the good case a correct process broadcasts a light certificate of O(1) bits
// (instead of O(n) of a certificate)
// and only actually sends the full certificate to nodes from which it receives a light certificate
// for a predecision other than the locally Decided one. The recipient of the certificate can then
// generate and broadcast the PoMs.
//
// ATTENTION: This module is intended to be used once per instance
// (to avoid replay attacks) and reinstantiated in a factory.
func NewModule(mc ModuleConfig, params *ModuleParams, logger logging.Logger) (modules.PassiveModule, error) {
	m := dsl.NewModule(mc.Self)

	state := &incommon.State{
		SignedPredecisions: make(map[t.NodeID]*accpbtypes.SignedPredecision),
		PredecisionNodeIDs: make(map[string][]t.NodeID),
		SignedPredecision:  nil,
		DecidedCertificate: nil,
		Predecided:         false,
		UnsentPoMs:         make([]*accpbtypes.PoM, 0),
		SentPoMs:           make(map[t.NodeID]*accpbtypes.PoM),
	}

	predecisions.IncludePredecisions(m, &mc, params, state, logger)
	fullcertificates.IncludeFullCertificate(m, &mc, params, state, logger)
	if params.LightCertificates {
		lightcertificates.IncludeLightCertificate(m, &mc, params, state, logger)
	}
	poms.IncludePoMs(m, &mc, params, state, logger)

	return m, nil
}

func NewReconfigurableModule(mc ModuleConfig, paramsTemplate ModuleParams, logger logging.Logger) modules.PassiveModule {
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}
	return factorymodule.New(
		mc.Self,
		factorymodule.DefaultParams(

			// This function will be called whenever the factory module
			// is asked to create a new instance of the accountabuility module.
			func(accID t.ModuleID, params *factorypbtypes.GeneratorParams) (modules.PassiveModule, error) {

				accParams := params.Type.(*factorypbtypes.GeneratorParams_AccModule).AccModule

				// Create a copy of basic module config with an adapted ID for the submodule.
				submc := mc
				submc.Self = accID

				// Fill in instance-specific parameters.
				moduleParams := paramsTemplate
				moduleParams.Membership = accParams.Membership

				// Create a new instance of the multisig collector.
				accountabilityModule, err := NewModule(
					submc,
					&moduleParams,
					logger,
				)
				if err != nil {
					return nil, err
				}
				return accountabilityModule, nil
			},
		),
		logger,
	)
}
