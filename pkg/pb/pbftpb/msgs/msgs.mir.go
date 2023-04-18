package pbftpbmsgs

import (
	types1 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/ordererpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func Preprepare(destModule types.ModuleID, sn types.SeqNr, view types.PBFTViewNr, data []uint8, aborted bool) *types1.Message {
	return &types1.Message{
		DestModule: destModule,
		Type: &types1.Message_Orderer{
			Orderer: &types2.Message{
				Type: &types2.Message_Pbft{
					Pbft: &types3.Message{
						Type: &types3.Message_Preprepare{
							Preprepare: &types3.Preprepare{
								Sn:      sn,
								View:    view,
								Data:    data,
								Aborted: aborted,
							},
						},
					},
				},
			},
		},
	}
}

func Prepare(destModule types.ModuleID, sn types.SeqNr, view types.PBFTViewNr, digest []uint8) *types1.Message {
	return &types1.Message{
		DestModule: destModule,
		Type: &types1.Message_Orderer{
			Orderer: &types2.Message{
				Type: &types2.Message_Pbft{
					Pbft: &types3.Message{
						Type: &types3.Message_Prepare{
							Prepare: &types3.Prepare{
								Sn:     sn,
								View:   view,
								Digest: digest,
							},
						},
					},
				},
			},
		},
	}
}

func Commit(destModule types.ModuleID, sn types.SeqNr, view types.PBFTViewNr, digest []uint8) *types1.Message {
	return &types1.Message{
		DestModule: destModule,
		Type: &types1.Message_Orderer{
			Orderer: &types2.Message{
				Type: &types2.Message_Pbft{
					Pbft: &types3.Message{
						Type: &types3.Message_Commit{
							Commit: &types3.Commit{
								Sn:     sn,
								View:   view,
								Digest: digest,
							},
						},
					},
				},
			},
		},
	}
}

func Done(destModule types.ModuleID, digests [][]uint8) *types1.Message {
	return &types1.Message{
		DestModule: destModule,
		Type: &types1.Message_Orderer{
			Orderer: &types2.Message{
				Type: &types2.Message_Pbft{
					Pbft: &types3.Message{
						Type: &types3.Message_Done{
							Done: &types3.Done{
								Digests: digests,
							},
						},
					},
				},
			},
		},
	}
}

func CatchUpRequest(destModule types.ModuleID, digest []uint8, sn types.SeqNr) *types1.Message {
	return &types1.Message{
		DestModule: destModule,
		Type: &types1.Message_Orderer{
			Orderer: &types2.Message{
				Type: &types2.Message_Pbft{
					Pbft: &types3.Message{
						Type: &types3.Message_CatchUpRequest{
							CatchUpRequest: &types3.CatchUpRequest{
								Digest: digest,
								Sn:     sn,
							},
						},
					},
				},
			},
		},
	}
}

func CatchUpResponse(destModule types.ModuleID, resp *types3.Preprepare) *types1.Message {
	return &types1.Message{
		DestModule: destModule,
		Type: &types1.Message_Orderer{
			Orderer: &types2.Message{
				Type: &types2.Message_Pbft{
					Pbft: &types3.Message{
						Type: &types3.Message_CatchUpResponse{
							CatchUpResponse: &types3.CatchUpResponse{
								Resp: resp,
							},
						},
					},
				},
			},
		},
	}
}

func SignedViewChange(destModule types.ModuleID, viewChange *types3.ViewChange, signature []uint8) *types1.Message {
	return &types1.Message{
		DestModule: destModule,
		Type: &types1.Message_Orderer{
			Orderer: &types2.Message{
				Type: &types2.Message_Pbft{
					Pbft: &types3.Message{
						Type: &types3.Message_SignedViewChange{
							SignedViewChange: &types3.SignedViewChange{
								ViewChange: viewChange,
								Signature:  signature,
							},
						},
					},
				},
			},
		},
	}
}

func PreprepareRequest(destModule types.ModuleID, digest []uint8, sn types.SeqNr) *types1.Message {
	return &types1.Message{
		DestModule: destModule,
		Type: &types1.Message_Orderer{
			Orderer: &types2.Message{
				Type: &types2.Message_Pbft{
					Pbft: &types3.Message{
						Type: &types3.Message_PreprepareRequest{
							PreprepareRequest: &types3.PreprepareRequest{
								Digest: digest,
								Sn:     sn,
							},
						},
					},
				},
			},
		},
	}
}

func MissingPreprepare(destModule types.ModuleID, preprepare *types3.Preprepare) *types1.Message {
	return &types1.Message{
		DestModule: destModule,
		Type: &types1.Message_Orderer{
			Orderer: &types2.Message{
				Type: &types2.Message_Pbft{
					Pbft: &types3.Message{
						Type: &types3.Message_MissingPreprepare{
							MissingPreprepare: &types3.MissingPreprepare{
								Preprepare: preprepare,
							},
						},
					},
				},
			},
		},
	}
}

func NewView(destModule types.ModuleID, view types.PBFTViewNr, viewChangeSenders []string, signedViewChanges []*types3.SignedViewChange, preprepareSeqNrs []types.SeqNr, preprepares []*types3.Preprepare) *types1.Message {
	return &types1.Message{
		DestModule: destModule,
		Type: &types1.Message_Orderer{
			Orderer: &types2.Message{
				Type: &types2.Message_Pbft{
					Pbft: &types3.Message{
						Type: &types3.Message_NewView{
							NewView: &types3.NewView{
								View:              view,
								ViewChangeSenders: viewChangeSenders,
								SignedViewChanges: signedViewChanges,
								PreprepareSeqNrs:  preprepareSeqNrs,
								Preprepares:       preprepares,
							},
						},
					},
				},
			},
		},
	}
}
