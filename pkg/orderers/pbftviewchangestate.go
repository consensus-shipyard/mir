package orderers

import (
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/iss/config"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/ordererspbftpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// pbftViewChangeState tracks the state of the view change sub-protocol in PBFT.
// It is only associated with a single view transition, e.g., the transition from view 1 to view 2.
// The transition from view 2 to view 3 will be tracked by a different instance of pbftViewChangeState.
type pbftViewChangeState struct {
	numNodes          int
	signedViewChanges map[t.NodeID]*ordererspbftpb.SignedViewChange

	// Digests of preprepares that need to be repropsed in a new view.
	// At initialization, an explicit nil entry is created for each sequence number of the segment.
	// Based on received ViewChange messages, each value will eventually be set to
	// - either a non-zero-length byte slice representing a digest to be re-proposed
	// - or a zero-length byte slice marking it safe to re-propose a special "null" certificate.
	reproposals map[t.SeqNr][]byte

	// Preprepare messages to be reproposed in the new view.
	// At initialization, an explicit nil entry is created for each sequence number of the segment.
	// Based on received ViewChange messages, each value will eventually be set to
	// a new Preprepare message to be re-proposed (with a correctly set view number).
	// This also holds for sequence numbers for which nothing was prepared in the previous view,
	// in which case the value is set to a Preprepare with an empty certificate and the "aborted" flag set.
	preprepares map[t.SeqNr]*ordererspbftpb.Preprepare

	prepreparedIDs map[t.SeqNr][]t.NodeID

	logger logging.Logger
}

func newPbftViewChangeState(seqNrs []t.SeqNr, membership []t.NodeID, logger logging.Logger) *pbftViewChangeState {
	reproposals := make(map[t.SeqNr][]byte)
	preprepares := make(map[t.SeqNr]*ordererspbftpb.Preprepare)
	for _, sn := range seqNrs {
		reproposals[sn] = nil
		preprepares[sn] = nil
	}

	return &pbftViewChangeState{
		numNodes:          len(membership),
		signedViewChanges: make(map[t.NodeID]*ordererspbftpb.SignedViewChange),
		reproposals:       reproposals,
		prepreparedIDs:    make(map[t.SeqNr][]t.NodeID),
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

func (vcState *pbftViewChangeState) AddSignedViewChange(svc *ordererspbftpb.SignedViewChange, from t.NodeID) {

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

func (vcState *pbftViewChangeState) SetEmptyPreprepares(view t.PBFTViewNr, proposals map[t.SeqNr][]byte) [][][]byte {

	// dataToHash will store the serialized form of newly created empty ("aborted") Preprepares.
	dataToHash := make([][][]byte, 0, len(vcState.reproposals))

	maputil.IterateSorted(vcState.reproposals, func(sn t.SeqNr, digest []byte) (cont bool) {
		if digest != nil && len(digest) == 0 {

			// If there is a pre-configured proposal, use it, otherwise use an empty byte slice.
			var data []byte
			if proposals[sn] != nil {
				data = proposals[sn]
			} else {
				data = []byte{}
			}

			// Create a new empty Preprepare message.
			vcState.preprepares[sn] = pbftPreprepareMsg(
				sn,
				view,
				data,
				true,
			)

			// Serialize the newly created Preprepare message for hashing.
			dataToHash = append(dataToHash, serializePreprepareForHashing(vcState.preprepares[sn]))
		}
		return true
	})

	// Return serialized Preprepares
	return dataToHash
}

func (vcState *pbftViewChangeState) SetEmptyPreprepareDigests(digests [][]byte) {
	// Index of the next digest to assign.
	i := 0

	// Iterate over all empty reproposals the exactly same way as when setting the corresponding "aborted" Preprepares.
	maputil.IterateSorted(vcState.reproposals, func(sn t.SeqNr, digest []byte) (cont bool) {

		// Assign the corresponding hash to each sequence number with an empty re-proposal.
		if digest != nil && len(digest) == 0 {
			vcState.reproposals[sn] = digests[i]
			i++
		}
		return true
	})

	// Sanity check (all digests must have been used)
	if i != len(digests) {
		panic("more digests than empty preprepares")
	}
}

func (vcState *pbftViewChangeState) SetLocalPreprepares(pbft *Orderer, view t.PBFTViewNr) {
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
			eventsOut.PushBack(events.SendMessage(
				moduleConfig.Net,
				OrdererMessage(
					PbftPreprepareRequestSBMessage(
						pbftPreprepareRequestMsg(sn, digest)),
					moduleConfig.Self),
				removeNodeID(vcState.prepreparedIDs[sn], "1"), // TODO be smarter about this eventually, not asking everyone at once.
			))
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
