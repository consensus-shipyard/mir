package common

import (
	accpbtypes "github.com/filecoin-project/mir/pkg/pb/accountabilitypb/types"
	isspbtypes "github.com/filecoin-project/mir/pkg/pb/isspb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

type State struct {
	SignedPredecisions map[t.NodeID]*accpbtypes.SignedPredecision // Map of received signed predicisions (including own's) with their signer as key.
	PredecisionNodeIDs map[string][]t.NodeID                      // Map of predecisions and the nodes that have signed them with the predecision as key,
	// used for fast verification of whether a predecision is predecided by a strong quorum.
	LocalPredecision   *LocalPredecision            // Decision locally decided
	DecidedCertificate *Certificate                 // Locally decided certificate (predecision and list of signatures with signers as key)
	Predecided         bool                         // Whether this process has received a predecided value from calling module.
	UnsentPoMs         []*accpbtypes.PoM            // List of PoMs not yet sent to the application.
	SentPoMs           map[t.NodeID]*accpbtypes.PoM // List of PoMs already sent to the application with the signer as key.
}

type LocalPredecision struct {
	SBDeliver         *isspbtypes.SBDeliver         // Actual payload of the local predecision.
	SignedPredecision *accpbtypes.SignedPredecision // Own signed predecision.
}

type Certificate struct {
	Decision   []byte
	Signatures map[t.NodeID][]byte
}
