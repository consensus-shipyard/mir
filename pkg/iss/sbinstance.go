/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package iss

import (
	availabilityevents "github.com/filecoin-project/mir/pkg/availability/events"
	"github.com/filecoin-project/mir/pkg/contextstore"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	"github.com/filecoin-project/mir/pkg/pb/contextstorepb"
	"github.com/filecoin-project/mir/pkg/pb/isspb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// sbInstance represents an instance of Sequenced Broadcast and is the type of each ISS orderer.
// Each orderer (being an sbInstance) is assigned a segment and is responsible for
// proposing and delivering request batches for all sequence numbers described by the segment.
type sbInstance interface {

	// ApplyEvent receives one event and applies it to the SB implementation's state machine,
	// potentially altering its state and producing a (potentially empty) list of more events
	// to be applied to other modules.
	// Since the SB instance is always part of ISS, it is only the ISS code that supplies events to this function.
	// The isspb.SBInstanceEvent type defines the events that can be exchanged between an SB instance and ISS.
	// The events returned from ApplyEvent must be produced by an sbEventService
	// injected to the SB instance at creation.
	ApplyEvent(event *isspb.SBInstanceEvent) *events.EventList

	// Segment returns the segment assigned to this SB instance.
	Segment() *segment
}

// ============================================================
// ISS methods handling SB instance events
// ============================================================

// applySBInstanceEvent applies one event produced by an orderer to the ISS state, potentially altering its state
// and producing a (potentially empty) list of events to be applied to other modules.
func (iss *ISS) applySBInstanceEvent(
	event *isspb.SBInstanceEvent,
	instance sbInstance,
) *events.EventList {
	switch e := event.Type.(type) {
	case *isspb.SBInstanceEvent_Deliver:
		return iss.applySBInstDeliver(instance, e.Deliver)
	case *isspb.SBInstanceEvent_CutBatch:
		return iss.applySBInstCutBatch(instance, t.NumRequests(e.CutBatch.MaxSize))
	case *isspb.SBInstanceEvent_ResurrectBatch:
		return iss.applySBInstResurrectBatch(e.ResurrectBatch)
	default:
		return instance.ApplyEvent(event)
	}
}

// applySBInstDeliver processes the event of an SB instance delivering a request batch (or the special abort value)
// for a sequence number. It creates a corresponding commitLog entry and requests the computation of its hash.
// Note that applySBInstDeliver does not yet insert the entry to the commitLog. This will be done later.
// Operation continues on reception of the HashResult event.
func (iss *ISS) applySBInstDeliver(instance sbInstance, deliver *isspb.SBDeliver) *events.EventList {

	// Create a new preliminary log entry based on the delivered batch and hash it.
	// Note that, although tempting, the hash used internally by the SB implementation cannot be re-used.
	// Apart from making the SB abstraction extremely leaky (reason enough not to do it), it would also be incorrect.
	// E.g., in PBFT, if the digest of the corresponding Preprepare message was used, the hashes at different nodes
	// might mismatch, if they commit in different PBFT views (and thus using different Preprepares).
	unhashedEntry := &CommitLogEntry{
		Sn:       t.SeqNr(deliver.Sn),
		CertData: deliver.CertData,
		Digest:   nil,
		Aborted:  deliver.Aborted,
		Suspect:  instance.Segment().Leader,
	}

	// Save the preliminary hash entry to a map where it can be looked up when the hash result arrives.
	iss.unhashedLogEntries[unhashedEntry.Sn] = unhashedEntry

	// Create a HashRequest for the commit log entry with the newly delivered hash.
	// The hash is required for state transfer.
	// Only after the hash is computed, the log entry can be stored in the log (and potentially delivered to the App).
	return events.ListOf(events.HashRequest(
		iss.moduleConfig.Hasher,
		[][][]byte{serializeLogEntryForHashing(unhashedEntry)},
		LogEntryHashOrigin(iss.moduleConfig.Self, unhashedEntry.Sn),
	))
}

// applySBInstCutBatch processes a request by an orderer for a new request batch that the orderer will propose.
// To this end, applySBInstCutBatch requests a new batch certificate from the availability layer.
func (iss *ISS) applySBInstCutBatch(instance sbInstance, maxBatchSize t.NumRequests) *events.EventList {
	return events.ListOf(availabilityevents.RequestCert(iss.moduleConfig.Avaliability, &availabilitypb.RequestCertOrigin{
		Module: iss.moduleConfig.Self.Pb(),
		Type: &availabilitypb.RequestCertOrigin_ContextStore{ContextStore: &contextstorepb.Origin{
			ItemID: iss.contextStore.Store(instance).Pb(),
		}},
	}))
}

func (iss *ISS) applyNewCert(newCert *availabilitypb.NewCert) (*events.EventList, error) {
	csID := contextstore.ItemID(
		newCert.Origin.Type.(*availabilitypb.RequestCertOrigin_ContextStore).ContextStore.ItemID,
	)
	instance := iss.contextStore.RecoverAndDispose(csID).(sbInstance)

	return instance.ApplyEvent(SBCertReadyEvent(newCert.Cert)), nil
}

// applySBInstResurrectBatch resurrects requests contained in a batch that was cut, but could not been proposed
// or committed. Through the ResurrectBatch event, an orderer "returns" a batch it was unable to order.
func (iss *ISS) applySBInstResurrectBatch(batch *requestpb.Batch) *events.EventList {

	// TODO: Implement resurrection (if appropriate).

	// No further actions to be performed.
	return events.EmptyList()
}
