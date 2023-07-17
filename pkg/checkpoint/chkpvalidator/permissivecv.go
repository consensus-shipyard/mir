package chkpvalidator

import (
	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/logging"

	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/crypto"
	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

type PermissiveCV struct {
	configOffset int
	ownID        t.NodeID
	hashImpl     crypto.HashImpl
	chkpVerifier checkpoint.Verifier
	logger       logging.Logger
}

// NewPermissiveCV returns a new PermissiveCV. This checkpoint validity checker
// simply believes membership information of checkpoints whose membership cannot be verified because
// the node does not know the membership of the relevant epoch yet/anymore.
func NewPermissiveCV(configOffset int, ownID t.NodeID, hashImpl crypto.HashImpl, chkpVerifier checkpoint.Verifier, logger logging.Logger) *PermissiveCV {
	return &PermissiveCV{
		configOffset: configOffset,
		ownID:        ownID,
		hashImpl:     hashImpl,
		chkpVerifier: chkpVerifier,
		logger:       logger,
	}
}

func (pcv *PermissiveCV) Verify(chkp *checkpointpbtypes.StableCheckpoint, epochNr tt.EpochNr, memberships []*trantorpbtypes.Membership) error {
	sc := checkpoint.StableCheckpointFromPb(chkp.Pb())

	// Check syntactic validity of the checkpoint.
	if err := sc.SyntacticCheck(pcv.configOffset); err != nil {
		return err
	}

	// We consider a checkpoint invalid if we are not part of its membership
	// (more precisely, membership of the epoch the checkpoint is at the start of).
	// Correct nodes should never send such checkpoints, but faulty ones could.
	if _, ok := sc.Memberships()[0].Nodes[pcv.ownID]; !ok {
		return es.Errorf("nodeID not in membership")
	}

	chkpMembership := sc.PreviousMembership() // This is wrong and it is a vulnerability, come back to fix (issue #384)
	chkpMembershipOffset := int(sc.Epoch()) - 1 - int(epochNr)

	if chkpMembershipOffset > pcv.configOffset {
		// cannot verify checkpoint signatures, too far ahead
		pcv.logger.Log(logging.LevelWarn, "-----------------------------------------------------\n",
			"ATTENTION: cannot verify membership of checkpoint, too far ahead, proceed with caution\n",
			"-----------------------------------------------------\n",
			"localEpoch", epochNr,
			"chkpEpoch", sc.Epoch(),
			"configOffset", pcv.configOffset,
		)
	} else {
		chkpMembership = memberships[chkpMembershipOffset]
	}

	return sc.Verify(pcv.configOffset, pcv.hashImpl, pcv.chkpVerifier, chkpMembership)
}
