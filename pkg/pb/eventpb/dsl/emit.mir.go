package eventpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types3 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	events "github.com/filecoin-project/mir/pkg/pb/eventpb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types5 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func Init(m dsl.Module, destModule types.ModuleID) {
	dsl.EmitMirEvent(m, events.Init(destModule))
}

func TimerDelay(m dsl.Module, destModule types.ModuleID, eventsToDelay []*types1.Event, delay types.TimeDuration) {
	dsl.EmitMirEvent(m, events.TimerDelay(destModule, eventsToDelay, delay))
}

func TimerRepeat(m dsl.Module, destModule types.ModuleID, eventsToRepeat []*types1.Event, delay types.TimeDuration, retentionIndex types.RetentionIndex) {
	dsl.EmitMirEvent(m, events.TimerRepeat(destModule, eventsToRepeat, delay, retentionIndex))
}

func TimerGarbageCollect(m dsl.Module, destModule types.ModuleID, retentionIndex types.RetentionIndex) {
	dsl.EmitMirEvent(m, events.TimerGarbageCollect(destModule, retentionIndex))
}

func SendMessage(m dsl.Module, destModule types.ModuleID, msg *types2.Message, destinations []types.NodeID) {
	dsl.EmitMirEvent(m, events.SendMessage(destModule, msg, destinations))
}

func MessageReceived(m dsl.Module, destModule types.ModuleID, from types.NodeID, msg *types2.Message) {
	dsl.EmitMirEvent(m, events.MessageReceived(destModule, from, msg))
}

func AppSnapshotRequest(m dsl.Module, destModule types.ModuleID, replyTo types.ModuleID) {
	dsl.EmitMirEvent(m, events.AppSnapshotRequest(destModule, replyTo))
}

func AppRestoreState(m dsl.Module, destModule types.ModuleID, checkpoint *types3.StableCheckpoint) {
	dsl.EmitMirEvent(m, events.AppRestoreState(destModule, checkpoint))
}

func NewEpoch(m dsl.Module, destModule types.ModuleID, epochNr types.EpochNr) {
	dsl.EmitMirEvent(m, events.NewEpoch(destModule, epochNr))
}

func NewConfig(m dsl.Module, destModule types.ModuleID, epochNr types.EpochNr, membership *types4.Membership) {
	dsl.EmitMirEvent(m, events.NewConfig(destModule, epochNr, membership))
}

func NewRequests(m dsl.Module, destModule types.ModuleID, requests []*types5.Request) {
	dsl.EmitMirEvent(m, events.NewRequests(destModule, requests))
}
