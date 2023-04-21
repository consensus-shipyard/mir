package orderers

import (
	"bytes"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/iss/config"
	"github.com/filecoin-project/mir/pkg/logging"
	commonpbtypes "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	eventpbevents "github.com/filecoin-project/mir/pkg/pb/eventpb/events"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	hasherpbevents "github.com/filecoin-project/mir/pkg/pb/hasherpb/events"
	isspbevents "github.com/filecoin-project/mir/pkg/pb/isspb/events"
	pbftpbmsgs "github.com/filecoin-project/mir/pkg/pb/pbftpb/msgs"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	"github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
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
	doneMessages map[t.NodeID]*pbftpbtypes.Done

	// For each received Done message, stores the IDs of nodes that sent it.
	doneMsgIndex map[string][]t.NodeID

	// Once enough Done messages have been received,
	// digests will contain the Preprepare digests for all sequence numbers in the segment.
	digests map[tt.SeqNr][]byte

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
		doneMessages: make(map[t.NodeID]*pbftpbtypes.Done),
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
func (chkp *pbftSegmentChkp) Digests() map[tt.SeqNr][]byte {
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
		if len(nodeIDs) >= config.StrongQuorum(numNodes) {
			return true
		}
	}

	return false
}

// NodeDone registers a Done message received from a node.
// Once NodeDone has been called with matching Done messages for a quorum of nodes,
// the instance-level checkpoint will become stable.
func (chkp *pbftSegmentChkp) NodeDone(nodeID t.NodeID, doneMsg *pbftpbtypes.Done, segment *Segment) {

	// Ignore duplicate Done messages.
	if _, ok := chkp.doneMessages[nodeID]; ok {
		return
	}

	// Store Done message
	chkp.doneMessages[nodeID] = doneMsg
	strKey := aggregateToString(doneMsg.Digests)
	chkp.doneMsgIndex[strKey] = append(chkp.doneMsgIndex[strKey], nodeID)

	// If a quorum of nodes has sent a Done message
	if len(chkp.doneMsgIndex[strKey]) >= config.WeakQuorum(len(segment.Membership.Nodes)) {

		// Save the IDs of the nodes that are done with the segment
		chkp.doneNodes = chkp.doneMsgIndex[strKey]

		// Save, for each sequence number of the segment,
		// the corresponding Preprepare digest of the committed certificate.
		if chkp.digests == nil {
			chkp.digests = make(map[tt.SeqNr][]byte, segment.Len())
			for i, sn := range segment.SeqNrs() {
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
func (orderer *Orderer) sendDoneMessages() *events.EventList {

	orderer.logger.Log(logging.LevelInfo, "Done with segment.")

	// Collect the preprepare digests of all committed certificates.
	digests := make([][]byte, 0, orderer.segment.Len())
	maputil.IterateSorted(orderer.slots[orderer.view], func(sn tt.SeqNr, slot *pbftSlot) bool {
		digests = append(digests, slot.Digest)
		return true
	})

	// Periodically send a Done message with the digests to all other nodes.
	return events.ListOf(eventpbevents.TimerRepeat(
		orderer.moduleConfig.Timer,
		[]*eventpbtypes.Event{transportpbevents.SendMessage(
			orderer.moduleConfig.Net,
			pbftpbmsgs.Done(orderer.moduleConfig.Self, digests),
			orderer.segment.NodeIDs(),
		)},
		types.Duration(orderer.config.DoneResendPeriod),
		tt.RetentionIndex(orderer.config.epochNr),
	).Pb())
}

// applyMsgDone applies a received Done message.
// Once enough Done messages have been applied, makes the protocol
// - stop participating in view changes and
// - set up a timer for fetching missing certificates.
func (orderer *Orderer) applyMsgDone(doneMsg *pbftpbtypes.Done, from t.NodeID) *events.EventList {

	// Register Done message.
	orderer.segmentCheckpoint.NodeDone(from, doneMsg, orderer.segment)

	// If more Done messages still need to be received or retransmission has already been requested, do nothing.
	doneNodes := orderer.segmentCheckpoint.DoneNodes()
	if doneNodes == nil || orderer.segmentCheckpoint.catchingUp {
		return events.EmptyList()
	}

	// If this was the last Done message required for a quorum,
	// set up a timer to ask for the missing committed certificates.
	// In case the requests get lost, they need to be periodically retransmitted.
	// Also, we still want to give the node some time to deliver the requests naturally before trying to catch up.
	// Thus, we pack a TimerRepeat OrdererEvent (that triggers the first "repetition" immediately) inside a TimerDelay.
	// We also set the catchingUp flag to prevent this code from executing more than once per PBFT instance.
	orderer.segmentCheckpoint.catchingUp = true

	return events.ListOf(eventpbevents.TimerDelay(
		orderer.moduleConfig.Timer,
		[]*eventpbtypes.Event{eventpbevents.TimerRepeat(
			orderer.moduleConfig.Timer,
			orderer.catchUpRequests(doneNodes, orderer.segmentCheckpoint.Digests()),
			types.Duration(orderer.config.CatchUpDelay),
			tt.RetentionIndex(orderer.config.epochNr))},
		types.Duration(orderer.config.CatchUpDelay),
	).Pb())

	// TODO: Requesting all missing certificates from all the nodes known to have them right away is quite an overkill,
	//       resulting in a huge waste of resources. Be smarter about it by, for example, only asking a few nodes first.
}

// catchUpRequests assembles and returns a list of Events
// representing requests for retransmission of committed certificates.
// The list contains one request for each slot of the segment that has not yet been committed.
func (orderer *Orderer) catchUpRequests(nodes []t.NodeID, digests map[tt.SeqNr][]byte) []*eventpbtypes.Event {

	catchUpRequests := make([]*eventpbtypes.Event, 0)

	// Deterministically iterate through all the (sequence number, certificate) pairs
	// received in a quorum of Done messages.
	maputil.IterateSorted(digests, func(sn tt.SeqNr, digest []byte) bool {

		// If no certificate has been committed for the sequence number, create a retransmission request.
		if !orderer.slots[orderer.view][sn].Committed {
			catchUpRequests = append(catchUpRequests, transportpbevents.SendMessage(
				orderer.moduleConfig.Net,
				pbftpbmsgs.CatchUpRequest(orderer.moduleConfig.Self, digest, sn),
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
func (orderer *Orderer) applyMsgCatchUpRequest(
	catchUpReq *pbftpbtypes.CatchUpRequest,
	from t.NodeID,
) *events.EventList {
	if preprepare := orderer.lookUpPreprepare(catchUpReq.Sn, catchUpReq.Digest); preprepare != nil {

		// If the requested Preprepare message is available, send it to the originator of the request.
		// No need for periodic re-transmission. The requester will re-transmit the request if needed.
		return events.ListOf(transportpbevents.SendMessage(
			orderer.moduleConfig.Net,
			pbftpbmsgs.CatchUpResponse(orderer.moduleConfig.Self, preprepare),
			[]t.NodeID{from}).Pb(),
		)
	}

	// If the requested Preprepare message is not available, ignore the request.
	return events.EmptyList()
}

// applyMsgCatchUpResponse applies a retransmitted missing committed certificate.
// It only requests hashing of the response,
// the actual handling of it being performed only when the hash result is available.
func (orderer *Orderer) applyMsgCatchUpResponse(preprepare *pbftpbtypes.Preprepare, _ t.NodeID) *events.EventList {

	// Ignore preprepare if received in the meantime.
	// This check is technically redundant, as it is (and must be) performed also after the Preprepare is hashed.
	// However, it might prevent some unnecessary hash computation if performed here as well.
	if orderer.slots[orderer.view][preprepare.Sn].Committed {
		return events.EmptyList()
	}

	return events.ListOf(hasherpbevents.Request(
		orderer.moduleConfig.Hasher,
		[]*commonpbtypes.HashData{serializePreprepareForHashing(preprepare)},
		HashOrigin(orderer.moduleConfig.Self, catchUpResponseHashOrigin(preprepare.Pb())),
	).Pb())
}

// applyCatchUpResponseHashResult processes a missing committed certificate when its hash becomes available.
// It is the final step of catching up with an instance-level checkpoint.
func (orderer *Orderer) applyCatchUpResponseHashResult(
	digest []byte,
	preprepare *pbftpbtypes.Preprepare,
) *events.EventList {

	eventsOut := events.EmptyList()

	// Convenience variables
	sn := preprepare.Sn
	slot := orderer.slots[orderer.view][sn]

	// Ignore preprepare if slot is already committed.
	if orderer.slots[orderer.view][preprepare.Sn].Committed {
		return events.EmptyList()
	}

	// Check whether the received certificate was actually requested (a faulty node might have sent it on its own).
	digests := orderer.segmentCheckpoint.Digests()
	if digests == nil {
		orderer.logger.Log(logging.LevelWarn, "Ignoring unsolicited CatchUpResponse.", "sn", sn)
		return events.EmptyList()
	}

	// Check whether the digest of the received message matches the requested one.
	if !bytes.Equal(digests[sn], digest) {
		orderer.logger.Log(logging.LevelWarn, "Ignoring CatchUpResponse with invalid digest.", "sn", sn)
		return events.EmptyList()
	}

	orderer.logger.Log(logging.LevelDebug, "Catching up with segment-level checkpoint.", "sn", sn)

	// Add the missing certificate, updating the corresponding Preprepare's view.
	// Note that copying a Preprepare with an updated view preserves its hash.
	slot.catchUp(copyPreprepareToNewView(preprepare, orderer.view), digest)

	// If all certificates have been committed (i.e. this is the last certificate to commit),
	// send a Done message to all other nodes.
	// This is required for liveness, see comments for pbftSegmentChkp.
	if orderer.allCommitted() {
		orderer.segmentCheckpoint.SetDone()
		eventsOut.PushBackList(orderer.sendDoneMessages())
	}

	// Deliver certificate.
	eventsOut.PushBack(isspbevents.SBDeliver(
		orderer.moduleConfig.Ord,
		sn,
		slot.Preprepare.Data,
		slot.Preprepare.Aborted,
		orderer.segment.Leader,
		orderer.moduleConfig.Self,
	).Pb())

	return eventsOut
}
