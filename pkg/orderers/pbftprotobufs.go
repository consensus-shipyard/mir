/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// This file provides constructors for protobuf messages (also used to represent events) used by the ISS PBFT Orderer.
// The primary purpose is convenience and improved readability of the PBFT code,
// As creating protobuf objects is rather verbose in Go.
// Moreover, in case the definitions of some protocol buffers change,
// this file should be the only one that will potentially need to change.
// TODO: When PBFT is moved to a different package, remove the Pbft prefix form the function names defined in this file.

// TODO: Write documentation comments for the functions in this file.
//       Part of the text can probably be copy-pasted from the documentation of the functions handling those events.

package orderers

import (
	"github.com/filecoin-project/mir/pkg/pb/ordererspb"
	"github.com/filecoin-project/mir/pkg/pb/ordererspbftpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ============================================================
// Events
// ============================================================

func PbftProposeTimeout(numProposals uint64) *ordererspb.SBInstanceEvent {
	return &ordererspb.SBInstanceEvent{Type: &ordererspb.SBInstanceEvent_PbftProposeTimeout{
		PbftProposeTimeout: numProposals,
	}}
}

func PbftViewChangeSNTimeout(view t.PBFTViewNr, numCommitted int) *ordererspb.SBInstanceEvent {
	return &ordererspb.SBInstanceEvent{Type: &ordererspb.SBInstanceEvent_PbftViewChangeSnTimeout{
		PbftViewChangeSnTimeout: &ordererspbftpb.VCSNTimeout{
			View:         view.Pb(),
			NumCommitted: uint64(numCommitted),
		},
	}}
}

func PbftViewChangeSegmentTimeout(view t.PBFTViewNr) *ordererspb.SBInstanceEvent {
	return &ordererspb.SBInstanceEvent{Type: &ordererspb.SBInstanceEvent_PbftViewChangeSegTimeout{
		PbftViewChangeSegTimeout: view.Pb(),
	}}
}

// ============================================================
// SB Instance Messages
// ============================================================

func PbftPreprepareSBMessage(content *ordererspbftpb.Preprepare) *ordererspb.SBInstanceMessage {
	return &ordererspb.SBInstanceMessage{Type: &ordererspb.SBInstanceMessage_PbftPreprepare{
		PbftPreprepare: content,
	}}
}

func PbftPrepareSBMessage(content *ordererspbftpb.Prepare) *ordererspb.SBInstanceMessage {
	return &ordererspb.SBInstanceMessage{Type: &ordererspb.SBInstanceMessage_PbftPrepare{
		PbftPrepare: content,
	}}
}

func PbftCommitSBMessage(content *ordererspbftpb.Commit) *ordererspb.SBInstanceMessage {
	return &ordererspb.SBInstanceMessage{Type: &ordererspb.SBInstanceMessage_PbftCommit{
		PbftCommit: content,
	}}
}

func PbftSignedViewChangeSBMessage(signedViewChange *ordererspbftpb.SignedViewChange) *ordererspb.SBInstanceMessage {
	return &ordererspb.SBInstanceMessage{Type: &ordererspb.SBInstanceMessage_PbftSignedViewChange{
		PbftSignedViewChange: signedViewChange,
	}}
}

func PbftPreprepareRequestSBMessage(preprepareRequest *ordererspbftpb.PreprepareRequest) *ordererspb.SBInstanceMessage {
	return &ordererspb.SBInstanceMessage{Type: &ordererspb.SBInstanceMessage_PbftPreprepareRequest{
		PbftPreprepareRequest: preprepareRequest,
	}}
}

func PbftMissingPreprepareSBMessage(preprepare *ordererspbftpb.Preprepare) *ordererspb.SBInstanceMessage {
	return &ordererspb.SBInstanceMessage{Type: &ordererspb.SBInstanceMessage_PbftMissingPreprepare{
		PbftMissingPreprepare: preprepare,
	}}
}

func PbftNewViewSBMessage(newView *ordererspbftpb.NewView) *ordererspb.SBInstanceMessage {
	return &ordererspb.SBInstanceMessage{Type: &ordererspb.SBInstanceMessage_PbftNewView{
		PbftNewView: newView,
	}}
}

func PbftDoneSBMessage(digests [][]byte) *ordererspb.SBInstanceMessage {
	return &ordererspb.SBInstanceMessage{Type: &ordererspb.SBInstanceMessage_PbftDone{
		PbftDone: &ordererspbftpb.Done{
			Digests: digests,
		},
	}}
}

func PbftCatchUpRequestSBMessage(sn t.SeqNr, digest []byte) *ordererspb.SBInstanceMessage {
	return &ordererspb.SBInstanceMessage{Type: &ordererspb.SBInstanceMessage_PbftCatchUpRequest{
		PbftCatchUpRequest: &ordererspbftpb.CatchUpRequest{
			Digest: digest,
			Sn:     sn.Pb(),
		},
	}}
}

func PbftCatchUpResponseSBMessage(preprepare *ordererspbftpb.Preprepare) *ordererspb.SBInstanceMessage {
	return &ordererspb.SBInstanceMessage{Type: &ordererspb.SBInstanceMessage_PbftCatchUpResponse{
		PbftCatchUpResponse: preprepare,
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
func pbftPreprepareMsg(sn t.SeqNr, view t.PBFTViewNr, certData []byte, aborted bool) *ordererspbftpb.Preprepare {
	return &ordererspbftpb.Preprepare{
		Sn:       sn.Pb(),
		View:     view.Pb(),
		CertData: certData,
		Aborted:  aborted,
	}
}

// pbftPrepareMsg returns a protocol buffer representing a Prepare message.
// Analogous pbftPreprepareMsg, but for a Prepare message.
func pbftPrepareMsg(sn t.SeqNr, view t.PBFTViewNr, digest []byte) *ordererspbftpb.Prepare {
	return &ordererspbftpb.Prepare{
		Sn:     sn.Pb(),
		View:   view.Pb(),
		Digest: digest,
	}
}

// pbftCommitMsg returns a protocol buffer representing a Commit message.
// Analogous pbftPreprepareMsg, but for a Commit message.
func pbftCommitMsg(sn t.SeqNr, view t.PBFTViewNr, digest []byte) *ordererspbftpb.Commit {
	return &ordererspbftpb.Commit{
		Sn:     sn.Pb(),
		View:   view.Pb(),
		Digest: digest,
	}
}

// pbftViewChangeMsg returns a protocol buffer representing a ViewChange message.
func pbftViewChangeMsg(view t.PBFTViewNr, pSet viewChangePSet, qSet viewChangeQSet) *ordererspbftpb.ViewChange {
	return &ordererspbftpb.ViewChange{
		View: view.Pb(),
		PSet: pSet.Pb(),
		QSet: qSet.Pb(),
	}
}

// pbftSignedViewChangeMsg returns a protocol buffer representing a SignedViewChange message.
func pbftSignedViewChangeMsg(viewChange *ordererspbftpb.ViewChange, signature []byte) *ordererspbftpb.SignedViewChange {
	return &ordererspbftpb.SignedViewChange{
		ViewChange: viewChange,
		Signature:  signature,
	}
}

func pbftPreprepareRequestMsg(sn t.SeqNr, digest []byte) *ordererspbftpb.PreprepareRequest {
	return &ordererspbftpb.PreprepareRequest{
		Digest: digest,
		Sn:     sn.Pb(),
	}
}

// pbftNewViewMsg returns a protocol buffer representing a NewView message.
func pbftNewViewMsg(
	view t.PBFTViewNr,
	viewChangeSenders []t.NodeID,
	viewChanges []*ordererspbftpb.SignedViewChange,
	preprepareSeqNrs []t.SeqNr,
	preprepares []*ordererspbftpb.Preprepare,
) *ordererspbftpb.NewView {
	return &ordererspbftpb.NewView{
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

func preprepareHashOrigin(preprepare *ordererspbftpb.Preprepare) *ordererspb.SBInstanceHashOrigin {
	return &ordererspb.SBInstanceHashOrigin{Type: &ordererspb.SBInstanceHashOrigin_PbftPreprepare{
		PbftPreprepare: preprepare,
	}}
}

func missingPreprepareHashOrigin(preprepare *ordererspbftpb.Preprepare) *ordererspb.SBInstanceHashOrigin {
	return &ordererspb.SBInstanceHashOrigin{Type: &ordererspb.SBInstanceHashOrigin_PbftMissingPreprepare{
		PbftMissingPreprepare: preprepare,
	}}
}

func viewChangeSignOrigin(viewChange *ordererspbftpb.ViewChange) *ordererspb.SBInstanceSignOrigin {
	return &ordererspb.SBInstanceSignOrigin{Type: &ordererspb.SBInstanceSignOrigin_PbftViewChange{
		PbftViewChange: viewChange,
	}}
}

func viewChangeSigVerOrigin(viewChange *ordererspbftpb.SignedViewChange) *ordererspb.SBInstanceSigVerOrigin {
	return &ordererspb.SBInstanceSigVerOrigin{Type: &ordererspb.SBInstanceSigVerOrigin_PbftSignedViewChange{
		PbftSignedViewChange: viewChange,
	}}
}

func emptyPreprepareHashOrigin(view t.PBFTViewNr) *ordererspb.SBInstanceHashOrigin {
	return &ordererspb.SBInstanceHashOrigin{Type: &ordererspb.SBInstanceHashOrigin_PbftEmptyPreprepares{
		PbftEmptyPreprepares: view.Pb(),
	}}
}

func newViewSigVerOrigin(newView *ordererspbftpb.NewView) *ordererspb.SBInstanceSigVerOrigin {
	return &ordererspb.SBInstanceSigVerOrigin{Type: &ordererspb.SBInstanceSigVerOrigin_PbftNewView{
		PbftNewView: newView,
	}}
}

func newViewHashOrigin(newView *ordererspbftpb.NewView) *ordererspb.SBInstanceHashOrigin {
	return &ordererspb.SBInstanceHashOrigin{Type: &ordererspb.SBInstanceHashOrigin_PbftNewView{
		PbftNewView: newView,
	}}
}

func catchUpResponseHashOrigin(preprepare *ordererspbftpb.Preprepare) *ordererspb.SBInstanceHashOrigin {
	return &ordererspb.SBInstanceHashOrigin{Type: &ordererspb.SBInstanceHashOrigin_PbftCatchUpResponse{
		PbftCatchUpResponse: preprepare,
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
func serializePreprepareForHashing(preprepare *ordererspbftpb.Preprepare) [][]byte {

	// Encode boolean Aborted field as one byte.
	aborted := byte(0)
	if preprepare.Aborted {
		aborted = 1
	}

	// TODO: Implement deterministic cert serialization and use a cert directly

	// Put everything together in a slice and return it.
	// Note that we do not include the view number,
	// as the view change protocol might compare hashes of Preprepares across vies.
	return [][]byte{t.SeqNr(preprepare.Sn).Bytes(), {aborted}, preprepare.CertData}
}

func serializeViewChangeForSigning(vc *ordererspbftpb.ViewChange) [][]byte {
	_ = &ordererspbftpb.ViewChange{
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
