package cryptopbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing events.

func UponEvent[W types.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponMirEvent[*types1.Event_Crypto](m, func(ev *types.Event) error {
		w, ok := ev.Type.(W)
		if !ok {
			return nil
		}

		return handler(w.Unwrap())
	})
}

func UponSignRequest(m dsl.Module, handler func(data [][]uint8, origin *types.SignOrigin) error) {
	UponEvent[*types.Event_SignRequest](m, func(ev *types.SignRequest) error {
		return handler(ev.Data, ev.Origin)
	})
}

func UponSignResult[C any](m dsl.Module, handler func(signature []uint8, context *C) error) {
	UponEvent[*types.Event_SignResult](m, func(ev *types.SignResult) error {
		originWrapper, ok := ev.Origin.Type.(*types.SignOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.Signature, context)
	})
}

func UponVerifySig(m dsl.Module, handler func(data *types.SigVerData, signature []uint8, origin *types.SigVerOrigin, nodeId types2.NodeID) error) {
	UponEvent[*types.Event_VerifySig](m, func(ev *types.VerifySig) error {
		return handler(ev.Data, ev.Signature, ev.Origin, ev.NodeId)
	})
}

func UponSigVerified[C any](m dsl.Module, handler func(nodeId types2.NodeID, valid bool, error error, context *C) error) {
	UponEvent[*types.Event_SigVerified](m, func(ev *types.SigVerified) error {
		originWrapper, ok := ev.Origin.Type.(*types.SigVerOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.NodeId, ev.Valid, ev.Error, context)
	})
}

func UponVerifySigs(m dsl.Module, handler func(data []*types.SigVerData, signatures [][]uint8, origin *types.SigVerOrigin, nodeIds []types2.NodeID) error) {
	UponEvent[*types.Event_VerifySigs](m, func(ev *types.VerifySigs) error {
		return handler(ev.Data, ev.Signatures, ev.Origin, ev.NodeIds)
	})
}

func UponSigsVerified[C any](m dsl.Module, handler func(nodeIds []types2.NodeID, valid []bool, errors []error, allOk bool, context *C) error) {
	UponEvent[*types.Event_SigsVerified](m, func(ev *types.SigsVerified) error {
		originWrapper, ok := ev.Origin.Type.(*types.SigVerOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(originWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.NodeIds, ev.Valid, ev.Errors, ev.AllOk, context)
	})
}
