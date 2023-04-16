/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// This file provides constructors for protobuf messages (also used to represent events) used by the ISS PBFT Orderer.
// The primary purpose is convenience and improved readability of the PBFT code,
// As creating protobuf objects is rather verbose in Go.
// Moreover, in case the definitions of some protocol buffers change,
// this file should be the only one that will potentially need to change.

// TODO: Write documentation comments for the functions in this file.
//       Part of the text can probably be copy-pasted from the documentation of the functions handling those events.

package orderers

import (
	commonpbtypes "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	"github.com/filecoin-project/mir/pkg/pb/ordererpb"
	"github.com/filecoin-project/mir/pkg/pb/pbftpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ============================================================
// Events
// ============================================================

func PbftProposeTimeout(numProposals uint64) *ordererpb.Event {
	return &ordererpb.Event{Type: &ordererpb.Event_Pbft{Pbft: &pbftpb.Event{
		Type: &pbftpb.Event_ProposeTimeout{ProposeTimeout: numProposals},
	}}}
}

func PbftViewChangeSNTimeout(view t.PBFTViewNr, numCommitted int) *ordererpb.Event {
	return &ordererpb.Event{Type: &ordererpb.Event_Pbft{
		Pbft: &pbftpb.Event{Type: &pbftpb.Event_ViewChangeSnTimeout{ViewChangeSnTimeout: &pbftpb.VCSNTimeout{
			View:         view.Pb(),
			NumCommitted: uint64(numCommitted),
		}}},
	}}
}

func PbftViewChangeSegmentTimeout(view t.PBFTViewNr) *ordererpb.Event {
	return &ordererpb.Event{Type: &ordererpb.Event_Pbft{
		Pbft: &pbftpb.Event{Type: &pbftpb.Event_ViewChangeSegTimeout{ViewChangeSegTimeout: view.Pb()}},
	}}
}

// ============================================================
// PBFT Messages
// ============================================================

func PbftPreprepareSBMessage(content *pbftpb.Preprepare) *ordererpb.Message {
	return &ordererpb.Message{Type: &ordererpb.Message_Pbft{
		Pbft: &pbftpb.Message{Type: &pbftpb.Message_Preprepare{Preprepare: content}},
	}}
}

func PbftPrepareSBMessage(content *pbftpb.Prepare) *ordererpb.Message {
	return &ordererpb.Message{Type: &ordererpb.Message_Pbft{
		Pbft: &pbftpb.Message{Type: &pbftpb.Message_Prepare{Prepare: content}},
	}}
}

func PbftCommitSBMessage(content *pbftpb.Commit) *ordererpb.Message {
	return &ordererpb.Message{Type: &ordererpb.Message_Pbft{
		Pbft: &pbftpb.Message{Type: &pbftpb.Message_Commit{Commit: content}},
	}}
}

func PbftSignedViewChangeSBMessage(signedViewChange *pbftpb.SignedViewChange) *ordererpb.Message {
	return &ordererpb.Message{Type: &ordererpb.Message_Pbft{
		Pbft: &pbftpb.Message{Type: &pbftpb.Message_SignedViewChange{SignedViewChange: signedViewChange}},
	}}
}

func PbftPreprepareRequestSBMessage(preprepareRequest *pbftpb.PreprepareRequest) *ordererpb.Message {
	return &ordererpb.Message{Type: &ordererpb.Message_Pbft{
		Pbft: &pbftpb.Message{Type: &pbftpb.Message_PreprepareRequest{PreprepareRequest: preprepareRequest}},
	}}
}

func PbftMissingPreprepareSBMessage(preprepare *pbftpb.Preprepare) *ordererpb.Message {
	return &ordererpb.Message{Type: &ordererpb.Message_Pbft{
		Pbft: &pbftpb.Message{Type: &pbftpb.Message_MissingPreprepare{MissingPreprepare: &pbftpb.MissingPreprepare{
			Preprepare: preprepare,
		}}},
	}}
}

func PbftNewViewSBMessage(newView *pbftpb.NewView) *ordererpb.Message {
	return &ordererpb.Message{Type: &ordererpb.Message_Pbft{
		Pbft: &pbftpb.Message{Type: &pbftpb.Message_NewView{NewView: newView}},
	}}
}

func PbftDoneSBMessage(digests [][]byte) *ordererpb.Message {
	return &ordererpb.Message{Type: &ordererpb.Message_Pbft{
		Pbft: &pbftpb.Message{Type: &pbftpb.Message_Done{Done: &pbftpb.Done{Digests: digests}}},
	}}
}

func PbftCatchUpRequestSBMessage(sn t.SeqNr, digest []byte) *ordererpb.Message {
	return &ordererpb.Message{Type: &ordererpb.Message_Pbft{
		Pbft: &pbftpb.Message{Type: &pbftpb.Message_CatchUpRequest{CatchUpRequest: &pbftpb.CatchUpRequest{
			Digest: digest,
			Sn:     sn.Pb(),
		}}},
	}}
}

func PbftCatchUpResponseSBMessage(preprepare *pbftpb.Preprepare) *ordererpb.Message {
	return &ordererpb.Message{Type: &ordererpb.Message_Pbft{
		Pbft: &pbftpb.Message{Type: &pbftpb.Message_CatchUpResponse{CatchUpResponse: &pbftpb.CatchUpResponse{
			Resp: preprepare,
		}}},
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
func pbftPreprepareMsg(sn t.SeqNr, view t.PBFTViewNr, data []byte, aborted bool) *pbftpb.Preprepare {
	return &pbftpb.Preprepare{
		Sn:      sn.Pb(),
		View:    view.Pb(),
		Data:    data,
		Aborted: aborted,
	}
}

// pbftPrepareMsg returns a protocol buffer representing a Prepare message.
// Analogous pbftPreprepareMsg, but for a Prepare message.
func pbftPrepareMsg(sn t.SeqNr, view t.PBFTViewNr, digest []byte) *pbftpb.Prepare {
	return &pbftpb.Prepare{
		Sn:     sn.Pb(),
		View:   view.Pb(),
		Digest: digest,
	}
}

// pbftCommitMsg returns a protocol buffer representing a Commit message.
// Analogous pbftPreprepareMsg, but for a Commit message.
func pbftCommitMsg(sn t.SeqNr, view t.PBFTViewNr, digest []byte) *pbftpb.Commit {
	return &pbftpb.Commit{
		Sn:     sn.Pb(),
		View:   view.Pb(),
		Digest: digest,
	}
}

// pbftViewChangeMsg returns a protocol buffer representing a ViewChange message.
func pbftViewChangeMsg(view t.PBFTViewNr, pSet viewChangePSet, qSet viewChangeQSet) *pbftpb.ViewChange {
	return &pbftpb.ViewChange{
		View: view.Pb(),
		PSet: pSet.Pb(),
		QSet: qSet.Pb(),
	}
}

// pbftSignedViewChangeMsg returns a protocol buffer representing a SignedViewChange message.
func pbftSignedViewChangeMsg(viewChange *pbftpb.ViewChange, signature []byte) *pbftpb.SignedViewChange {
	return &pbftpb.SignedViewChange{
		ViewChange: viewChange,
		Signature:  signature,
	}
}

func pbftPreprepareRequestMsg(sn t.SeqNr, digest []byte) *pbftpb.PreprepareRequest {
	return &pbftpb.PreprepareRequest{
		Digest: digest,
		Sn:     sn.Pb(),
	}
}

// pbftNewViewMsg returns a protocol buffer representing a NewView message.
func pbftNewViewMsg(
	view t.PBFTViewNr,
	viewChangeSenders []t.NodeID,
	viewChanges []*pbftpb.SignedViewChange,
	preprepareSeqNrs []t.SeqNr,
	preprepares []*pbftpb.Preprepare,
) *pbftpb.NewView {
	return &pbftpb.NewView{
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

func preprepareHashOrigin(preprepare *pbftpb.Preprepare) *ordererpb.HashOrigin {
	return &ordererpb.HashOrigin{Type: &ordererpb.HashOrigin_Pbft{
		Pbft: &pbftpb.HashOrigin{Type: &pbftpb.HashOrigin_Preprepare{Preprepare: preprepare}},
	}}
}

func missingPreprepareHashOrigin(preprepare *pbftpb.Preprepare) *ordererpb.HashOrigin {
	return &ordererpb.HashOrigin{Type: &ordererpb.HashOrigin_Pbft{
		Pbft: &pbftpb.HashOrigin{Type: &pbftpb.HashOrigin_MissingPreprepare{MissingPreprepare: preprepare}},
	}}
}

func viewChangeSignOrigin(viewChange *pbftpb.ViewChange) *ordererpb.SignOrigin {
	return &ordererpb.SignOrigin{Type: &ordererpb.SignOrigin_Pbft{
		Pbft: &pbftpb.SignOrigin{Type: &pbftpb.SignOrigin_ViewChange{
			ViewChange: viewChange,
		}},
	}}
}

func viewChangeSigVerOrigin(viewChange *pbftpb.SignedViewChange) *ordererpb.SigVerOrigin {
	return &ordererpb.SigVerOrigin{Type: &ordererpb.SigVerOrigin_Pbft{
		Pbft: &pbftpb.SigVerOrigin{Type: &pbftpb.SigVerOrigin_SignedViewChange{SignedViewChange: viewChange}},
	}}
}

func emptyPreprepareHashOrigin(view t.PBFTViewNr) *ordererpb.HashOrigin {
	return &ordererpb.HashOrigin{Type: &ordererpb.HashOrigin_Pbft{
		Pbft: &pbftpb.HashOrigin{Type: &pbftpb.HashOrigin_EmptyPreprepares{EmptyPreprepares: view.Pb()}},
	}}
}

func newViewSigVerOrigin(newView *pbftpb.NewView) *ordererpb.SigVerOrigin {
	return &ordererpb.SigVerOrigin{Type: &ordererpb.SigVerOrigin_Pbft{
		Pbft: &pbftpb.SigVerOrigin{Type: &pbftpb.SigVerOrigin_NewView{NewView: newView}},
	}}
}

func newViewHashOrigin(newView *pbftpb.NewView) *ordererpb.HashOrigin {
	return &ordererpb.HashOrigin{Type: &ordererpb.HashOrigin_Pbft{
		Pbft: &pbftpb.HashOrigin{Type: &pbftpb.HashOrigin_NewView{NewView: newView}},
	}}
}

func catchUpResponseHashOrigin(preprepare *pbftpb.Preprepare) *ordererpb.HashOrigin {
	return &ordererpb.HashOrigin{Type: &ordererpb.HashOrigin_Pbft{
		Pbft: &pbftpb.HashOrigin{Type: &pbftpb.HashOrigin_CatchUpResponse{CatchUpResponse: preprepare}},
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
func serializePreprepareForHashing(preprepare *pbftpb.Preprepare) *commonpbtypes.HashData {

	// Encode boolean Aborted field as one byte.
	aborted := byte(0)
	if preprepare.Aborted {
		aborted = 1
	}

	// Put everything together in a slice and return it.
	// Note that we do not include the view number,
	// as the view change protocol might compare hashes of Preprepares across vies.
	return &commonpbtypes.HashData{Data: [][]byte{t.SeqNr(preprepare.Sn).Bytes(), {aborted}, preprepare.Data}}
}

func serializeViewChangeForSigning(vc *pbftpb.ViewChange) [][]byte {
	_ = &pbftpb.ViewChange{
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
