package commonpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/commonpb/events"
	types2 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types1 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func ClientProgress(m dsl.Module, destModule types.ModuleID, progress map[types1.ClientID]*types2.DeliveredReqs) {
	dsl.EmitMirEvent(m, events.ClientProgress(destModule, progress))
}

func EpochConfig(m dsl.Module, destModule types.ModuleID, epochNr types1.EpochNr, firstSn types1.SeqNr, length uint64, memberships []*types2.Membership) {
	dsl.EmitMirEvent(m, events.EpochConfig(destModule, epochNr, firstSn, length, memberships))
}
