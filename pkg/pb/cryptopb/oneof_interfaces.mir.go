package cryptopb

import (
	checkpointpb "github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	contextstorepb "github.com/filecoin-project/mir/pkg/pb/contextstorepb"
	dslpb "github.com/filecoin-project/mir/pkg/pb/dslpb"
	ordererpb "github.com/filecoin-project/mir/pkg/pb/ordererpb"
)

type Event_Type = isEvent_Type

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func (w *Event_SignRequest) Unwrap() *SignRequest {
	return w.SignRequest
}

func (w *Event_SignResult) Unwrap() *SignResult {
	return w.SignResult
}

func (w *Event_VerifySig) Unwrap() *VerifySig {
	return w.VerifySig
}

func (w *Event_SigVerified) Unwrap() *SigVerified {
	return w.SigVerified
}

func (w *Event_VerifySigs) Unwrap() *VerifySigs {
	return w.VerifySigs
}

func (w *Event_SigsVerified) Unwrap() *SigsVerified {
	return w.SigsVerified
}

type SignOrigin_Type = isSignOrigin_Type

type SignOrigin_TypeWrapper[T any] interface {
	SignOrigin_Type
	Unwrap() *T
}

func (w *SignOrigin_ContextStore) Unwrap() *contextstorepb.Origin {
	return w.ContextStore
}

func (w *SignOrigin_Dsl) Unwrap() *dslpb.Origin {
	return w.Dsl
}

func (w *SignOrigin_Checkpoint) Unwrap() *checkpointpb.SignOrigin {
	return w.Checkpoint
}

func (w *SignOrigin_Sb) Unwrap() *ordererpb.SignOrigin {
	return w.Sb
}

type SigVerOrigin_Type = isSigVerOrigin_Type

type SigVerOrigin_TypeWrapper[T any] interface {
	SigVerOrigin_Type
	Unwrap() *T
}

func (w *SigVerOrigin_ContextStore) Unwrap() *contextstorepb.Origin {
	return w.ContextStore
}

func (w *SigVerOrigin_Dsl) Unwrap() *dslpb.Origin {
	return w.Dsl
}

func (w *SigVerOrigin_Checkpoint) Unwrap() *checkpointpb.SigVerOrigin {
	return w.Checkpoint
}

func (w *SigVerOrigin_Sb) Unwrap() *ordererpb.SigVerOrigin {
	return w.Sb
}
