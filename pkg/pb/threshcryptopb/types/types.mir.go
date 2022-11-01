package threshcryptopbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types1 "github.com/filecoin-project/mir/pkg/pb/contextstorepb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/dslpb/types"
	threshcryptopb "github.com/filecoin-project/mir/pkg/pb/threshcryptopb"
	types "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Event struct {
	Type Event_Type
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() threshcryptopb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb threshcryptopb.Event_Type) Event_Type {
	switch pb := pb.(type) {
	case *threshcryptopb.Event_SignShare:
		return &Event_SignShare{SignShare: SignShareFromPb(pb.SignShare)}
	case *threshcryptopb.Event_SignShareResult:
		return &Event_SignShareResult{SignShareResult: SignShareResultFromPb(pb.SignShareResult)}
	case *threshcryptopb.Event_VerifyShare:
		return &Event_VerifyShare{VerifyShare: VerifyShareFromPb(pb.VerifyShare)}
	case *threshcryptopb.Event_VerifyShareResult:
		return &Event_VerifyShareResult{VerifyShareResult: VerifyShareResultFromPb(pb.VerifyShareResult)}
	case *threshcryptopb.Event_VerifyFull:
		return &Event_VerifyFull{VerifyFull: VerifyFullFromPb(pb.VerifyFull)}
	case *threshcryptopb.Event_VerifyFullResult:
		return &Event_VerifyFullResult{VerifyFullResult: VerifyFullResultFromPb(pb.VerifyFullResult)}
	case *threshcryptopb.Event_Recover:
		return &Event_Recover{Recover: RecoverFromPb(pb.Recover)}
	case *threshcryptopb.Event_RecoverResult:
		return &Event_RecoverResult{RecoverResult: RecoverResultFromPb(pb.RecoverResult)}
	}
	return nil
}

type Event_SignShare struct {
	SignShare *SignShare
}

func (*Event_SignShare) isEvent_Type() {}

func (w *Event_SignShare) Unwrap() *SignShare {
	return w.SignShare
}

func (w *Event_SignShare) Pb() threshcryptopb.Event_Type {
	return &threshcryptopb.Event_SignShare{SignShare: (w.SignShare).Pb()}
}

func (*Event_SignShare) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.Event_SignShare]()}
}

type Event_SignShareResult struct {
	SignShareResult *SignShareResult
}

func (*Event_SignShareResult) isEvent_Type() {}

func (w *Event_SignShareResult) Unwrap() *SignShareResult {
	return w.SignShareResult
}

func (w *Event_SignShareResult) Pb() threshcryptopb.Event_Type {
	return &threshcryptopb.Event_SignShareResult{SignShareResult: (w.SignShareResult).Pb()}
}

func (*Event_SignShareResult) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.Event_SignShareResult]()}
}

type Event_VerifyShare struct {
	VerifyShare *VerifyShare
}

func (*Event_VerifyShare) isEvent_Type() {}

func (w *Event_VerifyShare) Unwrap() *VerifyShare {
	return w.VerifyShare
}

func (w *Event_VerifyShare) Pb() threshcryptopb.Event_Type {
	return &threshcryptopb.Event_VerifyShare{VerifyShare: (w.VerifyShare).Pb()}
}

func (*Event_VerifyShare) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.Event_VerifyShare]()}
}

type Event_VerifyShareResult struct {
	VerifyShareResult *VerifyShareResult
}

func (*Event_VerifyShareResult) isEvent_Type() {}

func (w *Event_VerifyShareResult) Unwrap() *VerifyShareResult {
	return w.VerifyShareResult
}

func (w *Event_VerifyShareResult) Pb() threshcryptopb.Event_Type {
	return &threshcryptopb.Event_VerifyShareResult{VerifyShareResult: (w.VerifyShareResult).Pb()}
}

func (*Event_VerifyShareResult) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.Event_VerifyShareResult]()}
}

type Event_VerifyFull struct {
	VerifyFull *VerifyFull
}

func (*Event_VerifyFull) isEvent_Type() {}

func (w *Event_VerifyFull) Unwrap() *VerifyFull {
	return w.VerifyFull
}

func (w *Event_VerifyFull) Pb() threshcryptopb.Event_Type {
	return &threshcryptopb.Event_VerifyFull{VerifyFull: (w.VerifyFull).Pb()}
}

func (*Event_VerifyFull) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.Event_VerifyFull]()}
}

type Event_VerifyFullResult struct {
	VerifyFullResult *VerifyFullResult
}

func (*Event_VerifyFullResult) isEvent_Type() {}

func (w *Event_VerifyFullResult) Unwrap() *VerifyFullResult {
	return w.VerifyFullResult
}

func (w *Event_VerifyFullResult) Pb() threshcryptopb.Event_Type {
	return &threshcryptopb.Event_VerifyFullResult{VerifyFullResult: (w.VerifyFullResult).Pb()}
}

func (*Event_VerifyFullResult) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.Event_VerifyFullResult]()}
}

type Event_Recover struct {
	Recover *Recover
}

func (*Event_Recover) isEvent_Type() {}

func (w *Event_Recover) Unwrap() *Recover {
	return w.Recover
}

func (w *Event_Recover) Pb() threshcryptopb.Event_Type {
	return &threshcryptopb.Event_Recover{Recover: (w.Recover).Pb()}
}

func (*Event_Recover) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.Event_Recover]()}
}

type Event_RecoverResult struct {
	RecoverResult *RecoverResult
}

func (*Event_RecoverResult) isEvent_Type() {}

func (w *Event_RecoverResult) Unwrap() *RecoverResult {
	return w.RecoverResult
}

func (w *Event_RecoverResult) Pb() threshcryptopb.Event_Type {
	return &threshcryptopb.Event_RecoverResult{RecoverResult: (w.RecoverResult).Pb()}
}

func (*Event_RecoverResult) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.Event_RecoverResult]()}
}

func EventFromPb(pb *threshcryptopb.Event) *Event {
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *threshcryptopb.Event {
	return &threshcryptopb.Event{
		Type: (m.Type).Pb(),
	}
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.Event]()}
}

type SignShare struct {
	Data   [][]uint8
	Origin *SignShareOrigin
}

func SignShareFromPb(pb *threshcryptopb.SignShare) *SignShare {
	return &SignShare{
		Data:   pb.Data,
		Origin: SignShareOriginFromPb(pb.Origin),
	}
}

func (m *SignShare) Pb() *threshcryptopb.SignShare {
	return &threshcryptopb.SignShare{
		Data:   m.Data,
		Origin: (m.Origin).Pb(),
	}
}

func (*SignShare) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.SignShare]()}
}

type SignShareResult struct {
	SignatureShare []uint8
	Origin         *SignShareOrigin
}

func SignShareResultFromPb(pb *threshcryptopb.SignShareResult) *SignShareResult {
	return &SignShareResult{
		SignatureShare: pb.SignatureShare,
		Origin:         SignShareOriginFromPb(pb.Origin),
	}
}

func (m *SignShareResult) Pb() *threshcryptopb.SignShareResult {
	return &threshcryptopb.SignShareResult{
		SignatureShare: m.SignatureShare,
		Origin:         (m.Origin).Pb(),
	}
}

func (*SignShareResult) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.SignShareResult]()}
}

type SignShareOrigin struct {
	Module types.ModuleID
	Type   SignShareOrigin_Type
}

type SignShareOrigin_Type interface {
	mirreflect.GeneratedType
	isSignShareOrigin_Type()
	Pb() threshcryptopb.SignShareOrigin_Type
}

type SignShareOrigin_TypeWrapper[T any] interface {
	SignShareOrigin_Type
	Unwrap() *T
}

func SignShareOrigin_TypeFromPb(pb threshcryptopb.SignShareOrigin_Type) SignShareOrigin_Type {
	switch pb := pb.(type) {
	case *threshcryptopb.SignShareOrigin_ContextStore:
		return &SignShareOrigin_ContextStore{ContextStore: types1.OriginFromPb(pb.ContextStore)}
	case *threshcryptopb.SignShareOrigin_Dsl:
		return &SignShareOrigin_Dsl{Dsl: types2.OriginFromPb(pb.Dsl)}
	}
	return nil
}

type SignShareOrigin_ContextStore struct {
	ContextStore *types1.Origin
}

func (*SignShareOrigin_ContextStore) isSignShareOrigin_Type() {}

func (w *SignShareOrigin_ContextStore) Unwrap() *types1.Origin {
	return w.ContextStore
}

func (w *SignShareOrigin_ContextStore) Pb() threshcryptopb.SignShareOrigin_Type {
	return &threshcryptopb.SignShareOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*SignShareOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.SignShareOrigin_ContextStore]()}
}

type SignShareOrigin_Dsl struct {
	Dsl *types2.Origin
}

func (*SignShareOrigin_Dsl) isSignShareOrigin_Type() {}

func (w *SignShareOrigin_Dsl) Unwrap() *types2.Origin {
	return w.Dsl
}

func (w *SignShareOrigin_Dsl) Pb() threshcryptopb.SignShareOrigin_Type {
	return &threshcryptopb.SignShareOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*SignShareOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.SignShareOrigin_Dsl]()}
}

func SignShareOriginFromPb(pb *threshcryptopb.SignShareOrigin) *SignShareOrigin {
	return &SignShareOrigin{
		Module: (types.ModuleID)(pb.Module),
		Type:   SignShareOrigin_TypeFromPb(pb.Type),
	}
}

func (m *SignShareOrigin) Pb() *threshcryptopb.SignShareOrigin {
	return &threshcryptopb.SignShareOrigin{
		Module: (string)(m.Module),
		Type:   (m.Type).Pb(),
	}
}

func (*SignShareOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.SignShareOrigin]()}
}

type VerifyShare struct {
	Data           [][]uint8
	SignatureShare []uint8
	NodeId         types.NodeID
	Origin         *VerifyShareOrigin
}

func VerifyShareFromPb(pb *threshcryptopb.VerifyShare) *VerifyShare {
	return &VerifyShare{
		Data:           pb.Data,
		SignatureShare: pb.SignatureShare,
		NodeId:         (types.NodeID)(pb.NodeId),
		Origin:         VerifyShareOriginFromPb(pb.Origin),
	}
}

func (m *VerifyShare) Pb() *threshcryptopb.VerifyShare {
	return &threshcryptopb.VerifyShare{
		Data:           m.Data,
		SignatureShare: m.SignatureShare,
		NodeId:         (string)(m.NodeId),
		Origin:         (m.Origin).Pb(),
	}
}

func (*VerifyShare) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.VerifyShare]()}
}

type VerifyShareResult struct {
	Ok     bool
	Error  string
	Origin *VerifyShareOrigin
}

func VerifyShareResultFromPb(pb *threshcryptopb.VerifyShareResult) *VerifyShareResult {
	return &VerifyShareResult{
		Ok:     pb.Ok,
		Error:  pb.Error,
		Origin: VerifyShareOriginFromPb(pb.Origin),
	}
}

func (m *VerifyShareResult) Pb() *threshcryptopb.VerifyShareResult {
	return &threshcryptopb.VerifyShareResult{
		Ok:     m.Ok,
		Error:  m.Error,
		Origin: (m.Origin).Pb(),
	}
}

func (*VerifyShareResult) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.VerifyShareResult]()}
}

type VerifyShareOrigin struct {
	Module types.ModuleID
	Type   VerifyShareOrigin_Type
}

type VerifyShareOrigin_Type interface {
	mirreflect.GeneratedType
	isVerifyShareOrigin_Type()
	Pb() threshcryptopb.VerifyShareOrigin_Type
}

type VerifyShareOrigin_TypeWrapper[T any] interface {
	VerifyShareOrigin_Type
	Unwrap() *T
}

func VerifyShareOrigin_TypeFromPb(pb threshcryptopb.VerifyShareOrigin_Type) VerifyShareOrigin_Type {
	switch pb := pb.(type) {
	case *threshcryptopb.VerifyShareOrigin_ContextStore:
		return &VerifyShareOrigin_ContextStore{ContextStore: types1.OriginFromPb(pb.ContextStore)}
	case *threshcryptopb.VerifyShareOrigin_Dsl:
		return &VerifyShareOrigin_Dsl{Dsl: types2.OriginFromPb(pb.Dsl)}
	}
	return nil
}

type VerifyShareOrigin_ContextStore struct {
	ContextStore *types1.Origin
}

func (*VerifyShareOrigin_ContextStore) isVerifyShareOrigin_Type() {}

func (w *VerifyShareOrigin_ContextStore) Unwrap() *types1.Origin {
	return w.ContextStore
}

func (w *VerifyShareOrigin_ContextStore) Pb() threshcryptopb.VerifyShareOrigin_Type {
	return &threshcryptopb.VerifyShareOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*VerifyShareOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.VerifyShareOrigin_ContextStore]()}
}

type VerifyShareOrigin_Dsl struct {
	Dsl *types2.Origin
}

func (*VerifyShareOrigin_Dsl) isVerifyShareOrigin_Type() {}

func (w *VerifyShareOrigin_Dsl) Unwrap() *types2.Origin {
	return w.Dsl
}

func (w *VerifyShareOrigin_Dsl) Pb() threshcryptopb.VerifyShareOrigin_Type {
	return &threshcryptopb.VerifyShareOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*VerifyShareOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.VerifyShareOrigin_Dsl]()}
}

func VerifyShareOriginFromPb(pb *threshcryptopb.VerifyShareOrigin) *VerifyShareOrigin {
	return &VerifyShareOrigin{
		Module: (types.ModuleID)(pb.Module),
		Type:   VerifyShareOrigin_TypeFromPb(pb.Type),
	}
}

func (m *VerifyShareOrigin) Pb() *threshcryptopb.VerifyShareOrigin {
	return &threshcryptopb.VerifyShareOrigin{
		Module: (string)(m.Module),
		Type:   (m.Type).Pb(),
	}
}

func (*VerifyShareOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.VerifyShareOrigin]()}
}

type VerifyFull struct {
	Data          [][]uint8
	FullSignature []uint8
	Origin        *VerifyFullOrigin
}

func VerifyFullFromPb(pb *threshcryptopb.VerifyFull) *VerifyFull {
	return &VerifyFull{
		Data:          pb.Data,
		FullSignature: pb.FullSignature,
		Origin:        VerifyFullOriginFromPb(pb.Origin),
	}
}

func (m *VerifyFull) Pb() *threshcryptopb.VerifyFull {
	return &threshcryptopb.VerifyFull{
		Data:          m.Data,
		FullSignature: m.FullSignature,
		Origin:        (m.Origin).Pb(),
	}
}

func (*VerifyFull) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.VerifyFull]()}
}

type VerifyFullResult struct {
	Ok     bool
	Error  string
	Origin *VerifyFullOrigin
}

func VerifyFullResultFromPb(pb *threshcryptopb.VerifyFullResult) *VerifyFullResult {
	return &VerifyFullResult{
		Ok:     pb.Ok,
		Error:  pb.Error,
		Origin: VerifyFullOriginFromPb(pb.Origin),
	}
}

func (m *VerifyFullResult) Pb() *threshcryptopb.VerifyFullResult {
	return &threshcryptopb.VerifyFullResult{
		Ok:     m.Ok,
		Error:  m.Error,
		Origin: (m.Origin).Pb(),
	}
}

func (*VerifyFullResult) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.VerifyFullResult]()}
}

type VerifyFullOrigin struct {
	Module types.ModuleID
	Type   VerifyFullOrigin_Type
}

type VerifyFullOrigin_Type interface {
	mirreflect.GeneratedType
	isVerifyFullOrigin_Type()
	Pb() threshcryptopb.VerifyFullOrigin_Type
}

type VerifyFullOrigin_TypeWrapper[T any] interface {
	VerifyFullOrigin_Type
	Unwrap() *T
}

func VerifyFullOrigin_TypeFromPb(pb threshcryptopb.VerifyFullOrigin_Type) VerifyFullOrigin_Type {
	switch pb := pb.(type) {
	case *threshcryptopb.VerifyFullOrigin_ContextStore:
		return &VerifyFullOrigin_ContextStore{ContextStore: types1.OriginFromPb(pb.ContextStore)}
	case *threshcryptopb.VerifyFullOrigin_Dsl:
		return &VerifyFullOrigin_Dsl{Dsl: types2.OriginFromPb(pb.Dsl)}
	}
	return nil
}

type VerifyFullOrigin_ContextStore struct {
	ContextStore *types1.Origin
}

func (*VerifyFullOrigin_ContextStore) isVerifyFullOrigin_Type() {}

func (w *VerifyFullOrigin_ContextStore) Unwrap() *types1.Origin {
	return w.ContextStore
}

func (w *VerifyFullOrigin_ContextStore) Pb() threshcryptopb.VerifyFullOrigin_Type {
	return &threshcryptopb.VerifyFullOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*VerifyFullOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.VerifyFullOrigin_ContextStore]()}
}

type VerifyFullOrigin_Dsl struct {
	Dsl *types2.Origin
}

func (*VerifyFullOrigin_Dsl) isVerifyFullOrigin_Type() {}

func (w *VerifyFullOrigin_Dsl) Unwrap() *types2.Origin {
	return w.Dsl
}

func (w *VerifyFullOrigin_Dsl) Pb() threshcryptopb.VerifyFullOrigin_Type {
	return &threshcryptopb.VerifyFullOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*VerifyFullOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.VerifyFullOrigin_Dsl]()}
}

func VerifyFullOriginFromPb(pb *threshcryptopb.VerifyFullOrigin) *VerifyFullOrigin {
	return &VerifyFullOrigin{
		Module: (types.ModuleID)(pb.Module),
		Type:   VerifyFullOrigin_TypeFromPb(pb.Type),
	}
}

func (m *VerifyFullOrigin) Pb() *threshcryptopb.VerifyFullOrigin {
	return &threshcryptopb.VerifyFullOrigin{
		Module: (string)(m.Module),
		Type:   (m.Type).Pb(),
	}
}

func (*VerifyFullOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.VerifyFullOrigin]()}
}

type Recover struct {
	Data            [][]uint8
	SignatureShares [][]uint8
	Origin          *RecoverOrigin
}

func RecoverFromPb(pb *threshcryptopb.Recover) *Recover {
	return &Recover{
		Data:            pb.Data,
		SignatureShares: pb.SignatureShares,
		Origin:          RecoverOriginFromPb(pb.Origin),
	}
}

func (m *Recover) Pb() *threshcryptopb.Recover {
	return &threshcryptopb.Recover{
		Data:            m.Data,
		SignatureShares: m.SignatureShares,
		Origin:          (m.Origin).Pb(),
	}
}

func (*Recover) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.Recover]()}
}

type RecoverResult struct {
	FullSignature []uint8
	Ok            bool
	Error         string
	Origin        *RecoverOrigin
}

func RecoverResultFromPb(pb *threshcryptopb.RecoverResult) *RecoverResult {
	return &RecoverResult{
		FullSignature: pb.FullSignature,
		Ok:            pb.Ok,
		Error:         pb.Error,
		Origin:        RecoverOriginFromPb(pb.Origin),
	}
}

func (m *RecoverResult) Pb() *threshcryptopb.RecoverResult {
	return &threshcryptopb.RecoverResult{
		FullSignature: m.FullSignature,
		Ok:            m.Ok,
		Error:         m.Error,
		Origin:        (m.Origin).Pb(),
	}
}

func (*RecoverResult) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.RecoverResult]()}
}

type RecoverOrigin struct {
	Module types.ModuleID
	Type   RecoverOrigin_Type
}

type RecoverOrigin_Type interface {
	mirreflect.GeneratedType
	isRecoverOrigin_Type()
	Pb() threshcryptopb.RecoverOrigin_Type
}

type RecoverOrigin_TypeWrapper[T any] interface {
	RecoverOrigin_Type
	Unwrap() *T
}

func RecoverOrigin_TypeFromPb(pb threshcryptopb.RecoverOrigin_Type) RecoverOrigin_Type {
	switch pb := pb.(type) {
	case *threshcryptopb.RecoverOrigin_ContextStore:
		return &RecoverOrigin_ContextStore{ContextStore: types1.OriginFromPb(pb.ContextStore)}
	case *threshcryptopb.RecoverOrigin_Dsl:
		return &RecoverOrigin_Dsl{Dsl: types2.OriginFromPb(pb.Dsl)}
	}
	return nil
}

type RecoverOrigin_ContextStore struct {
	ContextStore *types1.Origin
}

func (*RecoverOrigin_ContextStore) isRecoverOrigin_Type() {}

func (w *RecoverOrigin_ContextStore) Unwrap() *types1.Origin {
	return w.ContextStore
}

func (w *RecoverOrigin_ContextStore) Pb() threshcryptopb.RecoverOrigin_Type {
	return &threshcryptopb.RecoverOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*RecoverOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.RecoverOrigin_ContextStore]()}
}

type RecoverOrigin_Dsl struct {
	Dsl *types2.Origin
}

func (*RecoverOrigin_Dsl) isRecoverOrigin_Type() {}

func (w *RecoverOrigin_Dsl) Unwrap() *types2.Origin {
	return w.Dsl
}

func (w *RecoverOrigin_Dsl) Pb() threshcryptopb.RecoverOrigin_Type {
	return &threshcryptopb.RecoverOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*RecoverOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.RecoverOrigin_Dsl]()}
}

func RecoverOriginFromPb(pb *threshcryptopb.RecoverOrigin) *RecoverOrigin {
	return &RecoverOrigin{
		Module: (types.ModuleID)(pb.Module),
		Type:   RecoverOrigin_TypeFromPb(pb.Type),
	}
}

func (m *RecoverOrigin) Pb() *threshcryptopb.RecoverOrigin {
	return &threshcryptopb.RecoverOrigin{
		Module: (string)(m.Module),
		Type:   (m.Type).Pb(),
	}
}

func (*RecoverOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*threshcryptopb.RecoverOrigin]()}
}
