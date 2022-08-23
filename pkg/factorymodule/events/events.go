package events

import (
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/factorymodulepb"
	t "github.com/filecoin-project/mir/pkg/types"
)

func NewModule(
	dest t.ModuleID,
	id t.ModuleID,
	retentionIndex t.RetentionIndex,
	params *factorymodulepb.GeneratorParams,
) *eventpb.Event {
	return events.Factory(dest, &factorymodulepb.Factory{Type: &factorymodulepb.Factory_NewModule{
		NewModule: &factorymodulepb.NewModule{
			ModuleId:       id.Pb(),
			RetentionIndex: retentionIndex.Pb(),
			Params:         params,
		},
	}})
}

func GarbageCollect(
	dest t.ModuleID,
	retentionIndex t.RetentionIndex,
) *eventpb.Event {
	return events.Factory(dest, &factorymodulepb.Factory{Type: &factorymodulepb.Factory_GarbageCollect{
		GarbageCollect: &factorymodulepb.GarbageCollect{
			RetentionIndex: retentionIndex.Pb(),
		},
	}})
}

func EchoModuleParams(prefix string) *factorymodulepb.GeneratorParams {
	return &factorymodulepb.GeneratorParams{Type: &factorymodulepb.GeneratorParams_EchoTestModule{
		EchoTestModule: &factorymodulepb.EchoModuleParams{
			Prefix: prefix,
		},
	}}
}
