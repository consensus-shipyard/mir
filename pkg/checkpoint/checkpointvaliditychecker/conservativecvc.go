package checkpointvaliditychecker

import (
	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/crypto"
	issconfig "github.com/filecoin-project/mir/pkg/iss/config"
	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

type ConservativeCVC struct {
	issParams    *issconfig.ModuleParams
	ownID        t.NodeID
	hashImpl     crypto.HashImpl
	chkpVerifier checkpoint.Verifier
}

// NewConservativeCVC returns a new ConservativeCVC. This checkpoint validity checker simply rejects checkpoints whose signatures cannot be verified because
// the node does not know the membership of the relevant epoch yet.
func NewConservativeCVC(issParams *issconfig.ModuleParams, ownID t.NodeID, hashImpl crypto.HashImpl, chkpVerifier checkpoint.Verifier) *ConservativeCVC {
	return &ConservativeCVC{
		issParams:    issParams,
		ownID:        ownID,
		hashImpl:     hashImpl,
		chkpVerifier: chkpVerifier,
	}
}

func (cc *ConservativeCVC) Verify(chkp *checkpointpbtypes.StableCheckpoint, epochNr tt.EpochNr, memberships []*trantorpbtypes.Membership) error {
	sc := checkpoint.StableCheckpointFromPb(chkp.Pb())

	// Check syntactic validity of the checkpoint.
	if err := sc.SyntacticCheck(cc.issParams); err != nil {
		return err
	}

	// Ignore checkpoint if we are not part of its membership
	// (more precisely, membership of the epoch the checkpoint is at the start of).
	// Correct nodes should never send such checkpoints, but faulty ones could.
	if _, ok := sc.Memberships()[0].Nodes[cc.ownID]; !ok {
		return es.Errorf("nodeID not in membership")
	}

	// Check how far the received stable checkpoint is ahead of the local node's state.
	chkpMembershipOffset := int(sc.Epoch()) - 1 - int(epochNr)
	if chkpMembershipOffset <= 0 {
		// Ignore stable checkpoints that are not far enough
		// ahead of the current state of the local node.
		return es.Errorf("checkpoint not far ahead enough")
	}

	if chkpMembershipOffset > cc.issParams.ConfigOffset {
		// cannot verify checkpoint signatures, too far ahead
		return es.Errorf("checkpoint too far ahead")
	}

	chkpMembership := memberships[chkpMembershipOffset]

	return sc.Verify(cc.issParams, cc.hashImpl, cc.chkpVerifier, chkpMembership)
}
