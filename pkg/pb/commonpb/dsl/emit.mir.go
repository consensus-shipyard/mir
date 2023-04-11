package commonpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	commonpb "github.com/filecoin-project/mir/pkg/pb/commonpb"
	events "github.com/filecoin-project/mir/pkg/pb/commonpb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func ClientProgress(m dsl.Module, destModule types.ModuleID, progress map[string]*commonpb.DeliveredReqs) {
	dsl.EmitMirEvent(m, events.ClientProgress(destModule, progress))
}

func EpochConfig(m dsl.Module, destModule types.ModuleID, epochNr uint64, firstSn uint64, length uint64, memberships []*types1.Membership) {
	dsl.EmitMirEvent(m, events.EpochConfig(destModule, epochNr, firstSn, length, memberships))
}
