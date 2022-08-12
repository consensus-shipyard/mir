/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package serializing

import (
	"encoding/binary"

	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

// TODO: Write doc.

func RequestForHash(req *requestpb.Request) [][]byte {

	// Encode integer fields.
	clientIDBuf := []byte(req.ClientId)
	reqNoBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(reqNoBuf, req.ReqNo)

	// Note that the signature is *not* part of the hashed data.

	// Return serialized integers along with the request data itself.
	return [][]byte{clientIDBuf, reqNoBuf, req.Data}
}

func BatchForHash(batch *requestpb.Batch) [][]byte {

	// Allocate output slice.
	data := make([][]byte, len(batch.Requests))

	// Collect all request digests in the batch.
	for i, reqRef := range batch.Requests {
		data[i] = reqRef.Digest
	}

	// Return populated output slice.
	return data
}

func CheckpointForSig(epoch t.EpochNr, seqNr t.SeqNr, snapshotHash []byte) [][]byte {
	epochBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(epochBytes, uint64(epoch))

	snBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(snBytes, uint64(seqNr))

	return [][]byte{epochBytes, snBytes, snapshotHash}
}

func SnapshotForHash(snapshot *commonpb.StateSnapshot) [][]byte {

	// Append epoch and app data
	data := [][]byte{
		t.EpochNr(snapshot.Configuration.EpochNr).Bytes(),
		snapshot.AppData,
	}

	// Append membership.
	// Each string representing an ID and an address is explicitly terminated with a zero byte.
	// This ensures that the last byte of an ID and the first byte of an address are not interchangeable.
	for _, membership := range snapshot.Configuration.Memberships {
		maputil.IterateSorted(membership.Membership, func(id string, addr string) bool {
			data = append(data, []byte(id), []byte{0}, []byte(addr), []byte{0})
			return true
		})
	}

	return data
}

// Uint64ToBytes returns a 8-byte slice encoding an unsigned 64-bit integer.
func Uint64ToBytes(n uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, n)
	return buf
}

func BytesToUint64(data []byte) uint64 {
	return binary.LittleEndian.Uint64(data)
}
