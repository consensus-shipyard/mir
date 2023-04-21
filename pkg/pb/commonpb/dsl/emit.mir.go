package commonpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/commonpb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types2 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func ClientProgress(m dsl.Module, destModule types.ModuleID, progress map[string]*types1.DeliveredReqs) {
	dsl.EmitMirEvent(m, events.ClientProgress(destModule, progress))
}

func EpochConfig(m dsl.Module, destModule types.ModuleID, epochNr types2.EpochNr, firstSn types2.SeqNr, length uint64, memberships []*types1.Membership) {
	dsl.EmitMirEvent(m, events.EpochConfig(destModule, epochNr, firstSn, length, memberships))
}
