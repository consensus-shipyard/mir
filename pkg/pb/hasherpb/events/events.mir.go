package hasherpbevents

import (
	types1 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func Request(destModule types.ModuleID, data []*types1.HashData, origin *types2.HashOrigin) *types3.Event {
	return &types3.Event{
		DestModule: destModule,
		Type: &types3.Event_Hasher{
			Hasher: &types2.Event{
				Type: &types2.Event_Request{
					Request: &types2.Request{
						Data:   data,
						Origin: origin,
					},
				},
			},
		},
	}
}

func Result(destModule types.ModuleID, digests [][]uint8, origin *types2.HashOrigin) *types3.Event {
	return &types3.Event{
		DestModule: destModule,
		Type: &types3.Event_Hasher{
			Hasher: &types2.Event{
				Type: &types2.Event_Result{
					Result: &types2.Result{
						Digests: digests,
						Origin:  origin,
					},
				},
			},
		},
	}
}

func RequestOne(destModule types.ModuleID, data *types1.HashData, origin *types2.HashOrigin) *types3.Event {
	return &types3.Event{
		DestModule: destModule,
		Type: &types3.Event_Hasher{
			Hasher: &types2.Event{
				Type: &types2.Event_RequestOne{
					RequestOne: &types2.RequestOne{
						Data:   data,
						Origin: origin,
					},
				},
			},
		},
	}
}

func ResultOne(destModule types.ModuleID, digest []uint8, origin *types2.HashOrigin) *types3.Event {
	return &types3.Event{
		DestModule: destModule,
		Type: &types3.Event_Hasher{
			Hasher: &types2.Event{
				Type: &types2.Event_ResultOne{
					ResultOne: &types2.ResultOne{
						Digest: digest,
						Origin: origin,
					},
				},
			},
		},
	}
}
