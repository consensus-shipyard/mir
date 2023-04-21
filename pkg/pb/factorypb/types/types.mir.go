package factorypbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types2 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	factorypb "github.com/filecoin-project/mir/pkg/pb/factorypb"
	types4 "github.com/filecoin-project/mir/pkg/pb/ordererpb/types"
	types1 "github.com/filecoin-project/mir/pkg/trantor/types"
	types "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Event struct {
	Type Event_Type
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() factorypb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb factorypb.Event_Type) Event_Type {
	switch pb := pb.(type) {
	case *factorypb.Event_NewModule:
		return &Event_NewModule{NewModule: NewModuleFromPb(pb.NewModule)}
	case *factorypb.Event_GarbageCollect:
		return &Event_GarbageCollect{GarbageCollect: GarbageCollectFromPb(pb.GarbageCollect)}
	}
	return nil
}

type Event_NewModule struct {
	NewModule *NewModule
}

func (*Event_NewModule) isEvent_Type() {}

func (w *Event_NewModule) Unwrap() *NewModule {
	return w.NewModule
}

func (w *Event_NewModule) Pb() factorypb.Event_Type {
	return &factorypb.Event_NewModule{NewModule: (w.NewModule).Pb()}
}

func (*Event_NewModule) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorypb.Event_NewModule]()}
}

type Event_GarbageCollect struct {
	GarbageCollect *GarbageCollect
}

func (*Event_GarbageCollect) isEvent_Type() {}

func (w *Event_GarbageCollect) Unwrap() *GarbageCollect {
	return w.GarbageCollect
}

func (w *Event_GarbageCollect) Pb() factorypb.Event_Type {
	return &factorypb.Event_GarbageCollect{GarbageCollect: (w.GarbageCollect).Pb()}
}

func (*Event_GarbageCollect) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorypb.Event_GarbageCollect]()}
}

func EventFromPb(pb *factorypb.Event) *Event {
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *factorypb.Event {
	return &factorypb.Event{
		Type: (m.Type).Pb(),
	}
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorypb.Event]()}
}

type NewModule struct {
	ModuleId       types.ModuleID
	RetentionIndex types1.RetentionIndex
	Params         *GeneratorParams
}

func NewModuleFromPb(pb *factorypb.NewModule) *NewModule {
	return &NewModule{
		ModuleId:       (types.ModuleID)(pb.ModuleId),
		RetentionIndex: (types1.RetentionIndex)(pb.RetentionIndex),
		Params:         GeneratorParamsFromPb(pb.Params),
	}
}

func (m *NewModule) Pb() *factorypb.NewModule {
	return &factorypb.NewModule{
		ModuleId:       (string)(m.ModuleId),
		RetentionIndex: (uint64)(m.RetentionIndex),
		Params:         (m.Params).Pb(),
	}
}

func (*NewModule) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorypb.NewModule]()}
}

type GarbageCollect struct {
	RetentionIndex types1.RetentionIndex
}

func GarbageCollectFromPb(pb *factorypb.GarbageCollect) *GarbageCollect {
	return &GarbageCollect{
		RetentionIndex: (types1.RetentionIndex)(pb.RetentionIndex),
	}
}

func (m *GarbageCollect) Pb() *factorypb.GarbageCollect {
	return &factorypb.GarbageCollect{
		RetentionIndex: (uint64)(m.RetentionIndex),
	}
}

func (*GarbageCollect) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorypb.GarbageCollect]()}
}

type GeneratorParams struct {
	Type GeneratorParams_Type
}

type GeneratorParams_Type interface {
	mirreflect.GeneratedType
	isGeneratorParams_Type()
	Pb() factorypb.GeneratorParams_Type
}

type GeneratorParams_TypeWrapper[T any] interface {
	GeneratorParams_Type
	Unwrap() *T
}

func GeneratorParams_TypeFromPb(pb factorypb.GeneratorParams_Type) GeneratorParams_Type {
	switch pb := pb.(type) {
	case *factorypb.GeneratorParams_MultisigCollector:
		return &GeneratorParams_MultisigCollector{MultisigCollector: types2.InstanceParamsFromPb(pb.MultisigCollector)}
	case *factorypb.GeneratorParams_Checkpoint:
		return &GeneratorParams_Checkpoint{Checkpoint: types3.InstanceParamsFromPb(pb.Checkpoint)}
	case *factorypb.GeneratorParams_EchoTestModule:
		return &GeneratorParams_EchoTestModule{EchoTestModule: EchoModuleParamsFromPb(pb.EchoTestModule)}
	case *factorypb.GeneratorParams_PbftModule:
		return &GeneratorParams_PbftModule{PbftModule: types4.PBFTModuleFromPb(pb.PbftModule)}
	}
	return nil
}

type GeneratorParams_MultisigCollector struct {
	MultisigCollector *types2.InstanceParams
}

func (*GeneratorParams_MultisigCollector) isGeneratorParams_Type() {}

func (w *GeneratorParams_MultisigCollector) Unwrap() *types2.InstanceParams {
	return w.MultisigCollector
}

func (w *GeneratorParams_MultisigCollector) Pb() factorypb.GeneratorParams_Type {
	return &factorypb.GeneratorParams_MultisigCollector{MultisigCollector: (w.MultisigCollector).Pb()}
}

func (*GeneratorParams_MultisigCollector) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorypb.GeneratorParams_MultisigCollector]()}
}

type GeneratorParams_Checkpoint struct {
	Checkpoint *types3.InstanceParams
}

func (*GeneratorParams_Checkpoint) isGeneratorParams_Type() {}

func (w *GeneratorParams_Checkpoint) Unwrap() *types3.InstanceParams {
	return w.Checkpoint
}

func (w *GeneratorParams_Checkpoint) Pb() factorypb.GeneratorParams_Type {
	return &factorypb.GeneratorParams_Checkpoint{Checkpoint: (w.Checkpoint).Pb()}
}

func (*GeneratorParams_Checkpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorypb.GeneratorParams_Checkpoint]()}
}

type GeneratorParams_EchoTestModule struct {
	EchoTestModule *EchoModuleParams
}

func (*GeneratorParams_EchoTestModule) isGeneratorParams_Type() {}

func (w *GeneratorParams_EchoTestModule) Unwrap() *EchoModuleParams {
	return w.EchoTestModule
}

func (w *GeneratorParams_EchoTestModule) Pb() factorypb.GeneratorParams_Type {
	return &factorypb.GeneratorParams_EchoTestModule{EchoTestModule: (w.EchoTestModule).Pb()}
}

func (*GeneratorParams_EchoTestModule) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorypb.GeneratorParams_EchoTestModule]()}
}

type GeneratorParams_PbftModule struct {
	PbftModule *types4.PBFTModule
}

func (*GeneratorParams_PbftModule) isGeneratorParams_Type() {}

func (w *GeneratorParams_PbftModule) Unwrap() *types4.PBFTModule {
	return w.PbftModule
}

func (w *GeneratorParams_PbftModule) Pb() factorypb.GeneratorParams_Type {
	return &factorypb.GeneratorParams_PbftModule{PbftModule: (w.PbftModule).Pb()}
}

func (*GeneratorParams_PbftModule) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorypb.GeneratorParams_PbftModule]()}
}

func GeneratorParamsFromPb(pb *factorypb.GeneratorParams) *GeneratorParams {
	return &GeneratorParams{
		Type: GeneratorParams_TypeFromPb(pb.Type),
	}
}

func (m *GeneratorParams) Pb() *factorypb.GeneratorParams {
	return &factorypb.GeneratorParams{
		Type: (m.Type).Pb(),
	}
}

func (*GeneratorParams) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorypb.GeneratorParams]()}
}

type EchoModuleParams struct {
	Prefix string
}

func EchoModuleParamsFromPb(pb *factorypb.EchoModuleParams) *EchoModuleParams {
	return &EchoModuleParams{
		Prefix: pb.Prefix,
	}
}

func (m *EchoModuleParams) Pb() *factorypb.EchoModuleParams {
	return &factorypb.EchoModuleParams{
		Prefix: m.Prefix,
	}
}

func (*EchoModuleParams) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*factorypb.EchoModuleParams]()}
}
