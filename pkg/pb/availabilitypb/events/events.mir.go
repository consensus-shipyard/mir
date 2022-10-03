package availabilitypbevents

import (
	types1 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func CertVerified(destModule types.ModuleID, valid bool, err string, origin *types1.VerifyCertOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Availability{
			Availability: &types1.Event{
				Type: &types1.Event_CertVerified{
					CertVerified: &types1.CertVerified{
						Valid:  valid,
						Err:    err,
						Origin: origin,
					},
				},
			},
		},
	}
}
