package pbftpbmsgs

import (
	types2 "github.com/filecoin-project/mir/pkg/orderers/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/ordererpb/types"
	types5 "github.com/filecoin-project/mir/pkg/pb/pbftpb/types"
	types1 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func Preprepare(destModule types.ModuleID, sn types1.SeqNr, view types2.ViewNr, data []uint8, aborted bool) *types3.Message {
	return &types3.Message{
		DestModule: destModule,
		Type: &types3.Message_Orderer{
			Orderer: &types4.Message{
				Type: &types4.Message_Pbft{
					Pbft: &types5.Message{
						Type: &types5.Message_Preprepare{
							Preprepare: &types5.Preprepare{
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

func Prepare(destModule types.ModuleID, sn types1.SeqNr, view types2.ViewNr, digest []uint8) *types3.Message {
	return &types3.Message{
		DestModule: destModule,
		Type: &types3.Message_Orderer{
			Orderer: &types4.Message{
				Type: &types4.Message_Pbft{
					Pbft: &types5.Message{
						Type: &types5.Message_Prepare{
							Prepare: &types5.Prepare{
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

func Commit(destModule types.ModuleID, sn types1.SeqNr, view types2.ViewNr, digest []uint8) *types3.Message {
	return &types3.Message{
		DestModule: destModule,
		Type: &types3.Message_Orderer{
			Orderer: &types4.Message{
				Type: &types4.Message_Pbft{
					Pbft: &types5.Message{
						Type: &types5.Message_Commit{
							Commit: &types5.Commit{
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

func Done(destModule types.ModuleID, digests [][]uint8) *types3.Message {
	return &types3.Message{
		DestModule: destModule,
		Type: &types3.Message_Orderer{
			Orderer: &types4.Message{
				Type: &types4.Message_Pbft{
					Pbft: &types5.Message{
						Type: &types5.Message_Done{
							Done: &types5.Done{
								Digests: digests,
							},
						},
					},
				},
			},
		},
	}
}

func CatchUpRequest(destModule types.ModuleID, digest []uint8, sn types1.SeqNr) *types3.Message {
	return &types3.Message{
		DestModule: destModule,
		Type: &types3.Message_Orderer{
			Orderer: &types4.Message{
				Type: &types4.Message_Pbft{
					Pbft: &types5.Message{
						Type: &types5.Message_CatchUpRequest{
							CatchUpRequest: &types5.CatchUpRequest{
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

func CatchUpResponse(destModule types.ModuleID, resp *types5.Preprepare) *types3.Message {
	return &types3.Message{
		DestModule: destModule,
		Type: &types3.Message_Orderer{
			Orderer: &types4.Message{
				Type: &types4.Message_Pbft{
					Pbft: &types5.Message{
						Type: &types5.Message_CatchUpResponse{
							CatchUpResponse: &types5.CatchUpResponse{
								Resp: resp,
							},
						},
					},
				},
			},
		},
	}
}

func SignedViewChange(destModule types.ModuleID, viewChange *types5.ViewChange, signature []uint8) *types3.Message {
	return &types3.Message{
		DestModule: destModule,
		Type: &types3.Message_Orderer{
			Orderer: &types4.Message{
				Type: &types4.Message_Pbft{
					Pbft: &types5.Message{
						Type: &types5.Message_SignedViewChange{
							SignedViewChange: &types5.SignedViewChange{
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

func PreprepareRequest(destModule types.ModuleID, digest []uint8, sn types1.SeqNr) *types3.Message {
	return &types3.Message{
		DestModule: destModule,
		Type: &types3.Message_Orderer{
			Orderer: &types4.Message{
				Type: &types4.Message_Pbft{
					Pbft: &types5.Message{
						Type: &types5.Message_PreprepareRequest{
							PreprepareRequest: &types5.PreprepareRequest{
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

func MissingPreprepare(destModule types.ModuleID, preprepare *types5.Preprepare) *types3.Message {
	return &types3.Message{
		DestModule: destModule,
		Type: &types3.Message_Orderer{
			Orderer: &types4.Message{
				Type: &types4.Message_Pbft{
					Pbft: &types5.Message{
						Type: &types5.Message_MissingPreprepare{
							MissingPreprepare: &types5.MissingPreprepare{
								Preprepare: preprepare,
							},
						},
					},
				},
			},
		},
	}
}

func NewView(destModule types.ModuleID, view types2.ViewNr, viewChangeSenders []string, signedViewChanges []*types5.SignedViewChange, preprepareSeqNrs []types1.SeqNr, preprepares []*types5.Preprepare) *types3.Message {
	return &types3.Message{
		DestModule: destModule,
		Type: &types3.Message_Orderer{
			Orderer: &types4.Message{
				Type: &types4.Message_Pbft{
					Pbft: &types5.Message{
						Type: &types5.Message_NewView{
							NewView: &types5.NewView{
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
