package hasherpbevents

import (
	types2 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func Request(destModule types.ModuleID, data []*types1.HashData, origin *types1.HashOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Hasher{
			Hasher: &types1.Event{
				Type: &types1.Event_Request{
					Request: &types1.Request{
						Data:   data,
						Origin: origin,
					},
				},
			},
		},
	}
}

func Result(destModule types.ModuleID, digests [][]uint8, origin *types1.HashOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Hasher{
			Hasher: &types1.Event{
				Type: &types1.Event_Result{
					Result: &types1.Result{
						Digests: digests,
						Origin:  origin,
					},
				},
			},
		},
	}
}

func RequestOne(destModule types.ModuleID, data *types1.HashData, origin *types1.HashOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Hasher{
			Hasher: &types1.Event{
				Type: &types1.Event_RequestOne{
					RequestOne: &types1.RequestOne{
						Data:   data,
						Origin: origin,
					},
				},
			},
		},
	}
}

func ResultOne(destModule types.ModuleID, digest []uint8, origin *types1.HashOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Hasher{
			Hasher: &types1.Event{
				Type: &types1.Event_ResultOne{
					ResultOne: &types1.ResultOne{
						Digest: digest,
						Origin: origin,
					},
				},
			},
		},
	}
}
