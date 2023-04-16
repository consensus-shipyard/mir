package cryptopbevents

import (
	types1 "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func SignRequest(destModule types.ModuleID, data *types1.SignedData, origin *types1.SignOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Crypto{
			Crypto: &types1.Event{
				Type: &types1.Event_SignRequest{
					SignRequest: &types1.SignRequest{
						Data:   data,
						Origin: origin,
					},
				},
			},
		},
	}
}

func SignResult(destModule types.ModuleID, signature []uint8, origin *types1.SignOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Crypto{
			Crypto: &types1.Event{
				Type: &types1.Event_SignResult{
					SignResult: &types1.SignResult{
						Signature: signature,
						Origin:    origin,
					},
				},
			},
		},
	}
}

func VerifySig(destModule types.ModuleID, data *types1.SignedData, signature []uint8, origin *types1.SigVerOrigin, nodeId types.NodeID) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Crypto{
			Crypto: &types1.Event{
				Type: &types1.Event_VerifySig{
					VerifySig: &types1.VerifySig{
						Data:      data,
						Signature: signature,
						Origin:    origin,
						NodeId:    nodeId,
					},
				},
			},
		},
	}
}

func SigVerified(destModule types.ModuleID, origin *types1.SigVerOrigin, nodeId types.NodeID, error error) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Crypto{
			Crypto: &types1.Event{
				Type: &types1.Event_SigVerified{
					SigVerified: &types1.SigVerified{
						Origin: origin,
						NodeId: nodeId,
						Error:  error,
					},
				},
			},
		},
	}
}

func VerifySigs(destModule types.ModuleID, data []*types1.SignedData, signatures [][]uint8, origin *types1.SigVerOrigin, nodeIds []types.NodeID) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Crypto{
			Crypto: &types1.Event{
				Type: &types1.Event_VerifySigs{
					VerifySigs: &types1.VerifySigs{
						Data:       data,
						Signatures: signatures,
						Origin:     origin,
						NodeIds:    nodeIds,
					},
				},
			},
		},
	}
}

func SigsVerified(destModule types.ModuleID, origin *types1.SigVerOrigin, nodeIds []types.NodeID, errors []error, allOk bool) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Crypto{
			Crypto: &types1.Event{
				Type: &types1.Event_SigsVerified{
					SigsVerified: &types1.SigsVerified{
						Origin:  origin,
						NodeIds: nodeIds,
						Errors:  errors,
						AllOk:   allOk,
					},
				},
			},
		},
	}
}
