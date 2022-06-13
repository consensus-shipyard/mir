/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// TODO: Put the PBFT sub-protocol implementation in a separate package that the iss package imports.
//       When doing that, split the code meaningfully in multiple files.

package iss

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/messagebuffer"
	"github.com/filecoin-project/mir/pkg/pb/isspb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ============================================================
// PBFT orderer type and constructor
// ============================================================

// pbftInstance represents a PBFT orderer.
// It implements the sbInstance (instance of Sequenced broadcast) interface and thus can be used as an orderer for ISS.
type pbftInstance struct {

	// The ID of this node.
	ownID t.NodeID

	// PBFT-specific configuration parameters (e.g. view change timeout, etc.)
	config *PBFTConfig

	// The segment governing this SB instance, specifying the leader, the set of sequence numbers, the buckets, etc.
	segment *segment

	// Buffers representing a backlog of messages destined to future views.
	// A node that already transitioned to a newer view might send messages,
	// while this node is behind (still in an older view) and cannot process these messages yet.
	// Such messages end up in this buffer (if there is buffer space) for later processing.
	// The buffer is checked after each view change.
	messageBuffers map[t.NodeID]*messagebuffer.MessageBuffer

	// Tracks the state related to proposing batches.
	proposal pbftProposalState

	// For each view, slots contains one pbftSlot per sequence number this orderer is responsible for.
	// Each slot tracks the state of the agreement protocol for one sequence number.
	slots map[t.PBFTViewNr]map[t.SeqNr]*pbftSlot

	// Logger for outputting debugging messages.
	logger logging.Logger

	// ISS-provided event creator object.
	// All events produced by this pbftInstance must be created exclusively using the methods of eventService.
	// This ensures that the events are associated with this particular pbftInstance within the ISS protocol.
	eventService *sbEventService

	// PBFT view
	view t.PBFTViewNr

	// Flag indicating whether this node is currently performing a view change.
	// It is set on sending the ViewChange message and cleared on accepting a new view.
	inViewChange bool

	// For each view transition, stores the state of the corresponding PBFT view change sub-protocol.
	// The map itself is allocated on creation of the pbftInstance, but the entries are initialized lazily,
	// only when needed (when the node initiates a view change).
	viewChangeStates map[t.PBFTViewNr]*pbftViewChangeState
}

// newPbftInstance allocates and initializes a new instance of the PBFT orderer.
// It takes the following parameters:
// - ownID:              The ID of this node.
// - segment:            The segment governing this SB instance,
//                       specifying the leader, the set of sequence numbers, the buckets, etc.
// - numPendingRequests: The number of requests currently pending in the buckets
//                       assigned to the new instance (segment.BucketIDs) and ready to be proposed by this PBFT orderer.
//                       This is required for the orderer to know whether it make proposals right away.
// - config:             PBFT-specific configuration parameters.
// - eventService:       Event creator object enabling the orderer to produce events.
//                       All events this orderer creates will be created using the methods of the eventService.
//                       The eventService must be configured to produce events associated with this PBFT orderer,
//                       since the implementation of the orderer does not know its own identity at the level of ISS.
// - logger:             Logger for outputting debugging messages.
func newPbftInstance(
	ownID t.NodeID,
	segment *segment,
	numPendingRequests t.NumRequests,
	config *PBFTConfig,
	eventService *sbEventService,
	logger logging.Logger) *pbftInstance {

	// Set all the necessary fields of the new instance and return it.
	return &pbftInstance{
		ownID:   ownID,
		segment: segment,
		config:  config,
		slots:   make(map[t.PBFTViewNr]map[t.SeqNr]*pbftSlot),
		proposal: pbftProposalState{
			proposalsMade:      0,
			numPendingRequests: numPendingRequests,
			batchRequested:     false,
			batchRequestedView: 0,
			proposalTimeout:    0,
		},
		messageBuffers: messagebuffer.NewBuffers(
			removeNodeID(config.Membership, ownID), // Create a message buffer for everyone except for myself.
			config.MsgBufCapacity,                  // TODO: Configure this separately for ISS buffers and PBFT buffers.
			//       Even better, share the same buffers with ISS.
			logging.Decorate(logger, "Msgbuf: "),
		),
		logger:           logger,
		eventService:     eventService,
		view:             0,
		inViewChange:     false,
		viewChangeStates: make(map[t.PBFTViewNr]*pbftViewChangeState),
	}
}

// ============================================================
// SB Instance Interface implementation and event dispatching
// ============================================================

// ApplyEvent receives one event and applies it to the PBFT orderer state machine, potentially altering its state
// and producing a (potentially empty) list of more events.
func (pbft *pbftInstance) ApplyEvent(event *isspb.SBInstanceEvent) *events.EventList {
	switch e := event.Type.(type) {

	case *isspb.SBInstanceEvent_Init:
		return pbft.applyInit()
	case *isspb.SBInstanceEvent_PbftProposeTimeout:
		return pbft.applyProposeTimeout(int(e.PbftProposeTimeout))
	case *isspb.SBInstanceEvent_PbftViewChangeBatchTimeout:
		return pbft.applyViewChangeBatchTimeout(e.PbftViewChangeBatchTimeout)
	case *isspb.SBInstanceEvent_PbftViewChangeSegTimeout:
		return pbft.applyViewChangeSegmentTimeout(t.PBFTViewNr(e.PbftViewChangeSegTimeout))
	case *isspb.SBInstanceEvent_PendingRequests:
		return pbft.applyPendingRequests(t.NumRequests(e.PendingRequests.NumRequests))
	case *isspb.SBInstanceEvent_BatchReady:
		return pbft.applyBatchReady(e.BatchReady)
	case *isspb.SBInstanceEvent_RequestsReady:
		return pbft.applyRequestsReady(e.RequestsReady)
	case *isspb.SBInstanceEvent_HashResult:
		return pbft.applyHashResult(e.HashResult)
	case *isspb.SBInstanceEvent_SignResult:
		return pbft.applySignResult(e.SignResult)
	case *isspb.SBInstanceEvent_NodeSigsVerified:
		return pbft.applyNodeSigsVerified(e.NodeSigsVerified)
	case *isspb.SBInstanceEvent_PbftPersistPreprepare:
		return pbft.applyPbftPersistPreprepare(e.PbftPersistPreprepare)
	case *isspb.SBInstanceEvent_MessageReceived:
		return pbft.applyMessageReceived(e.MessageReceived.Msg, t.NodeID(e.MessageReceived.From))
	case *isspb.SBInstanceEvent_PbftPersistPrepare,
		*isspb.SBInstanceEvent_PbftPersistCommit,
		*isspb.SBInstanceEvent_PbftPersistSignedViewChange,
		*isspb.SBInstanceEvent_PbftPersistNewView:
		// TODO: Ignoring WAL loading for the moment.
		return &events.EventList{}
	default:
		// Panic if message type is not known.
		panic(fmt.Sprintf("unknown PBFT SB instance event type: %T", event.Type))
	}
}

func (pbft *pbftInstance) applyHashResult(result *isspb.SBHashResult) *events.EventList {
	// Depending on the origin of the hash result, continue processing where the hash was needed.
	switch origin := result.Origin.Type.(type) {
	case *isspb.SBInstanceHashOrigin_PbftPreprepare:
		return pbft.applyPreprepareHashResult(result.Digests[0], origin.PbftPreprepare)
	case *isspb.SBInstanceHashOrigin_PbftEmptyPreprepares:
		return pbft.applyEmptyPreprepareHashResult(result.Digests, t.PBFTViewNr(origin.PbftEmptyPreprepares))
	case *isspb.SBInstanceHashOrigin_PbftMissingPreprepare:
		return pbft.applyMissingPreprepareHashResult(result.Digests[0], origin.PbftMissingPreprepare)
	case *isspb.SBInstanceHashOrigin_PbftNewView:
		return pbft.applyNewViewHashResult(result.Digests, origin.PbftNewView)
	default:
		panic(fmt.Sprintf("unknown hash origin type: %T", origin))
	}
}

func (pbft *pbftInstance) applySignResult(result *isspb.SBSignResult) *events.EventList {
	// Depending on the origin of the sign result, continue processing where the signature was needed.
	switch origin := result.Origin.Type.(type) {
	case *isspb.SBInstanceSignOrigin_PbftViewChange:
		return pbft.applyViewChangeSignResult(result.Signature, origin.PbftViewChange)
	default:
		panic(fmt.Sprintf("unknown sign origin type: %T", origin))
	}
}

func (pbft *pbftInstance) applyNodeSigsVerified(result *isspb.SBNodeSigsVerified) *events.EventList {

	// Ignore events with invalid signatures.
	if !result.AllOk {
		pbft.logger.Log(logging.LevelWarn,
			"Ignoring invalid signature, ignoring event (with all signatures).",
			"from", result.NodeIds,
			"type", fmt.Sprintf("%T", result.Origin.Type),
			"errors", result.Errors,
		)
	}

	// Depending on the origin of the sign result, continue processing where the signature verification was needed.
	switch origin := result.Origin.Type.(type) {
	case *isspb.SBInstanceSigVerOrigin_PbftSignedViewChange:
		return pbft.applyVerifiedViewChange(origin.PbftSignedViewChange, t.NodeID(result.NodeIds[0]))
	case *isspb.SBInstanceSigVerOrigin_PbftNewView:
		return pbft.applyVerifiedNewView(origin.PbftNewView)
	default:
		panic(fmt.Sprintf("unknown signature verification origin type: %T", origin))
	}
}

// applyMessageReceived handles a received PBFT protocol message.
func (pbft *pbftInstance) applyMessageReceived(message *isspb.SBInstanceMessage, from t.NodeID) *events.EventList {

	// Based on the message type, call the appropriate handler method.
	switch msg := message.Type.(type) {
	case *isspb.SBInstanceMessage_PbftPreprepare:
		return pbft.applyMsgPreprepare(msg.PbftPreprepare, from)
	case *isspb.SBInstanceMessage_PbftPrepare:
		return pbft.applyMsgPrepare(msg.PbftPrepare, from)
	case *isspb.SBInstanceMessage_PbftCommit:
		return pbft.applyMsgCommit(msg.PbftCommit, from)
	case *isspb.SBInstanceMessage_PbftSignedViewChange:
		return pbft.applyMsgSignedViewChange(msg.PbftSignedViewChange, from)
	case *isspb.SBInstanceMessage_PbftPreprepareRequest:
		return pbft.applyMsgPreprepareRequest(msg.PbftPreprepareRequest, from)
	case *isspb.SBInstanceMessage_PbftMissingPreprepare:
		return pbft.applyMsgMissingPreprepare(msg.PbftMissingPreprepare, from)
	case *isspb.SBInstanceMessage_PbftNewView:
		return pbft.applyMsgNewView(msg.PbftNewView, from)
	default:
		panic(fmt.Sprintf("unknown ISS PBFT message type: %T", message.Type))
	}
}

// Segment returns the segment associated with this orderer.
func (pbft *pbftInstance) Segment() *segment {
	return pbft.segment
}

// Status returns a protobuf representation of the current state of the orderer that can be later printed.
// This functionality is meant mostly for debugging and is *not* meant to provide an interface for
// serializing and deserializing the whole protocol state.
func (pbft *pbftInstance) Status() *isspb.SBStatus {
	// TODO: Return actual status here, not just a stub.
	return &isspb.SBStatus{Leader: pbft.segment.Leader.Pb()}
}

// ============================================================
// General protocol logic (other specific parts in separate files)
// ============================================================

// canPropose returns true if the current state of the PBFT orderer allows for a new batch to be proposed.
// Note that "new batch" means a "fresh" batch proposed during normal operation outside of view change.
// Proposals part of a new view message during a view change do not call this function and are treated separately.
func (pbft *pbftInstance) canPropose() bool {
	return pbft.ownID == pbft.segment.Leader && // Only the leader can propose

		// No regular proposals can be made after a view change.
		// This is specific for the SB-version of PBFT used in ISS and deviates from the standard PBFT protocol.
		pbft.view == 0 &&

		// A new batch must not have been requested (if it has, we are already in the process of proposing).
		!pbft.proposal.batchRequested &&

		// There must still be a free sequence number for which a proposal can be made.
		pbft.proposal.proposalsMade < len(pbft.segment.SeqNrs) &&

		// Either the batch timeout must have passed, or there must be enough requests for a full batch.
		// The value 0 for config.MaxBatchSize means no limit on batch size,
		// i.e., a proposal cannot be triggered just by the number of pending requests.
		(pbft.proposal.proposalTimeout > pbft.proposal.proposalsMade ||
			(pbft.config.MaxBatchSize != 0 && pbft.proposal.numPendingRequests >= pbft.config.MaxBatchSize)) &&

		// No proposals can be made while in view change.
		!pbft.inViewChange
}

// applyInit takes all the actions resulting from the PBFT orderer's initial state.
// The Init event is expected to be the first event applied to the orderer,
// except for events read from the WAL at startup, which are expected to be applied even before the Init event.
func (pbft *pbftInstance) applyInit() *events.EventList {

	eventsOut := &events.EventList{}

	// Initialize the first PBFT view
	eventsOut.PushBackList(pbft.initView(0))

	// Set up timer for the first proposal.
	return eventsOut.PushBack(pbft.eventService.TimerDelay(
		t.TimeDuration(pbft.config.MaxProposeDelay),
		pbft.eventService.SBEvent(PbftProposeTimeout(1)),
	))

}

// numCommitted returns the number of slots that are already committed in the given view.
func (pbft *pbftInstance) numCommitted(view t.PBFTViewNr) int {
	numCommitted := 0
	for _, slot := range pbft.slots[view] {
		if slot.Committed {
			numCommitted++
		}
	}
	return numCommitted
}

// allCommitted returns true if all slots of this pbftInstance in the current view are in the committed state
// (i.e., have the committed flag set).
func (pbft *pbftInstance) allCommitted() bool {
	return pbft.numCommitted(pbft.view) == len(pbft.slots[pbft.view])
}

// applyPendingRequests processes a notification form ISS about the number of requests in buckets ready to be proposed.
func (pbft *pbftInstance) applyPendingRequests(numRequests t.NumRequests) *events.EventList {

	// Update the orderer's view on the number of pending requests.
	pbft.proposal.numPendingRequests = numRequests

	if pbft.canPropose() {
		// Start a new proposal if applicable (i.e. if the number of pending requests reached config.MaxBatchSize).
		return pbft.requestNewBatch()
	}

	return &events.EventList{}
}

func (pbft *pbftInstance) initView(view t.PBFTViewNr) *events.EventList {
	// Sanity check
	if view < pbft.view {
		panic(fmt.Sprintf("Starting a view (%d) older than the current one (%d)", view, pbft.view))
	}

	// Do not start the same view more than once.
	// View 0 is also started only once (the code makes sure that startView(0) is only called at initialization),
	// it's just that the default value of the variable is already 0 - that's why it needs an exception.
	if view != 0 && view == pbft.view {
		return &events.EventList{}
	}

	pbft.logger.Log(logging.LevelInfo, "Initializing new view.", "view", view)

	// Initialize PBFT slots for the new view, one for each sequence number.
	pbft.slots[view] = make(map[t.SeqNr]*pbftSlot)
	for _, sn := range pbft.segment.SeqNrs {

		// Create a fresh, empty slot.
		// For n being the membership size, f = (n-1) / 3
		pbft.slots[view][sn] = newPbftSlot((len(pbft.segment.Membership) - 1) / 3)

		// Except for initialization of view 0, carry over state from the previous view.
		if view > 0 {
			pbft.slots[view][sn].populateFromPrevious(pbft.slots[pbft.view][sn], view)
		}
	}

	// Set view change timeouts
	timerEvents := &events.EventList{}
	timerEvents.PushBack(pbft.eventService.TimerDelay(
		computeTimeout(t.TimeDuration(pbft.config.ViewChangeBatchTimeout), view),
		pbft.eventService.SBEvent(PbftViewChangeBatchTimeout(view, pbft.numCommitted(view)))),
	).PushBack(pbft.eventService.TimerDelay(
		computeTimeout(t.TimeDuration(pbft.config.ViewChangeSegmentTimeout), view),
		pbft.eventService.SBEvent(PbftViewChangeSegmentTimeout(view))),
	)

	pbft.view = view
	pbft.inViewChange = false

	return timerEvents
}

// ============================================================
// Auxiliary functions
// ============================================================

func primaryNode(seg *segment, view t.PBFTViewNr) t.NodeID {
	return seg.Membership[(leaderIndex(seg)+int(view))%len(seg.Membership)]
}

func leaderIndex(seg *segment) int {
	for i, nodeID := range seg.Membership {
		if nodeID == seg.Leader {
			return i
		}
	}
	panic("invalid segment: leader not in membership")
}

// computeTimeout adapts a view change timeout to the view in which it is used.
// This is to implement the doubling of timeouts on every view change.
func computeTimeout(timeout t.TimeDuration, view t.PBFTViewNr) t.TimeDuration {
	for view > 0 {
		timeout *= 2
		view--
	}
	return timeout
}
