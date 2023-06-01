package common

import (
	"github.com/filecoin-project/mir/pkg/orderers/common"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	types2 "github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/membutil"
)

// PbftSegmentChkp groups data structures pertaining to an instance-local checkpoint
// created when all slots have been committed.
// If correct nodes stop participating in the protocol immediately after having delivered all certificates in a segment,
// a minority of nodes might get stuck in higher views if they did not deliver all certificates yet,
// never finding enough support for finishing a view change.
// The high-level checkpoints (encompassing all segments) are not enough to resolve this problem,
// because multiple segments can block each other, each with its own majority of nodes that finish it,
// but with too few nodes that finished all segments.
// If, for example, each segment only has one node that fails to complete it,
// no high-level checkpoint can be constructed,
// as too few nodes will have delivered everything from all the segments.
// Thus, local instance-level checkpoints are required, so nodes can catch up on each segment separately.
type PbftSegmentChkp struct {

	// Saves all the Done messages received from other nodes.
	// They contain the hashes of all Preprepare messages, as committed by the respective nodes.
	// (A node sends a Done message when it commits everything in the segment.)
	doneReceived map[types.NodeID]struct{}

	// For each received Done message, stores the IDs of nodes that sent it.
	doneMsgIndex map[string][]types.NodeID

	// Once enough Done messages have been received,
	// digests will contain the Preprepare digests for all sequence numbers in the segment.
	digests map[types2.SeqNr][]byte

	// Set to true when sending the Done message.
	// This happens when the node locally commits all slots of the segment.
	done bool

	// Once enough Done messages have been received,
	// DoneNodes will contain IDs of nodes from which matching Done messages were received.
	doneNodes []types.NodeID

	// This flag is set once the retransmission of missing committed entries is requested.
	// It serves preventing redundant retransmission entries when more than a quorum of Done messages are received.
	CatchingUp bool
}

// NewPbftSegmentChkp returns a pointer to a new instance of PbftSegmentChkp
func NewPbftSegmentChkp() *PbftSegmentChkp {
	return &PbftSegmentChkp{
		doneReceived: make(map[types.NodeID]struct{}),
		doneMsgIndex: make(map[string][]types.NodeID),
	}
}

// SetDone marks the local checkpoint as done. It must be called after all slots of the segment have been committed.
// This is a necessary condition for the checkpoint to be considered stable.
func (chkp *PbftSegmentChkp) SetDone() {
	chkp.done = true
}

// Digests returns, for each sequence number of the associated segment, the digest of the committed certificate
// (more precisely, the digest of the corresponding Preprepare message).
// If the information is not yet available (not enough Done messages have been received), Digests returns nil.
func (chkp *PbftSegmentChkp) Digests() map[types2.SeqNr][]byte {
	return chkp.digests
}

// DoneNodes returns a list of IDs of Nodes from which a matching Done message has been received.
// If a quorum of Done messages has not yet been received, DoneNodes returns nil.
func (chkp *PbftSegmentChkp) DoneNodes() []types.NodeID {
	return chkp.doneNodes
}

// Stable returns true if the instance-level checkpoint is stable,
// i.e., if all slots have been committed and a strong quorum of matching Done messages has been received.
// This ensures that at least a weak quorum of correct nodes has a local checkpoint
// and thus evey correct node will be able to catch up.
func (chkp *PbftSegmentChkp) Stable(membership *trantorpbtypes.Membership) bool {

	// If not all slots are committed (i.e. no checkpoint is present locally), the checkpoint is not considered stable.
	if !chkp.done {
		return false
	}

	// Return true if a strong quorum of nodes sent the same Done message.
	for _, nodeIDs := range chkp.doneMsgIndex {
		if membutil.HaveStrongQuorum(membership, nodeIDs) {
			return true
		}
	}

	return false
}

// NodeDone registers a Done message received from a node.
// Once NodeDone has been called with matching Done messages for a quorum of nodes,
// the instance-level checkpoint will become stable.
func (chkp *PbftSegmentChkp) NodeDone(nodeID types.NodeID, doneDigests [][]byte, segment *common.Segment) {

	// Ignore duplicate Done messages.
	if _, ok := chkp.doneReceived[nodeID]; ok {
		return
	}

	// Store Done message
	chkp.doneReceived[nodeID] = struct{}{}
	strKey := chkp.aggregateToString(doneDigests)

	chkp.doneMsgIndex[strKey] = append(chkp.doneMsgIndex[strKey], nodeID)

	// If a quorum of nodes has sent a Done message
	if membutil.HaveWeakQuorum(segment.Membership, chkp.doneMsgIndex[strKey]) {

		// Save the IDs of the nodes that are done with the segment
		chkp.doneNodes = chkp.doneMsgIndex[strKey]

		// Save, for each sequence number of the segment,
		// the corresponding Preprepare digest of the committed certificate.
		if chkp.digests == nil {
			chkp.digests = make(map[types2.SeqNr][]byte, segment.Len())
			for i, sn := range segment.SeqNrs() {
				chkp.digests[sn] = doneDigests[i]
			}
		}
	}
}

// aggregateToString concatenates a list of byte arrays to a single string.
// Used for obtaining a string representation of the content of a Done message to use it as a map key
// (in doneMsgIndex)
func (chkp *PbftSegmentChkp) aggregateToString(digests [][]byte) string {
	str := ""
	for _, digest := range digests {
		str += string(digest)
	}

	return str
}
