package transportpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types1 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	events "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func SendMessage(m dsl.Module, destModule types.ModuleID, msg *types1.Message, destinations []types.NodeID) {
	dsl.EmitMirEvent(m, events.SendMessage(destModule, msg, destinations))
}

func MessageReceived(m dsl.Module, destModule types.ModuleID, from types.NodeID, msg *types1.Message) {
	dsl.EmitMirEvent(m, events.MessageReceived(destModule, from, msg))
}
