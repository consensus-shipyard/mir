package checkpoint

import (
	"encoding/binary"

	"github.com/filecoin-project/mir/pkg/pb/commonpb"
	commonpbtypes "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	"github.com/filecoin-project/mir/pkg/serializing"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

func serializeCheckpointForSig(epoch tt.EpochNr, seqNr tt.SeqNr, snapshotHash []byte) *cryptopbtypes.SignedData {
	epochBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(epochBytes, uint64(epoch))

	snBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(snBytes, uint64(seqNr))

	return &cryptopbtypes.SignedData{Data: [][]byte{epochBytes, snBytes, snapshotHash}}
}

func serializeSnapshotForHash(snapshot *commonpb.StateSnapshot) *commonpbtypes.HashData {
	return &commonpbtypes.HashData{Data: append(serializeEpochDataForHash(snapshot.EpochData), snapshot.AppData)}
}

func serializeEpochDataForHash(epochData *commonpb.EpochData) [][]byte {
	data := append(serializeEpochConfigForHash(epochData.EpochConfig), epochData.LeaderPolicy)
	data = append(data, serializeClientProgressForHash(epochData.ClientProgress)...)
	if len(epochData.PreviousMembership.Membership) != 0 {
		// In the initial checkpoint the PreviousMembership is an empty map.
		data = append(data, serializeMembershipsForHash([]*commonpb.Membership{epochData.PreviousMembership})...)
	}
	return data
}

func serializeEpochConfigForHash(epochConfig *commonpb.EpochConfig) [][]byte {

	// Add simple values.
	data := [][]byte{
		tt.EpochNr(epochConfig.EpochNr).Bytes(),
		serializing.Uint64ToBytes(epochConfig.Length),
		serializing.Uint64ToBytes(epochConfig.FirstSn),
	}

	// Append memberships.
	data = append(data, serializeMembershipsForHash(epochConfig.Memberships)...)

	return data
}

func serializeMembershipsForHash(memberships []*commonpb.Membership) [][]byte {
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

func serializeClientProgressForHash(clientProgress *commonpb.ClientProgress) [][]byte {
	var data [][]byte
	maputil.IterateSorted(
		clientProgress.Progress,
		func(clientID string, deliveredReqs *commonpb.DeliveredReqs) (cont bool) {
			// Append client ID and low watermark.
			data = append(data, []byte(tt.ClientID(clientID).Pb()), serializing.Uint64ToBytes(deliveredReqs.LowWm))

			// Append all request numbers delivered after the watermark.
			for _, reqNo := range deliveredReqs.Delivered {
				data = append(data, serializing.Uint64ToBytes(reqNo))
			}
			return true
		},
	)
	return data
}
