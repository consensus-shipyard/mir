package factorypb

import (
	mscpb "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb"
	checkpointpb "github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	ordererpb "github.com/filecoin-project/mir/pkg/pb/ordererpb"
)

type Event_Type = isEvent_Type

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func (w *Event_NewModule) Unwrap() *NewModule {
	return w.NewModule
}

func (w *Event_GarbageCollect) Unwrap() *GarbageCollect {
	return w.GarbageCollect
}

type GeneratorParams_Type = isGeneratorParams_Type

type GeneratorParams_TypeWrapper[T any] interface {
	GeneratorParams_Type
	Unwrap() *T
}

func (w *GeneratorParams_MultisigCollector) Unwrap() *mscpb.InstanceParams {
	return w.MultisigCollector
}

func (w *GeneratorParams_Checkpoint) Unwrap() *checkpointpb.InstanceParams {
	return w.Checkpoint
}

func (w *GeneratorParams_EchoTestModule) Unwrap() *EchoModuleParams {
	return w.EchoTestModule
}

func (w *GeneratorParams_PbftModule) Unwrap() *ordererpb.PBFTModule {
	return w.PbftModule
}
