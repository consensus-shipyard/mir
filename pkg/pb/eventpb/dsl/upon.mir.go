package eventpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types2 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing events.

func UponSendMessage(m dsl.Module, handler func(msg *types.Message, destinations []types1.NodeID) error) {
	dsl.UponMirEvent[*types2.Event_SendMessage](m, func(ev *types2.SendMessage) error {
		return handler(ev.Msg, ev.Destinations)
	})
}

func UponMessageReceived(m dsl.Module, handler func(from types1.NodeID, msg *types.Message) error) {
	dsl.UponMirEvent[*types2.Event_MessageReceived](m, func(ev *types2.MessageReceived) error {
		return handler(ev.From, ev.Msg)
	})
}
