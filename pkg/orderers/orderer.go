/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderers

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/util/issutil"

	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/ordererspb"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/messagebuffer"
	"github.com/filecoin-project/mir/pkg/pb/ordererspbftpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ============================================================
// PBFT Orderer type and constructor
// ============================================================

// The Segment type represents an ISS Segment.
// It is used to parametrize an orderer (i.e. the SB instance).
type Segment struct {

	// The leader node of the orderer.
	Leader t.NodeID

	// List of all nodes executing the orderer implementation.
	Membership []t.NodeID

	// List of sequence numbers for which the orderer is responsible.
	// This is the actual "segment" of the commit log.
	SeqNrs []t.SeqNr
}

// Orderer represents a PBFT Orderer.
// It implements the sbInstance (instance of Sequenced broadcast) interface and thus can be used as an Orderer for ISS.
type Orderer struct {

	// The ID of this node.
	ownID t.NodeID

	// PBFT-specific configuration parameters (e.g. view change timeout, etc.)
	config *PBFTConfig

	// IDs of modules the checkpoint tracker interacts with.
	moduleConfig *ModuleConfig

	// The segment governing this SB instance, specifying the leader, the set of sequence numbers, the buckets, etc.
	segment *Segment

	// Buffers representing a backlog of messages destined to future views.
	// A node that already transitioned to a newer view might send messages,
	// while this node is behind (still in an older view) and cannot process these messages yet.
	// Such messages end up in this buffer (if there is buffer space) for later processing.
	// The buffer is checked after each view change.
	messageBuffers map[t.NodeID]*messagebuffer.MessageBuffer

	// Tracks the state related to proposing availability certificates.
	proposal pbftProposalState

	// For each view, slots contains one pbftSlot per sequence number this Orderer is responsible for.
	// Each slot tracks the state of the agreement protocol for one sequence number.
	slots map[t.PBFTViewNr]map[t.SeqNr]*pbftSlot

	// Tracks the state of the segment-local checkpoint.
	segmentCheckpoint *pbftSegmentChkp

	// Logger for outputting debugging messages.
	logger logging.Logger

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

// NewOrdererModule allocates and initializes a new instance of the PBFT Orderer.
// It takes the following parameters:
//   - moduleConfig
//   - ownID: The ID of this node.
//   - segment: The segment governing this SB instance,
//     specifying the leader, the set of sequence numbers, the buckets, etc.
//   - config: PBFT-specific configuration parameters.
//   - eventService: OrdererEvent creator object enabling the Orderer to produce events.
//     All events this Orderer creates will be created using the methods of the eventService.
//     since the implementation of the Orderer does not know its own identity at the level of ISS.
//   - logger: Logger for outputting debugging messages.
func NewOrdererModule(
	moduleConfig *ModuleConfig,
	ownID t.NodeID,
	segment *Segment,
	config *PBFTConfig,
	logger logging.Logger) *Orderer {

	// Set all the necessary fields of the new instance and return it.
	return &Orderer{
		ownID:             ownID,
		segment:           segment,
		moduleConfig:      moduleConfig,
		config:            config,
		slots:             make(map[t.PBFTViewNr]map[t.SeqNr]*pbftSlot),
		segmentCheckpoint: newPbftSegmentChkp(),
		proposal: pbftProposalState{
			proposalsMade:     0,
			certRequested:     false,
			certRequestedView: 0,
			proposalTimeout:   0,
		},
		messageBuffers: messagebuffer.NewBuffers(
			issutil.RemoveNodeID(config.Membership, ownID), // Create a message buffer for everyone except for myself.
			config.MsgBufCapacity,
			//       Even better, share the same buffers with ISS.
			logging.Decorate(logger, "Msgbuf: "),
		),
		logger:           logger,
		view:             0,
		inViewChange:     false,
		viewChangeStates: make(map[t.PBFTViewNr]*pbftViewChangeState),
	}
}

func newOrdererConfig(issParams *issutil.ModuleParams, membership []t.NodeID, epochNr t.EpochNr) *PBFTConfig {

	// Return a new PBFT configuration with selected values from the ISS configuration.
	return &PBFTConfig{
		Membership:               membership,
		MaxProposeDelay:          issParams.MaxProposeDelay,
		MsgBufCapacity:           issParams.MsgBufCapacity,
		DoneResendPeriod:         issParams.PBFTDoneResendPeriod,
		CatchUpDelay:             issParams.PBFTCatchUpDelay,
		ViewChangeSNTimeout:      issParams.PBFTViewChangeSNTimeout,
		ViewChangeSegmentTimeout: issParams.PBFTViewChangeSegmentTimeout,
		ViewChangeResendPeriod:   issParams.PBFTViewChangeResendPeriod,
		epochNr:                  epochNr,
	}
}

// ApplyEvent receives one event and applies it to the PBFT Orderer state machine, potentially altering its state
// and producing a (potentially empty) list of more events.
func (orderer *Orderer) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch ev := event.Type.(type) {
	case *eventpb.Event_Init:
		return orderer.applyInit(), nil
	case *eventpb.Event_SignResult: // TODO Remove SBInstanceEvent_SignResult
		return orderer.applySignResult(ev.SignResult), nil
	case *eventpb.Event_MessageReceived:
		return orderer.applyMessageReceived(ev.MessageReceived), nil
	case *eventpb.Event_NodeSigsVerified:
		return orderer.applyNodeSigsVerified(ev.NodeSigsVerified), nil
	case *eventpb.Event_Availability:
		switch avEvent := ev.Availability.Type.(type) {
		case *availabilitypb.Event_NewCert:
			return orderer.applyCertReady(avEvent.NewCert.Cert)
		default:
			return nil, fmt.Errorf("unknown availability event type: %T", avEvent)
		}
	case *eventpb.Event_HashResult:
		return orderer.applyHashResult(ev), nil
	case *eventpb.Event_SbEvent:
		switch e := ev.SbEvent.Type.(type) {
		case *ordererspb.SBInstanceEvent_Init:
			return orderer.applyInit(), nil
		case *ordererspb.SBInstanceEvent_PbftProposeTimeout:
			return orderer.applyProposeTimeout(int(e.PbftProposeTimeout)), nil
		case *ordererspb.SBInstanceEvent_PbftViewChangeSnTimeout:
			return orderer.applyViewChangeSNTimeout(e.PbftViewChangeSnTimeout), nil
		case *ordererspb.SBInstanceEvent_PbftViewChangeSegTimeout:
			return orderer.applyViewChangeSegmentTimeout(t.PBFTViewNr(e.PbftViewChangeSegTimeout)), nil
		default:
			return nil, fmt.Errorf("unknown PBFT SB instance event type: %T", ev.SbEvent.Type)
		}
	default:
		return nil, fmt.Errorf("unknown Orderer event type: %T", event.Type)
	}

}

func (orderer *Orderer) ApplyEvents(evts *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsSequentially(evts, orderer.ApplyEvent)
}

func (orderer *Orderer) applyHashResult(result *eventpb.Event_HashResult) *events.EventList {
	// Depending on the origin of the hash result, continue processing where the hash was needed.

	origin := result.HashResult
	switch ev := origin.Origin.Type.(type) {
	case *eventpb.HashOrigin_Sb:
		switch e := ev.Sb.Type.(type) {
		case *ordererspb.SBInstanceHashOrigin_PbftPreprepare:
			return orderer.applyPreprepareHashResult(origin.Digests[0], e.PbftPreprepare)
		case *ordererspb.SBInstanceHashOrigin_PbftEmptyPreprepares:
			return orderer.applyEmptyPreprepareHashResult(origin.Digests, t.PBFTViewNr(e.PbftEmptyPreprepares))
		case *ordererspb.SBInstanceHashOrigin_PbftMissingPreprepare:
			return orderer.applyMissingPreprepareHashResult(origin.Digests[0], e.PbftMissingPreprepare)
		case *ordererspb.SBInstanceHashOrigin_PbftNewView:
			return orderer.applyNewViewHashResult(origin.Digests, e.PbftNewView)
		case *ordererspb.SBInstanceHashOrigin_PbftCatchUpResponse:
			return orderer.applyCatchUpResponseHashResult(origin.Digests[0], e.PbftCatchUpResponse)
		default:
			panic(fmt.Sprintf("unknown hash origin type: %T", e))
		}
	default:
		panic(fmt.Sprintf("unknown hash origin type: %T", ev))
	}
}

func (orderer *Orderer) applySignResult(result *eventpb.SignResult) *events.EventList {
	// Depending on the origin of the sign result, continue processing where the signature was needed.
	switch or := result.Origin.Type.(type) {
	case *eventpb.SignOrigin_Sb:
		switch origin := or.Sb.Type.(type) {
		case *ordererspb.SBInstanceSignOrigin_PbftViewChange:
			return orderer.applyViewChangeSignResult(result.Signature, origin.PbftViewChange)
		default:
			panic(fmt.Sprintf("unknown sign origin type: %T", origin))
		}
	default:
		panic(fmt.Sprintf("unknown sign origin type: %T", or))
	}
}

func (orderer *Orderer) applyNodeSigsVerified(result *eventpb.NodeSigsVerified) *events.EventList {

	ordererOrigin := result.Origin.Type.(*eventpb.SigVerOrigin_Sb).Sb

	// Ignore events with invalid signatures.
	if !result.AllOk {
		orderer.logger.Log(logging.LevelWarn,
			"Ignoring invalid signature, ignoring event (with all signatures).",
			"from", result.NodeIds,
			"type", fmt.Sprintf("%T", result.Origin.Type),
			"errors", result.Errors,
		)
	}

	// Depending on the origin of the sign result, continue processing where the signature verification was needed.
	switch origin := ordererOrigin.Type.(type) {
	case *ordererspb.SBInstanceSigVerOrigin_PbftSignedViewChange:
		return orderer.applyVerifiedViewChange(origin.PbftSignedViewChange, t.NodeID(result.NodeIds[0]))
	case *ordererspb.SBInstanceSigVerOrigin_PbftNewView:
		return orderer.applyVerifiedNewView(origin.PbftNewView)
	default:
		panic(fmt.Sprintf("unknown signature verification origin type: %T", origin))
	}
}

// applyMessageReceived handles a received PBFT protocol message.
func (orderer *Orderer) applyMessageReceived(messageReceived *eventpb.MessageReceived) *events.EventList {

	message := messageReceived.Msg
	from := t.NodeID(messageReceived.From)
	switch msg := message.Type.(type) {
	case *messagepb.Message_SbMessage:
		// Based on the message type, call the appropriate handler method.
		switch msg := msg.SbMessage.Type.(type) {
		case *ordererspb.SBInstanceMessage_PbftPreprepare:
			return orderer.applyMsgPreprepare(msg.PbftPreprepare, from)
		case *ordererspb.SBInstanceMessage_PbftPrepare:
			return orderer.applyMsgPrepare(msg.PbftPrepare, from)
		case *ordererspb.SBInstanceMessage_PbftCommit:
			return orderer.applyMsgCommit(msg.PbftCommit, from)
		case *ordererspb.SBInstanceMessage_PbftSignedViewChange:
			return orderer.applyMsgSignedViewChange(msg.PbftSignedViewChange, from)
		case *ordererspb.SBInstanceMessage_PbftPreprepareRequest:
			return orderer.applyMsgPreprepareRequest(msg.PbftPreprepareRequest, from)
		case *ordererspb.SBInstanceMessage_PbftMissingPreprepare:
			return orderer.applyMsgMissingPreprepare(msg.PbftMissingPreprepare, from)
		case *ordererspb.SBInstanceMessage_PbftNewView:
			return orderer.applyMsgNewView(msg.PbftNewView, from)
		case *ordererspb.SBInstanceMessage_PbftDone:
			return orderer.applyMsgDone(msg.PbftDone, from)
		case *ordererspb.SBInstanceMessage_PbftCatchUpRequest:
			return orderer.applyMsgCatchUpRequest(msg.PbftCatchUpRequest, from)
		case *ordererspb.SBInstanceMessage_PbftCatchUpResponse:
			return orderer.applyMsgCatchUpResponse(msg.PbftCatchUpResponse, from)
		default:
			panic(fmt.Sprintf("unknown ISS PBFT SB message type: %T", message.Type))
		}
	default:
		panic(fmt.Sprintf("unknown ISS PBFT message type: %T", message.Type))
	}

}

// Segment returns the segment associated with this Orderer.
func (orderer *Orderer) Segment() *Segment {
	return orderer.segment
}

// ============================================================
// General protocol logic (other specific parts in separate files)
// ============================================================

// canPropose returns true if the current state of the PBFT Orderer
// allows for a new availability certificate to be proposed.
// Note that "new availability certificate" means a "fresh" certificate
// proposed during normal operation outside of view change.
// Proposals part of a new view message during a view change do not call this function and are treated separately.
func (orderer *Orderer) canPropose() bool {
	return orderer.ownID == orderer.segment.Leader && // Only the leader can propose

		// No regular proposals can be made after a view change.
		// This is specific for the SB-version of PBFT used in ISS and deviates from the standard PBFT protocol.
		orderer.view == 0 &&

		// A new certificate must not have been requested (if it has, we are already in the process of proposing).
		!orderer.proposal.certRequested &&

		// There must still be a free sequence number for which a proposal can be made.
		orderer.proposal.proposalsMade < len(orderer.segment.SeqNrs) &&

		// The proposal timeout must have passed.
		orderer.proposal.proposalTimeout > orderer.proposal.proposalsMade &&

		// No proposals can be made while in view change.
		!orderer.inViewChange
}

// applyInit takes all the actions resulting from the PBFT Orderer's initial state.
// The Init event is expected to be the first event applied to the Orderer,
// except for events read from the WAL at startup, which are expected to be applied even before the Init event.
// (At this time, the WAL is not used. TODO: Update this when wal is implemented.)
func (orderer *Orderer) applyInit() *events.EventList {

	eventsOut := events.EmptyList()

	// Initialize the first PBFT view
	eventsOut.PushBackList(orderer.initView(0))

	event := events.TimerDelay(
		orderer.moduleConfig.Timer,
		[]*eventpb.Event{OrdererEvent(orderer.moduleConfig.Self,
			PbftProposeTimeout(1))},
		t.TimeDuration(orderer.config.MaxProposeDelay),
	)

	// Set up timer for the first proposal.
	return eventsOut.PushBack(event)

}

// numCommitted returns the number of slots that are already committed in the given view.
func (orderer *Orderer) numCommitted(view t.PBFTViewNr) int {
	numCommitted := 0
	for _, slot := range orderer.slots[view] {
		if slot.Committed {
			numCommitted++
		}
	}
	return numCommitted
}

// allCommitted returns true if all slots of this pbftInstance in the current view are in the committed state
// (i.e., have the committed flag set).
func (orderer *Orderer) allCommitted() bool {
	return orderer.numCommitted(orderer.view) == len(orderer.slots[orderer.view])
}

func (orderer *Orderer) initView(view t.PBFTViewNr) *events.EventList {
	// Sanity check
	if view < orderer.view {
		panic(fmt.Sprintf("Starting a view (%d) older than the current one (%d)", view, orderer.view))
	}

	// Do not start the same view more than once.
	// View 0 is also started only once (the code makes sure that startView(0) is only called at initialization),
	// it's just that the default value of the variable is already 0 - that's why it needs an exception.
	if view != 0 && view == orderer.view {
		return events.EmptyList()
	}

	orderer.logger.Log(logging.LevelInfo, "Initializing new view.", "view", view)

	// Initialize PBFT slots for the new view, one for each sequence number.
	orderer.slots[view] = make(map[t.SeqNr]*pbftSlot)
	for _, sn := range orderer.segment.SeqNrs {

		// Create a fresh, empty slot.
		// For n being the membership size, f = (n-1) / 3
		orderer.slots[view][sn] = newPbftSlot(len(orderer.segment.Membership))

		// Except for initialization of view 0, carry over state from the previous view.
		if view > 0 {
			orderer.slots[view][sn].populateFromPrevious(orderer.slots[orderer.view][sn], view)
		}
	}

	// Set view change timeouts
	timerEvents := events.EmptyList()
	timerEvents.PushBack(events.TimerDelay(
		orderer.moduleConfig.Timer,
		[]*eventpb.Event{OrdererEvent(orderer.moduleConfig.Self,
			PbftViewChangeSNTimeout(view, orderer.numCommitted(view)))},
		computeTimeout(t.TimeDuration(orderer.config.ViewChangeSNTimeout), view),
	)).PushBack(events.TimerDelay(
		orderer.moduleConfig.Timer,
		[]*eventpb.Event{OrdererEvent(orderer.moduleConfig.Self,
			PbftViewChangeSegmentTimeout(view))},
		computeTimeout(t.TimeDuration(orderer.config.ViewChangeSegmentTimeout), view),
	))

	orderer.view = view
	orderer.inViewChange = false

	return timerEvents
}

func (orderer *Orderer) lookUpPreprepare(sn t.SeqNr, digest []byte) *ordererspbftpb.Preprepare {
	// Traverse all views, starting in the current view.
	for view := orderer.view; ; view-- {

		// If the view exists (i.e. if this node ever entered the view)
		if slots, ok := orderer.slots[view]; ok {

			// If the given sequence number is part of the segment (might not be, in case of a corrupted message)
			if slot, ok := slots[sn]; ok {

				// If the slot contains a matching Preprepare
				if preprepare := slot.getPreprepare(digest); preprepare != nil {

					// Preprepare found, return it.
					return preprepare
				}
			}
		}

		// This check cannot be replaced by a (view >= 0) condition in the loop header and must appear here.
		// If the underlying type of t.PBFTViewNr is unsigned, view would underflow and we would loop forever.
		if view == 0 {
			return nil
		}
	}
}

// ============================================================
// Auxiliary functions
// ============================================================

func primaryNode(seg *Segment, view t.PBFTViewNr) t.NodeID {
	return seg.Membership[(leaderIndex(seg)+int(view))%len(seg.Membership)]
}

func leaderIndex(seg *Segment) int {
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

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (orderer *Orderer) ImplementsModule() {}

func strongQuorum(n int) int {
	// assuming n > 3f:
	//   return min q: 2q > n+f
	f := maxFaulty(n)
	return (n+f)/2 + 1
}

func weakQuorum(n int) int {
	// assuming n > 3f:
	//   return min q: q > f
	return maxFaulty(n) + 1
}

func maxFaulty(n int) int {
	// assuming n > 3f:
	//   return max f
	return (n - 1) / 3
}
