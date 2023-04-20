package checkpointpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types4 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	dsl1 "github.com/filecoin-project/mir/pkg/pb/isspb/dsl"
	types3 "github.com/filecoin-project/mir/pkg/pb/isspb/types"
	dsl2 "github.com/filecoin-project/mir/pkg/pb/messagepb/dsl"
	types5 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types1 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing net messages.

func UponStableCheckpointReceived(m dsl.Module, handler func(from types.NodeID, sn types1.SeqNr, snapshot *types2.StateSnapshot, cert map[types.NodeID][]uint8) error) {
	dsl1.UponISSMessageReceived[*types3.ISSMessage_StableCheckpoint](m, func(from types.NodeID, msg *types4.StableCheckpoint) error {
		return handler(from, msg.Sn, msg.Snapshot, msg.Cert)
	})
}

func UponMessageReceived[W types4.Message_TypeWrapper[M], M any](m dsl.Module, handler func(from types.NodeID, msg *M) error) {
	dsl2.UponMessageReceived[*types5.Message_Checkpoint](m, func(from types.NodeID, msg *types4.Message) error {
		w, ok := msg.Type.(W)
		if !ok {
			return nil
		}

		return handler(from, w.Unwrap())
	})
}

func UponCheckpointReceived(m dsl.Module, handler func(from types.NodeID, epoch types1.EpochNr, sn types1.SeqNr, snapshotHash []uint8, signature []uint8) error) {
	UponMessageReceived[*types4.Message_Checkpoint](m, func(from types.NodeID, msg *types4.Checkpoint) error {
		return handler(from, msg.Epoch, msg.Sn, msg.SnapshotHash, msg.Signature)
	})
}
