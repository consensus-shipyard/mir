package bcbdsl

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/pb/bcbpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func Request(m dsl.Module, dest t.ModuleID, data []byte) {
	dsl.EmitEvent(m, &eventpb.Event{
		DestModule: dest.Pb(),

		Type: &eventpb.Event_Bcb{
			Bcb: &bcbpb.Event{
				Type: &bcbpb.Event_Request{
					Request: &bcbpb.Request{
						Data: data,
					},
				},
			},
		},
	})
}

func Deliver(m dsl.Module, dest t.ModuleID, data []byte) {
	dsl.EmitEvent(m, &eventpb.Event{
		DestModule: dest.Pb(),

		Type: &eventpb.Event_Bcb{
			Bcb: &bcbpb.Event{
				Type: &bcbpb.Event_Deliver{
					Deliver: &bcbpb.Deliver{
						Data: data,
					},
				},
			},
		},
	})
}

// Module-specific dsl functions for processing events.

func UponRequest(m dsl.Module, handler func(data []byte) error) {
	dsl.RegisterEventHandler(m, func(ev *eventpb.Event_Bcb) error {
		requestEvWrapper, ok := ev.Bcb.Type.(*bcbpb.Event_Request)
		if !ok {
			return nil
		}
		return handler(requestEvWrapper.Request.Data)
	})
}

func UponDeliver(m dsl.Module, handler func(data []byte) error) {
	dsl.RegisterEventHandler(m, func(ev *eventpb.Event_Bcb) error {
		deliverEvWrapper, ok := ev.Bcb.Type.(*bcbpb.Event_Deliver)
		if !ok {
			return nil
		}
		return handler(deliverEvWrapper.Deliver.Data)
	})
}

func UponBCBMessageReceived(m dsl.Module, handler func(from t.NodeID, msg *bcbpb.Message) error) {
	dsl.UponMessageReceived(m, func(from t.NodeID, msg *messagepb.Message) error {
		cbMsgWrapper, ok := msg.Type.(*messagepb.Message_Bcb)
		if !ok {
			return nil
		}

		return handler(from, cbMsgWrapper.Bcb)
	})
}

func UponStartMessageReceived(m dsl.Module, handler func(from t.NodeID, data []byte) error) {
	UponBCBMessageReceived(m, func(from t.NodeID, msg *bcbpb.Message) error {
		startMsgWrapper, ok := msg.Type.(*bcbpb.Message_StartMessage)
		if !ok {
			return nil
		}

		return handler(from, startMsgWrapper.StartMessage.Data)
	})
}

func UponEchoMessageReceived(m dsl.Module, handler func(from t.NodeID, signature []byte) error) {
	UponBCBMessageReceived(m, func(from t.NodeID, msg *bcbpb.Message) error {
		echoMsgWrapper, ok := msg.Type.(*bcbpb.Message_EchoMessage)
		if !ok {
			return nil
		}

		return handler(from, echoMsgWrapper.EchoMessage.Signature)
	})
}

func UponFinalMessageReceived(
	m dsl.Module,
	handler func(from t.NodeID, data []byte, signers []t.NodeID, signatures [][]byte) error,
) {
	UponBCBMessageReceived(m, func(from t.NodeID, msg *bcbpb.Message) error {
		finalMsgWrapper, ok := msg.Type.(*bcbpb.Message_FinalMessage)
		if !ok {
			return nil
		}

		finalMsg := finalMsgWrapper.FinalMessage

		var signers []t.NodeID
		for _, node := range finalMsg.Signers {
			signers = append(signers, t.NodeID(node))
		}

		return handler(from, finalMsg.Data, signers, finalMsg.Signatures)
	})
}
