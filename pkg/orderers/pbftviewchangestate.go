package orderers

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/iss/config"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/orderers/types"
	hasherpbtypes "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	pbftpbmsgs "github.com/filecoin-project/mir/pkg/pb/pbftpb/msgs"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// pbftViewChangeState tracks the state of the view change sub-protocol in PBFT.
// It is only associated with a single view transition, e.g., the transition from view 1 to view 2.
// The transition from view 2 to view 3 will be tracked by a different instance of pbftViewChangeState.
type pbftViewChangeState struct {
	numNodes          int
	signedViewChanges map[t.NodeID]*pbftpbtypes.SignedViewChange

	// Digests of preprepares that need to be repropsed in a new view.
	// At initialization, an explicit nil entry is created for each sequence number of the segment.
	// Based on received ViewChange messages, each value will eventually be set to
	// - either a non-zero-length byte slice representing a digest to be re-proposed
	// - or a zero-length byte slice marking it safe to re-propose a special fresh value
	// (in the current implementation, the fresh value is a pre-configured proposal
	// or, if none has been specified, an empty byte slice).
	reproposals map[tt.SeqNr][]byte

	// Preprepare messages to be reproposed in the new view.
	// At initialization, an explicit nil entry is created for each sequence number of the segment.
	// Based on received ViewChange messages, each value will eventually be set to
	// a new Preprepare message to be re-proposed (with a correctly set view number).
	// This also holds for sequence numbers for which nothing was prepared in the previous view,
	// in which case the value is set to a Preprepare with an empty certificate and the "aborted" flag set.
	preprepares map[tt.SeqNr]*pbftpbtypes.Preprepare

	prepreparedIDs map[tt.SeqNr][]t.NodeID

	logger logging.Logger
}

func newPbftViewChangeState(seqNrs []tt.SeqNr, membership []t.NodeID, logger logging.Logger) *pbftViewChangeState {
	reproposals := make(map[tt.SeqNr][]byte)
	preprepares := make(map[tt.SeqNr]*pbftpbtypes.Preprepare)
	for _, sn := range seqNrs {
		reproposals[sn] = nil
		preprepares[sn] = nil
	}

	return &pbftViewChangeState{
		numNodes:          len(membership),
		signedViewChanges: make(map[t.NodeID]*pbftpbtypes.SignedViewChange),
		reproposals:       reproposals,
		prepreparedIDs:    make(map[tt.SeqNr][]t.NodeID),
		preprepares:       preprepares,
		logger:            logger,
	}
}

func (vcState *pbftViewChangeState) EnoughViewChanges() bool {
	for _, r := range vcState.reproposals {
		if r == nil {
			return false
		}
	}
	return true
}

func (vcState *pbftViewChangeState) AddSignedViewChange(svc *pbftpbtypes.SignedViewChange, from t.NodeID) {

	if vcState.EnoughViewChanges() {
		return
	}

	if _, ok := vcState.signedViewChanges[from]; ok {
		return
	}

	vcState.signedViewChanges[from] = svc

	if len(vcState.signedViewChanges) >= config.StrongQuorum(vcState.numNodes) {
		vcState.updateReproposals()
	}
}

// TODO: Make sure this procedure is fully deterministic (map iterations etc...)
func (vcState *pbftViewChangeState) updateReproposals() {
	pSets, qSets := reconstructPSetQSet(vcState.signedViewChanges, vcState.logger)

	for sn, r := range vcState.reproposals {
		if r == nil {
			vcState.reproposals[sn], vcState.prepreparedIDs[sn] = reproposal(pSets, qSets, sn, vcState.numNodes)
		}
	}
}

func (vcState *pbftViewChangeState) SetEmptyPreprepares(view types.ViewNr, proposals map[tt.SeqNr][]byte) []*hasherpbtypes.HashData {

	// dataToHash will store the serialized form of newly created empty ("aborted") Preprepares.
	dataToHash := make([]*hasherpbtypes.HashData, 0, len(vcState.reproposals))

	maputil.IterateSorted(vcState.reproposals, func(sn tt.SeqNr, digest []byte) (cont bool) {
		if digest != nil && len(digest) == 0 {

			// If there is a pre-configured proposal, use it, otherwise use an empty byte slice.
			var data []byte
			if proposals[sn] != nil {
				data = proposals[sn]
			} else {
				data = []byte{}
			}

			// Create a new empty Preprepare message.
			vcState.preprepares[sn] = &pbftpbtypes.Preprepare{sn, view, data, true} // nolint:govet

			// Serialize the newly created Preprepare message for hashing.
			dataToHash = append(dataToHash, serializePreprepareForHashing(vcState.preprepares[sn]))
		}
		return true
	})

	// Return serialized Preprepares
	return dataToHash
}

func (vcState *pbftViewChangeState) SetEmptyPreprepareDigests(digests [][]byte) error {
	// Index of the next digest to assign.
	i := 0

	// Iterate over all empty reproposals the exactly same way as when setting the corresponding "aborted" Preprepares.
	maputil.IterateSorted(vcState.reproposals, func(sn tt.SeqNr, digest []byte) (cont bool) {

		// Assign the corresponding hash to each sequence number with an empty re-proposal.
		if digest != nil && len(digest) == 0 {
			vcState.reproposals[sn] = digests[i]
			i++
		}
		return true
	})

	// Sanity check (all digests must have been used)
	if i != len(digests) {
		return fmt.Errorf("more digests than empty preprepares")
	}

	return nil
}

func (vcState *pbftViewChangeState) SetLocalPreprepares(pbft *Orderer, view types.ViewNr) {
	for sn, digest := range vcState.reproposals {
		if vcState.preprepares[sn] == nil && digest != nil && len(digest) > 0 {
			if preprepare := pbft.lookUpPreprepare(sn, digest); preprepare != nil {
				// The re-proposed Preprepare must have an updated view.
				// We create a copy of the found Preprepare
				vcState.preprepares[sn] = copyPreprepareToNewView(preprepare, view)
			}
		}
	}
}

// askForMissingPreprepares requests the Preprepare messages that are part of a new view.
// The new primary might have received a prepare certificate from other nodes in the ViewChange messages they sent
// and thus the new primary has to re-propose the corresponding availability certificate
// by including the corresponding Preprepare message in the NewView message.
// However, the new primary might not have all the corresponding Preprepare messages,
// in which case it calls this function.
// Note that the requests for missing Preprepare messages need not necessarily be periodically re-transmitted.
// If they are dropped, the new primary will simply never send a NewView message
// and will be succeeded by another primary after another view change.
func (vcState *pbftViewChangeState) askForMissingPreprepares(moduleConfig *ModuleConfig) *events.EventList {

	eventsOut := events.EmptyList()
	for sn, digest := range vcState.reproposals {
		if len(digest) > 0 && vcState.preprepares[sn] == nil {
			eventsOut.PushBack(transportpbevents.SendMessage(
				moduleConfig.Net,
				pbftpbmsgs.PreprepareRequest(moduleConfig.Self, digest, sn),
				vcState.prepreparedIDs[sn],
			).Pb()) // TODO be smarter about this eventually, not asking everyone at once.
		}
	}
	return eventsOut
}

func (vcState *pbftViewChangeState) HasAllPreprepares() bool {
	for _, preprepare := range vcState.preprepares {
		if preprepare == nil {
			return false
		}
	}
	return true
}
