package availabilitypbevents

import (
	types1 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func RequestCert(destModule types.ModuleID, origin *types1.RequestCertOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Availability{
			Availability: &types1.Event{
				Type: &types1.Event_RequestCert{
					RequestCert: &types1.RequestCert{
						Origin: origin,
					},
				},
			},
		},
	}
}

func NewCert(destModule types.ModuleID, cert *types1.Cert, origin *types1.RequestCertOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Availability{
			Availability: &types1.Event{
				Type: &types1.Event_NewCert{
					NewCert: &types1.NewCert{
						Cert:   cert,
						Origin: origin,
					},
				},
			},
		},
	}
}

func VerifyCert(destModule types.ModuleID, cert *types1.Cert, origin *types1.VerifyCertOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Availability{
			Availability: &types1.Event{
				Type: &types1.Event_VerifyCert{
					VerifyCert: &types1.VerifyCert{
						Cert:   cert,
						Origin: origin,
					},
				},
			},
		},
	}
}

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

func RequestTransactions(destModule types.ModuleID, cert *types1.Cert, origin *types1.RequestTransactionsOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Availability{
			Availability: &types1.Event{
				Type: &types1.Event_RequestTransactions{
					RequestTransactions: &types1.RequestTransactions{
						Cert:   cert,
						Origin: origin,
					},
				},
			},
		},
	}
}

func ProvideTransactions(destModule types.ModuleID, txs []*types3.Transaction, origin *types1.RequestTransactionsOrigin) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Availability{
			Availability: &types1.Event{
				Type: &types1.Event_ProvideTransactions{
					ProvideTransactions: &types1.ProvideTransactions{
						Txs:    txs,
						Origin: origin,
					},
				},
			},
		},
	}
}

func ComputeCert(destModule types.ModuleID) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Availability{
			Availability: &types1.Event{
				Type: &types1.Event_ComputeCert{
					ComputeCert: &types1.ComputeCert{},
				},
			},
		},
	}
}
