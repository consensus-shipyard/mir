package checkpointpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types3 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	dsl1 "github.com/filecoin-project/mir/pkg/pb/isspb/dsl"
	types2 "github.com/filecoin-project/mir/pkg/pb/isspb/types"
	dsl2 "github.com/filecoin-project/mir/pkg/pb/messagepb/dsl"
	types4 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing net messages.

func UponStableCheckpointReceived(m dsl.Module, handler func(from types.NodeID, sn types.SeqNr, snapshot *types1.StateSnapshot, cert map[types.NodeID][]uint8) error) {
	dsl1.UponISSMessageReceived[*types2.ISSMessage_StableCheckpoint](m, func(from types.NodeID, msg *types3.StableCheckpoint) error {
		return handler(from, msg.Sn, msg.Snapshot, msg.Cert)
	})
}

func UponMessageReceived[W types3.Message_TypeWrapper[M], M any](m dsl.Module, handler func(from types.NodeID, msg *M) error) {
	dsl2.UponMessageReceived[*types4.Message_Checkpoint](m, func(from types.NodeID, msg *types3.Message) error {
		w, ok := msg.Type.(W)
		if !ok {
			return nil
		}

		return handler(from, w.Unwrap())
	})
}

func UponCheckpointReceived(m dsl.Module, handler func(from types.NodeID, epoch types.EpochNr, sn types.SeqNr, snapshotHash []uint8, signature []uint8) error) {
	UponMessageReceived[*types3.Message_Checkpoint](m, func(from types.NodeID, msg *types3.Checkpoint) error {
		return handler(from, msg.Epoch, msg.Sn, msg.SnapshotHash, msg.Signature)
	})
}
