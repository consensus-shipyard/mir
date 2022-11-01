package threshcryptopbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/threshcryptopb/types"
	types2 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing events.

func UponEvent[W types.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponMirEvent[*types1.Event_ThreshCrypto](m, func(ev *types.Event) error {
		w, ok := ev.Type.(W)
		if !ok {
			return nil
		}

		return handler(w.Unwrap())
	})
}

func UponSignShare(m dsl.Module, handler func(data [][]uint8, origin *types.SignShareOrigin) error) {
	UponEvent[*types.Event_SignShare](m, func(ev *types.SignShare) error {
		return handler(ev.Data, ev.Origin)
	})
}

func UponSignShareResult[C any](m dsl.Module, handler func(signatureShare []uint8, context *C) error) {
	UponEvent[*types.Event_SignShareResult](m, func(ev *types.SignShareResult) error {
		originWrapper, ok := ev.Origin.Type.(*types.SignShareOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.SignatureShare, context)
	})
}

func UponVerifyShare(m dsl.Module, handler func(data [][]uint8, signatureShare []uint8, nodeId types2.NodeID, origin *types.VerifyShareOrigin) error) {
	UponEvent[*types.Event_VerifyShare](m, func(ev *types.VerifyShare) error {
		return handler(ev.Data, ev.SignatureShare, ev.NodeId, ev.Origin)
	})
}

func UponVerifyShareResult[C any](m dsl.Module, handler func(ok bool, error string, context *C) error) {
	UponEvent[*types.Event_VerifyShareResult](m, func(ev *types.VerifyShareResult) error {
		originWrapper, ok := ev.Origin.Type.(*types.VerifyShareOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.Ok, ev.Error, context)
	})
}

func UponVerifyFull(m dsl.Module, handler func(data [][]uint8, fullSignature []uint8, origin *types.VerifyFullOrigin) error) {
	UponEvent[*types.Event_VerifyFull](m, func(ev *types.VerifyFull) error {
		return handler(ev.Data, ev.FullSignature, ev.Origin)
	})
}

func UponVerifyFullResult[C any](m dsl.Module, handler func(ok bool, error string, context *C) error) {
	UponEvent[*types.Event_VerifyFullResult](m, func(ev *types.VerifyFullResult) error {
		originWrapper, ok := ev.Origin.Type.(*types.VerifyFullOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.Ok, ev.Error, context)
	})
}

func UponRecover(m dsl.Module, handler func(data [][]uint8, signatureShares [][]uint8, origin *types.RecoverOrigin) error) {
	UponEvent[*types.Event_Recover](m, func(ev *types.Recover) error {
		return handler(ev.Data, ev.SignatureShares, ev.Origin)
	})
}

func UponRecoverResult[C any](m dsl.Module, handler func(fullSignature []uint8, ok bool, error string, context *C) error) {
	UponEvent[*types.Event_RecoverResult](m, func(ev *types.RecoverResult) error {
		originWrapper, ok := ev.Origin.Type.(*types.RecoverOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.FullSignature, ev.Ok, ev.Error, context)
	})
}
