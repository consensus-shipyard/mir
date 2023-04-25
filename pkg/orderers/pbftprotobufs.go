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
	"github.com/filecoin-project/mir/pkg/orderers/types"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	hasherpbtypes "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	"github.com/filecoin-project/mir/pkg/pb/ordererpb"
	"github.com/filecoin-project/mir/pkg/pb/pbftpb"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
)

// ============================================================
// Events
// ============================================================

func PbftProposeTimeout(numProposals uint64) *ordererpb.Event {
	return &ordererpb.Event{Type: &ordererpb.Event_Pbft{Pbft: &pbftpb.Event{
		Type: &pbftpb.Event_ProposeTimeout{ProposeTimeout: numProposals},
	}}}
}

func PbftViewChangeSNTimeout(view types.ViewNr, numCommitted int) *ordererpb.Event {
	return &ordererpb.Event{Type: &ordererpb.Event_Pbft{
		Pbft: &pbftpb.Event{Type: &pbftpb.Event_ViewChangeSnTimeout{ViewChangeSnTimeout: &pbftpb.VCSNTimeout{
			View:         view.Pb(),
			NumCommitted: uint64(numCommitted),
		}}},
	}}
}

func PbftViewChangeSegmentTimeout(view types.ViewNr) *ordererpb.Event {
	return &ordererpb.Event{Type: &ordererpb.Event_Pbft{
		Pbft: &pbftpb.Event{Type: &pbftpb.Event_ViewChangeSegTimeout{ViewChangeSegTimeout: view.Pb()}},
	}}
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

func emptyPreprepareHashOrigin(view types.ViewNr) *ordererpb.HashOrigin {
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
func serializePreprepareForHashing(preprepare *pbftpbtypes.Preprepare) *hasherpbtypes.HashData {

	// Encode boolean Aborted field as one byte.
	aborted := byte(0)
	if preprepare.Aborted {
		aborted = 1
	}

	// Put everything together in a slice and return it.
	// Note that we do not include the view number,
	// as the view change protocol might compare hashes of Preprepares across vies.
	return &hasherpbtypes.HashData{Data: [][]byte{preprepare.Sn.Bytes(), {aborted}, preprepare.Data}}
}

func serializeViewChangeForSigning(vc *pbftpbtypes.ViewChange) *cryptopbtypes.SignedData {
	_ = &pbftpb.ViewChange{
		View: 0,
		PSet: nil,
		QSet: nil,
	}

	// Allocate result data structure.
	data := make([][]byte, 0)

	// Encode view number.
	data = append(data, vc.View.Bytes())

	// Encode P set.
	for _, pSetEntry := range vc.PSet {
		data = append(data, pSetEntry.Sn.Bytes())
		data = append(data, pSetEntry.View.Bytes())
		data = append(data, pSetEntry.Digest)
	}

	// Encode Q set.
	for _, qSetEntry := range vc.QSet {
		data = append(data, qSetEntry.Sn.Bytes())
		data = append(data, qSetEntry.View.Bytes())
		data = append(data, qSetEntry.Digest)
	}

	return &cryptopbtypes.SignedData{Data: data}
}
