package commonpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	commonpb "github.com/filecoin-project/mir/pkg/pb/commonpb"
	events "github.com/filecoin-project/mir/pkg/pb/commonpb/events"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func ClientProgress(m dsl.Module, destModule types.ModuleID, progress map[string]*commonpb.DeliveredReqs) {
	dsl.EmitMirEvent(m, events.ClientProgress(destModule, progress))
}
