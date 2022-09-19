package dsl

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/pb/dslpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	pb "github.com/filecoin-project/mir/pkg/pb/threshcryptopb"

	tce "github.com/filecoin-project/mir/pkg/threshcrypto/events"
	t "github.com/filecoin-project/mir/pkg/types"
)

// DSL functions for emitting events.

// SignShare emits a request event to (threshold-)sign the given message.
// The response should be processed using UponSignShareResult with the same context type C.
// C can be an arbitrary type and does not have to be serializable.
func SignShare[C any](m dsl.Module, dest t.ModuleID, data [][]byte, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &pb.SignShareOrigin{
		Module: m.ModuleID().Pb(),
		Type: &pb.SignShareOrigin_Dsl{
			Dsl: &dslpb.Origin{
				ContextID: contextID.Pb(),
			},
		},
	}

	dsl.EmitEvent(m, tce.SignShare(dest, data, origin))
}

// VerifyShare emits a signature share verification request event.
// The response should be processed using UponVerifyShareResult with the same context type C.
// C can be an arbitrary type and does not have to be serializable.
func VerifyShare[C any](m dsl.Module, dest t.ModuleID, data [][]byte, sigShare []byte, nodeID t.NodeID, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &pb.VerifyShareOrigin{
		Module: m.ModuleID().Pb(),
		Type: &pb.VerifyShareOrigin_Dsl{
			Dsl: &dslpb.Origin{
				ContextID: contextID.Pb(),
			},
		},
	}

	dsl.EmitEvent(m, tce.VerifyShare(dest, data, sigShare, nodeID, origin))
}

// VerifyFull emits a (full) (threshold) signature verification request event.
// The response should be processed using UponVerifyFullResult with the same context type C.
// C can be an arbitrary type and does not have to be serializable.
func VerifyFull[C any](m dsl.Module, dest t.ModuleID, data [][]byte, sigFull []byte, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &pb.VerifyFullOrigin{
		Module: m.ModuleID().Pb(),
		Type: &pb.VerifyFullOrigin_Dsl{
			Dsl: &dslpb.Origin{
				ContextID: contextID.Pb(),
			},
		},
	}

	dsl.EmitEvent(m, tce.VerifyFull(dest, data, sigFull, origin))
}

// Recover emits a signature recovery request event.
// The response should be processed using UponRecoverResult with the same context type C.
// C can be an arbitrary type and does not have to be serializable.
func Recover[C any](m dsl.Module, dest t.ModuleID, data [][]byte, sigShares [][]byte, context *C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &pb.RecoverOrigin{
		Module: m.ModuleID().Pb(),
		Type: &pb.RecoverOrigin_Dsl{
			Dsl: &dslpb.Origin{
				ContextID: contextID.Pb(),
			},
		},
	}

	dsl.EmitEvent(m, tce.Recover(dest, data, sigShares, origin))
}

// DSL functions for processing events.

// UponSignShareResult invokes handler when the module receives a response to
// a request made by SignShare with the same context type C.
func UponSignShareResult[C any](m dsl.Module, handler func(sigShare []byte, context *C) error) {
	UponEvent[*pb.Event_SignShareResult](m, func(ev *pb.SignShareResult) error {
		OriginWrapper, ok := ev.Origin.Type.(*pb.SignShareOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(OriginWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.SignatureShare, context)
	})
}

// UponVerifyShareResult invokes handler when the module receives a response to
// a request made by VerifyShare with the same context type C.
func UponVerifyShareResult[C any](m dsl.Module, handler func(ok bool, err string, context *C) error) {
	UponEvent[*pb.Event_VerifyShareResult](m, func(ev *pb.VerifyShareResult) error {
		OriginWrapper, ok := ev.Origin.Type.(*pb.VerifyShareOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(OriginWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.Ok, ev.Error, context)
	})
}

// UponVerifyFullResult invokes handler when the module receives a response to
// a request made by VerifyFull with the same context type C.
func UponVerifyFullResult[C any](m dsl.Module, handler func(ok bool, err string, context *C) error) {
	UponEvent[*pb.Event_VerifyFullResult](m, func(ev *pb.VerifyFullResult) error {
		OriginWrapper, ok := ev.Origin.Type.(*pb.VerifyFullOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(OriginWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.Ok, ev.Error, context)
	})
}

// UponRecoverResult invokes handler when the module receives a response to a request made by Recover
// with the same context type C.
func UponRecoverResult[C any](m dsl.Module, handler func(ok bool, fullSig []byte, err string, context *C) error) {
	UponEvent[*pb.Event_RecoverResult](m, func(ev *pb.RecoverResult) error {
		OriginWrapper, ok := ev.Origin.Type.(*pb.RecoverOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(dsl.ContextID(OriginWrapper.Dsl.ContextID))
		context, ok := contextRaw.(*C)
		if !ok {
			return nil
		}

		return handler(ev.Ok, ev.FullSignature, ev.Error, context)
	})
}

// UponEvent registers a handler for the given threshcrypto event type.
func UponEvent[EvWrapper pb.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponEvent[*eventpb.Event_ThreshCrypto](m, func(ev *pb.Event) error {
		evWrapper, ok := ev.Type.(EvWrapper)
		if !ok {
			return nil
		}
		return handler(evWrapper.Unwrap())
	})
}
