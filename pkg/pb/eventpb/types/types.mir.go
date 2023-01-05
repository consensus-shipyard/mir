package eventpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types7 "github.com/filecoin-project/mir/codegen/model/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types5 "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/bcbpb/types"
	checkpointpb "github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	types8 "github.com/filecoin-project/mir/pkg/pb/contextstorepb/types"
	types9 "github.com/filecoin-project/mir/pkg/pb/dslpb/types"
	eventpb "github.com/filecoin-project/mir/pkg/pb/eventpb"
	factorymodulepb "github.com/filecoin-project/mir/pkg/pb/factorymodulepb"
	isspb "github.com/filecoin-project/mir/pkg/pb/isspb"
	types2 "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	types10 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	ordererspb "github.com/filecoin-project/mir/pkg/pb/ordererspb"
	pingpongpb "github.com/filecoin-project/mir/pkg/pb/pingpongpb"
	types6 "github.com/filecoin-project/mir/pkg/pb/threshcryptopb/types"
	types "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

type Event struct {
	Type       Event_Type
	Next       []*Event
	DestModule types.ModuleID
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() eventpb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb eventpb.Event_Type) Event_Type {
	switch pb := pb.(type) {
	case *eventpb.Event_Init:
		return &Event_Init{Init: pb.Init}
	case *eventpb.Event_Tick:
		return &Event_Tick{Tick: pb.Tick}
	case *eventpb.Event_WalAppend:
		return &Event_WalAppend{WalAppend: pb.WalAppend}
	case *eventpb.Event_WalEntry:
		return &Event_WalEntry{WalEntry: pb.WalEntry}
	case *eventpb.Event_WalTruncate:
		return &Event_WalTruncate{WalTruncate: pb.WalTruncate}
	case *eventpb.Event_NewRequests:
		return &Event_NewRequests{NewRequests: pb.NewRequests}
	case *eventpb.Event_HashRequest:
		return &Event_HashRequest{HashRequest: pb.HashRequest}
	case *eventpb.Event_HashResult:
		return &Event_HashResult{HashResult: pb.HashResult}
	case *eventpb.Event_SignRequest:
		return &Event_SignRequest{SignRequest: SignRequestFromPb(pb.SignRequest)}
	case *eventpb.Event_SignResult:
		return &Event_SignResult{SignResult: SignResultFromPb(pb.SignResult)}
	case *eventpb.Event_VerifyNodeSigs:
		return &Event_VerifyNodeSigs{VerifyNodeSigs: VerifyNodeSigsFromPb(pb.VerifyNodeSigs)}
	case *eventpb.Event_NodeSigsVerified:
		return &Event_NodeSigsVerified{NodeSigsVerified: NodeSigsVerifiedFromPb(pb.NodeSigsVerified)}
	case *eventpb.Event_RequestReady:
		return &Event_RequestReady{RequestReady: pb.RequestReady}
	case *eventpb.Event_SendMessage:
		return &Event_SendMessage{SendMessage: SendMessageFromPb(pb.SendMessage)}
	case *eventpb.Event_MessageReceived:
		return &Event_MessageReceived{MessageReceived: MessageReceivedFromPb(pb.MessageReceived)}
	case *eventpb.Event_DeliverCert:
		return &Event_DeliverCert{DeliverCert: pb.DeliverCert}
	case *eventpb.Event_Iss:
		return &Event_Iss{Iss: pb.Iss}
	case *eventpb.Event_VerifyRequestSig:
		return &Event_VerifyRequestSig{VerifyRequestSig: pb.VerifyRequestSig}
	case *eventpb.Event_RequestSigVerified:
		return &Event_RequestSigVerified{RequestSigVerified: pb.RequestSigVerified}
	case *eventpb.Event_StoreVerifiedRequest:
		return &Event_StoreVerifiedRequest{StoreVerifiedRequest: pb.StoreVerifiedRequest}
	case *eventpb.Event_AppSnapshotRequest:
		return &Event_AppSnapshotRequest{AppSnapshotRequest: pb.AppSnapshotRequest}
	case *eventpb.Event_AppSnapshot:
		return &Event_AppSnapshot{AppSnapshot: pb.AppSnapshot}
	case *eventpb.Event_AppRestoreState:
		return &Event_AppRestoreState{AppRestoreState: pb.AppRestoreState}
	case *eventpb.Event_TimerDelay:
		return &Event_TimerDelay{TimerDelay: pb.TimerDelay}
	case *eventpb.Event_TimerRepeat:
		return &Event_TimerRepeat{TimerRepeat: pb.TimerRepeat}
	case *eventpb.Event_TimerGarbageCollect:
		return &Event_TimerGarbageCollect{TimerGarbageCollect: pb.TimerGarbageCollect}
	case *eventpb.Event_Bcb:
		return &Event_Bcb{Bcb: types1.EventFromPb(pb.Bcb)}
	case *eventpb.Event_Mempool:
		return &Event_Mempool{Mempool: types2.EventFromPb(pb.Mempool)}
	case *eventpb.Event_Availability:
		return &Event_Availability{Availability: types3.EventFromPb(pb.Availability)}
	case *eventpb.Event_NewEpoch:
		return &Event_NewEpoch{NewEpoch: pb.NewEpoch}
	case *eventpb.Event_NewConfig:
		return &Event_NewConfig{NewConfig: pb.NewConfig}
	case *eventpb.Event_Factory:
		return &Event_Factory{Factory: pb.Factory}
	case *eventpb.Event_BatchDb:
		return &Event_BatchDb{BatchDb: types4.EventFromPb(pb.BatchDb)}
	case *eventpb.Event_BatchFetcher:
		return &Event_BatchFetcher{BatchFetcher: types5.EventFromPb(pb.BatchFetcher)}
	case *eventpb.Event_ThreshCrypto:
		return &Event_ThreshCrypto{ThreshCrypto: types6.EventFromPb(pb.ThreshCrypto)}
	case *eventpb.Event_PingPong:
		return &Event_PingPong{PingPong: pb.PingPong}
	case *eventpb.Event_Checkpoint:
		return &Event_Checkpoint{Checkpoint: pb.Checkpoint}
	case *eventpb.Event_SbEvent:
		return &Event_SbEvent{SbEvent: pb.SbEvent}
	case *eventpb.Event_TestingString:
		return &Event_TestingString{TestingString: pb.TestingString}
	case *eventpb.Event_TestingUint:
		return &Event_TestingUint{TestingUint: pb.TestingUint}
	}
	return nil
}

type Event_Init struct {
	Init *eventpb.Init
}

func (*Event_Init) isEvent_Type() {}

func (w *Event_Init) Unwrap() *eventpb.Init {
	return w.Init
}

func (w *Event_Init) Pb() eventpb.Event_Type {
	return &eventpb.Event_Init{Init: w.Init}
}

func (*Event_Init) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Init]()}
}

type Event_Tick struct {
	Tick *eventpb.Tick
}

func (*Event_Tick) isEvent_Type() {}

func (w *Event_Tick) Unwrap() *eventpb.Tick {
	return w.Tick
}

func (w *Event_Tick) Pb() eventpb.Event_Type {
	return &eventpb.Event_Tick{Tick: w.Tick}
}

func (*Event_Tick) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Tick]()}
}

type Event_WalAppend struct {
	WalAppend *eventpb.WALAppend
}

func (*Event_WalAppend) isEvent_Type() {}

func (w *Event_WalAppend) Unwrap() *eventpb.WALAppend {
	return w.WalAppend
}

func (w *Event_WalAppend) Pb() eventpb.Event_Type {
	return &eventpb.Event_WalAppend{WalAppend: w.WalAppend}
}

func (*Event_WalAppend) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_WalAppend]()}
}

type Event_WalEntry struct {
	WalEntry *eventpb.WALEntry
}

func (*Event_WalEntry) isEvent_Type() {}

func (w *Event_WalEntry) Unwrap() *eventpb.WALEntry {
	return w.WalEntry
}

func (w *Event_WalEntry) Pb() eventpb.Event_Type {
	return &eventpb.Event_WalEntry{WalEntry: w.WalEntry}
}

func (*Event_WalEntry) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_WalEntry]()}
}

type Event_WalTruncate struct {
	WalTruncate *eventpb.WALTruncate
}

func (*Event_WalTruncate) isEvent_Type() {}

func (w *Event_WalTruncate) Unwrap() *eventpb.WALTruncate {
	return w.WalTruncate
}

func (w *Event_WalTruncate) Pb() eventpb.Event_Type {
	return &eventpb.Event_WalTruncate{WalTruncate: w.WalTruncate}
}

func (*Event_WalTruncate) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_WalTruncate]()}
}

type Event_NewRequests struct {
	NewRequests *eventpb.NewRequests
}

func (*Event_NewRequests) isEvent_Type() {}

func (w *Event_NewRequests) Unwrap() *eventpb.NewRequests {
	return w.NewRequests
}

func (w *Event_NewRequests) Pb() eventpb.Event_Type {
	return &eventpb.Event_NewRequests{NewRequests: w.NewRequests}
}

func (*Event_NewRequests) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_NewRequests]()}
}

type Event_HashRequest struct {
	HashRequest *eventpb.HashRequest
}

func (*Event_HashRequest) isEvent_Type() {}

func (w *Event_HashRequest) Unwrap() *eventpb.HashRequest {
	return w.HashRequest
}

func (w *Event_HashRequest) Pb() eventpb.Event_Type {
	return &eventpb.Event_HashRequest{HashRequest: w.HashRequest}
}

func (*Event_HashRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_HashRequest]()}
}

type Event_HashResult struct {
	HashResult *eventpb.HashResult
}

func (*Event_HashResult) isEvent_Type() {}

func (w *Event_HashResult) Unwrap() *eventpb.HashResult {
	return w.HashResult
}

func (w *Event_HashResult) Pb() eventpb.Event_Type {
	return &eventpb.Event_HashResult{HashResult: w.HashResult}
}

func (*Event_HashResult) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_HashResult]()}
}

type Event_SignRequest struct {
	SignRequest *SignRequest
}

func (*Event_SignRequest) isEvent_Type() {}

func (w *Event_SignRequest) Unwrap() *SignRequest {
	return w.SignRequest
}

func (w *Event_SignRequest) Pb() eventpb.Event_Type {
	return &eventpb.Event_SignRequest{SignRequest: (w.SignRequest).Pb()}
}

func (*Event_SignRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_SignRequest]()}
}

type Event_SignResult struct {
	SignResult *SignResult
}

func (*Event_SignResult) isEvent_Type() {}

func (w *Event_SignResult) Unwrap() *SignResult {
	return w.SignResult
}

func (w *Event_SignResult) Pb() eventpb.Event_Type {
	return &eventpb.Event_SignResult{SignResult: (w.SignResult).Pb()}
}

func (*Event_SignResult) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_SignResult]()}
}

type Event_VerifyNodeSigs struct {
	VerifyNodeSigs *VerifyNodeSigs
}

func (*Event_VerifyNodeSigs) isEvent_Type() {}

func (w *Event_VerifyNodeSigs) Unwrap() *VerifyNodeSigs {
	return w.VerifyNodeSigs
}

func (w *Event_VerifyNodeSigs) Pb() eventpb.Event_Type {
	return &eventpb.Event_VerifyNodeSigs{VerifyNodeSigs: (w.VerifyNodeSigs).Pb()}
}

func (*Event_VerifyNodeSigs) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_VerifyNodeSigs]()}
}

type Event_NodeSigsVerified struct {
	NodeSigsVerified *NodeSigsVerified
}

func (*Event_NodeSigsVerified) isEvent_Type() {}

func (w *Event_NodeSigsVerified) Unwrap() *NodeSigsVerified {
	return w.NodeSigsVerified
}

func (w *Event_NodeSigsVerified) Pb() eventpb.Event_Type {
	return &eventpb.Event_NodeSigsVerified{NodeSigsVerified: (w.NodeSigsVerified).Pb()}
}

func (*Event_NodeSigsVerified) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_NodeSigsVerified]()}
}

type Event_RequestReady struct {
	RequestReady *eventpb.RequestReady
}

func (*Event_RequestReady) isEvent_Type() {}

func (w *Event_RequestReady) Unwrap() *eventpb.RequestReady {
	return w.RequestReady
}

func (w *Event_RequestReady) Pb() eventpb.Event_Type {
	return &eventpb.Event_RequestReady{RequestReady: w.RequestReady}
}

func (*Event_RequestReady) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_RequestReady]()}
}

type Event_SendMessage struct {
	SendMessage *SendMessage
}

func (*Event_SendMessage) isEvent_Type() {}

func (w *Event_SendMessage) Unwrap() *SendMessage {
	return w.SendMessage
}

func (w *Event_SendMessage) Pb() eventpb.Event_Type {
	return &eventpb.Event_SendMessage{SendMessage: (w.SendMessage).Pb()}
}

func (*Event_SendMessage) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_SendMessage]()}
}

type Event_MessageReceived struct {
	MessageReceived *MessageReceived
}

func (*Event_MessageReceived) isEvent_Type() {}

func (w *Event_MessageReceived) Unwrap() *MessageReceived {
	return w.MessageReceived
}

func (w *Event_MessageReceived) Pb() eventpb.Event_Type {
	return &eventpb.Event_MessageReceived{MessageReceived: (w.MessageReceived).Pb()}
}

func (*Event_MessageReceived) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_MessageReceived]()}
}

type Event_DeliverCert struct {
	DeliverCert *eventpb.DeliverCert
}

func (*Event_DeliverCert) isEvent_Type() {}

func (w *Event_DeliverCert) Unwrap() *eventpb.DeliverCert {
	return w.DeliverCert
}

func (w *Event_DeliverCert) Pb() eventpb.Event_Type {
	return &eventpb.Event_DeliverCert{DeliverCert: w.DeliverCert}
}

func (*Event_DeliverCert) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_DeliverCert]()}
}

type Event_Iss struct {
	Iss *isspb.ISSEvent
}

func (*Event_Iss) isEvent_Type() {}

func (w *Event_Iss) Unwrap() *isspb.ISSEvent {
	return w.Iss
}

func (w *Event_Iss) Pb() eventpb.Event_Type {
	return &eventpb.Event_Iss{Iss: w.Iss}
}

func (*Event_Iss) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Iss]()}
}

type Event_VerifyRequestSig struct {
	VerifyRequestSig *eventpb.VerifyRequestSig
}

func (*Event_VerifyRequestSig) isEvent_Type() {}

func (w *Event_VerifyRequestSig) Unwrap() *eventpb.VerifyRequestSig {
	return w.VerifyRequestSig
}

func (w *Event_VerifyRequestSig) Pb() eventpb.Event_Type {
	return &eventpb.Event_VerifyRequestSig{VerifyRequestSig: w.VerifyRequestSig}
}

func (*Event_VerifyRequestSig) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_VerifyRequestSig]()}
}

type Event_RequestSigVerified struct {
	RequestSigVerified *eventpb.RequestSigVerified
}

func (*Event_RequestSigVerified) isEvent_Type() {}

func (w *Event_RequestSigVerified) Unwrap() *eventpb.RequestSigVerified {
	return w.RequestSigVerified
}

func (w *Event_RequestSigVerified) Pb() eventpb.Event_Type {
	return &eventpb.Event_RequestSigVerified{RequestSigVerified: w.RequestSigVerified}
}

func (*Event_RequestSigVerified) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_RequestSigVerified]()}
}

type Event_StoreVerifiedRequest struct {
	StoreVerifiedRequest *eventpb.StoreVerifiedRequest
}

func (*Event_StoreVerifiedRequest) isEvent_Type() {}

func (w *Event_StoreVerifiedRequest) Unwrap() *eventpb.StoreVerifiedRequest {
	return w.StoreVerifiedRequest
}

func (w *Event_StoreVerifiedRequest) Pb() eventpb.Event_Type {
	return &eventpb.Event_StoreVerifiedRequest{StoreVerifiedRequest: w.StoreVerifiedRequest}
}

func (*Event_StoreVerifiedRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_StoreVerifiedRequest]()}
}

type Event_AppSnapshotRequest struct {
	AppSnapshotRequest *eventpb.AppSnapshotRequest
}

func (*Event_AppSnapshotRequest) isEvent_Type() {}

func (w *Event_AppSnapshotRequest) Unwrap() *eventpb.AppSnapshotRequest {
	return w.AppSnapshotRequest
}

func (w *Event_AppSnapshotRequest) Pb() eventpb.Event_Type {
	return &eventpb.Event_AppSnapshotRequest{AppSnapshotRequest: w.AppSnapshotRequest}
}

func (*Event_AppSnapshotRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_AppSnapshotRequest]()}
}

type Event_AppSnapshot struct {
	AppSnapshot *eventpb.AppSnapshot
}

func (*Event_AppSnapshot) isEvent_Type() {}

func (w *Event_AppSnapshot) Unwrap() *eventpb.AppSnapshot {
	return w.AppSnapshot
}

func (w *Event_AppSnapshot) Pb() eventpb.Event_Type {
	return &eventpb.Event_AppSnapshot{AppSnapshot: w.AppSnapshot}
}

func (*Event_AppSnapshot) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_AppSnapshot]()}
}

type Event_AppRestoreState struct {
	AppRestoreState *eventpb.AppRestoreState
}

func (*Event_AppRestoreState) isEvent_Type() {}

func (w *Event_AppRestoreState) Unwrap() *eventpb.AppRestoreState {
	return w.AppRestoreState
}

func (w *Event_AppRestoreState) Pb() eventpb.Event_Type {
	return &eventpb.Event_AppRestoreState{AppRestoreState: w.AppRestoreState}
}

func (*Event_AppRestoreState) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_AppRestoreState]()}
}

type Event_TimerDelay struct {
	TimerDelay *eventpb.TimerDelay
}

func (*Event_TimerDelay) isEvent_Type() {}

func (w *Event_TimerDelay) Unwrap() *eventpb.TimerDelay {
	return w.TimerDelay
}

func (w *Event_TimerDelay) Pb() eventpb.Event_Type {
	return &eventpb.Event_TimerDelay{TimerDelay: w.TimerDelay}
}

func (*Event_TimerDelay) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_TimerDelay]()}
}

type Event_TimerRepeat struct {
	TimerRepeat *eventpb.TimerRepeat
}

func (*Event_TimerRepeat) isEvent_Type() {}

func (w *Event_TimerRepeat) Unwrap() *eventpb.TimerRepeat {
	return w.TimerRepeat
}

func (w *Event_TimerRepeat) Pb() eventpb.Event_Type {
	return &eventpb.Event_TimerRepeat{TimerRepeat: w.TimerRepeat}
}

func (*Event_TimerRepeat) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_TimerRepeat]()}
}

type Event_TimerGarbageCollect struct {
	TimerGarbageCollect *eventpb.TimerGarbageCollect
}

func (*Event_TimerGarbageCollect) isEvent_Type() {}

func (w *Event_TimerGarbageCollect) Unwrap() *eventpb.TimerGarbageCollect {
	return w.TimerGarbageCollect
}

func (w *Event_TimerGarbageCollect) Pb() eventpb.Event_Type {
	return &eventpb.Event_TimerGarbageCollect{TimerGarbageCollect: w.TimerGarbageCollect}
}

func (*Event_TimerGarbageCollect) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_TimerGarbageCollect]()}
}

type Event_Bcb struct {
	Bcb *types1.Event
}

func (*Event_Bcb) isEvent_Type() {}

func (w *Event_Bcb) Unwrap() *types1.Event {
	return w.Bcb
}

func (w *Event_Bcb) Pb() eventpb.Event_Type {
	return &eventpb.Event_Bcb{Bcb: (w.Bcb).Pb()}
}

func (*Event_Bcb) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Bcb]()}
}

type Event_Mempool struct {
	Mempool *types2.Event
}

func (*Event_Mempool) isEvent_Type() {}

func (w *Event_Mempool) Unwrap() *types2.Event {
	return w.Mempool
}

func (w *Event_Mempool) Pb() eventpb.Event_Type {
	return &eventpb.Event_Mempool{Mempool: (w.Mempool).Pb()}
}

func (*Event_Mempool) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Mempool]()}
}

type Event_Availability struct {
	Availability *types3.Event
}

func (*Event_Availability) isEvent_Type() {}

func (w *Event_Availability) Unwrap() *types3.Event {
	return w.Availability
}

func (w *Event_Availability) Pb() eventpb.Event_Type {
	return &eventpb.Event_Availability{Availability: (w.Availability).Pb()}
}

func (*Event_Availability) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Availability]()}
}

type Event_NewEpoch struct {
	NewEpoch *eventpb.NewEpoch
}

func (*Event_NewEpoch) isEvent_Type() {}

func (w *Event_NewEpoch) Unwrap() *eventpb.NewEpoch {
	return w.NewEpoch
}

func (w *Event_NewEpoch) Pb() eventpb.Event_Type {
	return &eventpb.Event_NewEpoch{NewEpoch: w.NewEpoch}
}

func (*Event_NewEpoch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_NewEpoch]()}
}

type Event_NewConfig struct {
	NewConfig *eventpb.NewConfig
}

func (*Event_NewConfig) isEvent_Type() {}

func (w *Event_NewConfig) Unwrap() *eventpb.NewConfig {
	return w.NewConfig
}

func (w *Event_NewConfig) Pb() eventpb.Event_Type {
	return &eventpb.Event_NewConfig{NewConfig: w.NewConfig}
}

func (*Event_NewConfig) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_NewConfig]()}
}

type Event_Factory struct {
	Factory *factorymodulepb.Factory
}

func (*Event_Factory) isEvent_Type() {}

func (w *Event_Factory) Unwrap() *factorymodulepb.Factory {
	return w.Factory
}

func (w *Event_Factory) Pb() eventpb.Event_Type {
	return &eventpb.Event_Factory{Factory: w.Factory}
}

func (*Event_Factory) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Factory]()}
}

type Event_BatchDb struct {
	BatchDb *types4.Event
}

func (*Event_BatchDb) isEvent_Type() {}

func (w *Event_BatchDb) Unwrap() *types4.Event {
	return w.BatchDb
}

func (w *Event_BatchDb) Pb() eventpb.Event_Type {
	return &eventpb.Event_BatchDb{BatchDb: (w.BatchDb).Pb()}
}

func (*Event_BatchDb) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_BatchDb]()}
}

type Event_BatchFetcher struct {
	BatchFetcher *types5.Event
}

func (*Event_BatchFetcher) isEvent_Type() {}

func (w *Event_BatchFetcher) Unwrap() *types5.Event {
	return w.BatchFetcher
}

func (w *Event_BatchFetcher) Pb() eventpb.Event_Type {
	return &eventpb.Event_BatchFetcher{BatchFetcher: (w.BatchFetcher).Pb()}
}

func (*Event_BatchFetcher) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_BatchFetcher]()}
}

type Event_ThreshCrypto struct {
	ThreshCrypto *types6.Event
}

func (*Event_ThreshCrypto) isEvent_Type() {}

func (w *Event_ThreshCrypto) Unwrap() *types6.Event {
	return w.ThreshCrypto
}

func (w *Event_ThreshCrypto) Pb() eventpb.Event_Type {
	return &eventpb.Event_ThreshCrypto{ThreshCrypto: (w.ThreshCrypto).Pb()}
}

func (*Event_ThreshCrypto) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_ThreshCrypto]()}
}

type Event_PingPong struct {
	PingPong *pingpongpb.Event
}

func (*Event_PingPong) isEvent_Type() {}

func (w *Event_PingPong) Unwrap() *pingpongpb.Event {
	return w.PingPong
}

func (w *Event_PingPong) Pb() eventpb.Event_Type {
	return &eventpb.Event_PingPong{PingPong: w.PingPong}
}

func (*Event_PingPong) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_PingPong]()}
}

type Event_Checkpoint struct {
	Checkpoint *checkpointpb.Event
}

func (*Event_Checkpoint) isEvent_Type() {}

func (w *Event_Checkpoint) Unwrap() *checkpointpb.Event {
	return w.Checkpoint
}

func (w *Event_Checkpoint) Pb() eventpb.Event_Type {
	return &eventpb.Event_Checkpoint{Checkpoint: w.Checkpoint}
}

func (*Event_Checkpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Checkpoint]()}
}

type Event_SbEvent struct {
	SbEvent *ordererspb.SBInstanceEvent
}

func (*Event_SbEvent) isEvent_Type() {}

func (w *Event_SbEvent) Unwrap() *ordererspb.SBInstanceEvent {
	return w.SbEvent
}

func (w *Event_SbEvent) Pb() eventpb.Event_Type {
	return &eventpb.Event_SbEvent{SbEvent: w.SbEvent}
}

func (*Event_SbEvent) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_SbEvent]()}
}

type Event_TestingString struct {
	TestingString *wrapperspb.StringValue
}

func (*Event_TestingString) isEvent_Type() {}

func (w *Event_TestingString) Unwrap() *wrapperspb.StringValue {
	return w.TestingString
}

func (w *Event_TestingString) Pb() eventpb.Event_Type {
	return &eventpb.Event_TestingString{TestingString: w.TestingString}
}

func (*Event_TestingString) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_TestingString]()}
}

type Event_TestingUint struct {
	TestingUint *wrapperspb.UInt64Value
}

func (*Event_TestingUint) isEvent_Type() {}

func (w *Event_TestingUint) Unwrap() *wrapperspb.UInt64Value {
	return w.TestingUint
}

func (w *Event_TestingUint) Pb() eventpb.Event_Type {
	return &eventpb.Event_TestingUint{TestingUint: w.TestingUint}
}

func (*Event_TestingUint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_TestingUint]()}
}

func EventFromPb(pb *eventpb.Event) *Event {
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
		Next: types7.ConvertSlice(pb.Next, func(t *eventpb.Event) *Event {
			return EventFromPb(t)
		}),
		DestModule: (types.ModuleID)(pb.DestModule),
	}
}

func (m *Event) Pb() *eventpb.Event {
	return &eventpb.Event{
		Type: (m.Type).Pb(),
		Next: types7.ConvertSlice(m.Next, func(t *Event) *eventpb.Event {
			return (t).Pb()
		}),
		DestModule: (string)(m.DestModule),
	}
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event]()}
}

type SignRequest struct {
	Data   [][]uint8
	Origin *SignOrigin
}

func SignRequestFromPb(pb *eventpb.SignRequest) *SignRequest {
	return &SignRequest{
		Data:   pb.Data,
		Origin: SignOriginFromPb(pb.Origin),
	}
}

func (m *SignRequest) Pb() *eventpb.SignRequest {
	return &eventpb.SignRequest{
		Data:   m.Data,
		Origin: (m.Origin).Pb(),
	}
}

func (*SignRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SignRequest]()}
}

type SignResult struct {
	Signature []uint8
	Origin    *SignOrigin
}

func SignResultFromPb(pb *eventpb.SignResult) *SignResult {
	return &SignResult{
		Signature: pb.Signature,
		Origin:    SignOriginFromPb(pb.Origin),
	}
}

func (m *SignResult) Pb() *eventpb.SignResult {
	return &eventpb.SignResult{
		Signature: m.Signature,
		Origin:    (m.Origin).Pb(),
	}
}

func (*SignResult) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SignResult]()}
}

type SignOrigin struct {
	Module types.ModuleID
	Type   SignOrigin_Type
}

type SignOrigin_Type interface {
	mirreflect.GeneratedType
	isSignOrigin_Type()
	Pb() eventpb.SignOrigin_Type
}

type SignOrigin_TypeWrapper[T any] interface {
	SignOrigin_Type
	Unwrap() *T
}

func SignOrigin_TypeFromPb(pb eventpb.SignOrigin_Type) SignOrigin_Type {
	switch pb := pb.(type) {
	case *eventpb.SignOrigin_ContextStore:
		return &SignOrigin_ContextStore{ContextStore: types8.OriginFromPb(pb.ContextStore)}
	case *eventpb.SignOrigin_Dsl:
		return &SignOrigin_Dsl{Dsl: types9.OriginFromPb(pb.Dsl)}
	case *eventpb.SignOrigin_Checkpoint:
		return &SignOrigin_Checkpoint{Checkpoint: pb.Checkpoint}
	case *eventpb.SignOrigin_Sb:
		return &SignOrigin_Sb{Sb: pb.Sb}
	}
	return nil
}

type SignOrigin_ContextStore struct {
	ContextStore *types8.Origin
}

func (*SignOrigin_ContextStore) isSignOrigin_Type() {}

func (w *SignOrigin_ContextStore) Unwrap() *types8.Origin {
	return w.ContextStore
}

func (w *SignOrigin_ContextStore) Pb() eventpb.SignOrigin_Type {
	return &eventpb.SignOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*SignOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SignOrigin_ContextStore]()}
}

type SignOrigin_Dsl struct {
	Dsl *types9.Origin
}

func (*SignOrigin_Dsl) isSignOrigin_Type() {}

func (w *SignOrigin_Dsl) Unwrap() *types9.Origin {
	return w.Dsl
}

func (w *SignOrigin_Dsl) Pb() eventpb.SignOrigin_Type {
	return &eventpb.SignOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*SignOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SignOrigin_Dsl]()}
}

type SignOrigin_Checkpoint struct {
	Checkpoint *checkpointpb.SignOrigin
}

func (*SignOrigin_Checkpoint) isSignOrigin_Type() {}

func (w *SignOrigin_Checkpoint) Unwrap() *checkpointpb.SignOrigin {
	return w.Checkpoint
}

func (w *SignOrigin_Checkpoint) Pb() eventpb.SignOrigin_Type {
	return &eventpb.SignOrigin_Checkpoint{Checkpoint: w.Checkpoint}
}

func (*SignOrigin_Checkpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SignOrigin_Checkpoint]()}
}

type SignOrigin_Sb struct {
	Sb *ordererspb.SBInstanceSignOrigin
}

func (*SignOrigin_Sb) isSignOrigin_Type() {}

func (w *SignOrigin_Sb) Unwrap() *ordererspb.SBInstanceSignOrigin {
	return w.Sb
}

func (w *SignOrigin_Sb) Pb() eventpb.SignOrigin_Type {
	return &eventpb.SignOrigin_Sb{Sb: w.Sb}
}

func (*SignOrigin_Sb) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SignOrigin_Sb]()}
}

func SignOriginFromPb(pb *eventpb.SignOrigin) *SignOrigin {
	return &SignOrigin{
		Module: (types.ModuleID)(pb.Module),
		Type:   SignOrigin_TypeFromPb(pb.Type),
	}
}

func (m *SignOrigin) Pb() *eventpb.SignOrigin {
	return &eventpb.SignOrigin{
		Module: (string)(m.Module),
		Type:   (m.Type).Pb(),
	}
}

func (*SignOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SignOrigin]()}
}

type SigVerData struct {
	Data [][]uint8
}

func SigVerDataFromPb(pb *eventpb.SigVerData) *SigVerData {
	return &SigVerData{
		Data: pb.Data,
	}
}

func (m *SigVerData) Pb() *eventpb.SigVerData {
	return &eventpb.SigVerData{
		Data: m.Data,
	}
}

func (*SigVerData) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SigVerData]()}
}

type VerifyNodeSigs struct {
	Data       []*SigVerData
	Signatures [][]uint8
	Origin     *SigVerOrigin
	NodeIds    []types.NodeID
}

func VerifyNodeSigsFromPb(pb *eventpb.VerifyNodeSigs) *VerifyNodeSigs {
	return &VerifyNodeSigs{
		Data: types7.ConvertSlice(pb.Data, func(t *eventpb.SigVerData) *SigVerData {
			return SigVerDataFromPb(t)
		}),
		Signatures: pb.Signatures,
		Origin:     SigVerOriginFromPb(pb.Origin),
		NodeIds: types7.ConvertSlice(pb.NodeIds, func(t string) types.NodeID {
			return (types.NodeID)(t)
		}),
	}
}

func (m *VerifyNodeSigs) Pb() *eventpb.VerifyNodeSigs {
	return &eventpb.VerifyNodeSigs{
		Data: types7.ConvertSlice(m.Data, func(t *SigVerData) *eventpb.SigVerData {
			return (t).Pb()
		}),
		Signatures: m.Signatures,
		Origin:     (m.Origin).Pb(),
		NodeIds: types7.ConvertSlice(m.NodeIds, func(t types.NodeID) string {
			return (string)(t)
		}),
	}
}

func (*VerifyNodeSigs) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.VerifyNodeSigs]()}
}

type NodeSigsVerified struct {
	Origin  *SigVerOrigin
	NodeIds []types.NodeID
	Valid   []bool
	Errors  []error
	AllOk   bool
}

func NodeSigsVerifiedFromPb(pb *eventpb.NodeSigsVerified) *NodeSigsVerified {
	return &NodeSigsVerified{
		Origin: SigVerOriginFromPb(pb.Origin),
		NodeIds: types7.ConvertSlice(pb.NodeIds, func(t string) types.NodeID {
			return (types.NodeID)(t)
		}),
		Valid: pb.Valid,
		Errors: types7.ConvertSlice(pb.Errors, func(t string) error {
			return types7.StringToError(t)
		}),
		AllOk: pb.AllOk,
	}
}

func (m *NodeSigsVerified) Pb() *eventpb.NodeSigsVerified {
	return &eventpb.NodeSigsVerified{
		Origin: (m.Origin).Pb(),
		NodeIds: types7.ConvertSlice(m.NodeIds, func(t types.NodeID) string {
			return (string)(t)
		}),
		Valid: m.Valid,
		Errors: types7.ConvertSlice(m.Errors, func(t error) string {
			return types7.ErrorToString(t)
		}),
		AllOk: m.AllOk,
	}
}

func (*NodeSigsVerified) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.NodeSigsVerified]()}
}

type SigVerOrigin struct {
	Module types.ModuleID
	Type   SigVerOrigin_Type
}

type SigVerOrigin_Type interface {
	mirreflect.GeneratedType
	isSigVerOrigin_Type()
	Pb() eventpb.SigVerOrigin_Type
}

type SigVerOrigin_TypeWrapper[T any] interface {
	SigVerOrigin_Type
	Unwrap() *T
}

func SigVerOrigin_TypeFromPb(pb eventpb.SigVerOrigin_Type) SigVerOrigin_Type {
	switch pb := pb.(type) {
	case *eventpb.SigVerOrigin_ContextStore:
		return &SigVerOrigin_ContextStore{ContextStore: types8.OriginFromPb(pb.ContextStore)}
	case *eventpb.SigVerOrigin_Iss:
		return &SigVerOrigin_Iss{Iss: pb.Iss}
	case *eventpb.SigVerOrigin_Dsl:
		return &SigVerOrigin_Dsl{Dsl: types9.OriginFromPb(pb.Dsl)}
	case *eventpb.SigVerOrigin_Checkpoint:
		return &SigVerOrigin_Checkpoint{Checkpoint: pb.Checkpoint}
	case *eventpb.SigVerOrigin_Sb:
		return &SigVerOrigin_Sb{Sb: pb.Sb}
	}
	return nil
}

type SigVerOrigin_ContextStore struct {
	ContextStore *types8.Origin
}

func (*SigVerOrigin_ContextStore) isSigVerOrigin_Type() {}

func (w *SigVerOrigin_ContextStore) Unwrap() *types8.Origin {
	return w.ContextStore
}

func (w *SigVerOrigin_ContextStore) Pb() eventpb.SigVerOrigin_Type {
	return &eventpb.SigVerOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*SigVerOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SigVerOrigin_ContextStore]()}
}

type SigVerOrigin_Iss struct {
	Iss *isspb.ISSSigVerOrigin
}

func (*SigVerOrigin_Iss) isSigVerOrigin_Type() {}

func (w *SigVerOrigin_Iss) Unwrap() *isspb.ISSSigVerOrigin {
	return w.Iss
}

func (w *SigVerOrigin_Iss) Pb() eventpb.SigVerOrigin_Type {
	return &eventpb.SigVerOrigin_Iss{Iss: w.Iss}
}

func (*SigVerOrigin_Iss) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SigVerOrigin_Iss]()}
}

type SigVerOrigin_Dsl struct {
	Dsl *types9.Origin
}

func (*SigVerOrigin_Dsl) isSigVerOrigin_Type() {}

func (w *SigVerOrigin_Dsl) Unwrap() *types9.Origin {
	return w.Dsl
}

func (w *SigVerOrigin_Dsl) Pb() eventpb.SigVerOrigin_Type {
	return &eventpb.SigVerOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*SigVerOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SigVerOrigin_Dsl]()}
}

type SigVerOrigin_Checkpoint struct {
	Checkpoint *checkpointpb.SigVerOrigin
}

func (*SigVerOrigin_Checkpoint) isSigVerOrigin_Type() {}

func (w *SigVerOrigin_Checkpoint) Unwrap() *checkpointpb.SigVerOrigin {
	return w.Checkpoint
}

func (w *SigVerOrigin_Checkpoint) Pb() eventpb.SigVerOrigin_Type {
	return &eventpb.SigVerOrigin_Checkpoint{Checkpoint: w.Checkpoint}
}

func (*SigVerOrigin_Checkpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SigVerOrigin_Checkpoint]()}
}

type SigVerOrigin_Sb struct {
	Sb *ordererspb.SBInstanceSigVerOrigin
}

func (*SigVerOrigin_Sb) isSigVerOrigin_Type() {}

func (w *SigVerOrigin_Sb) Unwrap() *ordererspb.SBInstanceSigVerOrigin {
	return w.Sb
}

func (w *SigVerOrigin_Sb) Pb() eventpb.SigVerOrigin_Type {
	return &eventpb.SigVerOrigin_Sb{Sb: w.Sb}
}

func (*SigVerOrigin_Sb) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SigVerOrigin_Sb]()}
}

func SigVerOriginFromPb(pb *eventpb.SigVerOrigin) *SigVerOrigin {
	return &SigVerOrigin{
		Module: (types.ModuleID)(pb.Module),
		Type:   SigVerOrigin_TypeFromPb(pb.Type),
	}
}

func (m *SigVerOrigin) Pb() *eventpb.SigVerOrigin {
	return &eventpb.SigVerOrigin{
		Module: (string)(m.Module),
		Type:   (m.Type).Pb(),
	}
}

func (*SigVerOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SigVerOrigin]()}
}

type SendMessage struct {
	Msg          *types10.Message
	Destinations []types.NodeID
}

func SendMessageFromPb(pb *eventpb.SendMessage) *SendMessage {
	return &SendMessage{
		Msg: types10.MessageFromPb(pb.Msg),
		Destinations: types7.ConvertSlice(pb.Destinations, func(t string) types.NodeID {
			return (types.NodeID)(t)
		}),
	}
}

func (m *SendMessage) Pb() *eventpb.SendMessage {
	return &eventpb.SendMessage{
		Msg: (m.Msg).Pb(),
		Destinations: types7.ConvertSlice(m.Destinations, func(t types.NodeID) string {
			return (string)(t)
		}),
	}
}

func (*SendMessage) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SendMessage]()}
}

type MessageReceived struct {
	From types.NodeID
	Msg  *types10.Message
}

func MessageReceivedFromPb(pb *eventpb.MessageReceived) *MessageReceived {
	return &MessageReceived{
		From: (types.NodeID)(pb.From),
		Msg:  types10.MessageFromPb(pb.Msg),
	}
}

func (m *MessageReceived) Pb() *eventpb.MessageReceived {
	return &eventpb.MessageReceived{
		From: (string)(m.From),
		Msg:  (m.Msg).Pb(),
	}
}

func (*MessageReceived) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.MessageReceived]()}
}
