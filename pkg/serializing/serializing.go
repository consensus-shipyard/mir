/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package serializing

import (
	"encoding/binary"

	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	commonpbtypes "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
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

func CheckpointForSig(epoch tt.EpochNr, seqNr tt.SeqNr, snapshotHash []byte) *cryptopbtypes.SignedData {
	epochBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(epochBytes, uint64(epoch))

	snBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(snBytes, uint64(seqNr))

	return &cryptopbtypes.SignedData{Data: [][]byte{epochBytes, snBytes, snapshotHash}}
}

func SnapshotForHash(snapshot *commonpb.StateSnapshot) *commonpbtypes.HashData {
	return &commonpbtypes.HashData{Data: append(EpochDataForHash(snapshot.EpochData), snapshot.AppData)}
}

func EpochDataForHash(epochData *commonpb.EpochData) [][]byte {
	data := append(EpochConfigForHash(epochData.EpochConfig), epochData.LeaderPolicy)
	data = append(data, ClientProgressForHash(epochData.ClientProgress)...)
	if len(epochData.PreviousMembership.Membership) != 0 {
		// In the initial checkpoint the PreviousMembership is an empty map.
		data = append(data, MembershipsForHash([]*commonpb.Membership{epochData.PreviousMembership})...)
	}
	return data
}

func EpochConfigForHash(epochConfig *commonpb.EpochConfig) [][]byte {

	// Add simple values.
	data := [][]byte{
		tt.EpochNr(epochConfig.EpochNr).Bytes(),
		Uint64ToBytes(epochConfig.Length),
		Uint64ToBytes(epochConfig.FirstSn),
	}

	// Append memberships.
	data = append(data, MembershipsForHash(epochConfig.Memberships)...)

	return data
}

func MembershipsForHash(memberships []*commonpb.Membership) [][]byte {
	var data [][]byte

	// Each string representing an ID and an address is explicitly terminated with a zero byte.
	// This ensures that the last byte of an ID and the first byte of an address are not interchangeable.
	for _, membership := range memberships {
		maputil.IterateSorted(membership.Membership, func(id string, addr string) bool {
			data = append(data, []byte(id), []byte{0}, []byte(addr), []byte{0})
			return true
		})
	}

	return data
}

func ClientProgressForHash(clientProgress *commonpb.ClientProgress) [][]byte {
	var data [][]byte
	maputil.IterateSorted(
		clientProgress.Progress,
		func(clientID string, deliveredReqs *commonpb.DeliveredReqs) (cont bool) {
			// Append client ID and low watermark.
			data = append(data, []byte(tt.ClientID(clientID).Pb()), Uint64ToBytes(deliveredReqs.LowWm))

			// Append all request numbers delivered after the watermark.
			for _, reqNo := range deliveredReqs.Delivered {
				data = append(data, Uint64ToBytes(reqNo))
			}
			return true
		},
	)
	return data
}

// Uint64ToBytes returns a 8-byte slice encoding an unsigned 64-bit integer.
func Uint64ToBytes(n uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, n)
	return buf
}

func Uint64FromBytes(bytes []byte) uint64 {
	return binary.LittleEndian.Uint64(bytes)
}
