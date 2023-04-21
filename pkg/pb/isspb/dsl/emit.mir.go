package isspbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types2 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	events "github.com/filecoin-project/mir/pkg/pb/isspb/events"
	types1 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func PushCheckpoint(m dsl.Module, destModule types.ModuleID) {
	dsl.EmitMirEvent(m, events.PushCheckpoint(destModule))
}

func SBDeliver(m dsl.Module, destModule types.ModuleID, sn types1.SeqNr, data []uint8, aborted bool, leader types.NodeID, instanceId types.ModuleID) {
	dsl.EmitMirEvent(m, events.SBDeliver(destModule, sn, data, aborted, leader, instanceId))
}

func DeliverCert(m dsl.Module, destModule types.ModuleID, sn types1.SeqNr, cert *types2.Cert) {
	dsl.EmitMirEvent(m, events.DeliverCert(destModule, sn, cert))
}

func NewConfig(m dsl.Module, destModule types.ModuleID, epochNr types1.EpochNr, membership *types3.Membership) {
	dsl.EmitMirEvent(m, events.NewConfig(destModule, epochNr, membership))
}
