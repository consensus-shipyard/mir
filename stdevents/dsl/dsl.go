package dsl

import (
	"time"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
)

func GarbageCollect(
	m dsl.Module,
	destModule stdtypes.ModuleID,
	retentionIndex stdtypes.RetentionIndex,
) {
	dsl.EmitEvent(m, stdevents.NewGarbageCollect(destModule, retentionIndex))
}

func Init(
	m dsl.Module,
	destModule stdtypes.ModuleID,
) {
	dsl.EmitEvent(m, stdevents.NewInit(destModule))
}

func MessageReceived(
	m dsl.Module,
	destModule stdtypes.ModuleID,
	sender stdtypes.NodeID,
	payload stdtypes.Serializable,
) {
	dsl.EmitEvent(m, stdevents.NewMessageReceived(destModule, sender, payload))
}

func NewSubmodule(
	m dsl.Module,
	destModule stdtypes.ModuleID,
	submoduleID stdtypes.ModuleID,
	params stdtypes.Serializable,
	retIdx stdtypes.RetentionIndex,
) {
	dsl.EmitEvent(m, stdevents.NewNewSubmodule(destModule, submoduleID, params, retIdx))
}

func Raw(
	m dsl.Module,
	destModule stdtypes.ModuleID,
	data []byte,
) {
	dsl.EmitEvent(m, stdevents.NewRaw(destModule, data))
}

func SendMessage(
	m dsl.Module,
	localDestModule stdtypes.ModuleID,
	remoteDestModule stdtypes.ModuleID,
	message stdtypes.Message,
	destNodes ...stdtypes.NodeID,
) {
	dsl.EmitEvent(m, stdevents.NewSendMessage(localDestModule, remoteDestModule, message, destNodes...))
}

func TestString(
	m dsl.Module,
	destModule stdtypes.ModuleID,
	value string,
) {
	dsl.EmitEvent(m, stdevents.NewTestString(destModule, value))
}

func TestUint64(
	m dsl.Module,
	destModule stdtypes.ModuleID,
	value uint64,
) {
	dsl.EmitEvent(m, stdevents.NewTestUint64(destModule, value))
}

func TimerDelay(
	m dsl.Module,
	destModule stdtypes.ModuleID,
	delay time.Duration,
	eventsToDelay ...stdtypes.Event,
) {
	dsl.EmitEvent(m, stdevents.NewTimerDelay(destModule, delay, eventsToDelay...))
}

func TimerRepeat(
	m dsl.Module,
	destModule stdtypes.ModuleID,
	period time.Duration,
	retIdx stdtypes.RetentionIndex,
	eventsToRepeat ...stdtypes.Event,
) {
	dsl.EmitEvent(m, stdevents.NewTimerRepeat(destModule, period, retIdx, eventsToRepeat...))
}
