package commonpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	dsl1 "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/dsl"
	types "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/types"
	commonpb "github.com/filecoin-project/mir/pkg/pb/commonpb"
	types1 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
)

// Module-specific dsl functions for processing events.

func UponClientProgress(m dsl.Module, handler func(progress map[string]*commonpb.DeliveredReqs) error) {
	dsl1.UponEvent[*types.Event_ClientProgress](m, func(ev *types1.ClientProgress) error {
		return handler(ev.Progress)
	})
}
