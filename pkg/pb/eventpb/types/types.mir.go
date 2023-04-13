package eventpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types11 "github.com/filecoin-project/mir/codegen/model/types"
	types5 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types6 "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/bcbpb/types"
	checkpointpb "github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	types8 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types16 "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	types13 "github.com/filecoin-project/mir/pkg/pb/contextstorepb/types"
	types14 "github.com/filecoin-project/mir/pkg/pb/dslpb/types"
	eventpb "github.com/filecoin-project/mir/pkg/pb/eventpb"
	types9 "github.com/filecoin-project/mir/pkg/pb/factorymodulepb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	types10 "github.com/filecoin-project/mir/pkg/pb/isspb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	types15 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	ordererspb "github.com/filecoin-project/mir/pkg/pb/ordererspb"
	pingpongpb "github.com/filecoin-project/mir/pkg/pb/pingpongpb"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
	types12 "github.com/filecoin-project/mir/pkg/pb/requestpb/types"
	types7 "github.com/filecoin-project/mir/pkg/pb/threshcryptopb/types"
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
		return &Event_Init{Init: InitFromPb(pb.Init)}
	case *eventpb.Event_Timer:
		return &Event_Timer{Timer: TimerEventFromPb(pb.Timer)}
	case *eventpb.Event_Hasher:
		return &Event_Hasher{Hasher: types1.EventFromPb(pb.Hasher)}
	case *eventpb.Event_Bcb:
		return &Event_Bcb{Bcb: types2.EventFromPb(pb.Bcb)}
	case *eventpb.Event_Mempool:
		return &Event_Mempool{Mempool: types3.EventFromPb(pb.Mempool)}
	case *eventpb.Event_Availability:
		return &Event_Availability{Availability: types4.EventFromPb(pb.Availability)}
	case *eventpb.Event_BatchDb:
		return &Event_BatchDb{BatchDb: types5.EventFromPb(pb.BatchDb)}
	case *eventpb.Event_BatchFetcher:
		return &Event_BatchFetcher{BatchFetcher: types6.EventFromPb(pb.BatchFetcher)}
	case *eventpb.Event_ThreshCrypto:
		return &Event_ThreshCrypto{ThreshCrypto: types7.EventFromPb(pb.ThreshCrypto)}
	case *eventpb.Event_PingPong:
		return &Event_PingPong{PingPong: pb.PingPong}
	case *eventpb.Event_Checkpoint:
		return &Event_Checkpoint{Checkpoint: types8.EventFromPb(pb.Checkpoint)}
	case *eventpb.Event_Factory:
		return &Event_Factory{Factory: types9.FactoryFromPb(pb.Factory)}
	case *eventpb.Event_SbEvent:
		return &Event_SbEvent{SbEvent: pb.SbEvent}
	case *eventpb.Event_Iss:
		return &Event_Iss{Iss: types10.ISSEventFromPb(pb.Iss)}
	case *eventpb.Event_NewRequests:
		return &Event_NewRequests{NewRequests: NewRequestsFromPb(pb.NewRequests)}
	case *eventpb.Event_SignRequest:
		return &Event_SignRequest{SignRequest: SignRequestFromPb(pb.SignRequest)}
	case *eventpb.Event_SignResult:
		return &Event_SignResult{SignResult: SignResultFromPb(pb.SignResult)}
	case *eventpb.Event_VerifyNodeSigs:
		return &Event_VerifyNodeSigs{VerifyNodeSigs: VerifyNodeSigsFromPb(pb.VerifyNodeSigs)}
	case *eventpb.Event_NodeSigsVerified:
		return &Event_NodeSigsVerified{NodeSigsVerified: NodeSigsVerifiedFromPb(pb.NodeSigsVerified)}
	case *eventpb.Event_SendMessage:
		return &Event_SendMessage{SendMessage: SendMessageFromPb(pb.SendMessage)}
	case *eventpb.Event_MessageReceived:
		return &Event_MessageReceived{MessageReceived: MessageReceivedFromPb(pb.MessageReceived)}
	case *eventpb.Event_DeliverCert:
		return &Event_DeliverCert{DeliverCert: DeliverCertFromPb(pb.DeliverCert)}
	case *eventpb.Event_VerifyRequestSig:
		return &Event_VerifyRequestSig{VerifyRequestSig: pb.VerifyRequestSig}
	case *eventpb.Event_RequestSigVerified:
		return &Event_RequestSigVerified{RequestSigVerified: pb.RequestSigVerified}
	case *eventpb.Event_StoreVerifiedRequest:
		return &Event_StoreVerifiedRequest{StoreVerifiedRequest: pb.StoreVerifiedRequest}
	case *eventpb.Event_AppSnapshotRequest:
		return &Event_AppSnapshotRequest{AppSnapshotRequest: AppSnapshotRequestFromPb(pb.AppSnapshotRequest)}
	case *eventpb.Event_AppSnapshot:
		return &Event_AppSnapshot{AppSnapshot: pb.AppSnapshot}
	case *eventpb.Event_AppRestoreState:
		return &Event_AppRestoreState{AppRestoreState: AppRestoreStateFromPb(pb.AppRestoreState)}
	case *eventpb.Event_NewEpoch:
		return &Event_NewEpoch{NewEpoch: NewEpochFromPb(pb.NewEpoch)}
	case *eventpb.Event_NewConfig:
		return &Event_NewConfig{NewConfig: NewConfigFromPb(pb.NewConfig)}
	case *eventpb.Event_TestingString:
		return &Event_TestingString{TestingString: pb.TestingString}
	case *eventpb.Event_TestingUint:
		return &Event_TestingUint{TestingUint: pb.TestingUint}
	}
	return nil
}

type Event_Init struct {
	Init *Init
}

func (*Event_Init) isEvent_Type() {}

func (w *Event_Init) Unwrap() *Init {
	return w.Init
}

func (w *Event_Init) Pb() eventpb.Event_Type {
	return &eventpb.Event_Init{Init: (w.Init).Pb()}
}

func (*Event_Init) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Init]()}
}

type Event_Timer struct {
	Timer *TimerEvent
}

func (*Event_Timer) isEvent_Type() {}

func (w *Event_Timer) Unwrap() *TimerEvent {
	return w.Timer
}

func (w *Event_Timer) Pb() eventpb.Event_Type {
	return &eventpb.Event_Timer{Timer: (w.Timer).Pb()}
}

func (*Event_Timer) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Timer]()}
}

type Event_Hasher struct {
	Hasher *types1.Event
}

func (*Event_Hasher) isEvent_Type() {}

func (w *Event_Hasher) Unwrap() *types1.Event {
	return w.Hasher
}

func (w *Event_Hasher) Pb() eventpb.Event_Type {
	return &eventpb.Event_Hasher{Hasher: (w.Hasher).Pb()}
}

func (*Event_Hasher) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Hasher]()}
}

type Event_Bcb struct {
	Bcb *types2.Event
}

func (*Event_Bcb) isEvent_Type() {}

func (w *Event_Bcb) Unwrap() *types2.Event {
	return w.Bcb
}

func (w *Event_Bcb) Pb() eventpb.Event_Type {
	return &eventpb.Event_Bcb{Bcb: (w.Bcb).Pb()}
}

func (*Event_Bcb) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Bcb]()}
}

type Event_Mempool struct {
	Mempool *types3.Event
}

func (*Event_Mempool) isEvent_Type() {}

func (w *Event_Mempool) Unwrap() *types3.Event {
	return w.Mempool
}

func (w *Event_Mempool) Pb() eventpb.Event_Type {
	return &eventpb.Event_Mempool{Mempool: (w.Mempool).Pb()}
}

func (*Event_Mempool) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Mempool]()}
}

type Event_Availability struct {
	Availability *types4.Event
}

func (*Event_Availability) isEvent_Type() {}

func (w *Event_Availability) Unwrap() *types4.Event {
	return w.Availability
}

func (w *Event_Availability) Pb() eventpb.Event_Type {
	return &eventpb.Event_Availability{Availability: (w.Availability).Pb()}
}

func (*Event_Availability) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Availability]()}
}

type Event_BatchDb struct {
	BatchDb *types5.Event
}

func (*Event_BatchDb) isEvent_Type() {}

func (w *Event_BatchDb) Unwrap() *types5.Event {
	return w.BatchDb
}

func (w *Event_BatchDb) Pb() eventpb.Event_Type {
	return &eventpb.Event_BatchDb{BatchDb: (w.BatchDb).Pb()}
}

func (*Event_BatchDb) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_BatchDb]()}
}

type Event_BatchFetcher struct {
	BatchFetcher *types6.Event
}

func (*Event_BatchFetcher) isEvent_Type() {}

func (w *Event_BatchFetcher) Unwrap() *types6.Event {
	return w.BatchFetcher
}

func (w *Event_BatchFetcher) Pb() eventpb.Event_Type {
	return &eventpb.Event_BatchFetcher{BatchFetcher: (w.BatchFetcher).Pb()}
}

func (*Event_BatchFetcher) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_BatchFetcher]()}
}

type Event_ThreshCrypto struct {
	ThreshCrypto *types7.Event
}

func (*Event_ThreshCrypto) isEvent_Type() {}

func (w *Event_ThreshCrypto) Unwrap() *types7.Event {
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
	Checkpoint *types8.Event
}

func (*Event_Checkpoint) isEvent_Type() {}

func (w *Event_Checkpoint) Unwrap() *types8.Event {
	return w.Checkpoint
}

func (w *Event_Checkpoint) Pb() eventpb.Event_Type {
	return &eventpb.Event_Checkpoint{Checkpoint: (w.Checkpoint).Pb()}
}

func (*Event_Checkpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Checkpoint]()}
}

type Event_Factory struct {
	Factory *types9.Factory
}

func (*Event_Factory) isEvent_Type() {}

func (w *Event_Factory) Unwrap() *types9.Factory {
	return w.Factory
}

func (w *Event_Factory) Pb() eventpb.Event_Type {
	return &eventpb.Event_Factory{Factory: (w.Factory).Pb()}
}

func (*Event_Factory) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Factory]()}
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

type Event_Iss struct {
	Iss *types10.ISSEvent
}

func (*Event_Iss) isEvent_Type() {}

func (w *Event_Iss) Unwrap() *types10.ISSEvent {
	return w.Iss
}

func (w *Event_Iss) Pb() eventpb.Event_Type {
	return &eventpb.Event_Iss{Iss: (w.Iss).Pb()}
}

func (*Event_Iss) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Iss]()}
}

type Event_NewRequests struct {
	NewRequests *NewRequests
}

func (*Event_NewRequests) isEvent_Type() {}

func (w *Event_NewRequests) Unwrap() *NewRequests {
	return w.NewRequests
}

func (w *Event_NewRequests) Pb() eventpb.Event_Type {
	return &eventpb.Event_NewRequests{NewRequests: (w.NewRequests).Pb()}
}

func (*Event_NewRequests) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_NewRequests]()}
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
	DeliverCert *DeliverCert
}

func (*Event_DeliverCert) isEvent_Type() {}

func (w *Event_DeliverCert) Unwrap() *DeliverCert {
	return w.DeliverCert
}

func (w *Event_DeliverCert) Pb() eventpb.Event_Type {
	return &eventpb.Event_DeliverCert{DeliverCert: (w.DeliverCert).Pb()}
}

func (*Event_DeliverCert) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_DeliverCert]()}
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
	AppSnapshotRequest *AppSnapshotRequest
}

func (*Event_AppSnapshotRequest) isEvent_Type() {}

func (w *Event_AppSnapshotRequest) Unwrap() *AppSnapshotRequest {
	return w.AppSnapshotRequest
}

func (w *Event_AppSnapshotRequest) Pb() eventpb.Event_Type {
	return &eventpb.Event_AppSnapshotRequest{AppSnapshotRequest: (w.AppSnapshotRequest).Pb()}
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
	AppRestoreState *AppRestoreState
}

func (*Event_AppRestoreState) isEvent_Type() {}

func (w *Event_AppRestoreState) Unwrap() *AppRestoreState {
	return w.AppRestoreState
}

func (w *Event_AppRestoreState) Pb() eventpb.Event_Type {
	return &eventpb.Event_AppRestoreState{AppRestoreState: (w.AppRestoreState).Pb()}
}

func (*Event_AppRestoreState) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_AppRestoreState]()}
}

type Event_NewEpoch struct {
	NewEpoch *NewEpoch
}

func (*Event_NewEpoch) isEvent_Type() {}

func (w *Event_NewEpoch) Unwrap() *NewEpoch {
	return w.NewEpoch
}

func (w *Event_NewEpoch) Pb() eventpb.Event_Type {
	return &eventpb.Event_NewEpoch{NewEpoch: (w.NewEpoch).Pb()}
}

func (*Event_NewEpoch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_NewEpoch]()}
}

type Event_NewConfig struct {
	NewConfig *NewConfig
}

func (*Event_NewConfig) isEvent_Type() {}

func (w *Event_NewConfig) Unwrap() *NewConfig {
	return w.NewConfig
}

func (w *Event_NewConfig) Pb() eventpb.Event_Type {
	return &eventpb.Event_NewConfig{NewConfig: (w.NewConfig).Pb()}
}

func (*Event_NewConfig) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_NewConfig]()}
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
		Next: types11.ConvertSlice(pb.Next, func(t *eventpb.Event) *Event {
			return EventFromPb(t)
		}),
		DestModule: (types.ModuleID)(pb.DestModule),
	}
}

func (m *Event) Pb() *eventpb.Event {
	return &eventpb.Event{
		Type: (m.Type).Pb(),
		Next: types11.ConvertSlice(m.Next, func(t *Event) *eventpb.Event {
			return (t).Pb()
		}),
		DestModule: (string)(m.DestModule),
	}
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event]()}
}

type Init struct{}

func InitFromPb(pb *eventpb.Init) *Init {
	return &Init{}
}

func (m *Init) Pb() *eventpb.Init {
	return &eventpb.Init{}
}

func (*Init) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Init]()}
}

type NewRequests struct {
	Requests []*types12.Request
}

func NewRequestsFromPb(pb *eventpb.NewRequests) *NewRequests {
	return &NewRequests{
		Requests: types11.ConvertSlice(pb.Requests, func(t *requestpb.Request) *types12.Request {
			return types12.RequestFromPb(t)
		}),
	}
}

func (m *NewRequests) Pb() *eventpb.NewRequests {
	return &eventpb.NewRequests{
		Requests: types11.ConvertSlice(m.Requests, func(t *types12.Request) *requestpb.Request {
			return (t).Pb()
		}),
	}
}

func (*NewRequests) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.NewRequests]()}
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
		return &SignOrigin_ContextStore{ContextStore: types13.OriginFromPb(pb.ContextStore)}
	case *eventpb.SignOrigin_Dsl:
		return &SignOrigin_Dsl{Dsl: types14.OriginFromPb(pb.Dsl)}
	case *eventpb.SignOrigin_Checkpoint:
		return &SignOrigin_Checkpoint{Checkpoint: pb.Checkpoint}
	case *eventpb.SignOrigin_Sb:
		return &SignOrigin_Sb{Sb: pb.Sb}
	}
	return nil
}

type SignOrigin_ContextStore struct {
	ContextStore *types13.Origin
}

func (*SignOrigin_ContextStore) isSignOrigin_Type() {}

func (w *SignOrigin_ContextStore) Unwrap() *types13.Origin {
	return w.ContextStore
}

func (w *SignOrigin_ContextStore) Pb() eventpb.SignOrigin_Type {
	return &eventpb.SignOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*SignOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SignOrigin_ContextStore]()}
}

type SignOrigin_Dsl struct {
	Dsl *types14.Origin
}

func (*SignOrigin_Dsl) isSignOrigin_Type() {}

func (w *SignOrigin_Dsl) Unwrap() *types14.Origin {
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
		Data: types11.ConvertSlice(pb.Data, func(t *eventpb.SigVerData) *SigVerData {
			return SigVerDataFromPb(t)
		}),
		Signatures: pb.Signatures,
		Origin:     SigVerOriginFromPb(pb.Origin),
		NodeIds: types11.ConvertSlice(pb.NodeIds, func(t string) types.NodeID {
			return (types.NodeID)(t)
		}),
	}
}

func (m *VerifyNodeSigs) Pb() *eventpb.VerifyNodeSigs {
	return &eventpb.VerifyNodeSigs{
		Data: types11.ConvertSlice(m.Data, func(t *SigVerData) *eventpb.SigVerData {
			return (t).Pb()
		}),
		Signatures: m.Signatures,
		Origin:     (m.Origin).Pb(),
		NodeIds: types11.ConvertSlice(m.NodeIds, func(t types.NodeID) string {
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
		NodeIds: types11.ConvertSlice(pb.NodeIds, func(t string) types.NodeID {
			return (types.NodeID)(t)
		}),
		Valid: pb.Valid,
		Errors: types11.ConvertSlice(pb.Errors, func(t string) error {
			return types11.StringToError(t)
		}),
		AllOk: pb.AllOk,
	}
}

func (m *NodeSigsVerified) Pb() *eventpb.NodeSigsVerified {
	return &eventpb.NodeSigsVerified{
		Origin: (m.Origin).Pb(),
		NodeIds: types11.ConvertSlice(m.NodeIds, func(t types.NodeID) string {
			return (string)(t)
		}),
		Valid: m.Valid,
		Errors: types11.ConvertSlice(m.Errors, func(t error) string {
			return types11.ErrorToString(t)
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
		return &SigVerOrigin_ContextStore{ContextStore: types13.OriginFromPb(pb.ContextStore)}
	case *eventpb.SigVerOrigin_Dsl:
		return &SigVerOrigin_Dsl{Dsl: types14.OriginFromPb(pb.Dsl)}
	case *eventpb.SigVerOrigin_Checkpoint:
		return &SigVerOrigin_Checkpoint{Checkpoint: pb.Checkpoint}
	case *eventpb.SigVerOrigin_Sb:
		return &SigVerOrigin_Sb{Sb: pb.Sb}
	}
	return nil
}

type SigVerOrigin_ContextStore struct {
	ContextStore *types13.Origin
}

func (*SigVerOrigin_ContextStore) isSigVerOrigin_Type() {}

func (w *SigVerOrigin_ContextStore) Unwrap() *types13.Origin {
	return w.ContextStore
}

func (w *SigVerOrigin_ContextStore) Pb() eventpb.SigVerOrigin_Type {
	return &eventpb.SigVerOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*SigVerOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SigVerOrigin_ContextStore]()}
}

type SigVerOrigin_Dsl struct {
	Dsl *types14.Origin
}

func (*SigVerOrigin_Dsl) isSigVerOrigin_Type() {}

func (w *SigVerOrigin_Dsl) Unwrap() *types14.Origin {
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
	Msg          *types15.Message
	Destinations []types.NodeID
}

func SendMessageFromPb(pb *eventpb.SendMessage) *SendMessage {
	return &SendMessage{
		Msg: types15.MessageFromPb(pb.Msg),
		Destinations: types11.ConvertSlice(pb.Destinations, func(t string) types.NodeID {
			return (types.NodeID)(t)
		}),
	}
}

func (m *SendMessage) Pb() *eventpb.SendMessage {
	return &eventpb.SendMessage{
		Msg: (m.Msg).Pb(),
		Destinations: types11.ConvertSlice(m.Destinations, func(t types.NodeID) string {
			return (string)(t)
		}),
	}
}

func (*SendMessage) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.SendMessage]()}
}

type MessageReceived struct {
	From types.NodeID
	Msg  *types15.Message
}

func MessageReceivedFromPb(pb *eventpb.MessageReceived) *MessageReceived {
	return &MessageReceived{
		From: (types.NodeID)(pb.From),
		Msg:  types15.MessageFromPb(pb.Msg),
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

type DeliverCert struct {
	Sn   types.SeqNr
	Cert *types4.Cert
}

func DeliverCertFromPb(pb *eventpb.DeliverCert) *DeliverCert {
	return &DeliverCert{
		Sn:   (types.SeqNr)(pb.Sn),
		Cert: types4.CertFromPb(pb.Cert),
	}
}

func (m *DeliverCert) Pb() *eventpb.DeliverCert {
	return &eventpb.DeliverCert{
		Sn:   (uint64)(m.Sn),
		Cert: (m.Cert).Pb(),
	}
}

func (*DeliverCert) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.DeliverCert]()}
}

type AppSnapshotRequest struct {
	ReplyTo types.ModuleID
}

func AppSnapshotRequestFromPb(pb *eventpb.AppSnapshotRequest) *AppSnapshotRequest {
	return &AppSnapshotRequest{
		ReplyTo: (types.ModuleID)(pb.ReplyTo),
	}
}

func (m *AppSnapshotRequest) Pb() *eventpb.AppSnapshotRequest {
	return &eventpb.AppSnapshotRequest{
		ReplyTo: (string)(m.ReplyTo),
	}
}

func (*AppSnapshotRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.AppSnapshotRequest]()}
}

type AppRestoreState struct {
	Checkpoint *types8.StableCheckpoint
}

func AppRestoreStateFromPb(pb *eventpb.AppRestoreState) *AppRestoreState {
	return &AppRestoreState{
		Checkpoint: types8.StableCheckpointFromPb(pb.Checkpoint),
	}
}

func (m *AppRestoreState) Pb() *eventpb.AppRestoreState {
	return &eventpb.AppRestoreState{
		Checkpoint: (m.Checkpoint).Pb(),
	}
}

func (*AppRestoreState) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.AppRestoreState]()}
}

type NewEpoch struct {
	EpochNr types.EpochNr
}

func NewEpochFromPb(pb *eventpb.NewEpoch) *NewEpoch {
	return &NewEpoch{
		EpochNr: (types.EpochNr)(pb.EpochNr),
	}
}

func (m *NewEpoch) Pb() *eventpb.NewEpoch {
	return &eventpb.NewEpoch{
		EpochNr: (uint64)(m.EpochNr),
	}
}

func (*NewEpoch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.NewEpoch]()}
}

type NewConfig struct {
	EpochNr    types.EpochNr
	Membership *types16.Membership
}

func NewConfigFromPb(pb *eventpb.NewConfig) *NewConfig {
	return &NewConfig{
		EpochNr:    (types.EpochNr)(pb.EpochNr),
		Membership: types16.MembershipFromPb(pb.Membership),
	}
}

func (m *NewConfig) Pb() *eventpb.NewConfig {
	return &eventpb.NewConfig{
		EpochNr:    (uint64)(m.EpochNr),
		Membership: (m.Membership).Pb(),
	}
}

func (*NewConfig) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.NewConfig]()}
}

type TimerEvent struct {
	Type TimerEvent_Type
}

type TimerEvent_Type interface {
	mirreflect.GeneratedType
	isTimerEvent_Type()
	Pb() eventpb.TimerEvent_Type
}

type TimerEvent_TypeWrapper[T any] interface {
	TimerEvent_Type
	Unwrap() *T
}

func TimerEvent_TypeFromPb(pb eventpb.TimerEvent_Type) TimerEvent_Type {
	switch pb := pb.(type) {
	case *eventpb.TimerEvent_Delay:
		return &TimerEvent_Delay{Delay: TimerDelayFromPb(pb.Delay)}
	case *eventpb.TimerEvent_Repeat:
		return &TimerEvent_Repeat{Repeat: TimerRepeatFromPb(pb.Repeat)}
	case *eventpb.TimerEvent_GarbageCollect:
		return &TimerEvent_GarbageCollect{GarbageCollect: TimerGarbageCollectFromPb(pb.GarbageCollect)}
	}
	return nil
}

type TimerEvent_Delay struct {
	Delay *TimerDelay
}

func (*TimerEvent_Delay) isTimerEvent_Type() {}

func (w *TimerEvent_Delay) Unwrap() *TimerDelay {
	return w.Delay
}

func (w *TimerEvent_Delay) Pb() eventpb.TimerEvent_Type {
	return &eventpb.TimerEvent_Delay{Delay: (w.Delay).Pb()}
}

func (*TimerEvent_Delay) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.TimerEvent_Delay]()}
}

type TimerEvent_Repeat struct {
	Repeat *TimerRepeat
}

func (*TimerEvent_Repeat) isTimerEvent_Type() {}

func (w *TimerEvent_Repeat) Unwrap() *TimerRepeat {
	return w.Repeat
}

func (w *TimerEvent_Repeat) Pb() eventpb.TimerEvent_Type {
	return &eventpb.TimerEvent_Repeat{Repeat: (w.Repeat).Pb()}
}

func (*TimerEvent_Repeat) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.TimerEvent_Repeat]()}
}

type TimerEvent_GarbageCollect struct {
	GarbageCollect *TimerGarbageCollect
}

func (*TimerEvent_GarbageCollect) isTimerEvent_Type() {}

func (w *TimerEvent_GarbageCollect) Unwrap() *TimerGarbageCollect {
	return w.GarbageCollect
}

func (w *TimerEvent_GarbageCollect) Pb() eventpb.TimerEvent_Type {
	return &eventpb.TimerEvent_GarbageCollect{GarbageCollect: (w.GarbageCollect).Pb()}
}

func (*TimerEvent_GarbageCollect) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.TimerEvent_GarbageCollect]()}
}

func TimerEventFromPb(pb *eventpb.TimerEvent) *TimerEvent {
	return &TimerEvent{
		Type: TimerEvent_TypeFromPb(pb.Type),
	}
}

func (m *TimerEvent) Pb() *eventpb.TimerEvent {
	return &eventpb.TimerEvent{
		Type: (m.Type).Pb(),
	}
}

func (*TimerEvent) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.TimerEvent]()}
}

type TimerDelay struct {
	Events []*Event
	Delay  uint64
}

func TimerDelayFromPb(pb *eventpb.TimerDelay) *TimerDelay {
	return &TimerDelay{
		Events: types11.ConvertSlice(pb.Events, func(t *eventpb.Event) *Event {
			return EventFromPb(t)
		}),
		Delay: pb.Delay,
	}
}

func (m *TimerDelay) Pb() *eventpb.TimerDelay {
	return &eventpb.TimerDelay{
		Events: types11.ConvertSlice(m.Events, func(t *Event) *eventpb.Event {
			return (t).Pb()
		}),
		Delay: m.Delay,
	}
}

func (*TimerDelay) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.TimerDelay]()}
}

type TimerRepeat struct {
	EventsToRepeat []*Event
	Delay          types.TimeDuration
	RetentionIndex types.RetentionIndex
}

func TimerRepeatFromPb(pb *eventpb.TimerRepeat) *TimerRepeat {
	return &TimerRepeat{
		EventsToRepeat: types11.ConvertSlice(pb.EventsToRepeat, func(t *eventpb.Event) *Event {
			return EventFromPb(t)
		}),
		Delay:          (types.TimeDuration)(pb.Delay),
		RetentionIndex: (types.RetentionIndex)(pb.RetentionIndex),
	}
}

func (m *TimerRepeat) Pb() *eventpb.TimerRepeat {
	return &eventpb.TimerRepeat{
		EventsToRepeat: types11.ConvertSlice(m.EventsToRepeat, func(t *Event) *eventpb.Event {
			return (t).Pb()
		}),
		Delay:          (uint64)(m.Delay),
		RetentionIndex: (uint64)(m.RetentionIndex),
	}
}

func (*TimerRepeat) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.TimerRepeat]()}
}

type TimerGarbageCollect struct {
	RetentionIndex types.RetentionIndex
}

func TimerGarbageCollectFromPb(pb *eventpb.TimerGarbageCollect) *TimerGarbageCollect {
	return &TimerGarbageCollect{
		RetentionIndex: (types.RetentionIndex)(pb.RetentionIndex),
	}
}

func (m *TimerGarbageCollect) Pb() *eventpb.TimerGarbageCollect {
	return &eventpb.TimerGarbageCollect{
		RetentionIndex: (uint64)(m.RetentionIndex),
	}
}

func (*TimerGarbageCollect) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.TimerGarbageCollect]()}
}
