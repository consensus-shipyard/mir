package isspbmsgs

import (
	types3 "github.com/filecoin-project/mir/pkg/pb/isspb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func RetransmitRequests(destModule types.ModuleID, requests []*types1.Request) *types2.Message {
	return &types2.Message{
		DestModule: destModule,
		Type: &types2.Message_Iss{
			Iss: &types3.ISSMessage{
				Type: &types3.ISSMessage_RetransmitRequests{
					RetransmitRequests: &types3.RetransmitRequests{
						Requests: requests,
					},
				},
			},
		},
	}
}
