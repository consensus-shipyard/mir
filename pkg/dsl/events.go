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

// SendMessage emits a request event to send a message over the network.
// The message should be processed on the receiving end using UponMessageReceived.
func SendMessage(m Module, destModule t.ModuleID, msg *messagepb.Message, dest []t.NodeID) {
	EmitEvent(m, events.SendMessage(destModule, msg, dest))
}

// SignRequest emits a request event to sign the given message.
// The response should be processed using UponSignResult with the same context type C.
func SignRequest[C any](m Module, destModule t.ModuleID, data [][]byte, context C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &eventpb.SignOrigin{
		Module: m.ModuleID().Pb(),
		Type: &eventpb.SignOrigin_Dsl{
			Dsl: &eventpb.DslOrigin{
				ContextID: contextID.Pb(),
			},
		},
	}
	EmitEvent(m, events.SignRequest(destModule, data, origin))
}

// VerifyNodeSigs emits a signature verification request event for a batch of signatures.
// The response should be processed using UponNodeSigsVerified with the same context type C.
func VerifyNodeSigs[C any](
	m Module,
	destModule t.ModuleID,
	data [][][]byte,
	signatures [][]byte,
	nodeIDs []t.NodeID,
	context C,
) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &eventpb.SigVerOrigin{
		Module: m.ModuleID().Pb(),
		Type: &eventpb.SigVerOrigin_Dsl{
			Dsl: &eventpb.DslOrigin{
				ContextID: contextID.Pb(),
			},
		},
	}

	EmitEvent(m, events.VerifyNodeSigs(destModule, data, signatures, nodeIDs, origin))
}

// VerifyOneNodeSig emits a signature verification request event for one signature.
// This is a wrapper around VerifyNodeSigs.
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

// HashRequest emits a request event to compute hashes of a batch of messages.
// The response should be processed using UponHashResult with the same context type C.
func HashRequest[C any](m Module, destModule t.ModuleID, data [][][]byte, context C) {
	contextID := m.DslHandle().StoreContext(context)

	origin := &eventpb.HashOrigin{
		Module: m.ModuleID().Pb(),
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

// UponSignResult invokes handler when the module receives a response to a request made by SignRequest with the same
// context type C.
func UponSignResult[C any](m Module, handler func(signature []byte, context C) error) {
	RegisterEventHandler(m, func(evTp *eventpb.Event_SignResult) error {
		res := evTp.SignResult

		dslOriginWrapper, ok := res.Origin.Type.(*eventpb.SignOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(ContextID(dslOriginWrapper.Dsl.ContextID))
		context, ok := contextRaw.(C)
		if !ok {
			return nil
		}

		return handler(res.Signature, context)
	})
}

// UponNodeSigsVerified invokes handler when the module receives a response to a request made by VerifyNodeSigs with
// the same context type C.
func UponNodeSigsVerified[C any](
	m Module,
	handler func(nodeIDs []t.NodeID, valid []bool, errs []error, allOK bool, context C) error,
) {
	RegisterEventHandler(m, func(evTp *eventpb.Event_NodeSigsVerified) error {
		ev := evTp.NodeSigsVerified

		dslOriginWrapper, ok := ev.Origin.Type.(*eventpb.SigVerOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(ContextID(dslOriginWrapper.Dsl.ContextID))
		context, ok := contextRaw.(C)
		if !ok {
			return nil
		}

		var nodeIds []t.NodeID
		for _, id := range ev.NodeIds {
			nodeIds = append(nodeIds, t.NodeID(id))
		}

		var errs []error
		for _, err := range ev.Errors {
			errs = append(errs, errors.New(err))
		}

		return handler(nodeIds, ev.Valid, errs, ev.AllOk, context)
	})
}

// UponOneNodeSigVerified is a wrapper around UponNodeSigsVerified that invokes handler on each response in a batch
// separately. May be useful in combination with VerifyOneNodeSig.
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

// UponHashResult invokes handler when the module receives a response to a request made by HashRequest with the same
// context type C.
func UponHashResult[C any](m Module, handler func(hashes [][]byte, context C) error) {
	RegisterEventHandler(m, func(evTp *eventpb.Event_HashResult) error {
		ev := evTp.HashResult

		dslOriginWrapper, ok := ev.Origin.Type.(*eventpb.HashOrigin_Dsl)
		if !ok {
			return nil
		}

		contextRaw := m.DslHandle().RecoverAndCleanupContext(ContextID(dslOriginWrapper.Dsl.ContextID))
		context, ok := contextRaw.(C)
		if !ok {
			return nil
		}

		return handler(ev.Digests, context)
	})
}

// UponMessageReceived invokes handler when the module receives a message over the network.
func UponMessageReceived(m Module, handler func(from t.NodeID, msg *messagepb.Message) error) {
	RegisterEventHandler(m, func(evTp *eventpb.Event_MessageReceived) error {
		ev := evTp.MessageReceived
		return handler(t.NodeID(ev.From), ev.Msg)
	})
}
