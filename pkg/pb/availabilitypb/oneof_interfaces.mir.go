package availabilitypb

import (
	mscpb "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb"
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

func (w *Event_ComputeCert) Unwrap() *ComputeCert {
	return w.ComputeCert
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

type RequestTransactionsOrigin_Type = isRequestTransactionsOrigin_Type

type RequestTransactionsOrigin_TypeWrapper[T any] interface {
	RequestTransactionsOrigin_Type
	Unwrap() *T
}

func (w *RequestTransactionsOrigin_ContextStore) Unwrap() *contextstorepb.Origin {
	return w.ContextStore
}

func (w *RequestTransactionsOrigin_Dsl) Unwrap() *dslpb.Origin {
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

type Cert_Type = isCert_Type

type Cert_TypeWrapper[T any] interface {
	Cert_Type
	Unwrap() *T
}

func (w *Cert_Mscs) Unwrap() *mscpb.Certs {
	return w.Mscs
}
