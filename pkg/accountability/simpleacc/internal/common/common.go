package common

import (
	accpbtypes "github.com/filecoin-project/mir/pkg/pb/accountabilitypb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

type State struct {
	SignedPredecisions map[t.NodeID]*accpbtypes.SignedPredecision // Map of received signed predicisions (including own's) with their signer as key.
	PredecisionNodeIDs map[string][]t.NodeID                      // Map of predecisions and the nodes that have signed them with the predecision as key,
	// used for fast verification of whether a predecision is predecided by a strong quorum.
	SignedPredecision  *accpbtypes.SignedPredecision              // Own signed predecision.
	DecidedCertificate map[t.NodeID]*accpbtypes.SignedPredecision // Locally decided certificate as a map with signer as key.
	Predecided         bool                                       // Whether this process has received a predecided value from calling module.
	UnsentPoMs         []*accpbtypes.PoM                          // List of PoMs not yet sent to the application.
	SentPoMs           map[t.NodeID]*accpbtypes.PoM               // List of PoMs already sent to the application with the signer as key.
}
