package availabilitypb

import (
	contextstorepb "github.com/filecoin-project/mir/pkg/pb/contextstorepb"
	dslpb "github.com/filecoin-project/mir/pkg/pb/dslpb"
)

type Event_Type = isEvent_Type

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func (w *Event_RequestCert) Unwrap() *RequestCert {
	return w.RequestCert
}

func (w *Event_NewCert) Unwrap() *NewCert {
	return w.NewCert
}

func (w *Event_VerifyCert) Unwrap() *VerifyCert {
	return w.VerifyCert
}

func (w *Event_CertVerified) Unwrap() *CertVerified {
	return w.CertVerified
}

func (w *Event_RequestTransactions) Unwrap() *RequestTransactions {
	return w.RequestTransactions
}

func (w *Event_ProvideTransactions) Unwrap() *ProvideTransactions {
	return w.ProvideTransactions
}

type RequestCertOrigin_Type = isRequestCertOrigin_Type

type RequestCertOrigin_TypeWrapper[T any] interface {
	RequestCertOrigin_Type
	Unwrap() *T
}

func (w *RequestCertOrigin_ContextStore) Unwrap() *contextstorepb.Origin {
	return w.ContextStore
}

func (w *RequestCertOrigin_Dsl) Unwrap() *dslpb.Origin {
	return w.Dsl
}

type VerifyCertOrigin_Type = isVerifyCertOrigin_Type

type VerifyCertOrigin_TypeWrapper[T any] interface {
	VerifyCertOrigin_Type
	Unwrap() *T
}

func (w *VerifyCertOrigin_ContextStore) Unwrap() *contextstorepb.Origin {
	return w.ContextStore
}

func (w *VerifyCertOrigin_Dsl) Unwrap() *dslpb.Origin {
	return w.Dsl
}
