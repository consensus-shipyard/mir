/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderers

import (
	"fmt"

	factorypbtypes "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	ordererpbtypes "github.com/filecoin-project/mir/pkg/pb/ordererpb/types"

	"github.com/filecoin-project/mir/pkg/dsl"
	issconfig "github.com/filecoin-project/mir/pkg/iss/config"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/messagebuffer"
	"github.com/filecoin-project/mir/pkg/modules"
	ot "github.com/filecoin-project/mir/pkg/orderers/types"
	availabilitypbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/dsl"
	availabilitypbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	cryptopbdsl "github.com/filecoin-project/mir/pkg/pb/cryptopb/dsl"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	hasherpbdsl "github.com/filecoin-project/mir/pkg/pb/hasherpb/dsl"
	pbftpbdsl "github.com/filecoin-project/mir/pkg/pb/pbftpb/dsl"
	pbftpbevents "github.com/filecoin-project/mir/pkg/pb/pbftpb/events"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
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
	moduleConfig ModuleConfig

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
	slots map[ot.ViewNr]map[tt.SeqNr]*pbftSlot

	// Tracks the state of the segment-local checkpoint.
	segmentCheckpoint *pbftSegmentChkp

	// Logger for outputting debugging messages.
	logger logging.Logger

	// PBFT view
	view ot.ViewNr

	// Flag indicating whether this node is currently performing a view change.
	// It is set on sending the ViewChange message and cleared on accepting a new view.
	inViewChange bool

	// For each view transition, stores the state of the corresponding PBFT view change sub-protocol.
	// The map itself is allocated on creation of the pbftInstance, but the entries are initialized lazily,
	// only when needed (when the node initiates a view change).
	viewChangeStates map[ot.ViewNr]*pbftViewChangeState

	// Contains the application-specific code for validating incoming proposals.
	externalValidator ValidityChecker
}

// NewOrdererModule allocates and initializes a new instance of the PBFT Orderer.
// It takes the following parameters:
//   - moduleConfig
//   - ownID: The ID of this node.
//   - segment: The segment governing this SB instance,
//     specifying the leader, the set of sequence numbers, etc.
//   - config: PBFT-specific configuration parameters.
//   - externalValidator: Contains the application-specific code for validating incoming proposals.
//   - logger: Logger for outputting debugging messages.

func NewOrdererModule(
	moduleConfig ModuleConfig,
	ownID t.NodeID,
	segment *Segment,
	config *PBFTConfig,
	externalValidator ValidityChecker,
	logger logging.Logger) modules.PassiveModule {

	// Set all the necessary fields of the new instance and return it.
	orderer := &Orderer{
		ownID:             ownID,
		segment:           segment,
		moduleConfig:      moduleConfig,
		config:            config,
		externalValidator: externalValidator,
		slots:             make(map[ot.ViewNr]map[tt.SeqNr]*pbftSlot),
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
		viewChangeStates: make(map[ot.ViewNr]*pbftViewChangeState),
	}
	m := dsl.NewModule(moduleConfig.Self)

	eventpbdsl.UponInit(m, func() error {
		return orderer.applyInit(m)
	})

	cryptopbdsl.UponSignResult(m, func(signature []byte, context *pbftpbtypes.ViewChange) error {
		orderer.applyViewChangeSignResult(m, signature, context)
		return nil
	})

	cryptopbdsl.UponSigVerified(m, func(nodeID t.NodeID, error error, context *pbftpbtypes.SignedViewChange) error {
		// Ignore events with invalid signatures.
		if error != nil {
			orderer.logger.Log(logging.LevelWarn,
				"Ignoring invalid signature, ignoring event.",
				"from", nodeID,
				"error", error,
			)
			return nil
		}
		// Verify all signers are part of the membership
		if !sliceutil.Contains(orderer.config.Membership, nodeID) {
			orderer.logger.Log(logging.LevelWarn,
				"Ignoring SigVerified message as it contains signatures from non members, ignoring event (with all signatures).",
				"from", nodeID,
			)
			return nil
		}

		orderer.applyVerifiedViewChange(m, context, nodeID)
		return nil
	})

	cryptopbdsl.UponSigsVerified(m, func(nodeIds []t.NodeID, errors []error, allOk bool, context *pbftpbtypes.NewView) error {
		// Ignore events with invalid signatures.
		if !allOk {
			orderer.logger.Log(logging.LevelWarn,
				"Ignoring invalid signature, ignoring event (with all signatures).",
				"from", nodeIds,
				"errors", errors,
			)
			return nil
		}

		// Verify all signers are part of the membership
		if !sliceutil.ContainsAll(orderer.config.Membership, nodeIds) {
			orderer.logger.Log(logging.LevelWarn,
				"Ignoring SigsVerified message as it contains signatures from non members, ignoring event (with all signatures).",
				"from", nodeIds,
			)
			return nil
		}
		orderer.applyVerifiedNewView(m, context)
		return nil
	})

	availabilitypbdsl.UponNewCert(m, func(cert *availabilitypbtypes.Cert, _ *struct{}) error {
		return orderer.applyCertReady(m, cert)
	})

	hasherpbdsl.UponResult(m, func(digests [][]byte, context *pbftpbtypes.Preprepare) error {
		orderer.applyPreprepareHashResult(m, digests[0], context)
		return nil
	})

	hasherpbdsl.UponResult(m, func(digests [][]byte, context *ot.ViewNr) error {
		return orderer.applyEmptyPreprepareHashResult(m, digests, *context)
	})

	hasherpbdsl.UponResult(m, func(digests [][]byte, context *MissingPreprepares) error {
		orderer.applyMissingPreprepareHashResult(
			m,
			digests[0],
			context.Preprepare,
		)
		return nil
	})

	hasherpbdsl.UponResult(m, func(digests [][]byte, context *pbftpbtypes.NewView) error {
		return orderer.applyNewViewHashResult(m, digests, context)
	})

	hasherpbdsl.UponResult(m, func(digests [][]byte, context *pbftpbtypes.CatchUpResponse) error {
		orderer.applyCatchUpResponseHashResult(
			m,
			digests[0],
			context.Resp,
		)
		return nil
	})

	pbftpbdsl.UponProposeTimeout(m, func(proposeTimeout uint64) error {
		return orderer.applyProposeTimeout(m, int(proposeTimeout))
	})

	pbftpbdsl.UponViewChangeSNTimeout(m, func(view ot.ViewNr, numCommitted uint64) error {
		return orderer.applyViewChangeSNTimeout(m, view, numCommitted)
	})

	pbftpbdsl.UponViewChangeSegTimeout(m, func(viewChangeSegTimeout uint64) error {
		return orderer.applyViewChangeSegmentTimeout(m, ot.ViewNr(viewChangeSegTimeout))
	})

	pbftpbdsl.UponPreprepareReceived(m, func(from t.NodeID, sn tt.SeqNr, view ot.ViewNr, data []byte, aborted bool) error {
		if !sliceutil.Contains(orderer.config.Membership, from) {
			orderer.logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		preprepare := &pbftpbtypes.Preprepare{
			Sn:      sn,
			View:    view,
			Data:    data,
			Aborted: aborted,
		}
		orderer.applyMsgPreprepare(m, preprepare, from)
		return nil
	})

	pbftpbdsl.UponPrepareReceived(m, func(from t.NodeID, sn tt.SeqNr, view ot.ViewNr, digest []byte) error {
		if !sliceutil.Contains(orderer.config.Membership, from) {
			orderer.logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		prepare := &pbftpbtypes.Prepare{
			Sn:     sn,
			View:   view,
			Digest: digest,
		}
		orderer.applyMsgPrepare(m, prepare, from)
		return nil
	})

	pbftpbdsl.UponCommitReceived(m, func(from t.NodeID, sn tt.SeqNr, view ot.ViewNr, digest []byte) error {
		if !sliceutil.Contains(orderer.config.Membership, from) {
			orderer.logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		commit := &pbftpbtypes.Commit{
			Sn:     sn,
			View:   view,
			Digest: digest,
		}
		orderer.applyMsgCommit(m, commit, from)
		return nil
	})

	pbftpbdsl.UponSignedViewChangeReceived(m, func(from t.NodeID, viewChange *pbftpbtypes.ViewChange, signature []byte) error {
		if !sliceutil.Contains(orderer.config.Membership, from) {
			orderer.logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		svc := &pbftpbtypes.SignedViewChange{
			ViewChange: viewChange,
			Signature:  signature,
		}
		orderer.applyMsgSignedViewChange(m, svc, from)
		return nil
	})

	pbftpbdsl.UponPreprepareRequestReceived(m, func(from t.NodeID, digest []byte, sn tt.SeqNr) error {
		if !sliceutil.Contains(orderer.config.Membership, from) {
			orderer.logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}

		orderer.applyMsgPreprepareRequest(m, digest, sn, from)
		return nil
	})

	pbftpbdsl.UponMissingPreprepareReceived(m, func(from t.NodeID, preprepare *pbftpbtypes.Preprepare) error {
		if !sliceutil.Contains(orderer.config.Membership, from) {
			orderer.logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		orderer.applyMsgMissingPreprepare(m, preprepare, from)
		return nil
	})

	pbftpbdsl.UponNewViewReceived(m, func(
		from t.NodeID,
		view ot.ViewNr,
		viewChangeSenders []string,
		signedViewChanges []*pbftpbtypes.SignedViewChange,
		preprepareSeqNrs []tt.SeqNr,
		preprepares []*pbftpbtypes.Preprepare,
	) error {

		if !sliceutil.Contains(orderer.config.Membership, from) {
			orderer.logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		nv := &pbftpbtypes.NewView{
			View:              view,
			ViewChangeSenders: viewChangeSenders,
			SignedViewChanges: signedViewChanges,
			PreprepareSeqNrs:  preprepareSeqNrs,
			Preprepares:       preprepares,
		}
		orderer.applyMsgNewView(m, nv, from)
		return nil
	})

	pbftpbdsl.UponDoneReceived(m, func(from t.NodeID, digests [][]byte) error {
		if !sliceutil.Contains(orderer.config.Membership, from) {
			orderer.logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		orderer.applyMsgDone(m, digests, from)
		return nil
	})

	pbftpbdsl.UponCatchUpRequestReceived(m, func(from t.NodeID, digest []uint8, sn tt.SeqNr) error {
		if !sliceutil.Contains(orderer.config.Membership, from) {
			orderer.logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		orderer.applyMsgCatchUpRequest(m, digest, sn, from)
		return nil
	})

	pbftpbdsl.UponCatchUpResponseReceived(m, func(from t.NodeID, resp *pbftpbtypes.Preprepare) error {
		if !sliceutil.Contains(orderer.config.Membership, from) {
			orderer.logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		orderer.applyMsgCatchUpResponse(m, resp, from)
		return nil
	})
	return m

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
func (orderer *Orderer) applyInit(m dsl.Module) error {

	// Initialize the first PBFT view
	var err error
	if err = orderer.initView(m, 0); err != nil {
		return err
	}

	eventpbdsl.TimerDelay(
		m,
		orderer.moduleConfig.Timer,
		[]*eventpbtypes.Event{pbftpbevents.ProposeTimeout(orderer.moduleConfig.Self, 1)},
		types.Duration(orderer.config.MaxProposeDelay),
	)

	// Set up timer for the first proposal.
	return nil

}

// numCommitted returns the number of slots that are already committed in the given view.
func (orderer *Orderer) numCommitted(view ot.ViewNr) int {
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

func (orderer *Orderer) initView(m dsl.Module, view ot.ViewNr) error {
	// Sanity check
	if view < orderer.view {
		return fmt.Errorf("starting a view (%d) older than the current one (%d)", view, orderer.view)
	}

	// Do not start the same view more than once.
	// View 0 is also started only once (the code makes sure that startView(0) is only called at initialization),
	// it's just that the default value of the variable is already 0 - that's why it needs an exception.
	if view != 0 && view == orderer.view {
		return nil
	}

	orderer.logger.Log(logging.LevelInfo, "Initializing new view.", "view", view)

	// Initialize PBFT slots for the new view, one for each sequence number.
	orderer.slots[view] = make(map[tt.SeqNr]*pbftSlot)
	for _, sn := range orderer.segment.SeqNrs() {

		// Create a fresh, empty slot.
		// For n being the membership size, f = (n-1) / 3
		orderer.slots[view][sn] = newPbftSlot(len(orderer.segment.Membership.Nodes))

		// Except for initialization of view 0, carry over state from the previous view.
		if view > 0 {
			orderer.slots[view][sn].populateFromPrevious(orderer.slots[orderer.view][sn], view)
		}
	}

	// Set view change timeouts
	eventpbdsl.TimerDelay(
		m,
		orderer.moduleConfig.Timer,
		[]*eventpbtypes.Event{pbftpbevents.ViewChangeSNTimeout(
			orderer.moduleConfig.Self,
			view,
			uint64(orderer.numCommitted(view)))},
		computeTimeout(types.Duration(orderer.config.ViewChangeSNTimeout), view),
	)
	eventpbdsl.TimerDelay(
		m,
		orderer.moduleConfig.Timer,
		[]*eventpbtypes.Event{pbftpbevents.ViewChangeSegTimeout(orderer.moduleConfig.Self, uint64(view))}, // TODO Update proto message to use ViewNr
		computeTimeout(types.Duration(orderer.config.ViewChangeSegmentTimeout), view),
	)

	orderer.view = view
	orderer.inViewChange = false

	return nil
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
func computeTimeout(timeout types.Duration, view ot.ViewNr) types.Duration {
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

func InstanceParams(
	segment *Segment,
	availabilityID t.ModuleID,
	epoch tt.EpochNr,
	validityCheckerType ValidityCheckerType,
) *factorypbtypes.GeneratorParams {
	return &factorypbtypes.GeneratorParams{Type: &factorypbtypes.GeneratorParams_PbftModule{
		PbftModule: &ordererpbtypes.PBFTModule{
			Segment:         segment.PbType(),
			AvailabilityId:  availabilityID.Pb(),
			Epoch:           epoch.Pb(),
			ValidityChecker: uint64(validityCheckerType),
		},
	}}
}

type MissingPreprepares struct {
	*pbftpbtypes.Preprepare
}

type CatchUpResponse struct {
	*pbftpbtypes.Preprepare
}
