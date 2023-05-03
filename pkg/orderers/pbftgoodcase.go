package orderers

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/dsl"
	apbdsl "github.com/filecoin-project/mir/pkg/pb/availabilitypb/dsl"
	availabilitypbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	eventpbdsl "github.com/filecoin-project/mir/pkg/pb/eventpb/dsl"
	hasherpbdsl "github.com/filecoin-project/mir/pkg/pb/hasherpb/dsl"
	pbftpbevents "github.com/filecoin-project/mir/pkg/pb/pbftpb/events"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"

	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/logging"
	orderertypes "github.com/filecoin-project/mir/pkg/orderers/types"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	hasherpbtypes "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	"github.com/filecoin-project/mir/pkg/pb/pbftpb"
	pbftpbmsgs "github.com/filecoin-project/mir/pkg/pb/pbftpb/msgs"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	"github.com/filecoin-project/mir/pkg/timer/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ============================================================
// Data structures and functions
// ============================================================

// pbftProposalState tracks the state of the pbftInstance related to proposing certificates.
// The proposal state is only used if this node is the leader of this instance of PBFT.
type pbftProposalState struct {

	// Tracks the number of proposals already made.
	// Used to calculate the sequence number of the next proposal (from the associated segment's sequence numbers)
	// and to stop proposing when a proposal has been made for all sequence numbers.
	proposalsMade int

	// Flag indicating whether a new certificate has been requested from ISS.
	// This effectively means that a proposal is in progress and no new proposals should be started.
	// This is required, since the PBFT implementation does not assemble availability certificates itself
	// (which, in general, do not exclusively concern the Orderer and are thus managed by ISS directly).
	// Instead, when the PBFT instance is ready to propose,
	// it requests a new certificate from the enclosing ISS implementation.
	certRequested bool

	// If certRequested is true, certRequestedView indicates the PBFT view
	// in which the PBFT protocol was when the certificate was requested.
	// When the certificate is ready, the protocol needs to check whether it is still in the same view
	// as when the certificate was requested.
	// If the view advanced in the meantime, the proposal must be aborted and the contained transactions resurrected
	// (i.e., put back in their respective ISS bucket queues).
	// If certRequested is false, certRequestedView must not be used.
	certRequestedView orderertypes.ViewNr

	// the number of proposals for which the timeout has passed.
	// If this number is higher than proposalsMade, a new certificate can be proposed.
	proposalTimeout int
}

// ============================================================
// OrdererEvent handling
// ============================================================

// applyProposeTimeout applies the event of the proposal timeout firing.
// It updates the proposal state accordingly and triggers a new proposal if possible.
func (orderer *Orderer) applyProposeTimeout(m dsl.Module, numProposals int) error {

	// If we are still waiting for this timeout
	if numProposals > orderer.proposal.proposalTimeout {

		// Save the number of proposals for which the timeout fired.
		orderer.proposal.proposalTimeout = numProposals

		// If this was the last bit missing, start a new proposal.
		if orderer.canPropose() {

			// If a proposal already has been set as a parameter of the segment, propose it directly.
			sn := orderer.segment.SeqNrs()[orderer.proposal.proposalsMade]
			if proposal := orderer.segment.Proposals[sn]; proposal != nil {
				return orderer.propose(m, proposal)
			}

			// Otherwise, obtain a fresh availability certificate first.
			orderer.requestNewCert(m)
		}
	}

	return nil
}

// requestNewCert asks (by means of a CertRequest event) the availability module to provide a new availability certificate.
// When the certificate is ready, it must be passed to the Orderer using the CertReady event.
func (orderer *Orderer) requestNewCert(m dsl.Module) {

	// Set a flag indicating that a certificate has been requested,
	// so that no new certificates will be requested before the reception of this one.
	// It will be cleared when CertReady is received.
	orderer.proposal.certRequested = true

	// Remember the view in which the certificate has been requested
	// to make sure that we are still in the same view when the certificate becomes ready.
	orderer.proposal.certRequestedView = orderer.view

	// Emit the CertRequest event.
	// Operation continues on reception of the CertReady event.
	apbdsl.RequestCert(m, orderer.moduleConfig.Ava, &struct{}{})
}

// applyCertReady processes a new availability certificate ready to be proposed.
// This event is triggered by availability module in response to the CertRequest event produced by this Orderer.
func (orderer *Orderer) applyCertReady(m dsl.Module, cert *availabilitypbtypes.Cert) error {

	// Clear flag that was set in requestNewCert(), so that new certificates can be requested if necessary.
	orderer.proposal.certRequested = false

	if orderer.proposal.certRequestedView == orderer.view {
		// If the protocol is still in the same PBFT view as when the certificate was requested,
		// propose the received certificate.

		certBytes, err := proto.Marshal(cert.Pb())
		if err != nil {
			return fmt.Errorf("error marshalling certificate: %w", err)
		}

		err = orderer.propose(m, certBytes)
		if err != nil {
			return fmt.Errorf("failed to propose: %w", err)
		}
	} else { // nolint:staticcheck,revive
		// If the PBFT view advanced since the certificate was requested,
		// do not propose the certificate and resurrect the transactions it contains.
		// eventsOut.PushBack(orderer.eventService.SBEvent(SBResurrectBatchEvent(batch.Batch)))
		// TODO: Implement resurrection!
	}

	return nil
}

// propose proposes a new availability certificate by sending a Preprepare message.
// propose assumes that the state of the PBFT Orderer allows sending a new proposal
// and does not perform any checks in this regard.
func (orderer *Orderer) propose(m dsl.Module, data []byte) error {

	// Update proposal counter.
	sn := orderer.segment.SeqNrs()[orderer.proposal.proposalsMade]
	orderer.proposal.proposalsMade++

	// Log debug message.
	orderer.logger.Log(logging.LevelDebug, "Proposing.",
		"sn", sn)

	// Send a Preprepare message.
	// No need for periodic re-transmission.
	// In the worst case, dropping of these messages may result in a view change, but will not compromise correctness.
	transportpbdsl.SendMessage(m,
		orderer.moduleConfig.Net,
		pbftpbmsgs.Preprepare(orderer.moduleConfig.Self, sn, orderer.view, data, false),
		orderer.segment.NodeIDs())

	// Set up a new timer for the next proposal.
	ordererEvent := pbftpbevents.ProposeTimeout(
		orderer.moduleConfig.Self,
		uint64(orderer.proposal.proposalsMade+1))

	eventpbdsl.TimerDelay(m,
		orderer.moduleConfig.Timer,
		[]*eventpbtypes.Event{ordererEvent},
		types.Duration(orderer.config.MaxProposeDelay),
	)

	return nil
}

// applyMsgPreprepare applies a received preprepare message.
// It performs the necessary checks and, if successful, submits it for hashing.
func (orderer *Orderer) applyMsgPreprepare(m dsl.Module, preprepare *pbftpbtypes.Preprepare, from t.NodeID) {

	// Convenience variable
	sn := preprepare.Sn

	// Preprocess message, looking up the corresponding pbftSlot.
	slot := orderer.preprocessMessage(sn, preprepare.View, preprepare.Pb(), from)
	if slot == nil {
		// If preprocessing does not return a pbftSlot, the message cannot be processed right now.
		return
	}

	// Check whether valid data has been proposed.
	if err := orderer.externalValidator.Check(preprepare.Data); err != nil {
		orderer.logger.Log(logging.LevelWarn, "Ignoring Preprepare message with invalid proposal.",
			"sn", sn, "from", from, "err", err)
		return
	}

	// Check that this is the first Preprepare message received.
	// Note that checking the orderer.Preprepared flag instead would be incorrect,
	// as that flag is only set after computing the Preprepare message's digest.
	if slot.Preprepare != nil {
		orderer.logger.Log(logging.LevelDebug, "Ignoring Preprepare message. Already preprepared or prepreparing.",
			"sn", sn, "from", from)
		return
	}

	// Check whether the sender is the legitimate leader of the segment in the current view.
	primary := orderer.segment.PrimaryNode(orderer.view)
	if from != primary {
		orderer.logger.Log(logging.LevelWarn, "Ignoring Preprepare message. Invalid Leader.",
			"expectedLeader", primary,
			"sender", from,
		)
		return
	}

	slot.Preprepare = preprepare

	// Request the computation of the hash of the Preprepare message.
	hasherpbdsl.Request(
		m,
		orderer.moduleConfig.Hasher,
		[]*hasherpbtypes.HashData{serializePreprepareForHashing(preprepare)},
		preprepareHashOrigin(preprepare.Pb()),
	)
}

func (orderer *Orderer) applyPreprepareHashResult(m dsl.Module, digest []byte, preprepare *pbftpb.Preprepare) {
	// Convenience variable.
	sn := tt.SeqNr(preprepare.Sn)

	// Stop processing the Preprepare if view advanced in the meantime.
	if orderertypes.ViewNr(preprepare.View) < orderer.view {
		return
	}

	// Save the digest of the Preprepare message and mark the slot as preprepared.
	slot := orderer.slots[orderer.view][sn]
	slot.Digest = digest
	slot.Preprepared = true

	// Send a Prepare message.
	transportpbdsl.SendMessage(
		m,
		orderer.moduleConfig.Net,
		pbftpbmsgs.Prepare(orderer.moduleConfig.Self, sn, orderer.view, digest),
		orderer.segment.NodeIDs(),
	)
	// Advance the state of the pbftSlot even more if necessary
	// (potentially sending a Commit message or even delivering).
	// This is required for the case when the Preprepare message arrives late.
	slot.advanceState(m, orderer, sn)
}

// applyMsgPrepare applies a received prepare message.
// It performs the necessary checks and, if successful,
// may trigger additional events like the sending of a Commit message.
func (orderer *Orderer) applyMsgPrepare(m dsl.Module, prepare *pbftpbtypes.Prepare, from t.NodeID) {

	// Convenience variable
	sn := prepare.Sn

	// Preprocess message, looking up the corresponding pbftSlot.
	slot := orderer.preprocessMessage(sn, prepare.View, prepare.Pb(), from)
	if slot == nil {
		// If preprocessing does not return a pbftSlot, the message cannot be processed right now.
		return
	}

	// Check if a Prepare message has already been received from this node in the current view.
	if _, ok := slot.PreparesDigest[from]; ok {
		orderer.logger.Log(logging.LevelDebug, "Ignoring Prepare message. Already received in this view.",
			"sn", sn, "from", from, "view", orderer.view)
		return
	}

	// Save the received Prepare message and advance the slot state
	// (potentially sending a Commit message or even delivering).
	slot.PreparesDigest[from] = prepare.Digest
	slot.advanceState(m, orderer, sn)
}

// applyMsgCommit applies a received commit message.
// It performs the necessary checks and, if successful,
// may trigger additional events like delivering the corresponding certificate.
func (orderer *Orderer) applyMsgCommit(m dsl.Module, commit *pbftpbtypes.Commit, from t.NodeID) {

	// Convenience variable
	sn := commit.Sn

	// Preprocess message, looking up the corresponding pbftSlot.
	slot := orderer.preprocessMessage(sn, commit.View, commit.Pb(), from)
	if slot == nil {
		// If preprocessing does not return a pbftSlot, the message cannot be processed right now.
		return
	}

	// Check if a Commit message has already been received from this node in the current view.
	if _, ok := slot.CommitsDigest[from]; ok {
		orderer.logger.Log(logging.LevelDebug, "Ignoring Commit message. Already received in this view.",
			"sn", sn, "from", from, "view", orderer.view)
		return
	}

	// Save the received Commit message and advance the slot state
	// (potentially delivering the corresponding certificate and its successors).
	slot.CommitsDigest[from] = commit.Digest
	slot.advanceState(m, orderer, sn)
}

// preprocessMessage performs basic checks on a PBFT protocol message based the associated view and sequence number.
// If the message is invalid or outdated, preprocessMessage simply returns nil.
// If the message is from a future view, preprocessMessage will try to buffer it for later processing and returns nil.
// If the message can be processed, preprocessMessage returns the pbftSlot tracking the corresponding state.
// The slot returned by preprocessMessage always belongs to the current view.
func (orderer *Orderer) preprocessMessage(sn tt.SeqNr, view orderertypes.ViewNr, msg proto.Message, from t.NodeID) *pbftSlot {

	if view < orderer.view {
		// Ignore messages from old views.
		orderer.logger.Log(logging.LevelDebug, "Ignoring message from old view.",
			"sn", sn, "from", from, "msgView", view, "localView", orderer.view, "type", fmt.Sprintf("%T", msg))
		return nil
	} else if view > orderer.view || orderer.inViewChange {
		// If message is from a future view, buffer it.
		orderer.logger.Log(logging.LevelDebug, "Buffering message.",
			"sn", sn, "from", from, "msgView", view, "localView", orderer.view, "type", fmt.Sprintf("%T", msg))
		orderer.messageBuffers[from].Store(msg)
		return nil
		// TODO: When view change is implemented, get the messages out of the buffer.
	} else if slot, ok := orderer.slots[orderer.view][sn]; !ok {
		// If the message is from the current view, look up the slot concerned by the message
		// and check if the sequence number is assigned to this PBFT instance.
		orderer.logger.Log(logging.LevelDebug, "Ignoring message. Wrong sequence number.",
			"sn", sn, "from", from, "msgView", view)
		return nil
	} else if slot.Committed {
		// Check if the slot has already been committed (this can happen with state transfer).
		return nil
	} else {
		return slot
	}
}
