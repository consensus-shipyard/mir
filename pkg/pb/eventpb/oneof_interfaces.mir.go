package eventpb

import (
	availabilitypb "github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	batchdbpb "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb"
	batchfetcherpb "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
	bcbpb "github.com/filecoin-project/mir/pkg/pb/bcbpb"
	checkpointpb "github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	contextstorepb "github.com/filecoin-project/mir/pkg/pb/contextstorepb"
	dslpb "github.com/filecoin-project/mir/pkg/pb/dslpb"
	factorypb "github.com/filecoin-project/mir/pkg/pb/factorypb"
	hasherpb "github.com/filecoin-project/mir/pkg/pb/hasherpb"
	isspb "github.com/filecoin-project/mir/pkg/pb/isspb"
	mempoolpb "github.com/filecoin-project/mir/pkg/pb/mempoolpb"
	ordererspb "github.com/filecoin-project/mir/pkg/pb/ordererspb"
	pingpongpb "github.com/filecoin-project/mir/pkg/pb/pingpongpb"
	threshcryptopb "github.com/filecoin-project/mir/pkg/pb/threshcryptopb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

type Event_Type = isEvent_Type

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func (w *Event_Init) Unwrap() *Init {
	return w.Init
}

func (w *Event_Timer) Unwrap() *TimerEvent {
	return w.Timer
}

func (w *Event_Hasher) Unwrap() *hasherpb.Event {
	return w.Hasher
}

func (w *Event_Bcb) Unwrap() *bcbpb.Event {
	return w.Bcb
}

func (w *Event_Mempool) Unwrap() *mempoolpb.Event {
	return w.Mempool
}

func (w *Event_Availability) Unwrap() *availabilitypb.Event {
	return w.Availability
}

func (w *Event_BatchDb) Unwrap() *batchdbpb.Event {
	return w.BatchDb
}

func (w *Event_BatchFetcher) Unwrap() *batchfetcherpb.Event {
	return w.BatchFetcher
}

func (w *Event_ThreshCrypto) Unwrap() *threshcryptopb.Event {
	return w.ThreshCrypto
}

func (w *Event_PingPong) Unwrap() *pingpongpb.Event {
	return w.PingPong
}

func (w *Event_Checkpoint) Unwrap() *checkpointpb.Event {
	return w.Checkpoint
}

func (w *Event_Factory) Unwrap() *factorypb.Event {
	return w.Factory
}

func (w *Event_Iss) Unwrap() *isspb.Event {
	return w.Iss
}

func (w *Event_SbEvent) Unwrap() *ordererspb.SBInstanceEvent {
	return w.SbEvent
}

func (w *Event_NewRequests) Unwrap() *NewRequests {
	return w.NewRequests
}

func (w *Event_SignRequest) Unwrap() *SignRequest {
	return w.SignRequest
}

func (w *Event_SignResult) Unwrap() *SignResult {
	return w.SignResult
}

func (w *Event_VerifyNodeSigs) Unwrap() *VerifyNodeSigs {
	return w.VerifyNodeSigs
}

func (w *Event_NodeSigsVerified) Unwrap() *NodeSigsVerified {
	return w.NodeSigsVerified
}

func (w *Event_SendMessage) Unwrap() *SendMessage {
	return w.SendMessage
}

func (w *Event_MessageReceived) Unwrap() *MessageReceived {
	return w.MessageReceived
}

func (w *Event_DeliverCert) Unwrap() *DeliverCert {
	return w.DeliverCert
}

func (w *Event_VerifyRequestSig) Unwrap() *VerifyRequestSig {
	return w.VerifyRequestSig
}

func (w *Event_RequestSigVerified) Unwrap() *RequestSigVerified {
	return w.RequestSigVerified
}

func (w *Event_StoreVerifiedRequest) Unwrap() *StoreVerifiedRequest {
	return w.StoreVerifiedRequest
}

func (w *Event_AppSnapshotRequest) Unwrap() *AppSnapshotRequest {
	return w.AppSnapshotRequest
}

func (w *Event_AppSnapshot) Unwrap() *AppSnapshot {
	return w.AppSnapshot
}

func (w *Event_AppRestoreState) Unwrap() *AppRestoreState {
	return w.AppRestoreState
}

func (w *Event_NewEpoch) Unwrap() *NewEpoch {
	return w.NewEpoch
}

func (w *Event_NewConfig) Unwrap() *NewConfig {
	return w.NewConfig
}

func (w *Event_TestingString) Unwrap() *wrapperspb.StringValue {
	return w.TestingString
}

func (w *Event_TestingUint) Unwrap() *wrapperspb.UInt64Value {
	return w.TestingUint
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

func (w *SignOrigin_Sb) Unwrap() *ordererspb.SBInstanceSignOrigin {
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

func (w *SigVerOrigin_Sb) Unwrap() *ordererspb.SBInstanceSigVerOrigin {
	return w.Sb
}

type TimerEvent_Type = isTimerEvent_Type

type TimerEvent_TypeWrapper[T any] interface {
	TimerEvent_Type
	Unwrap() *T
}

func (w *TimerEvent_Delay) Unwrap() *TimerDelay {
	return w.Delay
}

func (w *TimerEvent_Repeat) Unwrap() *TimerRepeat {
	return w.Repeat
}

func (w *TimerEvent_GarbageCollect) Unwrap() *TimerGarbageCollect {
	return w.GarbageCollect
}
