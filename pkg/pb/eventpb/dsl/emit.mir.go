package eventpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/eventpb/events"
	types "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func MessageReceived(m dsl.Module, next []*types.Event, destModule types1.ModuleID, from types1.NodeID, msg *types2.Message) {
	dsl.EmitMirEvent(m, events.MessageReceived(next, destModule, from, msg))
}
