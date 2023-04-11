package factorymodulepbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types1 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	factorymodulepb "github.com/filecoin-project/mir/pkg/pb/factorymodulepb"
	types3 "github.com/filecoin-project/mir/pkg/pb/ordererspb/types"
	types "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Factory struct {
	Type Factory_Type
}

type Factory_Type interface {
	mirreflect.GeneratedType
	isFactory_Type()
	Pb() factorymodulepb.Factory_Type
}

type Factory_TypeWrapper[T any] interface {
	Factory_Type
	Unwrap() *T
}

func Factory_TypeFromPb(pb factorymodulepb.Factory_Type) Factory_Type {
	switch pb := pb.(type) {
	case *factorymodulepb.Factory_NewModule:
		return &Factory_NewModule{NewModule: NewModuleFromPb(pb.NewModule)}
	case *factorymodulepb.Factory_GarbageCollect:
		return &Factory_GarbageCollect{GarbageCollect: GarbageCollectFromPb(pb.GarbageCollect)}
	}
	return nil
}

type Factory_NewModule struct {
	NewModule *NewModule
}

func (*Factory_NewModule) isFactory_Type() {}

func (w *Factory_NewModule) Unwrap() *NewModule {
	return w.NewModule
}

func (w *Factory_NewModule) Pb() factorymodulepb.Factory_Type {
	return &factorymodulepb.Factory_NewModule{NewModule: (w.NewModule).Pb()}
}

func (*Factory_NewModule) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorymodulepb.Factory_NewModule]()}
}

type Factory_GarbageCollect struct {
	GarbageCollect *GarbageCollect
}

func (*Factory_GarbageCollect) isFactory_Type() {}

func (w *Factory_GarbageCollect) Unwrap() *GarbageCollect {
	return w.GarbageCollect
}

func (w *Factory_GarbageCollect) Pb() factorymodulepb.Factory_Type {
	return &factorymodulepb.Factory_GarbageCollect{GarbageCollect: (w.GarbageCollect).Pb()}
}

func (*Factory_GarbageCollect) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorymodulepb.Factory_GarbageCollect]()}
}

func FactoryFromPb(pb *factorymodulepb.Factory) *Factory {
	return &Factory{
		Type: Factory_TypeFromPb(pb.Type),
	}
}

func (m *Factory) Pb() *factorymodulepb.Factory {
	return &factorymodulepb.Factory{
		Type: (m.Type).Pb(),
	}
}

func (*Factory) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorymodulepb.Factory]()}
}

type NewModule struct {
	ModuleId       types.ModuleID
	RetentionIndex types.RetentionIndex
	Params         *GeneratorParams
}

func NewModuleFromPb(pb *factorymodulepb.NewModule) *NewModule {
	return &NewModule{
		ModuleId:       (types.ModuleID)(pb.ModuleId),
		RetentionIndex: (types.RetentionIndex)(pb.RetentionIndex),
		Params:         GeneratorParamsFromPb(pb.Params),
	}
}

func (m *NewModule) Pb() *factorymodulepb.NewModule {
	return &factorymodulepb.NewModule{
		ModuleId:       (string)(m.ModuleId),
		RetentionIndex: (uint64)(m.RetentionIndex),
		Params:         (m.Params).Pb(),
	}
}

func (*NewModule) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorymodulepb.NewModule]()}
}

type GarbageCollect struct {
	RetentionIndex types.RetentionIndex
}

func GarbageCollectFromPb(pb *factorymodulepb.GarbageCollect) *GarbageCollect {
	return &GarbageCollect{
		RetentionIndex: (types.RetentionIndex)(pb.RetentionIndex),
	}
}

func (m *GarbageCollect) Pb() *factorymodulepb.GarbageCollect {
	return &factorymodulepb.GarbageCollect{
		RetentionIndex: (uint64)(m.RetentionIndex),
	}
}

func (*GarbageCollect) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorymodulepb.GarbageCollect]()}
}

type GeneratorParams struct {
	Type GeneratorParams_Type
}

type GeneratorParams_Type interface {
	mirreflect.GeneratedType
	isGeneratorParams_Type()
	Pb() factorymodulepb.GeneratorParams_Type
}

type GeneratorParams_TypeWrapper[T any] interface {
	GeneratorParams_Type
	Unwrap() *T
}

func GeneratorParams_TypeFromPb(pb factorymodulepb.GeneratorParams_Type) GeneratorParams_Type {
	switch pb := pb.(type) {
	case *factorymodulepb.GeneratorParams_MultisigCollector:
		return &GeneratorParams_MultisigCollector{MultisigCollector: types1.InstanceParamsFromPb(pb.MultisigCollector)}
	case *factorymodulepb.GeneratorParams_Checkpoint:
		return &GeneratorParams_Checkpoint{Checkpoint: types2.InstanceParamsFromPb(pb.Checkpoint)}
	case *factorymodulepb.GeneratorParams_EchoTestModule:
		return &GeneratorParams_EchoTestModule{EchoTestModule: EchoModuleParamsFromPb(pb.EchoTestModule)}
	case *factorymodulepb.GeneratorParams_PbftModule:
		return &GeneratorParams_PbftModule{PbftModule: types3.PBFTModuleFromPb(pb.PbftModule)}
	}
	return nil
}

type GeneratorParams_MultisigCollector struct {
	MultisigCollector *types1.InstanceParams
}

func (*GeneratorParams_MultisigCollector) isGeneratorParams_Type() {}

func (w *GeneratorParams_MultisigCollector) Unwrap() *types1.InstanceParams {
	return w.MultisigCollector
}

func (w *GeneratorParams_MultisigCollector) Pb() factorymodulepb.GeneratorParams_Type {
	return &factorymodulepb.GeneratorParams_MultisigCollector{MultisigCollector: (w.MultisigCollector).Pb()}
}

func (*GeneratorParams_MultisigCollector) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorymodulepb.GeneratorParams_MultisigCollector]()}
}

type GeneratorParams_Checkpoint struct {
	Checkpoint *types2.InstanceParams
}

func (*GeneratorParams_Checkpoint) isGeneratorParams_Type() {}

func (w *GeneratorParams_Checkpoint) Unwrap() *types2.InstanceParams {
	return w.Checkpoint
}

func (w *GeneratorParams_Checkpoint) Pb() factorymodulepb.GeneratorParams_Type {
	return &factorymodulepb.GeneratorParams_Checkpoint{Checkpoint: (w.Checkpoint).Pb()}
}

func (*GeneratorParams_Checkpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorymodulepb.GeneratorParams_Checkpoint]()}
}

type GeneratorParams_EchoTestModule struct {
	EchoTestModule *EchoModuleParams
}

func (*GeneratorParams_EchoTestModule) isGeneratorParams_Type() {}

func (w *GeneratorParams_EchoTestModule) Unwrap() *EchoModuleParams {
	return w.EchoTestModule
}

func (w *GeneratorParams_EchoTestModule) Pb() factorymodulepb.GeneratorParams_Type {
	return &factorymodulepb.GeneratorParams_EchoTestModule{EchoTestModule: (w.EchoTestModule).Pb()}
}

func (*GeneratorParams_EchoTestModule) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorymodulepb.GeneratorParams_EchoTestModule]()}
}

type GeneratorParams_PbftModule struct {
	PbftModule *types3.PBFTModule
}

func (*GeneratorParams_PbftModule) isGeneratorParams_Type() {}

func (w *GeneratorParams_PbftModule) Unwrap() *types3.PBFTModule {
	return w.PbftModule
}

func (w *GeneratorParams_PbftModule) Pb() factorymodulepb.GeneratorParams_Type {
	return &factorymodulepb.GeneratorParams_PbftModule{PbftModule: (w.PbftModule).Pb()}
}

func (*GeneratorParams_PbftModule) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorymodulepb.GeneratorParams_PbftModule]()}
}

func GeneratorParamsFromPb(pb *factorymodulepb.GeneratorParams) *GeneratorParams {
	return &GeneratorParams{
		Type: GeneratorParams_TypeFromPb(pb.Type),
	}
}

func (m *GeneratorParams) Pb() *factorymodulepb.GeneratorParams {
	return &factorymodulepb.GeneratorParams{
		Type: (m.Type).Pb(),
	}
}

func (*GeneratorParams) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorymodulepb.GeneratorParams]()}
}

type EchoModuleParams struct {
	Prefix string
}

func EchoModuleParamsFromPb(pb *factorymodulepb.EchoModuleParams) *EchoModuleParams {
	return &EchoModuleParams{
		Prefix: pb.Prefix,
	}
}

func (m *EchoModuleParams) Pb() *factorymodulepb.EchoModuleParams {
	return &factorymodulepb.EchoModuleParams{
		Prefix: m.Prefix,
	}
}

func (*EchoModuleParams) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorymodulepb.EchoModuleParams]()}
}
