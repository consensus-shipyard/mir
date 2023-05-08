/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
// TODO: Write documentation comments for the functions in this file.
//       Part of the text can probably be copy-pasted from the documentation of the functions handling those events.

package common

import (
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	hasherpbtypes "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	"github.com/filecoin-project/mir/pkg/pb/pbftpb"
	pbftpbtypes "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
)

// ============================================================
// Serialization
// ============================================================

// SerializePreprepareForHashing returns a slice of byte slices representing the contents of a Preprepare message
// that can be passed to the Hasher module.
// Note that the view number is *not* serialized, as hashes must be consistent across views.
// Even though the preprepare argument is a protocol buffer, this function is required to guarantee
// that the serialization is deterministic, since the protobuf native serialization does not provide this guarantee.
func SerializePreprepareForHashing(preprepare *pbftpbtypes.Preprepare) *hasherpbtypes.HashData {

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

func SerializeViewChangeForSigning(vc *pbftpbtypes.ViewChange) *cryptopbtypes.SignedData {
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
