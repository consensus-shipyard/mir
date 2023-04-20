package orderers

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	orderertypes "github.com/filecoin-project/mir/pkg/orderers/types"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	commonpbtypes "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	eventpbevents "github.com/filecoin-project/mir/pkg/pb/eventpb/events"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	hasherpbevents "github.com/filecoin-project/mir/pkg/pb/hasherpb/events"
	"github.com/filecoin-project/mir/pkg/pb/pbftpb"
	pbftpbmsgs "github.com/filecoin-project/mir/pkg/pb/pbftpb/msgs"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
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
	// If the view advanced in the meantime, the proposal must be aborted and the contained requests resurrected
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
func (orderer *Orderer) applyProposeTimeout(numProposals int) (*events.EventList, error) {

	// If we are still waiting for this timeout
	if numProposals > orderer.proposal.proposalTimeout {

		// Save the number of proposals for which the timeout fired.
		orderer.proposal.proposalTimeout = numProposals

		// If this was the last bit missing, start a new proposal.
		if orderer.canPropose() {

			// If a proposal already has been set as a parameter of the segment, propose it directly.
			sn := orderer.segment.SeqNrs()[orderer.proposal.proposalsMade]
			if proposal := orderer.segment.Proposals[sn]; proposal != nil {
				return orderer.propose(proposal)
			}

			// Otherwise, obtain a fresh availability certificate first.
			return orderer.requestNewCert(), nil
		}
	}

	return events.EmptyList(), nil
}

// requestNewCert asks (by means of a CertRequest event) the availability module to provide a new availability certificate.
// When the certificate is ready, it must be passed to the Orderer using the CertReady event.
func (orderer *Orderer) requestNewCert() *events.EventList {

	// Set a flag indicating that a certificate has been requested,
	// so that no new certificates will be requested before the reception of this one.
	// It will be cleared when CertReady is received.
	orderer.proposal.certRequested = true

	// Remember the view in which the certificate has been requested
	// to make sure that we are still in the same view when the certificate becomes ready.
	orderer.proposal.certRequestedView = orderer.view

	// Emit the CertRequest event.
	// Operation continues on reception of the CertReady event.
	return orderer.requestCertOrigin()
}

// applyCertReady processes a new availability certificate ready to be proposed.
// This event is triggered by availability module in response to the CertRequest event produced by this Orderer.
func (orderer *Orderer) applyCertReady(cert *availabilitypb.Cert) (*events.EventList, error) {
	eventsOut := events.EmptyList()

	// Clear flag that was set in requestNewCert(), so that new certificates can be requested if necessary.
	orderer.proposal.certRequested = false

	if orderer.proposal.certRequestedView == orderer.view {
		// If the protocol is still in the same PBFT view as when the certificate was requested,
		// propose the received certificate.

		certBytes, err := proto.Marshal(cert)
		if err != nil {
			return nil, fmt.Errorf("error marshalling certificate: %w", err)
		}

		l, err := orderer.propose(certBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to propose: %w", err)
		}
		eventsOut.PushBackList(l)
	} else { // nolint:staticcheck,revive
		// If the PBFT view advanced since the certificate was requested,
		// do not propose the certificate and resurrect the requests it contains.
		// eventsOut.PushBack(orderer.eventService.SBEvent(SBResurrectBatchEvent(batch.Batch)))
		// TODO: Implement resurrection!
	}

	return eventsOut, nil
}

// propose proposes a new availability certificate by sending a Preprepare message.
// propose assumes that the state of the PBFT Orderer allows sending a new proposal
// and does not perform any checks in this regard.
func (orderer *Orderer) propose(data []byte) (*events.EventList, error) {

	// Update proposal counter.
	sn := orderer.segment.SeqNrs()[orderer.proposal.proposalsMade]
	orderer.proposal.proposalsMade++

	// Log debug message.
	orderer.logger.Log(logging.LevelDebug, "Proposing.",
		"sn", sn)

	// Send a Preprepare message.
	// No need for periodic re-transmission.
	// In the worst case, dropping of these messages may result in a view change, but will not compromise correctness.
	msgSendEvent := transportpbevents.SendMessage(orderer.moduleConfig.Net,
		pbftpbmsgs.Preprepare(orderer.moduleConfig.Self, sn, orderer.view, data, false),
		orderer.segment.NodeIDs())

	// Set up a new timer for the next proposal.
	ordererEvent := eventpbtypes.EventFromPb(OrdererEvent(orderer.moduleConfig.Self,
		PbftProposeTimeout(uint64(orderer.proposal.proposalsMade+1))))

	timerEvent := eventpbevents.TimerDelay(orderer.moduleConfig.Timer,
		[]*eventpbtypes.Event{ordererEvent},
		types.Duration(orderer.config.MaxProposeDelay),
	).Pb()

	return events.ListOf(msgSendEvent.Pb(), timerEvent), nil
}

// applyMsgPreprepare applies a received preprepare message.
// It performs the necessary checks and, if successful,
// requests a confirmation from ISS that all contained requests have been received and authenticated.
func (orderer *Orderer) applyMsgPreprepare(preprepare *pbftpbtypes.Preprepare, from t.NodeID) *events.EventList {

	// Convenience variable
	sn := preprepare.Sn

	// Preprocess message, looking up the corresponding pbftSlot.
	slot := orderer.preprocessMessage(sn, preprepare.View, preprepare.Pb(), from)
	if slot == nil {
		// If preprocessing does not return a pbftSlot, the message cannot be processed right now.
		return events.EmptyList()
	}

	// Check whether valid data has been proposed.
	if err := orderer.externalValidator.Check(preprepare.Data); err != nil {
		orderer.logger.Log(logging.LevelWarn, "Ignoring Preprepare message with invalid proposal.",
			"sn", sn, "from", from, "err", err)
		return events.EmptyList()
	}

	// Check that this is the first Preprepare message received.
	// Note that checking the orderer.Preprepared flag instead would be incorrect,
	// as that flag is only set upon receiving the RequestsReady OrdererEvent.
	if slot.Preprepare != nil {
		orderer.logger.Log(logging.LevelDebug, "Ignoring Preprepare message. Already preprepared or prepreparing.",
			"sn", sn, "from", from)
		return events.EmptyList()
	}

	// Check whether the sender is the legitimate leader of the segment in the current view.
	primary := orderer.segment.PrimaryNode(orderer.view)
	if from != primary {
		orderer.logger.Log(logging.LevelWarn, "Ignoring Preprepare message. Invalid Leader.",
			"expectedLeader", primary,
			"sender", from,
		)
		return events.EmptyList()
	}

	slot.Preprepare = preprepare

	// Request the computation of the hash of the Preprepare message.
	return events.ListOf(hasherpbevents.Request(
		orderer.moduleConfig.Hasher,
		[]*commonpbtypes.HashData{serializePreprepareForHashing(preprepare)},
		HashOrigin(orderer.moduleConfig.Self, preprepareHashOrigin(preprepare.Pb())),
	).Pb())
}

func (orderer *Orderer) applyPreprepareHashResult(digest []byte, preprepare *pbftpb.Preprepare) *events.EventList {
	eventsOut := events.EmptyList()

	// Convenience variable.
	sn := tt.SeqNr(preprepare.Sn)

	// Stop processing the Preprepare if view advanced in the meantime.
	if orderertypes.ViewNr(preprepare.View) < orderer.view {
		return events.EmptyList()
	}

	// Save the digest of the Preprepare message and mark the slot as preprepared.
	slot := orderer.slots[orderer.view][sn]
	slot.Digest = digest
	slot.Preprepared = true

	// Send a Prepare message.
	eventsOut.PushBackList(orderer.sendPrepare(sn, orderer.view, digest))

	// Advance the state of the pbftSlot even more if necessary
	// (potentially sending a Commit message or even delivering).
	// This is required for the case when the Preprepare message arrives late.
	eventsOut.PushBackList(slot.advanceState(orderer, sn))

	return eventsOut
}

// sendPrepare creates an event for sending a Prepare message.
func (orderer *Orderer) sendPrepare(sn tt.SeqNr, view orderertypes.ViewNr, digest []byte) *events.EventList {

	// Return a list with a single element - the send event containing a PBFT prepare message.
	// No need for periodic re-transmission.
	// In the worst case, dropping of these messages may result in a view change, but will not compromise correctness.
	return events.ListOf(transportpbevents.SendMessage(
		orderer.moduleConfig.Net,
		pbftpbmsgs.Prepare(orderer.moduleConfig.Self, sn, view, digest),
		orderer.segment.NodeIDs(),
	).Pb())
}

// applyMsgPrepare applies a received prepare message.
// It performs the necessary checks and, if successful,
// may trigger additional events like the sending of a Commit message.
func (orderer *Orderer) applyMsgPrepare(prepare *pbftpbtypes.Prepare, from t.NodeID) *events.EventList {

	// Convenience variable
	sn := prepare.Sn

	// Preprocess message, looking up the corresponding pbftSlot.
	slot := orderer.preprocessMessage(sn, prepare.View, prepare.Pb(), from)
	if slot == nil {
		// If preprocessing does not return a pbftSlot, the message cannot be processed right now.
		return events.EmptyList()
	}

	// Check if a Prepare message has already been received from this node in the current view.
	if _, ok := slot.Prepares[from]; ok {
		orderer.logger.Log(logging.LevelDebug, "Ignoring Prepare message. Already received in this view.",
			"sn", sn, "from", from, "view", orderer.view)
		return events.EmptyList()
	}

	// Save the received Prepare message and advance the slot state
	// (potentially sending a Commit message or even delivering).
	slot.Prepares[from] = prepare
	return slot.advanceState(orderer, sn)
}

// sendCommit creates an event for sending a Commit message.
func (orderer *Orderer) sendCommit(sn tt.SeqNr, view orderertypes.ViewNr, digest []byte) *events.EventList {

	// Emit a send event with a PBFT Commit message.
	// No need for periodic re-transmission.
	// In the worst case, dropping of these messages may result in a view change, but will not compromise correctness.
	return events.ListOf(transportpbevents.SendMessage(
		orderer.moduleConfig.Net,
		pbftpbmsgs.Commit(orderer.moduleConfig.Self, sn, view, digest),
		orderer.segment.NodeIDs()).Pb(),
	)
}

// applyMsgCommit applies a received commit message.
// It performs the necessary checks and, if successful,
// may trigger additional events like delivering the corresponding certificate.
func (orderer *Orderer) applyMsgCommit(commit *pbftpbtypes.Commit, from t.NodeID) *events.EventList {

	// Convenience variable
	sn := commit.Sn

	// Preprocess message, looking up the corresponding pbftSlot.
	slot := orderer.preprocessMessage(sn, commit.View, commit.Pb(), from)
	if slot == nil {
		// If preprocessing does not return a pbftSlot, the message cannot be processed right now.
		return events.EmptyList()
	}

	// Check if a Commit message has already been received from this node in the current view.
	if _, ok := slot.Commits[from]; ok {
		orderer.logger.Log(logging.LevelDebug, "Ignoring Commit message. Already received in this view.",
			"sn", sn, "from", from, "view", orderer.view)
		return events.EmptyList()
	}

	// Save the received Commit message and advance the slot state
	// (potentially delivering the corresponding certificate and its successors).
	slot.Commits[from] = commit
	return slot.advanceState(orderer, sn)
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
