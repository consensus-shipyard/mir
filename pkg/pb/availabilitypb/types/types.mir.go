package availabilitypbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	availabilitypb "github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	types1 "github.com/filecoin-project/mir/pkg/pb/contextstorepb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/dslpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Event struct {
	Type Event_Type
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() availabilitypb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb availabilitypb.Event_Type) Event_Type {
	switch pb := pb.(type) {
	case *availabilitypb.Event_RequestCert:
		return &Event_RequestCert{RequestCert: pb.RequestCert}
	case *availabilitypb.Event_NewCert:
		return &Event_NewCert{NewCert: pb.NewCert}
	case *availabilitypb.Event_VerifyCert:
		return &Event_VerifyCert{VerifyCert: pb.VerifyCert}
	case *availabilitypb.Event_CertVerified:
		return &Event_CertVerified{CertVerified: CertVerifiedFromPb(pb.CertVerified)}
	case *availabilitypb.Event_RequestTransactions:
		return &Event_RequestTransactions{RequestTransactions: pb.RequestTransactions}
	case *availabilitypb.Event_ProvideTransactions:
		return &Event_ProvideTransactions{ProvideTransactions: pb.ProvideTransactions}
	}
	return nil
}

type Event_RequestCert struct {
	RequestCert *availabilitypb.RequestCert
}

func (*Event_RequestCert) isEvent_Type() {}

func (w *Event_RequestCert) Unwrap() *availabilitypb.RequestCert {
	return w.RequestCert
}

func (w *Event_RequestCert) Pb() availabilitypb.Event_Type {
	return &availabilitypb.Event_RequestCert{RequestCert: w.RequestCert}
}

func (*Event_RequestCert) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*availabilitypb.Event_RequestCert]()}
}

type Event_NewCert struct {
	NewCert *availabilitypb.NewCert
}

func (*Event_NewCert) isEvent_Type() {}

func (w *Event_NewCert) Unwrap() *availabilitypb.NewCert {
	return w.NewCert
}

func (w *Event_NewCert) Pb() availabilitypb.Event_Type {
	return &availabilitypb.Event_NewCert{NewCert: w.NewCert}
}

func (*Event_NewCert) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*availabilitypb.Event_NewCert]()}
}

type Event_VerifyCert struct {
	VerifyCert *availabilitypb.VerifyCert
}

func (*Event_VerifyCert) isEvent_Type() {}

func (w *Event_VerifyCert) Unwrap() *availabilitypb.VerifyCert {
	return w.VerifyCert
}

func (w *Event_VerifyCert) Pb() availabilitypb.Event_Type {
	return &availabilitypb.Event_VerifyCert{VerifyCert: w.VerifyCert}
}

func (*Event_VerifyCert) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*availabilitypb.Event_VerifyCert]()}
}

type Event_CertVerified struct {
	CertVerified *CertVerified
}

func (*Event_CertVerified) isEvent_Type() {}

func (w *Event_CertVerified) Unwrap() *CertVerified {
	return w.CertVerified
}

func (w *Event_CertVerified) Pb() availabilitypb.Event_Type {
	return &availabilitypb.Event_CertVerified{CertVerified: (w.CertVerified).Pb()}
}

func (*Event_CertVerified) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*availabilitypb.Event_CertVerified]()}
}

type Event_RequestTransactions struct {
	RequestTransactions *availabilitypb.RequestTransactions
}

func (*Event_RequestTransactions) isEvent_Type() {}

func (w *Event_RequestTransactions) Unwrap() *availabilitypb.RequestTransactions {
	return w.RequestTransactions
}

func (w *Event_RequestTransactions) Pb() availabilitypb.Event_Type {
	return &availabilitypb.Event_RequestTransactions{RequestTransactions: w.RequestTransactions}
}

func (*Event_RequestTransactions) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*availabilitypb.Event_RequestTransactions]()}
}

type Event_ProvideTransactions struct {
	ProvideTransactions *availabilitypb.ProvideTransactions
}

func (*Event_ProvideTransactions) isEvent_Type() {}

func (w *Event_ProvideTransactions) Unwrap() *availabilitypb.ProvideTransactions {
	return w.ProvideTransactions
}

func (w *Event_ProvideTransactions) Pb() availabilitypb.Event_Type {
	return &availabilitypb.Event_ProvideTransactions{ProvideTransactions: w.ProvideTransactions}
}

func (*Event_ProvideTransactions) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*availabilitypb.Event_ProvideTransactions]()}
}

func EventFromPb(pb *availabilitypb.Event) *Event {
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *availabilitypb.Event {
	return &availabilitypb.Event{
		Type: (m.Type).Pb(),
	}
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*availabilitypb.Event]()}
}

type CertVerified struct {
	Valid  bool
	Err    string
	Origin *VerifyCertOrigin
}

func CertVerifiedFromPb(pb *availabilitypb.CertVerified) *CertVerified {
	return &CertVerified{
		Valid:  pb.Valid,
		Err:    pb.Err,
		Origin: VerifyCertOriginFromPb(pb.Origin),
	}
}

func (m *CertVerified) Pb() *availabilitypb.CertVerified {
	return &availabilitypb.CertVerified{
		Valid:  m.Valid,
		Err:    m.Err,
		Origin: (m.Origin).Pb(),
	}
}

func (*CertVerified) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*availabilitypb.CertVerified]()}
}

type RequestCertOrigin struct {
	Module types.ModuleID
	Type   RequestCertOrigin_Type
}

type RequestCertOrigin_Type interface {
	mirreflect.GeneratedType
	isRequestCertOrigin_Type()
	Pb() availabilitypb.RequestCertOrigin_Type
}

type RequestCertOrigin_TypeWrapper[T any] interface {
	RequestCertOrigin_Type
	Unwrap() *T
}

func RequestCertOrigin_TypeFromPb(pb availabilitypb.RequestCertOrigin_Type) RequestCertOrigin_Type {
	switch pb := pb.(type) {
	case *availabilitypb.RequestCertOrigin_ContextStore:
		return &RequestCertOrigin_ContextStore{ContextStore: types1.OriginFromPb(pb.ContextStore)}
	case *availabilitypb.RequestCertOrigin_Dsl:
		return &RequestCertOrigin_Dsl{Dsl: types2.OriginFromPb(pb.Dsl)}
	}
	return nil
}

type RequestCertOrigin_ContextStore struct {
	ContextStore *types1.Origin
}

func (*RequestCertOrigin_ContextStore) isRequestCertOrigin_Type() {}

func (w *RequestCertOrigin_ContextStore) Unwrap() *types1.Origin {
	return w.ContextStore
}

func (w *RequestCertOrigin_ContextStore) Pb() availabilitypb.RequestCertOrigin_Type {
	return &availabilitypb.RequestCertOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*RequestCertOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*availabilitypb.RequestCertOrigin_ContextStore]()}
}

type RequestCertOrigin_Dsl struct {
	Dsl *types2.Origin
}

func (*RequestCertOrigin_Dsl) isRequestCertOrigin_Type() {}

func (w *RequestCertOrigin_Dsl) Unwrap() *types2.Origin {
	return w.Dsl
}

func (w *RequestCertOrigin_Dsl) Pb() availabilitypb.RequestCertOrigin_Type {
	return &availabilitypb.RequestCertOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*RequestCertOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*availabilitypb.RequestCertOrigin_Dsl]()}
}

func RequestCertOriginFromPb(pb *availabilitypb.RequestCertOrigin) *RequestCertOrigin {
	return &RequestCertOrigin{
		Module: (types.ModuleID)(pb.Module),
		Type:   RequestCertOrigin_TypeFromPb(pb.Type),
	}
}

func (m *RequestCertOrigin) Pb() *availabilitypb.RequestCertOrigin {
	return &availabilitypb.RequestCertOrigin{
		Module: (string)(m.Module),
		Type:   (m.Type).Pb(),
	}
}

func (*RequestCertOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*availabilitypb.RequestCertOrigin]()}
}

type VerifyCertOrigin struct {
	Module string
	Type   VerifyCertOrigin_Type
}

type VerifyCertOrigin_Type interface {
	mirreflect.GeneratedType
	isVerifyCertOrigin_Type()
	Pb() availabilitypb.VerifyCertOrigin_Type
}

type VerifyCertOrigin_TypeWrapper[T any] interface {
	VerifyCertOrigin_Type
	Unwrap() *T
}

func VerifyCertOrigin_TypeFromPb(pb availabilitypb.VerifyCertOrigin_Type) VerifyCertOrigin_Type {
	switch pb := pb.(type) {
	case *availabilitypb.VerifyCertOrigin_ContextStore:
		return &VerifyCertOrigin_ContextStore{ContextStore: types1.OriginFromPb(pb.ContextStore)}
	case *availabilitypb.VerifyCertOrigin_Dsl:
		return &VerifyCertOrigin_Dsl{Dsl: types2.OriginFromPb(pb.Dsl)}
	}
	return nil
}

type VerifyCertOrigin_ContextStore struct {
	ContextStore *types1.Origin
}

func (*VerifyCertOrigin_ContextStore) isVerifyCertOrigin_Type() {}

func (w *VerifyCertOrigin_ContextStore) Unwrap() *types1.Origin {
	return w.ContextStore
}

func (w *VerifyCertOrigin_ContextStore) Pb() availabilitypb.VerifyCertOrigin_Type {
	return &availabilitypb.VerifyCertOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*VerifyCertOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*availabilitypb.VerifyCertOrigin_ContextStore]()}
}

type VerifyCertOrigin_Dsl struct {
	Dsl *types2.Origin
}

func (*VerifyCertOrigin_Dsl) isVerifyCertOrigin_Type() {}

func (w *VerifyCertOrigin_Dsl) Unwrap() *types2.Origin {
	return w.Dsl
}

func (w *VerifyCertOrigin_Dsl) Pb() availabilitypb.VerifyCertOrigin_Type {
	return &availabilitypb.VerifyCertOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*VerifyCertOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*availabilitypb.VerifyCertOrigin_Dsl]()}
}

func VerifyCertOriginFromPb(pb *availabilitypb.VerifyCertOrigin) *VerifyCertOrigin {
	return &VerifyCertOrigin{
		Module: pb.Module,
		Type:   VerifyCertOrigin_TypeFromPb(pb.Type),
	}
}

func (m *VerifyCertOrigin) Pb() *availabilitypb.VerifyCertOrigin {
	return &availabilitypb.VerifyCertOrigin{
		Module: m.Module,
		Type:   (m.Type).Pb(),
	}
}

func (*VerifyCertOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*availabilitypb.VerifyCertOrigin]()}
}
