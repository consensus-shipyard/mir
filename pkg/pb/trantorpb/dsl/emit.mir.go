// Code generated by Mir codegen. DO NOT EDIT.

package trantorpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/trantorpb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	types "github.com/filecoin-project/mir/pkg/trantor/types"
	stdtypes "github.com/filecoin-project/mir/stdtypes"
)

// Module-specific dsl functions for emitting events.

func ClientProgress(m dsl.Module, destModule stdtypes.ModuleID, progress map[types.ClientID]*types1.DeliveredTXs) {
	dsl.EmitMirEvent(m, events.ClientProgress(destModule, progress))
}

func EpochConfig(m dsl.Module, destModule stdtypes.ModuleID, epochNr types.EpochNr, firstSn types.SeqNr, length uint64, memberships []*types1.Membership) {
	dsl.EmitMirEvent(m, events.EpochConfig(destModule, epochNr, firstSn, length, memberships))
}
