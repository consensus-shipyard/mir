/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// This file provides constructors for protobuf messages (also used to represent events) used by the ISS PBFT orderer.
// The primary purpose is convenience and improved readability of the PBFT code,
// As creating protobuf objects is rather verbose in Go.
// Moreover, in case the definitions of some protocol buffers change,
// this file should be the only one that will potentially need to change.
// TODO: When PBFT is moved to a different package, remove the Pbft prefix form the function names defined in this file.

// TODO: Write documentation comments for the functions in this file.
//       Part of the text can probably be copy-pasted from the documentation of the functions handling those events.

package iss

import (
	"github.com/filecoin-project/mir/pkg/pb/isspb"
	"github.com/filecoin-project/mir/pkg/pb/isspbftpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	"github.com/filecoin-project/mir/pkg/serializing"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ============================================================
// Events
// ============================================================

func PbftPersistPreprepare(preprepare *isspbftpb.Preprepare) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_PbftPersistPreprepare{
		PbftPersistPreprepare: preprepare,
	}}
}

func PbftPersistPrepare(prepare *isspbftpb.Prepare) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_PbftPersistPrepare{
		PbftPersistPrepare: prepare,
	}}
}

func PbftPersistCommit(commit *isspbftpb.Commit) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_PbftPersistCommit{
		PbftPersistCommit: commit,
	}}
}

func PbftPersistSignedViewChange(signedViewChange *isspbftpb.SignedViewChange) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_PbftPersistSignedViewChange{
		PbftPersistSignedViewChange: signedViewChange,
	}}
}

func PbftPersistNewView(newView *isspbftpb.NewView) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_PbftPersistNewView{
		PbftPersistNewView: newView,
	}}
}

// TODO: Generalize the Preprepare, Prepare, Commit, etc... persist events to one (PersistMessage)

func PbftProposeTimeout(numProposals uint64) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_PbftProposeTimeout{
		PbftProposeTimeout: numProposals,
	}}
}

func PbftViewChangeBatchTimeout(view t.PBFTViewNr, numCommitted int) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_PbftViewChangeBatchTimeout{
		PbftViewChangeBatchTimeout: &isspbftpb.VCBatchTimeout{
			View:         view.Pb(),
			NumCommitted: uint64(numCommitted),
		},
	}}
}

func PbftViewChangeSegmentTimeout(view t.PBFTViewNr) *isspb.SBInstanceEvent {
	return &isspb.SBInstanceEvent{Type: &isspb.SBInstanceEvent_PbftViewChangeSegTimeout{
		PbftViewChangeSegTimeout: view.Pb(),
	}}
}

// ============================================================
// SB Instance Messages
// ============================================================

func PbftPreprepareSBMessage(content *isspbftpb.Preprepare) *isspb.SBInstanceMessage {
	return &isspb.SBInstanceMessage{Type: &isspb.SBInstanceMessage_PbftPreprepare{
		PbftPreprepare: content,
	}}
}

func PbftPrepareSBMessage(content *isspbftpb.Prepare) *isspb.SBInstanceMessage {
	return &isspb.SBInstanceMessage{Type: &isspb.SBInstanceMessage_PbftPrepare{
		PbftPrepare: content,
	}}
}

func PbftCommitSBMessage(content *isspbftpb.Commit) *isspb.SBInstanceMessage {
	return &isspb.SBInstanceMessage{Type: &isspb.SBInstanceMessage_PbftCommit{
		PbftCommit: content,
	}}
}

func PbftSignedViewChangeSBMessage(signedViewChange *isspbftpb.SignedViewChange) *isspb.SBInstanceMessage {
	return &isspb.SBInstanceMessage{Type: &isspb.SBInstanceMessage_PbftSignedViewChange{
		PbftSignedViewChange: signedViewChange,
	}}
}

func PbftPreprepareRequestSBMessage(preprepareRequest *isspbftpb.PreprepareRequest) *isspb.SBInstanceMessage {
	return &isspb.SBInstanceMessage{Type: &isspb.SBInstanceMessage_PbftPreprepareRequest{
		PbftPreprepareRequest: preprepareRequest,
	}}
}

func PbftMissingPreprepareSBMessage(preprepare *isspbftpb.Preprepare) *isspb.SBInstanceMessage {
	return &isspb.SBInstanceMessage{Type: &isspb.SBInstanceMessage_PbftMissingPreprepare{
		PbftMissingPreprepare: preprepare,
	}}
}

func PbftNewViewSBMessage(newView *isspbftpb.NewView) *isspb.SBInstanceMessage {
	return &isspb.SBInstanceMessage{Type: &isspb.SBInstanceMessage_PbftNewView{
		PbftNewView: newView,
	}}
}

// ============================================================
// PBFT Message
// ============================================================

// pbftPreprepareMsg returns a protocol buffer representing a Preprepare message.
// Instead of constructing the protocol buffer directly in the code, this function serves the purpose
// of enforcing that all fields are explicitly set and none is forgotten.
// Should the structure of the message change (e.g. by augmenting it by new fields),
// using this function ensures that these the message is always constructed properly.
func pbftPreprepareMsg(sn t.SeqNr, view t.PBFTViewNr, batch *requestpb.Batch, aborted bool) *isspbftpb.Preprepare {
	return &isspbftpb.Preprepare{
		Sn:      sn.Pb(),
		View:    view.Pb(),
		Batch:   batch,
		Aborted: aborted,
	}
}

// pbftPrepareMsg returns a protocol buffer representing a Prepare message.
// Analogous pbftPreprepareMsg, but for a Prepare message.
func pbftPrepareMsg(sn t.SeqNr, view t.PBFTViewNr, digest []byte) *isspbftpb.Prepare {
	return &isspbftpb.Prepare{
		Sn:     sn.Pb(),
		View:   view.Pb(),
		Digest: digest,
	}
}

// pbftCommitMsg returns a protocol buffer representing a Commit message.
// Analogous pbftPreprepareMsg, but for a Commit message.
func pbftCommitMsg(sn t.SeqNr, view t.PBFTViewNr, digest []byte) *isspbftpb.Commit {
	return &isspbftpb.Commit{
		Sn:     sn.Pb(),
		View:   view.Pb(),
		Digest: digest,
	}
}

// pbftViewChangeMsg returns a protocol buffer representing a ViewChange message.
func pbftViewChangeMsg(view t.PBFTViewNr, pSet viewChangePSet, qSet viewChangeQSet) *isspbftpb.ViewChange {
	return &isspbftpb.ViewChange{
		View: view.Pb(),
		PSet: pSet.Pb(),
		QSet: qSet.Pb(),
	}
}

// pbftSignedViewChangeMsg returns a protocol buffer representing a SignedViewChange message.
func pbftSignedViewChangeMsg(viewChange *isspbftpb.ViewChange, signature []byte) *isspbftpb.SignedViewChange {
	return &isspbftpb.SignedViewChange{
		ViewChange: viewChange,
		Signature:  signature,
	}
}

func pbftPreprepareRequestMsg(sn t.SeqNr, digest []byte) *isspbftpb.PreprepareRequest {
	return &isspbftpb.PreprepareRequest{
		Digest: digest,
		Sn:     sn.Pb(),
	}
}

// pbftNewViewMsg returns a protocol buffer representing a NewView message.
func pbftNewViewMsg(
	view t.PBFTViewNr,
	viewChangeSenders []t.NodeID,
	viewChanges []*isspbftpb.SignedViewChange,
	preprepareSeqNrs []t.SeqNr,
	preprepares []*isspbftpb.Preprepare,
) *isspbftpb.NewView {
	return &isspbftpb.NewView{
		View:              view.Pb(),
		ViewChangeSenders: t.NodeIDSlicePb(viewChangeSenders),
		SignedViewChanges: viewChanges,
		PreprepareSeqNrs:  t.SeqNrSlicePb(preprepareSeqNrs),
		Preprepares:       preprepares,
	}
}

// ============================================================
// Hashing and signing origins
// ============================================================

func preprepareHashOrigin(preprepare *isspbftpb.Preprepare) *isspb.SBInstanceHashOrigin {
	return &isspb.SBInstanceHashOrigin{Type: &isspb.SBInstanceHashOrigin_PbftPreprepare{
		PbftPreprepare: preprepare,
	}}
}

func missingPreprepareHashOrigin(preprepare *isspbftpb.Preprepare) *isspb.SBInstanceHashOrigin {
	return &isspb.SBInstanceHashOrigin{Type: &isspb.SBInstanceHashOrigin_PbftMissingPreprepare{
		PbftMissingPreprepare: preprepare,
	}}
}

func viewChangeSignOrigin(viewChange *isspbftpb.ViewChange) *isspb.SBInstanceSignOrigin {
	return &isspb.SBInstanceSignOrigin{Type: &isspb.SBInstanceSignOrigin_PbftViewChange{
		PbftViewChange: viewChange,
	}}
}

func viewChangeSigVerOrigin(viewChange *isspbftpb.SignedViewChange) *isspb.SBInstanceSigVerOrigin {
	return &isspb.SBInstanceSigVerOrigin{Type: &isspb.SBInstanceSigVerOrigin_PbftSignedViewChange{
		PbftSignedViewChange: viewChange,
	}}
}

func emptyPreprepareHashOrigin(view t.PBFTViewNr) *isspb.SBInstanceHashOrigin {
	return &isspb.SBInstanceHashOrigin{Type: &isspb.SBInstanceHashOrigin_PbftEmptyPreprepares{
		PbftEmptyPreprepares: view.Pb(),
	}}
}

func newViewSigVerOrigin(newView *isspbftpb.NewView) *isspb.SBInstanceSigVerOrigin {
	return &isspb.SBInstanceSigVerOrigin{Type: &isspb.SBInstanceSigVerOrigin_PbftNewView{
		PbftNewView: newView,
	}}
}

func newViewHashOrigin(newView *isspbftpb.NewView) *isspb.SBInstanceHashOrigin {
	return &isspb.SBInstanceHashOrigin{Type: &isspb.SBInstanceHashOrigin_PbftNewView{
		PbftNewView: newView,
	}}
}

// ============================================================
// Serialization
// ============================================================

// serializePreprepareForHashing returns a slice of byte slices representing the contents of a Preprepare message
// that can be passed to the Hasher module.
// Note that the view number is *not* serialized, as hashes must be consistent across views.
// Even though the preprepare argument is a protocol buffer, this function is required to guarantee
// that the serialization is deterministic, since the protobuf native serialization does not provide this guarantee.
func serializePreprepareForHashing(preprepare *isspbftpb.Preprepare) [][]byte {

	// Encode boolean Aborted field as one byte.
	aborted := byte(0)
	if preprepare.Aborted {
		aborted = 1
	}

	// Encode the batch content.
	batchData := serializing.BatchForHash(preprepare.Batch)

	// Put everything together in a slice and return it.
	// Note that we do not include the view number,
	// as the view change protocol might compare hashes of Preprepares across vies.
	data := make([][]byte, 0, len(preprepare.Batch.Requests)+3)
	data = append(data, t.SeqNr(preprepare.Sn).Bytes(), []byte{aborted})
	data = append(data, batchData...)
	return data
}

func serializeViewChangeForSigning(vc *isspbftpb.ViewChange) [][]byte {
	_ = &isspbftpb.ViewChange{
		View: 0,
		PSet: nil,
		QSet: nil,
	}

	// Allocate result data structure.
	data := make([][]byte, 0)

	// Encode view number.
	data = append(data, t.PBFTViewNr(vc.View).Bytes())

	// Encode P set.
	for _, pSetEntry := range vc.PSet {
		data = append(data, t.SeqNr(pSetEntry.Sn).Bytes())
		data = append(data, t.PBFTViewNr(pSetEntry.View).Bytes())
		data = append(data, pSetEntry.Digest)
	}

	// Encode Q set.
	for _, qSetEntry := range vc.QSet {
		data = append(data, t.SeqNr(qSetEntry.Sn).Bytes())
		data = append(data, t.PBFTViewNr(qSetEntry.View).Bytes())
		data = append(data, qSetEntry.Digest)
	}

	return data
}
