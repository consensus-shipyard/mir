package commonpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	dsl1 "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/dsl"
	types1 "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/types"
	dsl2 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/dsl"
	types3 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types2 "github.com/filecoin-project/mir/pkg/trantor/types"
)

// Module-specific dsl functions for processing events.

func UponClientProgress(m dsl.Module, handler func(progress map[string]*types.DeliveredReqs) error) {
	dsl1.UponEvent[*types1.Event_ClientProgress](m, func(ev *types.ClientProgress) error {
		return handler(ev.Progress)
	})
}

func UponEpochConfig(m dsl.Module, handler func(epochNr types2.EpochNr, firstSn types2.SeqNr, length uint64, memberships []*types.Membership) error) {
	dsl2.UponEvent[*types3.Event_EpochConfig](m, func(ev *types.EpochConfig) error {
		return handler(ev.EpochNr, ev.FirstSn, ev.Length, ev.Memberships)
	})
}
