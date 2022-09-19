package dsl

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/pb/dslpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	pb "github.com/filecoin-project/mir/pkg/pb/threshcryptopb"

	tce "github.com/filecoin-project/mir/pkg/threshcrypto/events"
	t "github.com/filecoin-project/mir/pkg/types"
)

// TODO: document

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

func UponEvent[EvWrapper pb.Event_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponEvent[*eventpb.Event_ThreshCrypto](m, func(ev *pb.Event) error {
		evWrapper, ok := ev.Type.(EvWrapper)
		if !ok {
			return nil
		}
		return handler(evWrapper.Unwrap())
	})
}
