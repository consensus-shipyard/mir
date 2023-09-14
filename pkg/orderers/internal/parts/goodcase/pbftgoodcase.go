package goodcase

import (
	"fmt"

	es "github.com/go-errors/errors"

	common2 "github.com/filecoin-project/mir/pkg/orderers/common"
	"github.com/filecoin-project/mir/pkg/orderers/internal/parts/catchup"
	isspbdsl "github.com/filecoin-project/mir/pkg/pb/isspb/dsl"
	"github.com/filecoin-project/mir/pkg/pb/pbftpb"

	availabilitypbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	pbftpbdsl "github.com/filecoin-project/mir/pkg/pb/pbftpb/dsl"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"

	ot "github.com/filecoin-project/mir/pkg/orderers/types"

	"github.com/filecoin-project/mir/pkg/orderers/internal/common"

	"github.com/filecoin-project/mir/pkg/dsl"
	apbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/dsl"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	hasherpbdsl "github.com/filecoin-project/mir/pkg/pb/hasherpb/dsl"
	ppvpbdsl "github.com/filecoin-project/mir/pkg/pb/ordererpb/pprepvalidatorpb/dsl"
	pbftpbevents "github.com/filecoin-project/mir/pkg/pb/pbftpb/events"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"

	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/logging"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	pbftpbmsgs "github.com/filecoin-project/mir/pkg/pb/pbftpb/msgs"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	"github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

func IncludeGoodCase(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	logger logging.Logger,
) {

	// UponInit take all the actions resulting from the PBFT State's initial state.
	// The Init event is expected to be the first event applied to the State,
	// except for events read from the WAL at startup, which are expected to be applied even before the Init event.
	// (At this time, the WAL is not used. TODO: Update this when wal is implemented.)
	eventpbdsl.UponInit(m, func() error {
		// Initialize the first PBFT view
		if err := state.InitView(m, params, moduleConfig, 0, logger); err != nil {
			return err
		}

		// Set up timer for the first proposal.
		// TODO: change names, timeout is handled by the mempool
		pbftpbdsl.ProposeTimeout(m, moduleConfig.Self, 1)

		return nil
	})

	// UponNewCert process the new availability certificate ready to be proposed.
	// This event is triggered by the availability module in response to the CertRequest event produced by this State.
	apbdsl.UponNewCert(m, func(cert *availabilitypbtypes.Cert, _ *struct{}) error {
		// Clear flag that was set in requestNewCert(), so that new certificates can be requested if necessary.
		state.Proposal.CertRequested = false

		if state.Proposal.CertRequestedView == state.View {
			// If the protocol is still in the same PBFT view as when the certificate was requested,
			// propose the received certificate.

			certBytes, err := proto.Marshal(cert.Pb())
			if err != nil {
				return es.Errorf("error marshalling certificate: %w", err)
			}

			err = propose(m, state, params, moduleConfig, certBytes, logger)
			if err != nil {
				return es.Errorf("failed to propose: %w", err)
			}
		} else { // nolint:staticcheck,revive
			// If the PBFT view advanced since the certificate was requested,
			// do not propose the certificate and resurrect the transactions it contains.
			// eventsOut.PushBack(state.eventService.SBEvent(SBResurrectBatchEvent(batch.Batch)))
			// TODO: Implement resurrection!
		}

		return nil
	})

	// UponResultOne process the preprepare message and send a Prepare message to all nodes.
	hasherpbdsl.UponResultOne(m, func(digest []byte, context *pbftpbtypes.Preprepare) error {
		// Stop processing the Preprepare if view advanced in the meantime.
		if context.View < state.View {
			return nil
		}

		// Save the digest of the Preprepare message and mark the slot as preprepared.
		slot := state.Slots[state.View][context.Sn]
		slot.Digest = digest
		slot.Preprepared = true

		// Send a Prepare message.
		transportpbdsl.SendMessage(
			m,
			moduleConfig.Net,
			pbftpbmsgs.Prepare(moduleConfig.Self, context.Sn, state.View, digest),
			state.Segment.NodeIDs(),
		)
		// Advance the state of the PbftSlot even more if necessary
		// (potentially sending a Commit message or even delivering).
		// This is required for the case when the Preprepare message arrives late.
		return advanceSlotState(m, state, params, moduleConfig, slot, context.Sn, logger)
	})

	pbftpbdsl.UponProposeTimeout(m, func(proposeTimeout uint64) error {
		return applyProposeTimeout(m, state, params, moduleConfig, int(proposeTimeout), logger)
	})

	pbftpbdsl.UponPreprepareReceived(m, func(from t.NodeID, sn tt.SeqNr, view ot.ViewNr, data []byte, aborted bool) error {
		if !sliceutil.Contains(params.Config.Membership, from) {
			logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		preprepare := &pbftpbtypes.Preprepare{
			Sn:      sn,
			View:    view,
			Data:    data,
			Aborted: aborted,
		}

		return ApplyMsgPreprepare(m, moduleConfig, preprepare, from)
	})

	ppvpbdsl.UponPreprepareValidated(m, func(err error, c *validatePreprepareContext) error {
		ApplyMsgPreprepareValidated(m, state, moduleConfig, c.preprepare, c.from, logger, err)
		return nil
	})

	pbftpbdsl.UponPrepareReceived(m, func(from t.NodeID, sn tt.SeqNr, view ot.ViewNr, digest []byte) error {
		if !sliceutil.Contains(params.Config.Membership, from) {
			logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}

		prepare := &pbftpbtypes.Prepare{
			Sn:     sn,
			View:   view,
			Digest: digest,
		}
		return applyMsgPrepare(m, state, params, moduleConfig, prepare, from, logger)
	})

	pbftpbdsl.UponCommitReceived(m, func(from t.NodeID, sn tt.SeqNr, view ot.ViewNr, digest []byte) error {
		if !sliceutil.Contains(params.Config.Membership, from) {
			logger.Log(logging.LevelWarn, "sender %s is not a member.\n", from)
			return nil
		}
		commit := &pbftpbtypes.Commit{
			Sn:     sn,
			View:   view,
			Digest: digest,
		}

		return applyMsgCommit(m, state, params, moduleConfig, commit, from, logger)
	})

}

// applyProposeTimeout applies the event of the proposal timeout firing.
// It updates the proposal state accordingly and triggers a new proposal if possible.
func applyProposeTimeout(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	numProposals int,
	logger logging.Logger,
) error {

	// If we are still waiting for this timeout
	if numProposals > state.Proposal.ProposalTimeout {

		// Save the number of proposals for which the timeout fired.
		state.Proposal.ProposalTimeout = numProposals

		// If this was the last bit missing, start a new proposal.
		if canPropose(state, params) {

			// If a proposal already has been set as a parameter of the segment, propose it directly.
			sn := state.Segment.SeqNrs()[state.Proposal.ProposalsMade]
			if proposal := state.Segment.Proposals[sn]; proposal != nil {
				return propose(m, state, params, moduleConfig, proposal, logger)
			}

			// Otherwise, obtain a fresh availability certificate first.
			requestNewCert(m, state, moduleConfig)
		}
	}

	return nil
}

// requestNewCert asks (by means of a CertRequest event) the availability module to provide a new availability certificate.
// When the certificate is ready, it must be passed to the State using the NewCert event.
func requestNewCert(m dsl.Module, state *common.State, moduleConfig common2.ModuleConfig) {

	// Set a flag indicating that a certificate has been requested,
	// so that no new certificates will be requested before the reception of this one.
	// It will be cleared when NewCert is received.
	state.Proposal.CertRequested = true

	// Remember the view in which the certificate has been requested
	// to make sure that we are still in the same view when the certificate becomes ready.
	state.Proposal.CertRequestedView = state.View

	// Emit the CertRequest event.
	// Operation continues on reception of the NewCert event.
	apbdsl.RequestCert(m, moduleConfig.Ava, &struct{}{})
}

// propose proposes a new availability certificate by sending a Preprepare message.
// propose assumes that the state of the PBFT State allows sending a new proposal
// and does not perform any checks in this regard.
func propose(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	data []byte,
	logger logging.Logger,
) error {

	// Update proposal counter.
	sn := state.Segment.SeqNrs()[state.Proposal.ProposalsMade]
	state.Proposal.ProposalsMade++

	// Log debug message.
	logger.Log(logging.LevelDebug, "Proposing.", "sn", sn)

	// Send a Preprepare message.
	// No need for periodic re-transmission.
	// In the worst case, dropping of these messages may result in a view change, but will not compromise correctness.
	transportpbdsl.SendMessage(m,
		moduleConfig.Net,
		pbftpbmsgs.Preprepare(moduleConfig.Self, sn, state.View, data, false),
		state.Segment.NodeIDs())

	// Set up a new timer for the next proposal.
	timeoutEvent := pbftpbevents.ProposeTimeout(
		moduleConfig.Self,
		uint64(state.Proposal.ProposalsMade+1))

	eventpbdsl.TimerDelay(m,
		moduleConfig.Timer,
		[]*eventpbtypes.Event{timeoutEvent},
		types.Duration(params.Config.MaxProposeDelay),
	)

	return nil
}

// ApplyMsgPreprepare applies a received preprepare message.
// It performs the necessary checks and, if successful, submits it for hashing.
func ApplyMsgPreprepare(
	m dsl.Module,
	moduleConfig common2.ModuleConfig,
	preprepare *pbftpbtypes.Preprepare,
	from t.NodeID,
) error {
	ppvpbdsl.ValidatePreprepare(m,
		moduleConfig.PPrepValidator,
		preprepare,
		&validatePreprepareContext{preprepare, from})

	return nil
}

func ApplyMsgPreprepareValidated(
	m dsl.Module,
	state *common.State,
	moduleConfig common2.ModuleConfig,
	preprepare *pbftpbtypes.Preprepare,
	from t.NodeID,
	logger logging.Logger,
	err error,
) {

	// Convenience variable
	sn := preprepare.Sn

	if err != nil {
		logger.Log(logging.LevelWarn, "Ignoring Preprepare message with invalid proposal.",
			"sn", sn, "from", from, "err", err)
		return
	}

	// Preprocess message, looking up the corresponding pbftSlot.
	slot := preprocessMessage(state, sn, preprepare.View, preprepare.Pb(), from, logger)
	if slot == nil {
		// If preprocessing does not return a pbftSlot, the message cannot be processed right now.
		return
	}

	// Check that this is the first Preprepare message received.
	// Note that checking the orderer.Preprepared flag instead would be incorrect,
	// as that flag is only set after computing the Preprepare message's digest.
	if slot.Preprepare != nil {
		logger.Log(logging.LevelDebug, "Ignoring Preprepare message. Already preprepared or prepreparing.",
			"sn", sn, "from", from)
		return
	}

	// Check whether the sender is the legitimate leader of the segment in the current view.
	primary := state.Segment.PrimaryNode(state.View)
	if from != primary {
		logger.Log(logging.LevelWarn, "Ignoring Preprepare message. Invalid Leader.",
			"expectedLeader", primary,
			"sender", from,
		)
		return
	}

	slot.Preprepare = preprepare

	// Request the computation of the hash of the Preprepare message.
	hasherpbdsl.RequestOne(
		m,
		moduleConfig.Hasher,
		common.SerializePreprepareForHashing(preprepare),
		preprepare,
	)
}

// applyMsgPrepare applies a received prepare message.
// It performs the necessary checks and, if successful,
// may trigger additional events like the sending of a Commit message.
func applyMsgPrepare(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	prepare *pbftpbtypes.Prepare,
	from t.NodeID,
	logger logging.Logger,
) error {

	if prepare.Digest == nil {
		logger.Log(logging.LevelWarn, "Ignoring Prepare message with nil digest.")
		return nil
	}

	// Convenience variable
	sn := prepare.Sn

	// Preprocess message, looking up the corresponding pbftSlot.
	slot := preprocessMessage(state, sn, prepare.View, prepare.Pb(), from, logger)
	if slot == nil {
		// If preprocessing does not return a pbftSlot, the message cannot be processed right now.
		return nil
	}

	// Check if a Prepare message has already been received from this node in the current view.
	if _, ok := slot.PrepareDigests[from]; ok {
		logger.Log(logging.LevelDebug, "Ignoring Prepare message. Already received in this view.",
			"sn", sn, "from", from, "view", state.View)
		return nil
	}

	// Save the received Prepare message and advance the slot state
	// (potentially sending a Commit message or even delivering).
	slot.PrepareDigests[from] = prepare.Digest
	return advanceSlotState(m, state, params, moduleConfig, slot, sn, logger)
}

// applyMsgCommit applies a received commit message.
// It performs the necessary checks and, if successful,
// may trigger additional events like delivering the corresponding certificate.
func applyMsgCommit(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	commit *pbftpbtypes.Commit,
	from t.NodeID,
	logger logging.Logger,
) error {

	if commit.Digest == nil {
		logger.Log(logging.LevelWarn, "Ignoring Commit message with nil digest.")
		return nil
	}

	// Convenience variable
	sn := commit.Sn

	// Preprocess message, looking up the corresponding pbftSlot.
	slot := preprocessMessage(state, sn, commit.View, commit.Pb(), from, logger)
	if slot == nil {
		// If preprocessing does not return a pbftSlot, the message cannot be processed right now.
		return nil
	}

	// Check if a Commit message has already been received from this node in the current view.
	if _, ok := slot.CommitDigests[from]; ok {
		logger.Log(logging.LevelDebug, "Ignoring Commit message. Already received in this view.",
			"sn", sn, "from", from, "view", state.View)
		return nil
	}

	// Save the received Commit message and advance the slot state
	// (potentially delivering the corresponding certificate and its successors).
	slot.CommitDigests[from] = commit.Digest
	return advanceSlotState(m, state, params, moduleConfig, slot, sn, logger)
}

func ApplyBufferedMsg(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	msgPb proto.Message,
	from t.NodeID,
	logger logging.Logger,
) error {
	switch msg := msgPb.(type) {
	case *pbftpb.Preprepare:
		return ApplyMsgPreprepare(m, moduleConfig, pbftpbtypes.PreprepareFromPb(msg), from)
	case *pbftpb.Prepare:
		return applyMsgPrepare(m, state, params, moduleConfig, pbftpbtypes.PrepareFromPb(msg), from, logger)
	case *pbftpb.Commit:
		return applyMsgCommit(m, state, params, moduleConfig, pbftpbtypes.CommitFromPb(msg), from, logger)
	default:
		return nil
	}
}

// preprocessMessage performs basic checks on a PBFT protocol message based the associated view and sequence number.
// If the message is invalid or outdated, preprocessMessage simply returns nil.
// If the message is from a future view, preprocessMessage will try to buffer it for later processing and returns nil.
// If the message can be processed, preprocessMessage returns the pbftSlot tracking the corresponding state.
// The slot returned by preprocessMessage always belongs to the current view.
func preprocessMessage(
	state *common.State,
	sn tt.SeqNr,
	view ot.ViewNr,
	msg proto.Message,
	from t.NodeID,
	logger logging.Logger,
) *common.PbftSlot {

	if view < state.View {
		// Ignore messages from old views.
		logger.Log(logging.LevelDebug, "Ignoring message from old view.",
			"sn", sn, "from", from, "msgView", view, "localView", state.View, "type", fmt.Sprintf("%T", msg))
		return nil
	} else if view > state.View || state.InViewChange {
		// If message is from a future view, buffer it.
		logger.Log(logging.LevelDebug, "Buffering message.",
			"sn", sn, "from", from, "msgView", view, "localView", state.View, "type", fmt.Sprintf("%T", msg))
		state.MessageBuffers[from].Store(msg)
		return nil
		// TODO: When view change is implemented, get the messages out of the buffer.
	} else if slot, ok := state.Slots[state.View][sn]; !ok {
		// If the message is from the current view, look up the slot concerned by the message
		// and check if the sequence number is assigned to this PBFT instance.
		logger.Log(logging.LevelDebug, "Ignoring message. Wrong sequence number.",
			"sn", sn, "from", from, "msgView", view)
		return nil
	} else if slot.Committed {
		// Check if the slot has already been committed (this can happen with state transfer).
		return nil
	} else {
		return slot
	}
}

// canPropose returns true if the current state of the PBFT State
// allows for a new availability certificate to be proposed.
// Note that "new availability certificate" means a "fresh" certificate
// proposed during normal operation outside of view change.
// Proposals part of a new view message during a view change do not call this function and are treated separately.
func canPropose(state *common.State, params *common.ModuleParams) bool {
	return params.OwnID == state.Segment.Leader && // Only the leader can propose

		// No regular proposals can be made after a view change.
		// This is specific for the SB-version of PBFT used in ISS and deviates from the standard PBFT protocol.
		state.View == 0 &&

		// A new certificate must not have been requested (if it has, we are already in the process of proposing).
		!state.Proposal.CertRequested &&

		// There must still be a free sequence number for which a proposal can be made.
		state.Proposal.ProposalsMade < state.Segment.Len() &&

		// The proposal timeout must have passed.
		state.Proposal.ProposalTimeout > state.Proposal.ProposalsMade &&

		// No proposals can be made while in view change.
		!state.InViewChange
}

// advanceSlotState checks whether the state of the PbftSlot can be advanced.
// If it can, advanceSlotState updates the state of the PbftSlot and returns a list of Events that result from it.
// Requires the PBFT instance as an argument to use it to generate the proper events.
func advanceSlotState(
	m dsl.Module,
	state *common.State,
	params *common.ModuleParams,
	moduleConfig common2.ModuleConfig,
	slot *common.PbftSlot,
	sn tt.SeqNr,
	logger logging.Logger,
) error {
	// If the slot just became prepared, send the Commit message.
	if !slot.Prepared && slot.CheckPrepared() {
		slot.Prepared = true

		transportpbdsl.SendMessage(
			m,
			moduleConfig.Net,
			pbftpbmsgs.Commit(moduleConfig.Self, sn, state.View, slot.Digest),
			state.Segment.NodeIDs())
	}

	// If the slot just became committed, reset SN timeout and deliver the certificate.
	if !slot.Committed && slot.CheckCommitted() {

		// Mark slot as committed.
		slot.Committed = true

		// Set a new SN timeout (unless the segment is finished with a stable checkpoint).
		// Note that we set a new SN timeout even if everything has already been committed.
		// In such a case, we expect a quorum of other nodes to also commit everything and send a Done message
		// before the timeout fires.
		// The timeout event contains the current view and the number of committed slots.
		// It will be ignored if any of those values change by the time the timer fires
		// or if a quorum of nodes confirms having committed all certificates.
		if !state.SegmentCheckpoint.Stable(state.Segment.Membership) {
			eventpbdsl.TimerDelay(
				m,
				moduleConfig.Timer,
				[]*eventpbtypes.Event{pbftpbevents.ViewChangeSNTimeout(
					moduleConfig.Self,
					state.View,
					uint64(state.NumCommitted(state.View)))},
				types.Duration(params.Config.ViewChangeSNTimeout))
		}

		// If all certificates have been committed (i.e. this is the last certificate to commit),
		// send a Done message to all other nodes.
		// This is required for liveness, see comments for pbftSegmentChkp.
		if state.AllCommitted() {
			state.SegmentCheckpoint.SetDone()
			catchup.SendDoneMessages(m, state, params, moduleConfig, logger)
		}

		// Deliver availability certificate (will be verified by ISS)
		isspbdsl.SBDeliver(
			m,
			moduleConfig.Ord,
			sn,
			slot.Preprepare.Data,
			slot.Preprepare.Aborted,
			state.Segment.Leader,
			moduleConfig.Self,
		)

		// start the next sn immediately
		return applyProposeTimeout(m, state, params, moduleConfig, int(sn)+1, logger)
	}

	return nil
}

type validatePreprepareContext struct {
	preprepare *pbftpbtypes.Preprepare
	from       t.NodeID
}
