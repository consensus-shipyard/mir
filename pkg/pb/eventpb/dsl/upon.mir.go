package eventpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types2 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing events.

func UponMessageReceived(m dsl.Module, handler func(from types.NodeID, msg *types1.Message) error) {
	dsl.UponMirEvent[*types2.Event_MessageReceived](m, func(ev *types2.MessageReceived) error {
		return handler(ev.From, ev.Msg)
	})
}
