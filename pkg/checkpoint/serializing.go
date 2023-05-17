package checkpoint

import (
	"encoding/binary"

	es "github.com/go-errors/errors"

	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	hasherpbtypes "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/serializing"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/membutil"
)

func serializeCheckpointForSig(epoch tt.EpochNr, seqNr tt.SeqNr, snapshotHash []byte) *cryptopbtypes.SignedData {
	epochBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(epochBytes, uint64(epoch))

	snBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(snBytes, uint64(seqNr))

	return &cryptopbtypes.SignedData{Data: [][]byte{epochBytes, snBytes, snapshotHash}}
}

func serializeSnapshotForHash(snapshot *trantorpbtypes.StateSnapshot) (*hasherpbtypes.HashData, error) {
	epochData, err := serializeEpochDataForHash(snapshot.EpochData)
	if err != nil {
		return nil, es.Errorf("failed serializing epoch data: %w", err)
	}
	return &hasherpbtypes.HashData{Data: append(epochData, snapshot.AppData)}, nil
}

func serializeEpochDataForHash(epochData *trantorpbtypes.EpochData) ([][]byte, error) {
	data, err := serializeEpochConfigForHash(epochData.EpochConfig)
	if err != nil {
		return nil, es.Errorf("failed serializing epoch config: %w", err)
	}
	data = append(data, epochData.LeaderPolicy)
	data = append(data, serializeClientProgressForHash(epochData.ClientProgress)...)
	if len(epochData.PreviousMembership.Nodes) != 0 {
		// In the initial checkpoint the PreviousMembership is an empty map.
		prevMembData, err := membutil.Serialize(epochData.PreviousMembership)
		if err != nil {
			return nil, es.Errorf("failed serializing previous membership: %w", err)
		}
		data = append(data, prevMembData)
	}
	return data, nil
}

func serializeEpochConfigForHash(epochConfig *trantorpbtypes.EpochConfig) ([][]byte, error) {

	// Add simple values.
	data := [][]byte{
		epochConfig.EpochNr.Bytes(),
		serializing.Uint64ToBytes(epochConfig.Length),
		epochConfig.FirstSn.Bytes(),
	}

	// Append memberships.
	d, err := serializeMembershipsForHash(epochConfig.Memberships)
	if err != nil {
		return nil, err
	}
	data = append(data, d...)

	return data, nil
}

func serializeMembershipsForHash(memberships []*trantorpbtypes.Membership) ([][]byte, error) {
	data := make([][]byte, 0)

	// Each string representing an ID and an address is explicitly terminated with a zero byte.
	// This ensures that the last byte of an ID and the first byte of an address are not interchangeable.
	for _, membership := range memberships {
		d, err := membutil.Serialize(membership)
		if err != nil {
			return nil, es.Errorf("could not serialize membership: %w", err)
		}
		data = append(data, d)
	}

	return data, nil
}

func serializeClientProgressForHash(clientProgress *trantorpbtypes.ClientProgress) [][]byte {
	var data [][]byte
	maputil.IterateSorted(
		clientProgress.Progress,
		func(clientID tt.ClientID, deliveredTXs *trantorpbtypes.DeliveredTXs) (cont bool) {
			// Append client ID and low watermark.
			data = append(data, []byte(clientID.Pb()), serializing.Uint64ToBytes(deliveredTXs.LowWm))

			// Append all transaction numbers delivered after the watermark.
			for _, txNo := range deliveredTXs.Delivered {
				data = append(data, serializing.Uint64ToBytes(txNo))
			}
			return true
		},
	)
	return data
}
