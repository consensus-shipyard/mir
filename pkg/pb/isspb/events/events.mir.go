package isspbevents

import (
	types4 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types5 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/isspb/types"
	types3 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func PushCheckpoint(destModule types.ModuleID) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Iss{
			Iss: &types2.Event{
				Type: &types2.Event_PushCheckpoint{
					PushCheckpoint: &types2.PushCheckpoint{},
				},
			},
		},
	}
}

func SBDeliver(destModule types.ModuleID, sn types3.SeqNr, data []uint8, aborted bool, leader types.NodeID, instanceId types.ModuleID) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Iss{
			Iss: &types2.Event{
				Type: &types2.Event_SbDeliver{
					SbDeliver: &types2.SBDeliver{
						Sn:         sn,
						Data:       data,
						Aborted:    aborted,
						Leader:     leader,
						InstanceId: instanceId,
					},
				},
			},
		},
	}
}

func DeliverCert(destModule types.ModuleID, sn types3.SeqNr, cert *types4.Cert) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Iss{
			Iss: &types2.Event{
				Type: &types2.Event_DeliverCert{
					DeliverCert: &types2.DeliverCert{
						Sn:   sn,
						Cert: cert,
					},
				},
			},
		},
	}
}

func NewConfig(destModule types.ModuleID, epochNr types3.EpochNr, membership *types5.Membership) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Iss{
			Iss: &types2.Event{
				Type: &types2.Event_NewConfig{
					NewConfig: &types2.NewConfig{
						EpochNr:    epochNr,
						Membership: membership,
					},
				},
			},
		},
	}
}
