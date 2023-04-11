package factorymodulepb

import (
	mscpb "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb"
	checkpointpb "github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	ordererspb "github.com/filecoin-project/mir/pkg/pb/ordererspb"
)

type Factory_Type = isFactory_Type

type Factory_TypeWrapper[T any] interface {
	Factory_Type
	Unwrap() *T
}

func (w *Factory_NewModule) Unwrap() *NewModule {
	return w.NewModule
}

func (w *Factory_GarbageCollect) Unwrap() *GarbageCollect {
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

func (w *GeneratorParams_PbftModule) Unwrap() *ordererspb.PBFTModule {
	return w.PbftModule
}
