package common

import (
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ModuleConfig sets the module ids. All replicas are expected to use identical module configurations.
type ModuleConfig struct {
	Self t.ModuleID // id of this module, used to uniquely identify an instance of the accountability module.
	// It prevents cross-instance signature replay attack and should be unique across all executions.

	Ordering t.ModuleID // provides Predecisions
	App      t.ModuleID // receives Decisions and/or PoMs
	Crypto   t.ModuleID // provides cryptographic primitives
	Net      t.ModuleID // provides network primitives
}

// ModuleParams sets the values for the parameters of an instance of the protocol.
// All replicas are expected to use identical module parameters.
type ModuleParams struct {
	Membership        *trantorpbtypes.Membership // the list of participating nodes
	LightCertificates bool
}
