package availabilitypbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/availabilitypb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func RequestCert[C any](m dsl.Module, destModule types.ModuleID, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types1.RequestCertOrigin{
		Module: m.ModuleID(),
		Type:   &types1.RequestCertOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.RequestCert(destModule, origin))
}

func NewCert(m dsl.Module, destModule types.ModuleID, cert *types1.Cert, origin *types1.RequestCertOrigin) {
	dsl.EmitMirEvent(m, events.NewCert(destModule, cert, origin))
}

func VerifyCert[C any](m dsl.Module, destModule types.ModuleID, cert *types1.Cert, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types1.VerifyCertOrigin{
		Module: m.ModuleID(),
		Type:   &types1.VerifyCertOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.VerifyCert(destModule, cert, origin))
}

func CertVerified(m dsl.Module, destModule types.ModuleID, valid bool, err string, origin *types1.VerifyCertOrigin) {
	dsl.EmitMirEvent(m, events.CertVerified(destModule, valid, err, origin))
}

func RequestTransactions[C any](m dsl.Module, destModule types.ModuleID, cert *types1.Cert, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &types1.RequestTransactionsOrigin{
		Module: m.ModuleID(),
		Type:   &types1.RequestTransactionsOrigin_Dsl{Dsl: dsl.MirOrigin(contextID)},
	}

	dsl.EmitMirEvent(m, events.RequestTransactions(destModule, cert, origin))
}

func ProvideTransactions(m dsl.Module, destModule types.ModuleID, txs []*types2.Transaction, origin *types1.RequestTransactionsOrigin) {
	dsl.EmitMirEvent(m, events.ProvideTransactions(destModule, txs, origin))
}

func ComputeCert(m dsl.Module, destModule types.ModuleID) {
	dsl.EmitMirEvent(m, events.ComputeCert(destModule))
}
