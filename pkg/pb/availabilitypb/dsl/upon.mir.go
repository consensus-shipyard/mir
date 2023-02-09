package availabilitypbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
)

// Module-specific dsl functions for processing events.

func UponEvent[W types.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponMirEvent[*types1.Event_Availability](m, func(ev *types.Event) error {
		w, ok := ev.Type.(W)
		if !ok {
			return nil
		}

		return handler(w.Unwrap())
	})
}

func UponRequestCert(m dsl.Module, handler func(origin *types.RequestCertOrigin) error) {
	UponEvent[*types.Event_RequestCert](m, func(ev *types.RequestCert) error {
		return handler(ev.Origin)
	})
}

func UponNewCert[C any](m dsl.Module, handler func(cert *types.Cert, context *C) error) {
	UponEvent[*types.Event_NewCert](m, func(ev *types.NewCert) error {
		originWrapper, ok := ev.Origin.Type.(*types.RequestCertOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.Cert, context)
	})
}

func UponVerifyCert(m dsl.Module, handler func(cert *types.Cert, origin *types.VerifyCertOrigin) error) {
	UponEvent[*types.Event_VerifyCert](m, func(ev *types.VerifyCert) error {
		return handler(ev.Cert, ev.Origin)
	})
}

func UponCertVerified[C any](m dsl.Module, handler func(valid bool, err string, context *C) error) {
	UponEvent[*types.Event_CertVerified](m, func(ev *types.CertVerified) error {
		originWrapper, ok := ev.Origin.Type.(*types.VerifyCertOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.Valid, ev.Err, context)
	})
}

func UponRequestTransactions(m dsl.Module, handler func(cert *types.Cert, origin *types.RequestTransactionsOrigin) error) {
	UponEvent[*types.Event_RequestTransactions](m, func(ev *types.RequestTransactions) error {
		return handler(ev.Cert, ev.Origin)
	})
}

func UponProvideTransactions[C any](m dsl.Module, handler func(txs []*types2.Request, context *C) error) {
	UponEvent[*types.Event_ProvideTransactions](m, func(ev *types.ProvideTransactions) error {
		originWrapper, ok := ev.Origin.Type.(*types.RequestTransactionsOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.Txs, context)
	})
}

func UponComputeCert(m dsl.Module, handler func() error) {
	UponEvent[*types.Event_ComputeCert](m, func(ev *types.ComputeCert) error {
		return handler()
	})
}
