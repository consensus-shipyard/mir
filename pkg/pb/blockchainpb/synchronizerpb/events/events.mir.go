// Code generated by Mir codegen. DO NOT EDIT.

package synchronizerpbevents

import (
	types3 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/synchronizerpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func SyncRequest(destModule types.ModuleID, orphanBlock *types1.Block, leaveIds []uint64) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Synchronizer{
			Synchronizer: &types3.Event{
				Type: &types3.Event_SyncRequest{
					SyncRequest: &types3.SyncRequest{
						OrphanBlock: orphanBlock,
						LeaveIds:    leaveIds,
					},
				},
			},
		},
	}
}
