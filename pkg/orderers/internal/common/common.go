package common

import (
	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/messagebuffer"
	"github.com/filecoin-project/mir/pkg/orderers/common"
	ot "github.com/filecoin-project/mir/pkg/orderers/types"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	pbftpbevents "github.com/filecoin-project/mir/pkg/pb/pbftpb/events"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	timertypes "github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/types"
)

type ModuleParams struct {
	// The ID of this node.
	OwnID types.NodeID

	// PBFT-specific configuration parameters (e.g. view change timeout, etc.)
	Config *PBFTConfig
}

// State represents a PBFT State.
// It implements the sbInstance (instance of Sequenced broadcast) interface and thus can be used as an State for ISS.
type State struct {

	// The segment governing this SB instance, specifying the leader, the set of sequence numbers, etc.
	Segment *common.Segment

	// Buffers representing a backlog of messages destined to future views.
	// A node that already transitioned to a newer view might send messages,
	// while this node is behind (still in an older view) and cannot process these messages yet.
	// Such messages end up in this buffer (if there is buffer space) for later processing.
	// The buffer is checked after each view change.
	MessageBuffers map[types.NodeID]*messagebuffer.MessageBuffer

	// Tracks the state related to proposing availability certificates.
	Proposal PbftProposalState

	// For each view, slots contains one PbftSlot per sequence number this State is responsible for.
	// Each slot tracks the state of the agreement protocol for one sequence number.
	Slots map[ot.ViewNr]map[tt.SeqNr]*PbftSlot

	// Tracks the state of the segment-local checkpoint.
	SegmentCheckpoint *PbftSegmentChkp

	// PBFT view
	View ot.ViewNr

	// Flag indicating whether this node is currently performing a view change.
	// It is set on sending the ViewChange message and cleared on accepting a new view.
	InViewChange bool

	// For each view transition, stores the state of the corresponding PBFT view change sub-protocol.
	// The map itself is allocated on creation of the pbftInstance, but the entries are initialized lazily,
	// only when needed (when the node initiates a view change).
	ViewChangeStates map[ot.ViewNr]*PbftViewChangeState
}

// PbftProposalState tracks the state of the pbftInstance related to proposing certificates.
// The proposal state is only used if this node is the leader of this instance of PBFT.
type PbftProposalState struct {

	// Tracks the number of proposals already made.
	// Used to calculate the sequence number of the next proposal (from the associated segment's sequence numbers)
	// and to stop proposing when a proposal has been made for all sequence numbers.
	ProposalsMade int

	// Flag indicating whether a new certificate has been requested from ISS.
	// This effectively means that a proposal is in progress and no new proposals should be started.
	// This is required, since the PBFT implementation does not assemble availability certificates itself
	// (which, in general, do not exclusively concern the State and are thus managed by ISS directly).
	// Instead, when the PBFT instance is ready to propose,
	// it requests a new certificate from the enclosing ISS implementation.
	CertRequested bool

	// If CertRequested is true, CertRequestedView indicates the PBFT view
	// in which the PBFT protocol was when the certificate was requested.
	// When the certificate is ready, the protocol needs to check whether it is still in the same view
	// as when the certificate was requested.
	// If the view advanced in the meantime, the proposal must be aborted and the contained transactions resurrected
	// If CertRequested is false, CertRequestedView must not be used.
	CertRequestedView ot.ViewNr

	// the number of proposals for which the timeout has passed.
	// If this number is higher than proposalsMade, a new certificate can be proposed.
	ProposalTimeout int
}

// ============================================================
// General protocol logic (other specific parts in separate files)
// ============================================================

// NumCommitted returns the number of slots that are already committed in the given view.
func (state *State) NumCommitted(view ot.ViewNr) int {
	numCommitted := 0
	for _, slot := range state.Slots[view] {
		if slot.Committed {
			numCommitted++
		}
	}
	return numCommitted
}

func (state *State) InitView(
	m dsl.Module,
	params *ModuleParams,
	moduleConfig common.ModuleConfig,
	view ot.ViewNr,
	logger logging.Logger,
) error {

	// Sanity check
	if view < state.View {
		return es.Errorf("starting a view (%d) older than the current one (%d)", view, state.View)
	}

	// Do not start the same view more than once.
	// View 0 is also started only once (the code makes sure that startView(0) is only called at initialization),
	// it's just that the default value of the variable is already 0 - that's why it needs an exception.
	if view != 0 && view == state.View {
		return nil
	}

	logger.Log(logging.LevelInfo, "Initializing new view.", "view", view)

	// Initialize PBFT slots for the new view, one for each sequence number.
	state.Slots[view] = make(map[tt.SeqNr]*PbftSlot)
	for _, sn := range state.Segment.SeqNrs() {

		// Create a fresh, empty slot.
		// For n being the Membership size, f = (n-1) / 3
		state.Slots[view][sn] = NewPbftSlot(state.Segment.Membership)

		// Except for initialization of view 0, carry over state from the previous view.
		if view > 0 {
			state.Slots[view][sn].PopulateFromPrevious(state.Slots[state.View][sn], view)
		}
	}

	// Set view change timeouts
	eventpbdsl.TimerDelay(
		m,
		moduleConfig.Timer,
		[]*eventpbtypes.Event{pbftpbevents.ViewChangeSNTimeout(
			moduleConfig.Self,
			view,
			uint64(state.NumCommitted(view)))},
		computeTimeout(timertypes.Duration(params.Config.ViewChangeSNTimeout), view),
	)
	eventpbdsl.TimerDelay(
		m,
		moduleConfig.Timer,
		[]*eventpbtypes.Event{pbftpbevents.ViewChangeSegTimeout(moduleConfig.Self, uint64(view))}, // TODO Update proto message to use ViewNr
		computeTimeout(timertypes.Duration(params.Config.ViewChangeSegmentTimeout), view),
	)

	state.View = view

	return nil
}

// AllCommitted returns true if all slots of this pbftInstance in the current view are in the committed state
// (i.e., have the committed flag set).
func (state *State) AllCommitted() bool {
	return state.NumCommitted(state.View) == len(state.Slots[state.View])
}

func (state *State) LookUpPreprepare(sn tt.SeqNr, digest []byte) *pbftpbtypes.Preprepare {
	// Traverse all views, starting in the current view.
	for view := state.View; ; view-- {

		// If the view exists (i.e. if this node ever entered the view)
		if slots, ok := state.Slots[view]; ok {

			// If the given sequence number is part of the segment (might not be, in case of a corrupted message)
			if slot, ok := slots[sn]; ok {

				// If the slot contains a matching Preprepare
				if preprepare := slot.GetPreprepare(digest); preprepare != nil {

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
func computeTimeout(timeout timertypes.Duration, view ot.ViewNr) timertypes.Duration {
	timeout *= 1 << view
	return timeout
}
