package dsl

import (
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/pkg/errors"
)

// Dsl functions for emitting events.
// TODO: add missing event types.
// TODO: consider generating this code automatically using a protoc plugin.

func SendMessage(m Module, destModule t.ModuleID, msg *messagepb.Message, dest []t.NodeID) {
	EmitEvent(m, events.SendMessage(destModule, msg, dest))
}

func SignRequest[C any](m Module, destModule t.ModuleID, data [][]byte, context C) {
	contextID := m.GetDslHandle().StoreContext(context)

	origin := &eventpb.SignOrigin{
		Module: m.GetModuleID().Pb(),
		Type: &eventpb.SignOrigin_Dsl{
			Dsl: &eventpb.DslOrigin{
				ContextID: contextID.Pb(),
			},
		},
	}
	EmitEvent(m, events.SignRequest(destModule, data, origin))
}

func VerifyOneNodeSig[C any](
	m Module,
	destModule t.ModuleID,
	data [][]byte,
	signature []byte,
	nodeID t.NodeID,
	context C,
) {
	VerifyNodeSigs(m, destModule, [][][]byte{data}, [][]byte{signature}, []t.NodeID{nodeID}, context)
}

// VerifyNodeSigs emits a signature verification event for a batch of signatures.
func VerifyNodeSigs[C any](
	m Module,
	destModule t.ModuleID,
	data [][][]byte,
	signatures [][]byte,
	nodeIDs []t.NodeID,
	context C,
) {
	contextID := m.GetDslHandle().StoreContext(context)

	origin := &eventpb.SigVerOrigin{
		Module: m.GetModuleID().Pb(),
		Type: &eventpb.SigVerOrigin_Dsl{
			Dsl: &eventpb.DslOrigin{
				ContextID: contextID.Pb(),
			},
		},
	}

	EmitEvent(m, events.VerifyNodeSigs(destModule, data, signatures, nodeIDs, origin))
}

func HashRequest[C any](m Module, destModule t.ModuleID, data [][][]byte, context C) {
	contextID := m.GetDslHandle().StoreContext(context)

	origin := &eventpb.HashOrigin{
		Module: m.GetModuleID().Pb(),
		Type: &eventpb.HashOrigin_Dsl{
			Dsl: &eventpb.DslOrigin{
				ContextID: contextID.Pb(),
			},
		},
	}

	EmitEvent(m, events.HashRequest(destModule, data, origin))
}

// Dsl functions for processing events
// TODO: consider generating this code automatically using a protoc plugin.

func UponSignResult[C any](m Module, handler func(signature []byte, context C) error) {
	RegisterEventHandler(m, func(evTp *eventpb.Event_SignResult) error {
		res := evTp.SignResult

		dslOriginWrapper, ok := res.Origin.Type.(*eventpb.SignOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.GetDslHandle().RecoverAndCleanupContext(ContextID(dslOriginWrapper.Dsl.ContextID))
		context, ok := contextRaw.(C)
		if !ok {
			return nil
		}

		return handler(res.Signature, context)
	})
}

func UponNodeSigsVerified[C any](
	m Module,
	handler func(nodeIDs []t.NodeID, valid []bool, errs []error, allOK bool, context C) error,
) {
	RegisterEventHandler(m, func(evTp *eventpb.Event_NodeSigsVerified) error {
		res := evTp.NodeSigsVerified

		dslOriginWrapper, ok := res.Origin.Type.(*eventpb.SigVerOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.GetDslHandle().RecoverAndCleanupContext(ContextID(dslOriginWrapper.Dsl.ContextID))
		context, ok := contextRaw.(C)
		if !ok {
			return nil
		}

		var nodeIds []t.NodeID
		for _, id := range res.NodeIds {
			nodeIds = append(nodeIds, t.NodeID(id))
		}

		var errs []error
		for _, err := range res.Errors {
			errs = append(errs, errors.New(err))
		}

		return handler(nodeIds, res.Valid, errs, res.AllOk, context)
	})
}

func UponOneNodeSigVerified[C any](m Module, handler func(nodeID t.NodeID, valid bool, err error, context C) error) {
	UponNodeSigsVerified(m, func(nodeIDs []t.NodeID, valid []bool, errs []error, allOK bool, context C) error {
		for i := range nodeIDs {
			err := handler(nodeIDs[i], valid[i], errs[i], context)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func UponMessageReceived(m Module, handler func(from t.NodeID, msg *messagepb.Message) error) {
	RegisterEventHandler(m, func(evTp *eventpb.Event_MessageReceived) error {
		ev := evTp.MessageReceived
		return handler(t.NodeID(ev.From), ev.Msg)
	})
}
