package pbftpbmsgs

import (
	types1 "github.com/filecoin-project/mir/pkg/orderers/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/ordererpb/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func Preprepare(destModule types.ModuleID, sn types.SeqNr, view types1.ViewNr, data []uint8, aborted bool) *types2.Message {
	return &types2.Message{
		DestModule: destModule,
		Type: &types2.Message_Orderer{
			Orderer: &types3.Message{
				Type: &types3.Message_Pbft{
					Pbft: &types4.Message{
						Type: &types4.Message_Preprepare{
							Preprepare: &types4.Preprepare{
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

func Prepare(destModule types.ModuleID, sn types.SeqNr, view types1.ViewNr, digest []uint8) *types2.Message {
	return &types2.Message{
		DestModule: destModule,
		Type: &types2.Message_Orderer{
			Orderer: &types3.Message{
				Type: &types3.Message_Pbft{
					Pbft: &types4.Message{
						Type: &types4.Message_Prepare{
							Prepare: &types4.Prepare{
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

func Commit(destModule types.ModuleID, sn types.SeqNr, view types1.ViewNr, digest []uint8) *types2.Message {
	return &types2.Message{
		DestModule: destModule,
		Type: &types2.Message_Orderer{
			Orderer: &types3.Message{
				Type: &types3.Message_Pbft{
					Pbft: &types4.Message{
						Type: &types4.Message_Commit{
							Commit: &types4.Commit{
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

func Done(destModule types.ModuleID, digests [][]uint8) *types2.Message {
	return &types2.Message{
		DestModule: destModule,
		Type: &types2.Message_Orderer{
			Orderer: &types3.Message{
				Type: &types3.Message_Pbft{
					Pbft: &types4.Message{
						Type: &types4.Message_Done{
							Done: &types4.Done{
								Digests: digests,
							},
						},
					},
				},
			},
		},
	}
}

func CatchUpRequest(destModule types.ModuleID, digest []uint8, sn types.SeqNr) *types2.Message {
	return &types2.Message{
		DestModule: destModule,
		Type: &types2.Message_Orderer{
			Orderer: &types3.Message{
				Type: &types3.Message_Pbft{
					Pbft: &types4.Message{
						Type: &types4.Message_CatchUpRequest{
							CatchUpRequest: &types4.CatchUpRequest{
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

func CatchUpResponse(destModule types.ModuleID, resp *types4.Preprepare) *types2.Message {
	return &types2.Message{
		DestModule: destModule,
		Type: &types2.Message_Orderer{
			Orderer: &types3.Message{
				Type: &types3.Message_Pbft{
					Pbft: &types4.Message{
						Type: &types4.Message_CatchUpResponse{
							CatchUpResponse: &types4.CatchUpResponse{
								Resp: resp,
							},
						},
					},
				},
			},
		},
	}
}

func SignedViewChange(destModule types.ModuleID, viewChange *types4.ViewChange, signature []uint8) *types2.Message {
	return &types2.Message{
		DestModule: destModule,
		Type: &types2.Message_Orderer{
			Orderer: &types3.Message{
				Type: &types3.Message_Pbft{
					Pbft: &types4.Message{
						Type: &types4.Message_SignedViewChange{
							SignedViewChange: &types4.SignedViewChange{
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

func PreprepareRequest(destModule types.ModuleID, digest []uint8, sn types.SeqNr) *types2.Message {
	return &types2.Message{
		DestModule: destModule,
		Type: &types2.Message_Orderer{
			Orderer: &types3.Message{
				Type: &types3.Message_Pbft{
					Pbft: &types4.Message{
						Type: &types4.Message_PreprepareRequest{
							PreprepareRequest: &types4.PreprepareRequest{
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

func MissingPreprepare(destModule types.ModuleID, preprepare *types4.Preprepare) *types2.Message {
	return &types2.Message{
		DestModule: destModule,
		Type: &types2.Message_Orderer{
			Orderer: &types3.Message{
				Type: &types3.Message_Pbft{
					Pbft: &types4.Message{
						Type: &types4.Message_MissingPreprepare{
							MissingPreprepare: &types4.MissingPreprepare{
								Preprepare: preprepare,
							},
						},
					},
				},
			},
		},
	}
}

func NewView(destModule types.ModuleID, view types1.ViewNr, viewChangeSenders []string, signedViewChanges []*types4.SignedViewChange, preprepareSeqNrs []types.SeqNr, preprepares []*types4.Preprepare) *types2.Message {
	return &types2.Message{
		DestModule: destModule,
		Type: &types2.Message_Orderer{
			Orderer: &types3.Message{
				Type: &types3.Message_Pbft{
					Pbft: &types4.Message{
						Type: &types4.Message_NewView{
							NewView: &types4.NewView{
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
