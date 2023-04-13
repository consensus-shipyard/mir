package eventpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types4 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types5 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types6 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing events.

func UponInit(m dsl.Module, handler func() error) {
	dsl.UponMirEvent[*types.Event_Init](m, func(ev *types.Init) error {
		return handler()
	})
}

func UponTimerEvent[W types.TimerEvent_TypeWrapper[Ev], Ev any](m dsl.Module, handler func(ev *Ev) error) {
	dsl.UponMirEvent[*types.Event_Timer](m, func(ev *types.TimerEvent) error {
		w, ok := ev.Type.(W)
		if !ok {
			return nil
		}

		return handler(w.Unwrap())
	})
}

func UponTimerDelay(m dsl.Module, handler func(evts []*types.Event, delay uint64) error) {
	UponTimerEvent[*types.TimerEvent_Delay](m, func(ev *types.TimerDelay) error {
		return handler(ev.Evts, ev.Delay)
	})
}

func UponTimerRepeat(m dsl.Module, handler func(eventsToRepeat []*types.Event, delay types1.TimeDuration, retentionIndex types1.RetentionIndex) error) {
	UponTimerEvent[*types.TimerEvent_Repeat](m, func(ev *types.TimerRepeat) error {
		return handler(ev.EventsToRepeat, ev.Delay, ev.RetentionIndex)
	})
}

func UponTimerGarbageCollect(m dsl.Module, handler func(retentionIndex types1.RetentionIndex) error) {
	UponTimerEvent[*types.TimerEvent_GarbageCollect](m, func(ev *types.TimerGarbageCollect) error {
		return handler(ev.RetentionIndex)
	})
}

func UponNewRequests(m dsl.Module, handler func(requests []*types2.Request) error) {
	dsl.UponMirEvent[*types.Event_NewRequests](m, func(ev *types.NewRequests) error {
		return handler(ev.Requests)
	})
}

func UponSignRequest(m dsl.Module, handler func(data [][]uint8, origin *types.SignOrigin) error) {
	dsl.UponMirEvent[*types.Event_SignRequest](m, func(ev *types.SignRequest) error {
		return handler(ev.Data, ev.Origin)
	})
}

func UponSignResult[C any](m dsl.Module, handler func(signature []uint8, context *C) error) {
	dsl.UponMirEvent[*types.Event_SignResult](m, func(ev *types.SignResult) error {
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

func UponVerifyNodeSigs(m dsl.Module, handler func(data []*types.SigVerData, signatures [][]uint8, origin *types.SigVerOrigin, nodeIds []types1.NodeID) error) {
	dsl.UponMirEvent[*types.Event_VerifyNodeSigs](m, func(ev *types.VerifyNodeSigs) error {
		return handler(ev.Data, ev.Signatures, ev.Origin, ev.NodeIds)
	})
}

func UponNodeSigsVerified[C any](m dsl.Module, handler func(nodeIds []types1.NodeID, valid []bool, errors []error, allOk bool, context *C) error) {
	dsl.UponMirEvent[*types.Event_NodeSigsVerified](m, func(ev *types.NodeSigsVerified) error {
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

func UponSendMessage(m dsl.Module, handler func(msg *types3.Message, destinations []types1.NodeID) error) {
	dsl.UponMirEvent[*types.Event_SendMessage](m, func(ev *types.SendMessage) error {
		return handler(ev.Msg, ev.Destinations)
	})
}

func UponMessageReceived(m dsl.Module, handler func(from types1.NodeID, msg *types3.Message) error) {
	dsl.UponMirEvent[*types.Event_MessageReceived](m, func(ev *types.MessageReceived) error {
		return handler(ev.From, ev.Msg)
	})
}

func UponDeliverCert(m dsl.Module, handler func(sn types1.SeqNr, cert *types4.Cert) error) {
	dsl.UponMirEvent[*types.Event_DeliverCert](m, func(ev *types.DeliverCert) error {
		return handler(ev.Sn, ev.Cert)
	})
}

func UponAppSnapshotRequest(m dsl.Module, handler func(replyTo types1.ModuleID) error) {
	dsl.UponMirEvent[*types.Event_AppSnapshotRequest](m, func(ev *types.AppSnapshotRequest) error {
		return handler(ev.ReplyTo)
	})
}

func UponAppRestoreState(m dsl.Module, handler func(checkpoint *types5.StableCheckpoint) error) {
	dsl.UponMirEvent[*types.Event_AppRestoreState](m, func(ev *types.AppRestoreState) error {
		return handler(ev.Checkpoint)
	})
}

func UponNewEpoch(m dsl.Module, handler func(epochNr types1.EpochNr) error) {
	dsl.UponMirEvent[*types.Event_NewEpoch](m, func(ev *types.NewEpoch) error {
		return handler(ev.EpochNr)
	})
}

func UponNewConfig(m dsl.Module, handler func(epochNr types1.EpochNr, membership *types6.Membership) error) {
	dsl.UponMirEvent[*types.Event_NewConfig](m, func(ev *types.NewConfig) error {
		return handler(ev.EpochNr, ev.Membership)
	})
}
