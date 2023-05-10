package common

import (
	"bytes"
	"fmt"
	"sort"

	ot "github.com/filecoin-project/mir/pkg/orderers/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"

	"github.com/filecoin-project/mir/pkg/iss/config"
	"github.com/filecoin-project/mir/pkg/logging"
	hasherpbtypes "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// PbftViewChangeState tracks the state of the view change sub-protocol in PBFT.
// It is only associated with a single view transition, e.g., the transition from view 1 to view 2.
// The transition from view 2 to view 3 will be tracked by a different instance of PbftViewChangeState.
type PbftViewChangeState struct {
	NumNodes          int
	SignedViewChanges map[t.NodeID]*pbftpbtypes.SignedViewChange

	// Digests of Preprepares that need to be repropsed in a new view.
	// At initialization, an explicit nil entry is created for each sequence number of the segment.
	// Based on received ViewChange messages, each value will eventually be set to
	// - either a non-zero-length byte slice representing a digest to be re-proposed
	// - or a zero-length byte slice marking it safe to re-propose a special fresh value
	// (in the current implementation, the fresh value is a pre-configured proposal
	// or, if none has been specified, an empty byte slice).
	Reproposals map[tt.SeqNr][]byte

	// Preprepare messages to be reproposed in the new view.
	// At initialization, an explicit nil entry is created for each sequence number of the segment.
	// Based on received ViewChange messages, each value will eventually be set to
	// a new Preprepare message to be re-proposed (with a correctly set view number).
	// This also holds for sequence numbers for which nothing was prepared in the previous view,
	// in which case the value is set to a fresh Preprepare (as configured) and the "aborted" flag set.
	Preprepares map[tt.SeqNr]*pbftpbtypes.Preprepare

	PrepreparedIDs map[tt.SeqNr][]t.NodeID
}

func NewPbftViewChangeState(seqNrs []tt.SeqNr, membership []t.NodeID) *PbftViewChangeState {
	reproposals := make(map[tt.SeqNr][]byte)
	preprepares := make(map[tt.SeqNr]*pbftpbtypes.Preprepare)
	for _, sn := range seqNrs {
		reproposals[sn] = nil
		preprepares[sn] = nil
	}

	return &PbftViewChangeState{
		NumNodes:          len(membership),
		SignedViewChanges: make(map[t.NodeID]*pbftpbtypes.SignedViewChange),
		Reproposals:       reproposals,
		PrepreparedIDs:    make(map[tt.SeqNr][]t.NodeID),
		Preprepares:       preprepares,
	}
}

func (vcState *PbftViewChangeState) EnoughViewChanges() bool {
	for _, r := range vcState.Reproposals {
		if r == nil {
			return false
		}
	}
	return true
}

func (vcState *PbftViewChangeState) AddSignedViewChange(svc *pbftpbtypes.SignedViewChange, from t.NodeID, logger logging.Logger) {

	if vcState.EnoughViewChanges() {
		return
	}

	if _, ok := vcState.SignedViewChanges[from]; ok {
		return
	}

	vcState.SignedViewChanges[from] = svc

	if len(vcState.SignedViewChanges) >= config.StrongQuorum(vcState.NumNodes) {
		vcState.updateReproposals(logger)
	}
}

// TODO: Make sure this procedure is fully deterministic (map iterations etc...)
func (vcState *PbftViewChangeState) updateReproposals(logger logging.Logger) {
	pSets, qSets := reconstructPSetQSet(vcState.SignedViewChanges, logger)

	for sn, r := range vcState.Reproposals {
		if r == nil {
			vcState.Reproposals[sn], vcState.PrepreparedIDs[sn] = reproposal(pSets, qSets, sn, vcState.NumNodes)
		}
	}
}

func (vcState *PbftViewChangeState) SetEmptyPreprepares(view ot.ViewNr, proposals map[tt.SeqNr][]byte) []*hasherpbtypes.HashData {

	// dataToHash will store the serialized form of newly created empty ("aborted") Preprepares.
	dataToHash := make([]*hasherpbtypes.HashData, 0, len(vcState.Reproposals))

	maputil.IterateSorted(vcState.Reproposals, func(sn tt.SeqNr, digest []byte) (cont bool) {
		if digest != nil && len(digest) == 0 {

			// If there is a pre-configured proposal, use it, otherwise use an empty byte slice.
			var data []byte
			if proposals[sn] != nil {
				data = proposals[sn]
			} else {
				data = []byte{}
			}

			// Create a new empty Preprepare message.
			vcState.Preprepares[sn] = &pbftpbtypes.Preprepare{sn, view, data, true} // nolint:govet

			// Serialize the newly created Preprepare message for hashing.
			dataToHash = append(dataToHash, SerializePreprepareForHashing(vcState.Preprepares[sn]))
		}
		return true
	})

	// Return serialized Preprepares
	return dataToHash
}

func (vcState *PbftViewChangeState) SetEmptyPreprepareDigests(digests [][]byte) error {
	// Index of the next digest to assign.
	i := 0

	// Iterate over all empty Reproposals the exactly same way as when setting the corresponding "aborted" Preprepares.
	maputil.IterateSorted(vcState.Reproposals, func(sn tt.SeqNr, digest []byte) (cont bool) {

		// Assign the corresponding hash to each sequence number with an empty re-proposal.
		if digest != nil && len(digest) == 0 {
			vcState.Reproposals[sn] = digests[i]
			i++
		}
		return true
	})

	// Sanity check (all digests must have been used)
	if i != len(digests) {
		return fmt.Errorf("more digests than empty Preprepares")
	}

	return nil
}

func (vcState *PbftViewChangeState) SetLocalPreprepares(state *State, view ot.ViewNr) {
	for sn, digest := range vcState.Reproposals {
		if vcState.Preprepares[sn] == nil && digest != nil && len(digest) > 0 {
			if preprepare := state.LookUpPreprepare(sn, digest); preprepare != nil {
				// The re-proposed Preprepare must have an updated view.
				// We create a copy of the found Preprepare
				prepareCopy := *preprepare
				preprepare.View = view
				vcState.Preprepares[sn] = &prepareCopy
			}
		}
	}
}

func (vcState *PbftViewChangeState) HasAllPreprepares() bool {
	for _, preprepare := range vcState.Preprepares {
		if preprepare == nil {
			return false
		}
	}
	return true
}

// ============================================================
// NewView message construction
// ============================================================

func noPrepareConflictsA1(
	pSets map[t.NodeID]ViewChangePSet,
	sn tt.SeqNr,
	digest []byte,
	view ot.ViewNr,
	numNodes int,
) bool {
	numNonConflicting := 0

	for _, pSet := range pSets {
		if entry, ok := pSet[sn]; !ok {
			numNonConflicting++
		} else {
			if entry.View < view ||
				(entry.View == view && bytes.Equal(entry.Digest, digest)) {
				numNonConflicting++
			}
		}
	}

	return numNonConflicting >= config.StrongQuorum(numNodes)
}

func enoughPrepreparesA2(
	qSets map[t.NodeID]ViewChangeQSet,
	sn tt.SeqNr,
	digest []byte,
	view ot.ViewNr,
	numNodes int,
) (bool, []t.NodeID) {

	numPrepares := 0
	nodeIDs := make([]t.NodeID, 0, numNodes)

	for nodeID, qSet := range qSets {
		if snEntry, ok := qSet[sn]; ok {
			if snEntry[string(digest)] >= view {
				numPrepares++
				nodeIDs = append(nodeIDs, nodeID)
			}
		}
	}

	return numPrepares >= config.WeakQuorum(numNodes), nodeIDs
}

func nothingPreparedB(pSets map[t.NodeID]ViewChangePSet, sn tt.SeqNr, numNodes int) bool {
	nothingPrepared := 0

	for _, pSet := range pSets {
		if _, ok := pSet[sn]; !ok {
			nothingPrepared++
		}
	}

	return nothingPrepared >= config.StrongQuorum(numNodes)
}

// ViewChangePSet represents the P set of a PBFT view change message.
// For each sequence number, it holds the digest of the last prepared value,
// along with the view in which it was prepared.
type ViewChangePSet map[tt.SeqNr]*pbftpbtypes.PSetEntry

func reconstructPSet(entries []*pbftpbtypes.PSetEntry) (ViewChangePSet, error) {
	pSet := make(ViewChangePSet)
	for _, entry := range entries {

		// There can be at most one entry per sequence number. Otherwise, the set is not valid.
		if _, ok := pSet[entry.Sn]; ok {
			return nil, fmt.Errorf("invalid Pset: conflicting prepare entries")
		}

		pSet[entry.Sn] = entry
	}

	return pSet, nil
}

// The Q set of a PBFT view change message.
// For each sequence number, it holds the digests (encoded as string map keys)
// of all values preprepared for that sequence number,
// along with the latest view in which each of them was preprepared.
type ViewChangeQSet map[tt.SeqNr]map[string]ot.ViewNr

func reconstructQSet(entries []*pbftpbtypes.QSetEntry) (ViewChangeQSet, error) {
	qSet := make(ViewChangeQSet)
	for _, entry := range entries {

		var snEntry map[string]ot.ViewNr
		if sne, ok := qSet[entry.Sn]; ok {
			snEntry = sne
		} else {
			snEntry = make(map[string]ot.ViewNr)
			qSet[entry.Sn] = snEntry
		}

		// There can be at most one entry per digest and sequence number. Otherwise, the set is not valid.
		if _, ok := snEntry[string(entry.Digest)]; ok {
			return nil, fmt.Errorf("invalid Qset: conflicting preprepare entries")
		}

		snEntry[string(entry.Digest)] = entry.View

	}

	return qSet, nil

}

func reconstructPSetQSet(
	signedViewChanges map[t.NodeID]*pbftpbtypes.SignedViewChange,
	logger logging.Logger,
) (map[t.NodeID]ViewChangePSet, map[t.NodeID]ViewChangeQSet) {
	pSets := make(map[t.NodeID]ViewChangePSet)
	qSets := make(map[t.NodeID]ViewChangeQSet)

	for nodeID, svc := range signedViewChanges {
		var err error
		var pSet ViewChangePSet
		var qSet ViewChangeQSet

		pSet, err = reconstructPSet(svc.ViewChange.PSet)
		if err != nil {
			logger.Log(logging.LevelWarn, "could not reconstruct PSet for PBFT view change", "err", err)
			continue
		}

		qSet, err = reconstructQSet(svc.ViewChange.QSet)
		if err != nil {
			logger.Log(logging.LevelWarn, "could not reconstruct QSet for PBFT view change", "err", err)
			continue
		}

		pSets[nodeID] = pSet
		qSets[nodeID] = qSet
	}

	return pSets, qSets
}

func reproposal(
	pSets map[t.NodeID]ViewChangePSet,
	qSets map[t.NodeID]ViewChangeQSet,
	sn tt.SeqNr,
	numNodes int,
) ([]byte, []t.NodeID) {

	if nothingPreparedB(pSets, sn, numNodes) {

		return []byte{}, nil

	}

	for _, pSet := range pSets {
		if entry, ok := pSet[sn]; ok {
			a2, prepreparedIDs := enoughPrepreparesA2(qSets, sn, entry.Digest, entry.View, numNodes)
			if noPrepareConflictsA1(pSets, sn, entry.Digest, entry.View, numNodes) && a2 {

				return entry.Digest, prepreparedIDs

			}
		}
	}

	return nil, nil
}

// PbType returns a protobuf type representation (not a raw protobuf, but the generated type) of a viewChangePSet,
// Where all entries are stored in a simple list.
// The list is sorted for repeatability.
func (pSet ViewChangePSet) PbType() []*pbftpbtypes.PSetEntry {

	list := make([]*pbftpbtypes.PSetEntry, 0, len(pSet))

	for _, pEntry := range pSet {
		list = append(list, pEntry)
	}

	sort.Slice(list, func(i int, j int) bool {
		if list[i].Sn != list[j].Sn {
			return list[i].Sn < list[j].Sn
		} else if list[i].View != list[j].View {
			return list[i].View < list[j].View
		} else {
			return bytes.Compare(list[i].Digest, list[j].Digest) < 0
		}
	})

	return list
}

// PbType returns a protobuf tye representation (not a raw protobuf, but the generated type) of a ViewChangeQSet,
// where all entries, represented as (sn, view, digest) tuples, are stored in a simple list.
// The list is sorted for repeatability.
func (qSet ViewChangeQSet) PbType() []*pbftpbtypes.QSetEntry {

	list := make([]*pbftpbtypes.QSetEntry, 0, len(qSet))

	for sn, qEntry := range qSet {
		for digest, view := range qEntry {
			list = append(list, &pbftpbtypes.QSetEntry{
				Sn:     sn,
				View:   view,
				Digest: []byte(digest),
			})
		}
	}

	sort.Slice(list, func(i int, j int) bool {
		if list[i].Sn != list[j].Sn {
			return list[i].Sn < list[j].Sn
		} else if list[i].View != list[j].View {
			return list[i].View < list[j].View
		} else {
			return bytes.Compare(list[i].Digest, list[j].Digest) < 0
		}
	})

	return list
}
