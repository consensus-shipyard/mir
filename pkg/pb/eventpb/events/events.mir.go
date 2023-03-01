package eventpbevents

import (
	types5 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types6 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func Init(destModule types.ModuleID) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Init{
			Init: &types1.Init{},
		},
	}
}

func NewRequests(destModule types.ModuleID, requests []*types2.Request) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_NewRequests{
			NewRequests: &types1.NewRequests{
				Requests: requests,
			},
		},
	}
}

func HashRequest(destModule types.ModuleID, data []*types3.HashData, origin *types1.HashOrigin) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_HashRequest{
			HashRequest: &types1.HashRequest{
				Data:   data,
				Origin: origin,
			},
		},
	}
}

func HashResult(destModule types.ModuleID, digests [][]uint8, origin *types1.HashOrigin) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_HashResult{
			HashResult: &types1.HashResult{
				Digests: digests,
				Origin:  origin,
			},
		},
	}
}

func SignRequest(destModule types.ModuleID, data [][]uint8, origin *types1.SignOrigin) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_SignRequest{
			SignRequest: &types1.SignRequest{
				Data:   data,
				Origin: origin,
			},
		},
	}
}

func SignResult(destModule types.ModuleID, signature []uint8, origin *types1.SignOrigin) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_SignResult{
			SignResult: &types1.SignResult{
				Signature: signature,
				Origin:    origin,
			},
		},
	}
}

func VerifyNodeSigs(destModule types.ModuleID, data []*types1.SigVerData, signatures [][]uint8, origin *types1.SigVerOrigin, nodeIds []types.NodeID) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_VerifyNodeSigs{
			VerifyNodeSigs: &types1.VerifyNodeSigs{
				Data:       data,
				Signatures: signatures,
				Origin:     origin,
				NodeIds:    nodeIds,
			},
		},
	}
}

func NodeSigsVerified(destModule types.ModuleID, origin *types1.SigVerOrigin, nodeIds []types.NodeID, valid []bool, errors []error, allOk bool) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_NodeSigsVerified{
			NodeSigsVerified: &types1.NodeSigsVerified{
				Origin:  origin,
				NodeIds: nodeIds,
				Valid:   valid,
				Errors:  errors,
				AllOk:   allOk,
			},
		},
	}
}

func SendMessage(destModule types.ModuleID, msg *types4.Message, destinations []types.NodeID) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_SendMessage{
			SendMessage: &types1.SendMessage{
				Msg:          msg,
				Destinations: destinations,
			},
		},
	}
}

func MessageReceived(destModule types.ModuleID, from types.NodeID, msg *types4.Message) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_MessageReceived{
			MessageReceived: &types1.MessageReceived{
				From: from,
				Msg:  msg,
			},
		},
	}
}

func DeliverCert(destModule types.ModuleID, sn types.SeqNr, cert *types5.Cert) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_DeliverCert{
			DeliverCert: &types1.DeliverCert{
				Sn:   sn,
				Cert: cert,
			},
		},
	}
}

func AppSnapshotRequest(destModule types.ModuleID, replyTo types.ModuleID) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_AppSnapshotRequest{
			AppSnapshotRequest: &types1.AppSnapshotRequest{
				ReplyTo: replyTo,
			},
		},
	}
}

func AppRestoreState(destModule types.ModuleID, checkpoint *types6.StableCheckpoint) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_AppRestoreState{
			AppRestoreState: &types1.AppRestoreState{
				Checkpoint: checkpoint,
			},
		},
	}
}

func NewEpoch(destModule types.ModuleID, epochNr types.EpochNr) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_NewEpoch{
			NewEpoch: &types1.NewEpoch{
				EpochNr: epochNr,
			},
		},
	}
}
