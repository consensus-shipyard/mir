package threshcryptopbevents

import (
	types "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/threshcryptopb/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

func SignShare(next []*types.Event, destModule types1.ModuleID, data [][]uint8, origin *types2.SignShareOrigin) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_ThreshCrypto{
			ThreshCrypto: &types2.Event{
				Type: &types2.Event_SignShare{
					SignShare: &types2.SignShare{
						Data:   data,
						Origin: origin,
					},
				},
			},
		},
	}
}

func SignShareResult(next []*types.Event, destModule types1.ModuleID, signatureShare []uint8, origin *types2.SignShareOrigin) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_ThreshCrypto{
			ThreshCrypto: &types2.Event{
				Type: &types2.Event_SignShareResult{
					SignShareResult: &types2.SignShareResult{
						SignatureShare: signatureShare,
						Origin:         origin,
					},
				},
			},
		},
	}
}

func VerifyShare(next []*types.Event, destModule types1.ModuleID, data [][]uint8, signatureShare []uint8, nodeId types1.NodeID, origin *types2.VerifyShareOrigin) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_ThreshCrypto{
			ThreshCrypto: &types2.Event{
				Type: &types2.Event_VerifyShare{
					VerifyShare: &types2.VerifyShare{
						Data:           data,
						SignatureShare: signatureShare,
						NodeId:         nodeId,
						Origin:         origin,
					},
				},
			},
		},
	}
}

func VerifyShareResult(next []*types.Event, destModule types1.ModuleID, ok bool, error string, origin *types2.VerifyShareOrigin) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_ThreshCrypto{
			ThreshCrypto: &types2.Event{
				Type: &types2.Event_VerifyShareResult{
					VerifyShareResult: &types2.VerifyShareResult{
						Ok:     ok,
						Error:  error,
						Origin: origin,
					},
				},
			},
		},
	}
}

func VerifyFull(next []*types.Event, destModule types1.ModuleID, data [][]uint8, fullSignature []uint8, origin *types2.VerifyFullOrigin) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_ThreshCrypto{
			ThreshCrypto: &types2.Event{
				Type: &types2.Event_VerifyFull{
					VerifyFull: &types2.VerifyFull{
						Data:          data,
						FullSignature: fullSignature,
						Origin:        origin,
					},
				},
			},
		},
	}
}

func VerifyFullResult(next []*types.Event, destModule types1.ModuleID, ok bool, error string, origin *types2.VerifyFullOrigin) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_ThreshCrypto{
			ThreshCrypto: &types2.Event{
				Type: &types2.Event_VerifyFullResult{
					VerifyFullResult: &types2.VerifyFullResult{
						Ok:     ok,
						Error:  error,
						Origin: origin,
					},
				},
			},
		},
	}
}

func Recover(next []*types.Event, destModule types1.ModuleID, data [][]uint8, signatureShares [][]uint8, origin *types2.RecoverOrigin) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_ThreshCrypto{
			ThreshCrypto: &types2.Event{
				Type: &types2.Event_Recover{
					Recover: &types2.Recover{
						Data:            data,
						SignatureShares: signatureShares,
						Origin:          origin,
					},
				},
			},
		},
	}
}

func RecoverResult(next []*types.Event, destModule types1.ModuleID, fullSignature []uint8, ok bool, error string, origin *types2.RecoverOrigin) *types.Event {
	return &types.Event{
		Next:       next,
		DestModule: destModule,
		Type: &types.Event_ThreshCrypto{
			ThreshCrypto: &types2.Event{
				Type: &types2.Event_RecoverResult{
					RecoverResult: &types2.RecoverResult{
						FullSignature: fullSignature,
						Ok:            ok,
						Error:         error,
						Origin:        origin,
					},
				},
			},
		},
	}
}
