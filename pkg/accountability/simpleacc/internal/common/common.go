package common

import (
	accpbtypes "github.com/filecoin-project/mir/pkg/pb/accountabilitypb/types"
	isspbtypes "github.com/filecoin-project/mir/pkg/pb/isspb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

type State struct {

	// Map of received signed predicisions (including own) with their signer as key.
	SignedPredecisions map[t.NodeID]*accpbtypes.SignedPredecision

	// Map of predecisions and the nodes that have signed them with the predecision as key,
	PredecisionNodeIDs map[string][]t.NodeID

	// --------------------------------------------------------------------------------
	// Used for fast verification of whether a predecision is predecided by a strong quorum.

	// Decision locally decided
	LocalPredecision *LocalPredecision

	// Locally decided certificate (predecision and list of signatures with signers as key)
	DecidedCertificate *accpbtypes.FullCertificate

	// Whether this process has received a predecided value from calling module.
	Predecided bool

	// List of PoMs not yet sent to the application.
	UnhandledPoMs []*accpbtypes.PoM

	// List of PoMs already sent to the application with the signer as key.
	HandledPoMs map[t.NodeID]*accpbtypes.PoM

	// Map of light certificates with the signer as key, buffered if no local decision made yet.
	LightCertificates map[t.NodeID][]byte
}

type LocalPredecision struct {
	SBDeliver         *isspbtypes.SBDeliver         // Actual payload of the local predecision.
	SignedPredecision *accpbtypes.SignedPredecision // Own signed predecision.
}
