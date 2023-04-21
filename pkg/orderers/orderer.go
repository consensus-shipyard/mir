/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderers

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/events"
	issconfig "github.com/filecoin-project/mir/pkg/iss/config"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/messagebuffer"
	"github.com/filecoin-project/mir/pkg/modules"
	types2 "github.com/filecoin-project/mir/pkg/orderers/types"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	"github.com/filecoin-project/mir/pkg/pb/cryptopb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	eventpbevents "github.com/filecoin-project/mir/pkg/pb/eventpb/events"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	"github.com/filecoin-project/mir/pkg/pb/hasherpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/pb/ordererpb"
	ordererpbtypes "github.com/filecoin-project/mir/pkg/pb/ordererpb/types"
	"github.com/filecoin-project/mir/pkg/pb/pbftpb"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	"github.com/filecoin-project/mir/pkg/pb/transportpb"
	"github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

// ============================================================
// PBFT Orderer type and constructor
// ============================================================

// ValidityChecker is the interface of an external checker of validity of proposed data.
// Each orderer is provided with an object implementing this interface
// and applies its Check method to all received proposals.
type ValidityChecker interface {

	// Check returns nil if the provided proposal data is valid, a non-nil error otherwise.
	Check(data []byte) error
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
	slots map[types2.ViewNr]map[tt.SeqNr]*pbftSlot

	// Tracks the state of the segment-local checkpoint.
	segmentCheckpoint *pbftSegmentChkp

	// Logger for outputting debugging messages.
	logger logging.Logger

	// PBFT view
	view types2.ViewNr

	// Flag indicating whether this node is currently performing a view change.
	// It is set on sending the ViewChange message and cleared on accepting a new view.
	inViewChange bool

	// For each view transition, stores the state of the corresponding PBFT view change sub-protocol.
	// The map itself is allocated on creation of the pbftInstance, but the entries are initialized lazily,
	// only when needed (when the node initiates a view change).
	viewChangeStates map[types2.ViewNr]*pbftViewChangeState

	// Contains the application-specific code for validating incoming proposals.
	externalValidator ValidityChecker
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
	externalValidator ValidityChecker,
	logger logging.Logger) *Orderer {

	// Set all the necessary fields of the new instance and return it.
	return &Orderer{
		ownID:             ownID,
		segment:           segment,
		moduleConfig:      moduleConfig,
		config:            config,
		externalValidator: externalValidator,
		slots:             make(map[types2.ViewNr]map[tt.SeqNr]*pbftSlot),
		segmentCheckpoint: newPbftSegmentChkp(),
		proposal: pbftProposalState{
			proposalsMade:     0,
			certRequested:     false,
			certRequestedView: 0,
			proposalTimeout:   0,
		},
		messageBuffers: messagebuffer.NewBuffers(
			removeNodeID(config.Membership, ownID), // Create a message buffer for everyone except for myself.
			config.MsgBufCapacity,
			//       Even better, share the same buffers with ISS.
			logging.Decorate(logger, "Msgbuf: "),
		),
		logger:           logger,
		view:             0,
		inViewChange:     false,
		viewChangeStates: make(map[types2.ViewNr]*pbftViewChangeState),
	}
}

func newOrdererConfig(issParams *issconfig.ModuleParams, membership []t.NodeID, epochNr tt.EpochNr) *PBFTConfig {

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
	case *eventpb.Event_Crypto:
		switch e := ev.Crypto.Type.(type) {
		case *cryptopb.Event_SignResult:
			return orderer.applySignResult(e.SignResult), nil
		case *cryptopb.Event_SigsVerified:
			return orderer.applyNodeSigsVerified(e.SigsVerified), nil
		default:
			return nil, fmt.Errorf("unknown crypto event type: %T", e)
		}
	case *eventpb.Event_Transport:
		switch e := ev.Transport.Type.(type) {
		case *transportpb.Event_MessageReceived:
			return orderer.applyMessageReceived(e.MessageReceived), nil
		default:
			return nil, fmt.Errorf("unknown transport event type: %T", e)
		}
	case *eventpb.Event_Availability:
		switch avEvent := ev.Availability.Type.(type) {
		case *availabilitypb.Event_NewCert:
			return orderer.applyCertReady(avEvent.NewCert.Cert)
		default:
			return nil, fmt.Errorf("unknown availability event type: %T", avEvent)
		}
	case *eventpb.Event_Hasher:
		switch e := ev.Hasher.Type.(type) {
		case *hasherpb.Event_Result:
			return orderer.applyHashResult(e.Result), nil
		default:
			return nil, fmt.Errorf("unknown Hasher event: %T", e)
		}
	case *eventpb.Event_Orderer:
		switch e := ev.Orderer.Type.(type) {
		case *ordererpb.Event_Pbft:
			switch e := e.Pbft.Type.(type) {
			case *pbftpb.Event_ProposeTimeout:
				return orderer.applyProposeTimeout(int(e.ProposeTimeout))
			case *pbftpb.Event_ViewChangeSnTimeout:
				return orderer.applyViewChangeSNTimeout(e.ViewChangeSnTimeout), nil
			case *pbftpb.Event_ViewChangeSegTimeout:
				return orderer.applyViewChangeSegmentTimeout(types2.ViewNr(e.ViewChangeSegTimeout)), nil
			default:
				return nil, fmt.Errorf("unknown PBFT event type: %T", e)
			}
		default:
			return nil, fmt.Errorf("unknown orderer event type: %T", e)
		}
	default:
		return nil, fmt.Errorf("unknown event type: %T", event.Type)
	}

}

func (orderer *Orderer) ApplyEvents(evts *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsSequentially(evts, orderer.ApplyEvent)
}

func (orderer *Orderer) applyHashResult(result *hasherpb.Result) *events.EventList {
	// Depending on the origin of the hash result, continue processing where the hash was needed.

	switch ev := result.Origin.Type.(type) {
	case *hasherpb.HashOrigin_Sb:
		switch e := ev.Sb.Type.(type) {
		case *ordererpb.HashOrigin_Pbft:
			switch o := e.Pbft.Type.(type) {
			case *pbftpb.HashOrigin_Preprepare:
				return orderer.applyPreprepareHashResult(result.Digests[0], o.Preprepare)
			case *pbftpb.HashOrigin_EmptyPreprepares:
				return orderer.applyEmptyPreprepareHashResult(result.Digests, types2.ViewNr(o.EmptyPreprepares))
			case *pbftpb.HashOrigin_MissingPreprepare:
				return orderer.applyMissingPreprepareHashResult(
					result.Digests[0],
					pbftpbtypes.PreprepareFromPb(o.MissingPreprepare),
				)
			case *pbftpb.HashOrigin_NewView:
				return orderer.applyNewViewHashResult(result.Digests, pbftpbtypes.NewViewFromPb(o.NewView))
			case *pbftpb.HashOrigin_CatchUpResponse:
				return orderer.applyCatchUpResponseHashResult(
					result.Digests[0],
					pbftpbtypes.PreprepareFromPb(o.CatchUpResponse),
				)
			default:
				panic(fmt.Sprintf("unknown hash origin type: %T", o))
			}
		default:
			panic(fmt.Sprintf("unknown hash origin type: %T", e))
		}
	default:
		panic(fmt.Sprintf("unknown hash origin type: %T", ev))
	}
}

func (orderer *Orderer) applySignResult(result *cryptopb.SignResult) *events.EventList {
	// Depending on the origin of the sign result, continue processing where the signature was needed.
	switch or := result.Origin.Type.(type) {
	case *cryptopb.SignOrigin_Sb:
		switch origin := or.Sb.Type.(type) {
		case *ordererpb.SignOrigin_Pbft:
			switch origin := origin.Pbft.Type.(type) {
			case *pbftpb.SignOrigin_ViewChange:
				return orderer.applyViewChangeSignResult(result.Signature, origin.ViewChange)
			default:
				panic(fmt.Sprintf("unknown PBFT sign origin type: %T", origin))
			}
		default:
			panic(fmt.Sprintf("unknown sign origin type: %T", origin))
		}
	default:
		panic(fmt.Sprintf("unknown sign origin type: %T", or))
	}
}

func (orderer *Orderer) applyNodeSigsVerified(result *cryptopb.SigsVerified) *events.EventList {

	ordererOrigin := result.Origin.Type.(*cryptopb.SigVerOrigin_Sb).Sb

	// Ignore events with invalid signatures.
	if !result.AllOk {
		orderer.logger.Log(logging.LevelWarn,
			"Ignoring invalid signature, ignoring event (with all signatures).",
			"from", result.NodeIds,
			"type", fmt.Sprintf("%T", result.Origin.Type),
			"errors", result.Errors,
		)
		return events.EmptyList()
	}

	// Verify all signers are part of the membership
	if !sliceutil.ContainsAll(orderer.config.Membership, t.NodeIDSlice(result.NodeIds)) {
		orderer.logger.Log(logging.LevelWarn,
			"Ignoring message as it contains signatures from non members, ignoring event (with all signatures).",
			"from", result.NodeIds,
			"type", fmt.Sprintf("%T", result.Origin.Type),
		)
		return events.EmptyList()
	}

	// Depending on the origin of the sign result, continue processing where the signature verification was needed.
	switch origin := ordererOrigin.Type.(type) {
	case *ordererpb.SigVerOrigin_Pbft:
		switch origin := origin.Pbft.Type.(type) {
		case *pbftpb.SigVerOrigin_SignedViewChange:
			return orderer.applyVerifiedViewChange(origin.SignedViewChange, t.NodeID(result.NodeIds[0]))
		case *pbftpb.SigVerOrigin_NewView:
			return orderer.applyVerifiedNewView(origin.NewView)
		default:
			panic(fmt.Sprintf("unknown PBFT signature verification origin type: %T", origin))
		}
	default:
		panic(fmt.Sprintf("unknown signature verification origin type: %T", origin))
	}
}

// applyMessageReceived handles a received PBFT protocol message.
func (orderer *Orderer) applyMessageReceived(messageReceived *transportpb.MessageReceived) *events.EventList {

	message := messageReceived.Msg
	from := t.NodeID(messageReceived.From)

	// check if from is part of the membership
	if !sliceutil.Contains(orderer.config.Membership, from) {
		orderer.logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
		return events.EmptyList()
	}

	switch msg := message.Type.(type) {
	case *messagepb.Message_Orderer:
		// Based on the message type, call the appropriate handler method.
		switch msg := ordererpbtypes.MessageFromPb(msg.Orderer).Type.(type) {
		case *ordererpbtypes.Message_Pbft:
			switch msg := msg.Pbft.Type.(type) {
			case *pbftpbtypes.Message_Preprepare:
				return orderer.applyMsgPreprepare(msg.Preprepare, from)
			case *pbftpbtypes.Message_Prepare:
				return orderer.applyMsgPrepare(msg.Prepare, from)
			case *pbftpbtypes.Message_Commit:
				return orderer.applyMsgCommit(msg.Commit, from)
			case *pbftpbtypes.Message_SignedViewChange:
				return orderer.applyMsgSignedViewChange(msg.SignedViewChange, from)
			case *pbftpbtypes.Message_PreprepareRequest:
				return orderer.applyMsgPreprepareRequest(msg.PreprepareRequest, from)
			case *pbftpbtypes.Message_MissingPreprepare:
				return orderer.applyMsgMissingPreprepare(msg.MissingPreprepare.Preprepare, from)
			case *pbftpbtypes.Message_NewView:
				return orderer.applyMsgNewView(msg.NewView, from)
			case *pbftpbtypes.Message_Done:
				return orderer.applyMsgDone(msg.Done, from)
			case *pbftpbtypes.Message_CatchUpRequest:
				return orderer.applyMsgCatchUpRequest(msg.CatchUpRequest, from)
			case *pbftpbtypes.Message_CatchUpResponse:
				return orderer.applyMsgCatchUpResponse(msg.CatchUpResponse.Resp, from)
			default:
				panic(fmt.Sprintf("unknown PBFT message type: %T", message.Type))
			}
		default:
			panic(fmt.Sprintf("unknown orderer message type: %T", message.Type))
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
		orderer.proposal.proposalsMade < orderer.segment.Len() &&

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

	event := eventpbevents.TimerDelay(
		orderer.moduleConfig.Timer,
		[]*eventpbtypes.Event{eventpbtypes.EventFromPb(OrdererEvent(orderer.moduleConfig.Self, PbftProposeTimeout(1)))},
		types.Duration(orderer.config.MaxProposeDelay),
	)

	// Set up timer for the first proposal.
	return eventsOut.PushBack(event.Pb())

}

// numCommitted returns the number of slots that are already committed in the given view.
func (orderer *Orderer) numCommitted(view types2.ViewNr) int {
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

func (orderer *Orderer) initView(view types2.ViewNr) *events.EventList {
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
	orderer.slots[view] = make(map[tt.SeqNr]*pbftSlot)
	for _, sn := range orderer.segment.SeqNrs() {

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
	timerEvents.PushBack(eventpbevents.TimerDelay(
		orderer.moduleConfig.Timer,
		[]*eventpbtypes.Event{eventpbtypes.EventFromPb(OrdererEvent(orderer.moduleConfig.Self,
			PbftViewChangeSNTimeout(view, orderer.numCommitted(view))))},
		computeTimeout(types.Duration(orderer.config.ViewChangeSNTimeout), view),
	).Pb()).PushBack(eventpbevents.TimerDelay(
		orderer.moduleConfig.Timer,
		[]*eventpbtypes.Event{eventpbtypes.EventFromPb(OrdererEvent(orderer.moduleConfig.Self,
			PbftViewChangeSegmentTimeout(view)))},
		computeTimeout(types.Duration(orderer.config.ViewChangeSegmentTimeout), view),
	).Pb())

	orderer.view = view
	orderer.inViewChange = false

	return timerEvents
}

func (orderer *Orderer) lookUpPreprepare(sn tt.SeqNr, digest []byte) *pbftpbtypes.Preprepare {
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
		// If the underlying type of ViewNr is unsigned, view would underflow and we would loop forever.
		if view == 0 {
			return nil
		}
	}
}

// ============================================================
// Auxiliary functions
// ============================================================

// computeTimeout adapts a view change timeout to the view in which it is used.
// This is to implement the doubling of timeouts on every view change.
func computeTimeout(timeout types.Duration, view types2.ViewNr) types.Duration {
	timeout *= 1 << view
	return timeout
}

// removeNodeID removes a node ID from a list of node IDs.
// Takes a membership list and a Node ID and returns a new list of nodeIDs containing all IDs from the membership list,
// except for (if present) the specified nID.
// This is useful for obtaining the list of "other nodes" by removing the own ID from the membership.
func removeNodeID(membership []t.NodeID, nID t.NodeID) []t.NodeID {

	// Allocate the new node list.
	others := make([]t.NodeID, 0, len(membership))

	// Add all membership IDs except for the specified one.
	for _, nodeID := range membership {
		if nodeID != nID {
			others = append(others, nodeID)
		}
	}

	// Return the new list.
	return others
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (orderer *Orderer) ImplementsModule() {}
