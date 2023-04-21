package checkpoint

import (
	"encoding/binary"

	commonpbtypes "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	"github.com/filecoin-project/mir/pkg/serializing"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

func serializeCheckpointForSig(epoch tt.EpochNr, seqNr tt.SeqNr, snapshotHash []byte) *cryptopbtypes.SignedData {
	epochBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(epochBytes, uint64(epoch))

	snBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(snBytes, uint64(seqNr))

	return &cryptopbtypes.SignedData{Data: [][]byte{epochBytes, snBytes, snapshotHash}}
}

func serializeSnapshotForHash(snapshot *commonpbtypes.StateSnapshot) *commonpbtypes.HashData {
	return &commonpbtypes.HashData{Data: append(serializeEpochDataForHash(snapshot.EpochData), snapshot.AppData)}
}

func serializeEpochDataForHash(epochData *commonpbtypes.EpochData) [][]byte {
	data := append(serializeEpochConfigForHash(epochData.EpochConfig), epochData.LeaderPolicy)
	data = append(data, serializeClientProgressForHash(epochData.ClientProgress)...)
	if len(epochData.PreviousMembership.Nodes) != 0 {
		// In the initial checkpoint the PreviousMembership is an empty map.
		data = append(data, serializeMembershipsForHash([]*commonpbtypes.Membership{epochData.PreviousMembership})...)
	}
	return data
}

func serializeEpochConfigForHash(epochConfig *commonpbtypes.EpochConfig) [][]byte {

	// Add simple values.
	data := [][]byte{
		epochConfig.EpochNr.Bytes(),
		serializing.Uint64ToBytes(epochConfig.Length),
		epochConfig.FirstSn.Bytes(),
	}

	// Append memberships.
	data = append(data, serializeMembershipsForHash(epochConfig.Memberships)...)

	return data
}

func serializeMembershipsForHash(memberships []*commonpbtypes.Membership) [][]byte {
	var data [][]byte

	// Each string representing an ID and an address is explicitly terminated with a zero byte.
	// This ensures that the last byte of an ID and the first byte of an address are not interchangeable.
	for _, membership := range memberships {
		maputil.IterateSorted(membership.Nodes, func(id t.NodeID, identity *commonpbtypes.NodeIdentity) bool {
			data = append(data, id.Bytes(), []byte{0})
			data = append(data, serializeNodeIdentityForHash(identity)...)
			return true
		})
	}

	return data
}

func serializeNodeIdentityForHash(identity *commonpbtypes.NodeIdentity) [][]byte {

	// TODO: Using {0} as a field separator is technically not right,
	//   since one field ending with 0 still yields the same byte string as another field beginning with 0.
	return [][]byte{
		identity.Id.Bytes(),
		{0},
		[]byte(identity.Addr),
		{0},
		identity.Key,
		{0},
		serializing.Uint64ToBytes(identity.Weight),
		{0},
	}
}

func serializeClientProgressForHash(clientProgress *commonpbtypes.ClientProgress) [][]byte {
	var data [][]byte
	maputil.IterateSorted(
		clientProgress.Progress,
		func(clientID tt.ClientID, deliveredReqs *commonpbtypes.DeliveredReqs) (cont bool) {
			// Append client ID and low watermark.
			data = append(data, []byte(clientID.Pb()), serializing.Uint64ToBytes(deliveredReqs.LowWm))

			// Append all request numbers delivered after the watermark.
			for _, reqNo := range deliveredReqs.Delivered {
				data = append(data, serializing.Uint64ToBytes(reqNo))
			}
			return true
		},
	)
	return data
}
