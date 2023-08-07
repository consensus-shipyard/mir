package common

import (
	incommon "github.com/filecoin-project/mir/pkg/accountability/simpleacc/internal/common"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	accpbtypes "github.com/filecoin-project/mir/pkg/pb/accountabilitypb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	timertypes "github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig struct {
	Self t.ModuleID // id of this module, used to uniquely identify an instance of the accountability module.
	// It prevents cross-instance signature replay attack and should be unique across all executions.

	Ordering t.ModuleID // provides Predecisions
	App      t.ModuleID // receives Decisions and/or PoMs
	Crypto   t.ModuleID // provides cryptographic primitives
	Timer    t.ModuleID // provides Timing primitives
	Net      t.ModuleID // provides network primitives
}

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams struct {
	Membership        *trantorpbtypes.Membership // The list of participating nodes.
	LightCertificates bool
	ResendFrequency   timertypes.Duration // Frequency with which messages in the critical path are re-sent
	RetentionIndex    tt.RetentionIndex
	PoMsHandler       func(m dsl.Module, // Function to be called when PoMs detected.
		mc *ModuleConfig,
		params *ModuleParams,
		state *incommon.State,
		poms []*accpbtypes.PoM,
		logger logging.Logger)
}
