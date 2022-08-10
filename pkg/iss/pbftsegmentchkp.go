package iss

import (
	"bytes"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/isspbftpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// pbftSegmentChkp groups data structures pertaining to an instance-local checkpoint
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
type pbftSegmentChkp struct {

	// Saves all the Done messages received from other nodes.
	// They contain the hashes of all Preprepare messages, as committed by the respective nodes.
	// (A node sends a Done message when it commits everything in the segment.)
	doneMessages map[t.NodeID]*isspbftpb.Done

	// For each received Done message, stores the IDs of nodes that sent it.
	doneMsgIndex map[string][]t.NodeID

	// Once enough Done messages have been received,
	// digests will contain the Preprepare digests for all sequence numbers in the segment.
	digests map[t.SeqNr][]byte

	// Set to true when sending the Done message.
	// This happens when the node locally commits all slots of the segment.
	done bool

	// Once enough Done messages have been received,
	// DoneNodes will contain IDs of nodes from which matching Done messages were received.
	doneNodes []t.NodeID

	// This flag is set once the retransmission of missing committed requests is requested.
	// It serves preventing redundant retransmission requests when more than a quorum of Done messages are received.
	catchingUp bool
}

// newPbftSegmentChkp returns a pointer to a new instance of pbftSegmentChkp
func newPbftSegmentChkp() *pbftSegmentChkp {
	return &pbftSegmentChkp{
		doneMessages: make(map[t.NodeID]*isspbftpb.Done),
		doneMsgIndex: make(map[string][]t.NodeID),
	}
}

// SetDone marks the local checkpoint as done. It must be called after all slots of the segment have been committed.
// This is a necessary condition for the checkpoint to be considered stable.
func (chkp *pbftSegmentChkp) SetDone() {
	chkp.done = true
}

// Digests returns, for each sequence number of the associated segment, the digest of the committed certificate
// (more precisely, the digest of the corresponding Preprepare message).
// If the information is not yet available (not enough Done messages have been received), Digests returns nil.
func (chkp *pbftSegmentChkp) Digests() map[t.SeqNr][]byte {
	return chkp.digests
}

// DoneNodes returns a list of IDs of Nodes from which a matching Done message has been received.
// If a quorum of Done messages has not yet been received, DoneNodes returns nil.
func (chkp *pbftSegmentChkp) DoneNodes() []t.NodeID {
	return chkp.doneNodes
}

// Stable returns true if the instance-level checkpoint is stable,
// i.e., if all slots have been committed and a strong quorum of matching Done messages has been received.
// This ensures that at least a weak quorum of correct nodes has a local checkpoint
// and thus evey correct node will be able to catch up.
func (chkp *pbftSegmentChkp) Stable(numNodes int) bool {

	// If not all slots are committed (i.e. no checkpoint is present locally), the checkpoint is not considered stable.
	if !chkp.done {
		return false
	}

	// Return true if a strong quorum of nodes sent the same Done message.
	for _, nodeIDs := range chkp.doneMsgIndex {
		if len(nodeIDs) >= strongQuorum(numNodes) {
			return true
		}
	}

	return false
}

// NodeDone registers a Done message received from a node.
// Once NodeDone has been called with matching Done messages for a quorum of nodes,
// the instance-level checkpoint will become stable.
func (chkp *pbftSegmentChkp) NodeDone(nodeID t.NodeID, doneMsg *isspbftpb.Done, segment *segment) {

	// Ignore duplicate Done messages.
	if _, ok := chkp.doneMessages[nodeID]; ok {
		return
	}

	// Store Done message
	chkp.doneMessages[nodeID] = doneMsg
	strKey := aggregateToString(doneMsg.Digests)
	chkp.doneMsgIndex[strKey] = append(chkp.doneMsgIndex[strKey], nodeID)

	// If a quorum of nodes has sent a Done message
	if len(chkp.doneMsgIndex[strKey]) >= weakQuorum(len(segment.Membership)) {

		// Save the IDs of the nodes that are done with the segment
		chkp.doneNodes = chkp.doneMsgIndex[strKey]

		// Save, for each sequence number of the segment,
		// the corresponding Preprepare digest of the committed certificate.
		if chkp.digests == nil {
			chkp.digests = make(map[t.SeqNr][]byte, len(segment.SeqNrs))
			for i, sn := range segment.SeqNrs {
				chkp.digests[sn] = doneMsg.Digests[i]
			}
		}
	}
}

// aggregateToString concatenates a list of byte arrays to a single string.
// Used for obtaining a string representation of the content of a Done message to use it as a map key
// (in doneMsgIndex)
func aggregateToString(digests [][]byte) string {
	str := ""
	for _, digest := range digests {
		str += string(digest)
	}

	return str
}

// sendDoneMessages sends a Done message to all other nodes as part of the instance-level checkpoint subprotocol.
// This method is called when all slots have been committed.
func (pbft *pbftInstance) sendDoneMessages() *events.EventList {

	pbft.logger.Log(logging.LevelInfo, "Done with segment.")

	// Collect the preprepare digests of all committed certificates.
	digests := make([][]byte, 0, len(pbft.segment.SeqNrs))
	maputil.IterateSorted(pbft.slots[pbft.view], func(sn t.SeqNr, slot *pbftSlot) bool {
		digests = append(digests, slot.Digest)
		return true
	})

	// Periodically send a Done message with the digests to all other nodes.
	return events.ListOf(pbft.eventService.TimerRepeat(
		t.TimeDuration(pbft.config.DoneResendPeriod),
		pbft.eventService.SendMessage(PbftDoneSBMessage(digests), pbft.segment.Membership),
	))
}

// applyMsgDone applies a received Done message.
// Once enough Done messages have been applied, makes the protocol
// - stop participating in view changes and
// - set up a timer for fetching missing certificates.
func (pbft *pbftInstance) applyMsgDone(doneMsg *isspbftpb.Done, from t.NodeID) *events.EventList {

	// Register Done message.
	pbft.segmentCheckpoint.NodeDone(from, doneMsg, pbft.segment)

	// If more Done messages still need to be received or retransmission has already been requested, do nothing.
	doneNodes := pbft.segmentCheckpoint.DoneNodes()
	if doneNodes == nil || pbft.segmentCheckpoint.catchingUp {
		return events.EmptyList()
	}

	// If this was the last Done message required for a quorum,
	// set up a timer to ask for the missing committed certificates.
	// In case the requests get lost, they need to be periodically retransmitted.
	// Also, we still want to give the node some time to deliver the requests naturally before trying to catch up.
	// Thus, we pack a TimerRepeat Event (that triggers the first "repetition" immediately) inside a TimerDelay.
	// We also set the catchingUp flag to prevent this code from executing more than once per PBFT instance.
	pbft.segmentCheckpoint.catchingUp = true
	return events.ListOf(pbft.eventService.TimerDelay(
		t.TimeDuration(pbft.config.CatchUpDelay),
		pbft.eventService.TimerRepeat(
			t.TimeDuration(pbft.config.CatchUpDelay),
			pbft.catchUpRequests(doneNodes, pbft.segmentCheckpoint.Digests())...,
		),
	))

	// TODO: Requesting all missing certificates from all the nodes known to have them right away is quite an overkill,
	//       resulting in a huge waste of resources. Be smarter about it by, for example, only asking a few nodes first.
}

// catchUpRequests assembles and returns a list of Events
// representing requests for retransmission of committed certificates.
// The list contains one request for each slot of the segment that has not yet been committed.
func (pbft *pbftInstance) catchUpRequests(nodes []t.NodeID, digests map[t.SeqNr][]byte) []*eventpb.Event {

	catchUpRequests := make([]*eventpb.Event, 0)

	// Deterministically iterate through all the (sequence number, certificate) pairs
	// received in a quorum of Done messages.
	maputil.IterateSorted(digests, func(sn t.SeqNr, digest []byte) bool {

		// If no certificate has been committed for the sequence number, create a retransmission request.
		if !pbft.slots[pbft.view][sn].Committed {
			catchUpRequests = append(catchUpRequests, pbft.eventService.SendMessage(
				PbftCatchUpRequestSBMessage(sn, digest),
				nodes,
			))
		}
		return true
	})

	return catchUpRequests
}

// applyMsgCatchUpRequest applies a request for retransmitting a missing committed certificate.
// It looks up the requested certificate (more precisely, the corresponding Preprepare message)
// by its sequence number and digest and sends it to the originator of the request inside a CatchUpResponse message.
// If no matching Preprepare is found, does nothing.
func (pbft *pbftInstance) applyMsgCatchUpRequest(
	catchUpReq *isspbftpb.CatchUpRequest,
	from t.NodeID,
) *events.EventList {
	if preprepare := pbft.lookUpPreprepare(t.SeqNr(catchUpReq.Sn), catchUpReq.Digest); preprepare != nil {

		// If the requested Preprepare message is available, send it to the originator of the request.
		// No need for periodic re-transmission. The requester will re-transmit the request if needed.
		return events.ListOf(pbft.eventService.SendMessage(PbftCatchUpResponseSBMessage(preprepare), []t.NodeID{from}))
	}

	// If the requested Preprepare message is not available, ignore the request.
	return events.EmptyList()
}

// applyMsgCatchUpResponse applies a retransmitted missing committed certificate.
// It only requests hashing of the response,
// the actual handling of it being performed only when the hash result is available.
func (pbft *pbftInstance) applyMsgCatchUpResponse(preprepare *isspbftpb.Preprepare, _ t.NodeID) *events.EventList {

	// Ignore preprepare if received in the meantime.
	// This check is technically redundant, as it is (and must be) performed also after the Preprepare is hashed.
	// However, it might prevent some unnecessary hash computation if performed here as well.
	if pbft.slots[pbft.view][t.SeqNr(preprepare.Sn)].Committed {
		return events.EmptyList()
	}

	// Request a hash of the received preprepare message.
	hashRequest := pbft.eventService.HashRequest(
		[][][]byte{serializePreprepareForHashing(preprepare)},
		catchUpResponseHashOrigin(preprepare),
	)
	return events.ListOf(hashRequest)
}

// applyCatchUpResponseHashResult processes a missing committed certificate when its hash becomes available.
// It is the final step of catching up with an instance-level checkpoint.
func (pbft *pbftInstance) applyCatchUpResponseHashResult(
	digest []byte,
	preprepare *isspbftpb.Preprepare,
) *events.EventList {

	eventsOut := events.EmptyList()

	// Convenience variables
	sn := t.SeqNr(preprepare.Sn)
	slot := pbft.slots[pbft.view][sn]

	// Ignore preprepare if slot is already committed.
	if pbft.slots[pbft.view][t.SeqNr(preprepare.Sn)].Committed {
		return events.EmptyList()
	}

	// Check whether the received certificate was actually requested (a faulty node might have sent it on its own).
	digests := pbft.segmentCheckpoint.Digests()
	if digests == nil {
		pbft.logger.Log(logging.LevelWarn, "Ignoring unsolicited CatchUpResponse.", "sn", sn)
		return events.EmptyList()
	}

	// Check whether the digest of the received message matches the requested one.
	if !bytes.Equal(digests[sn], digest) {
		pbft.logger.Log(logging.LevelWarn, "Ignoring CatchUpResponse with invalid digest.", "sn", sn)
		return events.EmptyList()
	}

	pbft.logger.Log(logging.LevelDebug, "Catching up with segment-level checkpoint.", "sn", sn)

	// Add the missing certificate, updating the corresponding Preprepare's view.
	// Note that copying a Preprepare with an updated view preserves its hash.
	slot.catchUp(copyPreprepareToNewView(preprepare, pbft.view), digest)

	// If all certificates have been committed (i.e. this is the last certificate to commit),
	// send a Done message to all other nodes.
	// This is required for liveness, see comments for pbftSegmentChkp.
	if pbft.allCommitted() {
		pbft.segmentCheckpoint.SetDone()
		eventsOut.PushBackList(pbft.sendDoneMessages())
	}

	// Deliver certificate.
	eventsOut.PushBack(pbft.eventService.SBEvent(SBDeliverEvent(
		sn,
		slot.Preprepare.CertData,
		slot.Preprepare.Aborted,
	)))

	return eventsOut
}
