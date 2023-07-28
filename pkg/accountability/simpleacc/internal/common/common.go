package common

import (
	accpbtypes "github.com/filecoin-project/mir/pkg/pb/accountabilitypb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

type State struct {
	SignedPredecisions map[t.NodeID]*accpbtypes.SignedPredecision
	PredecisionCount   map[string][]t.NodeID
	SignedPredecision  *accpbtypes.SignedPredecision
	DecidedCertificate map[t.NodeID]*accpbtypes.SignedPredecision
	Predecided         bool
	UnsentPoMs         []*accpbtypes.PoM
	SentPoMs           map[t.NodeID]*accpbtypes.PoM
}
