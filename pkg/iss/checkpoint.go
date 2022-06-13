/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// TODO: Eventually make the checkpoint tracker a separate package.
//       Then, use an EventService for producing Events.

// TODO: Finish writing proper comments in this file.

package iss

import (
	"bytes"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/isspb"
	"github.com/filecoin-project/mir/pkg/serializing"
	t "github.com/filecoin-project/mir/pkg/types"
)

// checkpointTracker represents the state associated with a single instance of the checkpoint protocol
// (establishing a single stable checkpoint).
type checkpointTracker struct {
	logging.Logger

	// The ID of the node executing this instance of the protocol.
	ownID t.NodeID

	// Epoch to which this checkpoint belongs.
	// It is always the epoch the checkpoint's associated sequence number (seqNr) is part of.
	epoch t.EpochNr

	// Sequence number associated with this checkpoint protocol instance.
	// This checkpoint encompasses seqNr sequence numbers,
	// i.e., seqNr is the first sequence number *not* encompassed by this checkpoint.
	// One can imagine that the checkpoint represents the state of the system just before seqNr,
	// i.e., "between" seqNr-1 and seqNr.
	seqNr t.SeqNr

	// The IDs of nodes to execute this instance of the checkpoint protocol.
	membership []t.NodeID

	// Application snapshot data associated with this checkpoint.
	appSnapshot []byte

	// Hash of the application snapshot data associated with this checkpoint.
	appSnapshotHash []byte

	// Set of (potentially invalid) nodes' signatures.
	signatures map[t.NodeID][]byte

	// Set of nodes from which a valid Checkpoint messages has been received.
	confirmations map[t.NodeID]struct{}

	// Set of Checkpoint messages that were received ahead of time.
	pendingMessages map[t.NodeID]*isspb.Checkpoint
}

// newCheckpointTracker allocates and returns a new instance of a checkpointTracker associated with sequence number sn.
func newCheckpointTracker(ownID t.NodeID, sn t.SeqNr, epoch t.EpochNr, logger logging.Logger) *checkpointTracker {
	return &checkpointTracker{
		Logger:          logger,
		ownID:           ownID,
		seqNr:           sn,
		epoch:           epoch,
		signatures:      make(map[t.NodeID][]byte),
		confirmations:   make(map[t.NodeID]struct{}),
		pendingMessages: make(map[t.NodeID]*isspb.Checkpoint),
		// the membership field will be set later by iss.startCheckpoint
		// the appSnapshot field will be set by ProcessAppSnapshot
	}
}

// Start initiates the checkpoint protocol among nodes in membership.
// The checkpoint to be produced encompasses all currently delivered sequence numbers.
// If Start is called during epoch transition,
// it must be called with the old epoch's membership.
func (ct *checkpointTracker) Start(membership []t.NodeID) *events.EventList {

	// Save the membership this instance of the checkpoint protocol will use.
	// This is required in case where the membership changes before the checkpoint sub-protocol finishes.
	// That is also why the content of the Membership slice needs to be copied.
	ct.membership = make([]t.NodeID, len(membership))
	copy(ct.membership, membership)

	// Request a snapshot of the application state.
	// TODO: also get a snapshot of the shared state
	return (&events.EventList{}).PushBack(events.AppSnapshotRequest("app", "iss", ct.epoch))
}

func (ct *checkpointTracker) ProcessAppSnapshot(snapshot []byte) *events.EventList {

	// Save received snapshot
	ct.appSnapshot = snapshot

	// Initiate computing the hash of the snapshot
	hashEvent := events.HashRequest("hasher", [][][]byte{{snapshot}}, AppSnapshotHashOrigin(ct.epoch))

	return (&events.EventList{}).PushBack(hashEvent)
}

func (ct *checkpointTracker) ProcessAppSnapshotHash(snapshotHash []byte) *events.EventList {

	// Save the received snapshot hash
	ct.appSnapshotHash = snapshotHash

	// Request signature
	sigData := serializing.CheckpointForSig(ct.epoch, ct.seqNr, snapshotHash)
	sigEvent := events.SignRequest("crypto", sigData, CheckpointSignOrigin(ct.epoch))

	return (&events.EventList{}).PushBack(sigEvent)
}

func (ct *checkpointTracker) ProcessCheckpointSignResult(signature []byte) *events.EventList {

	// Save received own checkpoint signature
	ct.signatures[ct.ownID] = signature
	ct.confirmations[ct.ownID] = struct{}{}

	// Write Checkpoint to WAL
	persistEvent := PersistCheckpointEvent(ct.seqNr, ct.appSnapshot, ct.appSnapshotHash, signature)
	walEvent := events.WALAppend("wal", persistEvent, t.WALRetIndex(ct.epoch))

	// Send a checkpoint message to all nodes after persisting checkpoint to the WAL.
	// TODO: Implement checkpoint message retransmission.
	m := CheckpointMessage(ct.epoch, ct.seqNr, ct.appSnapshotHash, signature)
	walEvent.FollowUp(events.SendMessage(m, ct.membership))

	// Apply pending Checkpoint messages
	for s, m := range ct.pendingMessages {
		walEvent.FollowUps(ct.applyMessage(m, s).Slice())
	}
	ct.pendingMessages = nil

	// Return resulting WALEvent (with the SendMessage event appended).
	return (&events.EventList{}).PushBack(walEvent)
}

func (ct *checkpointTracker) applyMessage(msg *isspb.Checkpoint, source t.NodeID) *events.EventList {

	// If checkpoint is already stable, ignore message.
	if ct.stable() {
		return &events.EventList{}
	}

	// Check snapshot hash
	if ct.appSnapshotHash == nil {
		// The message is received too early, put it aside
		ct.pendingMessages[source] = msg
		return &events.EventList{}
	} else if !bytes.Equal(ct.appSnapshotHash, msg.AppSnapshotHash) {
		// Snapshot hash mismatch
		ct.Log(logging.LevelWarn, "Ignoring Checkpoint message. Mismatching app snapshot hash.", "source", source)
		return &events.EventList{}
	}

	// TODO: Only accept messages from nodes in membership.
	//       This might be more tricky than it seems, especially when the membership is not yet initialized.

	// Ignore duplicate messages.
	if _, ok := ct.signatures[source]; ok {
		return &events.EventList{}
	}
	ct.signatures[source] = msg.Signature

	// Verify signature of the sender.
	sigData := serializing.CheckpointForSig(ct.epoch, ct.seqNr, ct.appSnapshotHash)
	verifySigEvent := events.VerifyNodeSigs(
		"crypto",
		[][][]byte{sigData},
		[][]byte{msg.Signature},
		[]t.NodeID{source},
		CheckpointSigVerOrigin(ct.epoch),
	)

	return (&events.EventList{}).PushBack(verifySigEvent)
}

func (ct *checkpointTracker) ProcessSigVerified(valid bool, err string, source t.NodeID) *events.EventList {

	if !valid {
		ct.Log(logging.LevelWarn, "Ignoring Checkpoint message. Invalid signature.", "source", source, "error", err)
		ct.signatures[source] = nil
		return &events.EventList{}
	}

	// Note the reception of a valid Checkpoint message from node `source`.
	ct.confirmations[source] = struct{}{}

	// If, after having applied this message, the checkpoint became stable, produce the necessary events.
	if ct.stable() {
		return ct.announceStable()
	}

	return &events.EventList{}
}

func (ct *checkpointTracker) stable() bool {
	return ct.appSnapshot != nil && len(ct.confirmations) >= strongQuorum(len(ct.membership))
}

func (ct *checkpointTracker) announceStable() *events.EventList {

	// Assemble a multisig certificate from the received signatures.
	cert := make(map[string][]byte)
	for node := range ct.confirmations {
		cert[node.Pb()] = ct.signatures[node]
	}

	// Create a stable checkpoint object.
	stableCheckpoint := &isspb.StableCheckpoint{
		Epoch:       ct.epoch.Pb(),
		Sn:          ct.seqNr.Pb(),
		AppSnapshot: ct.appSnapshot,
		Cert:        cert,
	}

	// First persist the checkpoint in the WAL, then announce it to the protocol.
	persistEvent := events.WALAppend("wal", PersistStableCheckpointEvent(stableCheckpoint), t.WALRetIndex(ct.epoch))
	persistEvent.FollowUp(StableCheckpointEvent(stableCheckpoint))
	return (&events.EventList{}).PushBack(persistEvent)
}
