package availabilitypbevents

import (
	types2 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

func CertVerified(next []*types.Event, destModule types1.ModuleID, valid bool, err string, origin *types2.VerifyCertOrigin) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_Availability{
			Availability: &types2.Event{
				Type: &types2.Event_CertVerified{
					CertVerified: &types2.CertVerified{
						Valid:  valid,
						Err:    err,
						Origin: origin,
					},
				},
			},
		},
	}
}
