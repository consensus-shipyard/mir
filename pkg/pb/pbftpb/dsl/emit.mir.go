package pbftpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types1 "github.com/filecoin-project/mir/pkg/orderers/types"
	events "github.com/filecoin-project/mir/pkg/pb/pbftpb/events"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func ProposeTimeout(m dsl.Module, destModule types.ModuleID, proposeTimeout uint64) {
	dsl.EmitMirEvent(m, events.ProposeTimeout(destModule, proposeTimeout))
}

func VCSNTimeout(m dsl.Module, destModule types.ModuleID, view types1.ViewNr, numCommitted uint64) {
	dsl.EmitMirEvent(m, events.VCSNTimeout(destModule, view, numCommitted))
}

func ViewChangeSegTimeout(m dsl.Module, destModule types.ModuleID, viewChangeSegTimeout uint64) {
	dsl.EmitMirEvent(m, events.ViewChangeSegTimeout(destModule, viewChangeSegTimeout))
}
