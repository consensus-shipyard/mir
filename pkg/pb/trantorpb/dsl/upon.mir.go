package trantorpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	dsl1 "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/dsl"
	types2 "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/types"
	dsl2 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/dsl"
	types3 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	types "github.com/filecoin-project/mir/pkg/trantor/types"
)

// Module-specific dsl functions for processing events.

func UponClientProgress(m dsl.Module, handler func(progress map[types.ClientID]*types1.DeliveredTXs) error) {
	dsl1.UponEvent[*types2.Event_ClientProgress](m, func(ev *types1.ClientProgress) error {
		return handler(ev.Progress)
	})
}

func UponEpochConfig(m dsl.Module, handler func(epochNr types.EpochNr, firstSn types.SeqNr, length uint64, memberships []*types1.Membership) error) {
	dsl2.UponEvent[*types3.Event_EpochConfig](m, func(ev *types1.EpochConfig) error {
		return handler(ev.EpochNr, ev.FirstSn, ev.Length, ev.Memberships)
	})
}
