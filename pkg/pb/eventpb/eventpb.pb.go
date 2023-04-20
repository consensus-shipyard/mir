// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.14.0
// source: eventpb/eventpb.proto

package eventpb

import (
	apppb "github.com/filecoin-project/mir/pkg/pb/apppb"
	availabilitypb "github.com/filecoin-project/mir/pkg/pb/availabilitypb"
	batchdbpb "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb"
	batchfetcherpb "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
	bcbpb "github.com/filecoin-project/mir/pkg/pb/bcbpb"
	checkpointpb "github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	cryptopb "github.com/filecoin-project/mir/pkg/pb/cryptopb"
	factorypb "github.com/filecoin-project/mir/pkg/pb/factorypb"
	hasherpb "github.com/filecoin-project/mir/pkg/pb/hasherpb"
	isspb "github.com/filecoin-project/mir/pkg/pb/isspb"
	mempoolpb "github.com/filecoin-project/mir/pkg/pb/mempoolpb"
	_ "github.com/filecoin-project/mir/pkg/pb/mir"
	ordererpb "github.com/filecoin-project/mir/pkg/pb/ordererpb"
	pingpongpb "github.com/filecoin-project/mir/pkg/pb/pingpongpb"
	threshcryptopb "github.com/filecoin-project/mir/pkg/pb/threshcryptopb"
	transportpb "github.com/filecoin-project/mir/pkg/pb/transportpb"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Event represents a state event to be injected into the state machine
type Event struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DestModule string `protobuf:"bytes,1,opt,name=dest_module,json=destModule,proto3" json:"dest_module,omitempty"`
	// Types that are assignable to Type:
	//	*Event_Init
	//	*Event_Timer
	//	*Event_Hasher
	//	*Event_Bcb
	//	*Event_Mempool
	//	*Event_Availability
	//	*Event_BatchDb
	//	*Event_BatchFetcher
	//	*Event_ThreshCrypto
	//	*Event_Checkpoint
	//	*Event_Factory
	//	*Event_Iss
	//	*Event_Orderer
	//	*Event_Crypto
	//	*Event_App
	//	*Event_Transport
	//	*Event_PingPong
	//	*Event_TestingString
	//	*Event_TestingUint
	Type isEvent_Type `protobuf_oneof:"type"`
	// A list of follow-up events to process after this event has been processed.
	// This field is used if events need to be processed in a particular order.
	// For example, a message sending event must only be processed
	// after the corresponding entry has been persisted in the write-ahead log (WAL).
	// In this case, the WAL append event would be this event
	// and the next field would contain the message sending event.
	// (This is a hypothetical example, the WAL functionality is not implemented at a moment.)
	Next []*Event `protobuf:"bytes,400,rep,name=next,proto3" json:"next,omitempty"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eventpb_eventpb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_eventpb_eventpb_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_eventpb_eventpb_proto_rawDescGZIP(), []int{0}
}

func (x *Event) GetDestModule() string {
	if x != nil {
		return x.DestModule
	}
	return ""
}

func (m *Event) GetType() isEvent_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Event) GetInit() *Init {
	if x, ok := x.GetType().(*Event_Init); ok {
		return x.Init
	}
	return nil
}

func (x *Event) GetTimer() *TimerEvent {
	if x, ok := x.GetType().(*Event_Timer); ok {
		return x.Timer
	}
	return nil
}

func (x *Event) GetHasher() *hasherpb.Event {
	if x, ok := x.GetType().(*Event_Hasher); ok {
		return x.Hasher
	}
	return nil
}

func (x *Event) GetBcb() *bcbpb.Event {
	if x, ok := x.GetType().(*Event_Bcb); ok {
		return x.Bcb
	}
	return nil
}

func (x *Event) GetMempool() *mempoolpb.Event {
	if x, ok := x.GetType().(*Event_Mempool); ok {
		return x.Mempool
	}
	return nil
}

func (x *Event) GetAvailability() *availabilitypb.Event {
	if x, ok := x.GetType().(*Event_Availability); ok {
		return x.Availability
	}
	return nil
}

func (x *Event) GetBatchDb() *batchdbpb.Event {
	if x, ok := x.GetType().(*Event_BatchDb); ok {
		return x.BatchDb
	}
	return nil
}

func (x *Event) GetBatchFetcher() *batchfetcherpb.Event {
	if x, ok := x.GetType().(*Event_BatchFetcher); ok {
		return x.BatchFetcher
	}
	return nil
}

func (x *Event) GetThreshCrypto() *threshcryptopb.Event {
	if x, ok := x.GetType().(*Event_ThreshCrypto); ok {
		return x.ThreshCrypto
	}
	return nil
}

func (x *Event) GetCheckpoint() *checkpointpb.Event {
	if x, ok := x.GetType().(*Event_Checkpoint); ok {
		return x.Checkpoint
	}
	return nil
}

func (x *Event) GetFactory() *factorypb.Event {
	if x, ok := x.GetType().(*Event_Factory); ok {
		return x.Factory
	}
	return nil
}

func (x *Event) GetIss() *isspb.Event {
	if x, ok := x.GetType().(*Event_Iss); ok {
		return x.Iss
	}
	return nil
}

func (x *Event) GetOrderer() *ordererpb.Event {
	if x, ok := x.GetType().(*Event_Orderer); ok {
		return x.Orderer
	}
	return nil
}

func (x *Event) GetCrypto() *cryptopb.Event {
	if x, ok := x.GetType().(*Event_Crypto); ok {
		return x.Crypto
	}
	return nil
}

func (x *Event) GetApp() *apppb.Event {
	if x, ok := x.GetType().(*Event_App); ok {
		return x.App
	}
	return nil
}

func (x *Event) GetTransport() *transportpb.Event {
	if x, ok := x.GetType().(*Event_Transport); ok {
		return x.Transport
	}
	return nil
}

func (x *Event) GetPingPong() *pingpongpb.Event {
	if x, ok := x.GetType().(*Event_PingPong); ok {
		return x.PingPong
	}
	return nil
}

func (x *Event) GetTestingString() *wrapperspb.StringValue {
	if x, ok := x.GetType().(*Event_TestingString); ok {
		return x.TestingString
	}
	return nil
}

func (x *Event) GetTestingUint() *wrapperspb.UInt64Value {
	if x, ok := x.GetType().(*Event_TestingUint); ok {
		return x.TestingUint
	}
	return nil
}

func (x *Event) GetNext() []*Event {
	if x != nil {
		return x.Next
	}
	return nil
}

type isEvent_Type interface {
	isEvent_Type()
}

type Event_Init struct {
	// Special global event produced by the runtime itself and sent to each module on initialization.
	Init *Init `protobuf:"bytes,2,opt,name=init,proto3,oneof"`
}

type Event_Timer struct {
	// Timer events are recursive and must be defined in this file for protobuf-specific reasons, see below.
	Timer *TimerEvent `protobuf:"bytes,3,opt,name=timer,proto3,oneof"`
}

type Event_Hasher struct {
	// Module-specific events
	Hasher *hasherpb.Event `protobuf:"bytes,10,opt,name=hasher,proto3,oneof"`
}

type Event_Bcb struct {
	Bcb *bcbpb.Event `protobuf:"bytes,11,opt,name=bcb,proto3,oneof"`
}

type Event_Mempool struct {
	Mempool *mempoolpb.Event `protobuf:"bytes,12,opt,name=mempool,proto3,oneof"`
}

type Event_Availability struct {
	Availability *availabilitypb.Event `protobuf:"bytes,13,opt,name=availability,proto3,oneof"`
}

type Event_BatchDb struct {
	BatchDb *batchdbpb.Event `protobuf:"bytes,14,opt,name=batch_db,json=batchDb,proto3,oneof"`
}

type Event_BatchFetcher struct {
	BatchFetcher *batchfetcherpb.Event `protobuf:"bytes,15,opt,name=batch_fetcher,json=batchFetcher,proto3,oneof"`
}

type Event_ThreshCrypto struct {
	ThreshCrypto *threshcryptopb.Event `protobuf:"bytes,16,opt,name=thresh_crypto,json=threshCrypto,proto3,oneof"`
}

type Event_Checkpoint struct {
	Checkpoint *checkpointpb.Event `protobuf:"bytes,17,opt,name=checkpoint,proto3,oneof"`
}

type Event_Factory struct {
	Factory *factorypb.Event `protobuf:"bytes,18,opt,name=factory,proto3,oneof"`
}

type Event_Iss struct {
	Iss *isspb.Event `protobuf:"bytes,19,opt,name=iss,proto3,oneof"`
}

type Event_Orderer struct {
	Orderer *ordererpb.Event `protobuf:"bytes,20,opt,name=orderer,proto3,oneof"`
}

type Event_Crypto struct {
	Crypto *cryptopb.Event `protobuf:"bytes,21,opt,name=crypto,proto3,oneof"`
}

type Event_App struct {
	App *apppb.Event `protobuf:"bytes,22,opt,name=app,proto3,oneof"`
}

type Event_Transport struct {
	Transport *transportpb.Event `protobuf:"bytes,23,opt,name=transport,proto3,oneof"`
}

type Event_PingPong struct {
	// Events for code samples
	PingPong *pingpongpb.Event `protobuf:"bytes,200,opt,name=ping_pong,json=pingPong,proto3,oneof"`
}

type Event_TestingString struct {
	// for unit-tests
	TestingString *wrapperspb.StringValue `protobuf:"bytes,301,opt,name=testingString,proto3,oneof"`
}

type Event_TestingUint struct {
	TestingUint *wrapperspb.UInt64Value `protobuf:"bytes,302,opt,name=testingUint,proto3,oneof"`
}

func (*Event_Init) isEvent_Type() {}

func (*Event_Timer) isEvent_Type() {}

func (*Event_Hasher) isEvent_Type() {}

func (*Event_Bcb) isEvent_Type() {}

func (*Event_Mempool) isEvent_Type() {}

func (*Event_Availability) isEvent_Type() {}

func (*Event_BatchDb) isEvent_Type() {}

func (*Event_BatchFetcher) isEvent_Type() {}

func (*Event_ThreshCrypto) isEvent_Type() {}

func (*Event_Checkpoint) isEvent_Type() {}

func (*Event_Factory) isEvent_Type() {}

func (*Event_Iss) isEvent_Type() {}

func (*Event_Orderer) isEvent_Type() {}

func (*Event_Crypto) isEvent_Type() {}

func (*Event_App) isEvent_Type() {}

func (*Event_Transport) isEvent_Type() {}

func (*Event_PingPong) isEvent_Type() {}

func (*Event_TestingString) isEvent_Type() {}

func (*Event_TestingUint) isEvent_Type() {}

type Init struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Init) Reset() {
	*x = Init{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eventpb_eventpb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Init) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Init) ProtoMessage() {}

func (x *Init) ProtoReflect() protoreflect.Message {
	mi := &file_eventpb_eventpb_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Init.ProtoReflect.Descriptor instead.
func (*Init) Descriptor() ([]byte, []int) {
	return file_eventpb_eventpb_proto_rawDescGZIP(), []int{1}
}

type TimerEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//	*TimerEvent_Delay
	//	*TimerEvent_Repeat
	//	*TimerEvent_GarbageCollect
	Type isTimerEvent_Type `protobuf_oneof:"Type"`
}

func (x *TimerEvent) Reset() {
	*x = TimerEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eventpb_eventpb_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimerEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimerEvent) ProtoMessage() {}

func (x *TimerEvent) ProtoReflect() protoreflect.Message {
	mi := &file_eventpb_eventpb_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimerEvent.ProtoReflect.Descriptor instead.
func (*TimerEvent) Descriptor() ([]byte, []int) {
	return file_eventpb_eventpb_proto_rawDescGZIP(), []int{2}
}

func (m *TimerEvent) GetType() isTimerEvent_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *TimerEvent) GetDelay() *TimerDelay {
	if x, ok := x.GetType().(*TimerEvent_Delay); ok {
		return x.Delay
	}
	return nil
}

func (x *TimerEvent) GetRepeat() *TimerRepeat {
	if x, ok := x.GetType().(*TimerEvent_Repeat); ok {
		return x.Repeat
	}
	return nil
}

func (x *TimerEvent) GetGarbageCollect() *TimerGarbageCollect {
	if x, ok := x.GetType().(*TimerEvent_GarbageCollect); ok {
		return x.GarbageCollect
	}
	return nil
}

type isTimerEvent_Type interface {
	isTimerEvent_Type()
}

type TimerEvent_Delay struct {
	Delay *TimerDelay `protobuf:"bytes,1,opt,name=delay,proto3,oneof"`
}

type TimerEvent_Repeat struct {
	Repeat *TimerRepeat `protobuf:"bytes,2,opt,name=repeat,proto3,oneof"`
}

type TimerEvent_GarbageCollect struct {
	GarbageCollect *TimerGarbageCollect `protobuf:"bytes,3,opt,name=garbage_collect,json=garbageCollect,proto3,oneof"`
}

func (*TimerEvent_Delay) isTimerEvent_Type() {}

func (*TimerEvent_Repeat) isTimerEvent_Type() {}

func (*TimerEvent_GarbageCollect) isTimerEvent_Type() {}

type TimerDelay struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// TODO: The field name must not be `events`, since it conflicts with a package name in the generated code.
	//       This is a bug in the Mir code generator that should be fixed.
	EventsToDelay []*Event `protobuf:"bytes,1,rep,name=events_to_delay,json=eventsToDelay,proto3" json:"events_to_delay,omitempty"`
	Delay         uint64   `protobuf:"varint,2,opt,name=delay,proto3" json:"delay,omitempty"`
}

func (x *TimerDelay) Reset() {
	*x = TimerDelay{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eventpb_eventpb_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimerDelay) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimerDelay) ProtoMessage() {}

func (x *TimerDelay) ProtoReflect() protoreflect.Message {
	mi := &file_eventpb_eventpb_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimerDelay.ProtoReflect.Descriptor instead.
func (*TimerDelay) Descriptor() ([]byte, []int) {
	return file_eventpb_eventpb_proto_rawDescGZIP(), []int{3}
}

func (x *TimerDelay) GetEventsToDelay() []*Event {
	if x != nil {
		return x.EventsToDelay
	}
	return nil
}

func (x *TimerDelay) GetDelay() uint64 {
	if x != nil {
		return x.Delay
	}
	return 0
}

type TimerRepeat struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EventsToRepeat []*Event `protobuf:"bytes,1,rep,name=events_to_repeat,json=eventsToRepeat,proto3" json:"events_to_repeat,omitempty"`
	Delay          uint64   `protobuf:"varint,2,opt,name=delay,proto3" json:"delay,omitempty"`
	RetentionIndex uint64   `protobuf:"varint,3,opt,name=retention_index,json=retentionIndex,proto3" json:"retention_index,omitempty"`
}

func (x *TimerRepeat) Reset() {
	*x = TimerRepeat{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eventpb_eventpb_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimerRepeat) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimerRepeat) ProtoMessage() {}

func (x *TimerRepeat) ProtoReflect() protoreflect.Message {
	mi := &file_eventpb_eventpb_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimerRepeat.ProtoReflect.Descriptor instead.
func (*TimerRepeat) Descriptor() ([]byte, []int) {
	return file_eventpb_eventpb_proto_rawDescGZIP(), []int{4}
}

func (x *TimerRepeat) GetEventsToRepeat() []*Event {
	if x != nil {
		return x.EventsToRepeat
	}
	return nil
}

func (x *TimerRepeat) GetDelay() uint64 {
	if x != nil {
		return x.Delay
	}
	return 0
}

func (x *TimerRepeat) GetRetentionIndex() uint64 {
	if x != nil {
		return x.RetentionIndex
	}
	return 0
}

type TimerGarbageCollect struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RetentionIndex uint64 `protobuf:"varint,1,opt,name=retention_index,json=retentionIndex,proto3" json:"retention_index,omitempty"`
}

func (x *TimerGarbageCollect) Reset() {
	*x = TimerGarbageCollect{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eventpb_eventpb_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimerGarbageCollect) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimerGarbageCollect) ProtoMessage() {}

func (x *TimerGarbageCollect) ProtoReflect() protoreflect.Message {
	mi := &file_eventpb_eventpb_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimerGarbageCollect.ProtoReflect.Descriptor instead.
func (*TimerGarbageCollect) Descriptor() ([]byte, []int) {
	return file_eventpb_eventpb_proto_rawDescGZIP(), []int{5}
}

func (x *TimerGarbageCollect) GetRetentionIndex() uint64 {
	if x != nil {
		return x.RetentionIndex
	}
	return 0
}

var File_eventpb_eventpb_proto protoreflect.FileDescriptor

var file_eventpb_eventpb_proto_rawDesc = []byte{
	0x0a, 0x15, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x70,
	0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x70, 0x62,
	0x1a, 0x11, 0x61, 0x70, 0x70, 0x70, 0x62, 0x2f, 0x61, 0x70, 0x70, 0x70, 0x62, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x11, 0x69, 0x73, 0x73, 0x70, 0x62, 0x2f, 0x69, 0x73, 0x73, 0x70, 0x62,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x11, 0x62, 0x63, 0x62, 0x70, 0x62, 0x2f, 0x62, 0x63,
	0x62, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x6d, 0x65, 0x6d, 0x70, 0x6f,
	0x6f, 0x6c, 0x70, 0x62, 0x2f, 0x6d, 0x65, 0x6d, 0x70, 0x6f, 0x6f, 0x6c, 0x70, 0x62, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x23, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x69, 0x6c, 0x69,
	0x74, 0x79, 0x70, 0x62, 0x2f, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74,
	0x79, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x66, 0x61, 0x63, 0x74, 0x6f,
	0x72, 0x79, 0x70, 0x62, 0x2f, 0x66, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x70, 0x62, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x28, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x69, 0x6c, 0x69,
	0x74, 0x79, 0x70, 0x62, 0x2f, 0x62, 0x61, 0x74, 0x63, 0x68, 0x64, 0x62, 0x70, 0x62, 0x2f, 0x62,
	0x61, 0x74, 0x63, 0x68, 0x64, 0x62, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x23,
	0x62, 0x61, 0x74, 0x63, 0x68, 0x66, 0x65, 0x74, 0x63, 0x68, 0x65, 0x72, 0x70, 0x62, 0x2f, 0x62,
	0x61, 0x74, 0x63, 0x68, 0x66, 0x65, 0x74, 0x63, 0x68, 0x65, 0x72, 0x70, 0x62, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x23, 0x74, 0x68, 0x72, 0x65, 0x73, 0x68, 0x63, 0x72, 0x79, 0x70, 0x74,
	0x6f, 0x70, 0x62, 0x2f, 0x74, 0x68, 0x72, 0x65, 0x73, 0x68, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f,
	0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x70, 0x69, 0x6e, 0x67, 0x70, 0x6f,
	0x6e, 0x67, 0x70, 0x62, 0x2f, 0x70, 0x69, 0x6e, 0x67, 0x70, 0x6f, 0x6e, 0x67, 0x70, 0x62, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x70, 0x62, 0x2f, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x70, 0x62,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x72, 0x70,
	0x62, 0x2f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x72, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x17, 0x68, 0x61, 0x73, 0x68, 0x65, 0x72, 0x70, 0x62, 0x2f, 0x68, 0x61, 0x73, 0x68,
	0x65, 0x72, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x63, 0x72, 0x79, 0x70,
	0x74, 0x6f, 0x70, 0x62, 0x2f, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x70, 0x62, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1d, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x70, 0x62,
	0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x1c, 0x6d, 0x69, 0x72, 0x2f, 0x63, 0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e, 0x5f,
	0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xc7, 0x08, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x57, 0x0a, 0x0b, 0x64, 0x65,
	0x73, 0x74, 0x5f, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42,
	0x36, 0x82, 0xa6, 0x1d, 0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x4d,
	0x6f, 0x64, 0x75, 0x6c, 0x65, 0x49, 0x44, 0x52, 0x0a, 0x64, 0x65, 0x73, 0x74, 0x4d, 0x6f, 0x64,
	0x75, 0x6c, 0x65, 0x12, 0x23, 0x0a, 0x04, 0x69, 0x6e, 0x69, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x0d, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x49, 0x6e, 0x69, 0x74,
	0x48, 0x00, 0x52, 0x04, 0x69, 0x6e, 0x69, 0x74, 0x12, 0x2b, 0x0a, 0x05, 0x74, 0x69, 0x6d, 0x65,
	0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x70,
	0x62, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x72, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x05,
	0x74, 0x69, 0x6d, 0x65, 0x72, 0x12, 0x29, 0x0a, 0x06, 0x68, 0x61, 0x73, 0x68, 0x65, 0x72, 0x18,
	0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x68, 0x61, 0x73, 0x68, 0x65, 0x72, 0x70, 0x62,
	0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x06, 0x68, 0x61, 0x73, 0x68, 0x65, 0x72,
	0x12, 0x20, 0x0a, 0x03, 0x62, 0x63, 0x62, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e,
	0x62, 0x63, 0x62, 0x70, 0x62, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x03, 0x62,
	0x63, 0x62, 0x12, 0x2c, 0x0a, 0x07, 0x6d, 0x65, 0x6d, 0x70, 0x6f, 0x6f, 0x6c, 0x18, 0x0c, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x6d, 0x65, 0x6d, 0x70, 0x6f, 0x6f, 0x6c, 0x70, 0x62, 0x2e,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x07, 0x6d, 0x65, 0x6d, 0x70, 0x6f, 0x6f, 0x6c,
	0x12, 0x3b, 0x0a, 0x0c, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79,
	0x18, 0x0d, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62,
	0x69, 0x6c, 0x69, 0x74, 0x79, 0x70, 0x62, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52,
	0x0c, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x12, 0x2d, 0x0a,
	0x08, 0x62, 0x61, 0x74, 0x63, 0x68, 0x5f, 0x64, 0x62, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x10, 0x2e, 0x62, 0x61, 0x74, 0x63, 0x68, 0x64, 0x62, 0x70, 0x62, 0x2e, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x48, 0x00, 0x52, 0x07, 0x62, 0x61, 0x74, 0x63, 0x68, 0x44, 0x62, 0x12, 0x3c, 0x0a, 0x0d,
	0x62, 0x61, 0x74, 0x63, 0x68, 0x5f, 0x66, 0x65, 0x74, 0x63, 0x68, 0x65, 0x72, 0x18, 0x0f, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x62, 0x61, 0x74, 0x63, 0x68, 0x66, 0x65, 0x74, 0x63, 0x68,
	0x65, 0x72, 0x70, 0x62, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x0c, 0x62, 0x61,
	0x74, 0x63, 0x68, 0x46, 0x65, 0x74, 0x63, 0x68, 0x65, 0x72, 0x12, 0x3c, 0x0a, 0x0d, 0x74, 0x68,
	0x72, 0x65, 0x73, 0x68, 0x5f, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x18, 0x10, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x15, 0x2e, 0x74, 0x68, 0x72, 0x65, 0x73, 0x68, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f,
	0x70, 0x62, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x0c, 0x74, 0x68, 0x72, 0x65,
	0x73, 0x68, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x12, 0x35, 0x0a, 0x0a, 0x63, 0x68, 0x65, 0x63,
	0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x11, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x63,
	0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x48, 0x00, 0x52, 0x0a, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x12,
	0x2c, 0x0a, 0x07, 0x66, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x18, 0x12, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x10, 0x2e, 0x66, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x70, 0x62, 0x2e, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x48, 0x00, 0x52, 0x07, 0x66, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x12, 0x20, 0x0a,
	0x03, 0x69, 0x73, 0x73, 0x18, 0x13, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x69, 0x73, 0x73,
	0x70, 0x62, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x03, 0x69, 0x73, 0x73, 0x12,
	0x2c, 0x0a, 0x07, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x72, 0x18, 0x14, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x10, 0x2e, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x72, 0x70, 0x62, 0x2e, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x48, 0x00, 0x52, 0x07, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x72, 0x12, 0x29, 0x0a,
	0x06, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x18, 0x15, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e,
	0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x70, 0x62, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x48, 0x00,
	0x52, 0x06, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x12, 0x20, 0x0a, 0x03, 0x61, 0x70, 0x70, 0x18,
	0x16, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x61, 0x70, 0x70, 0x70, 0x62, 0x2e, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x03, 0x61, 0x70, 0x70, 0x12, 0x32, 0x0a, 0x09, 0x74, 0x72,
	0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x17, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x48, 0x00, 0x52, 0x09, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x31,
	0x0a, 0x09, 0x70, 0x69, 0x6e, 0x67, 0x5f, 0x70, 0x6f, 0x6e, 0x67, 0x18, 0xc8, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x11, 0x2e, 0x70, 0x69, 0x6e, 0x67, 0x70, 0x6f, 0x6e, 0x67, 0x70, 0x62, 0x2e,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x08, 0x70, 0x69, 0x6e, 0x67, 0x50, 0x6f, 0x6e,
	0x67, 0x12, 0x45, 0x0a, 0x0d, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x72, 0x69,
	0x6e, 0x67, 0x18, 0xad, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x69,
	0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x0d, 0x74, 0x65, 0x73, 0x74, 0x69,
	0x6e, 0x67, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x41, 0x0a, 0x0b, 0x74, 0x65, 0x73, 0x74,
	0x69, 0x6e, 0x67, 0x55, 0x69, 0x6e, 0x74, 0x18, 0xae, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x55, 0x49, 0x6e, 0x74, 0x36, 0x34, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x48, 0x00, 0x52, 0x0b,
	0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x55, 0x69, 0x6e, 0x74, 0x12, 0x29, 0x0a, 0x04, 0x6e,
	0x65, 0x78, 0x74, 0x18, 0x90, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x65, 0x76, 0x65,
	0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x42, 0x04, 0x90, 0xa6, 0x1d, 0x01,
	0x52, 0x04, 0x6e, 0x65, 0x78, 0x74, 0x3a, 0x04, 0x88, 0xa6, 0x1d, 0x01, 0x42, 0x0c, 0x0a, 0x04,
	0x74, 0x79, 0x70, 0x65, 0x12, 0x04, 0x80, 0xa6, 0x1d, 0x01, 0x22, 0x0c, 0x0a, 0x04, 0x49, 0x6e,
	0x69, 0x74, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x22, 0xc6, 0x01, 0x0a, 0x0a, 0x54, 0x69, 0x6d,
	0x65, 0x72, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x2b, 0x0a, 0x05, 0x64, 0x65, 0x6c, 0x61, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x70, 0x62,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x72, 0x44, 0x65, 0x6c, 0x61, 0x79, 0x48, 0x00, 0x52, 0x05, 0x64,
	0x65, 0x6c, 0x61, 0x79, 0x12, 0x2e, 0x0a, 0x06, 0x72, 0x65, 0x70, 0x65, 0x61, 0x74, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x54,
	0x69, 0x6d, 0x65, 0x72, 0x52, 0x65, 0x70, 0x65, 0x61, 0x74, 0x48, 0x00, 0x52, 0x06, 0x72, 0x65,
	0x70, 0x65, 0x61, 0x74, 0x12, 0x47, 0x0a, 0x0f, 0x67, 0x61, 0x72, 0x62, 0x61, 0x67, 0x65, 0x5f,
	0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x72, 0x47, 0x61, 0x72,
	0x62, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x48, 0x00, 0x52, 0x0e, 0x67,
	0x61, 0x72, 0x62, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x3a, 0x04, 0x90,
	0xa6, 0x1d, 0x01, 0x42, 0x0c, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x04, 0x80, 0xa6, 0x1d,
	0x01, 0x22, 0x9c, 0x01, 0x0a, 0x0a, 0x54, 0x69, 0x6d, 0x65, 0x72, 0x44, 0x65, 0x6c, 0x61, 0x79,
	0x12, 0x36, 0x0a, 0x0f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x5f, 0x74, 0x6f, 0x5f, 0x64, 0x65,
	0x6c, 0x61, 0x79, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x70, 0x62, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x0d, 0x65, 0x76, 0x65, 0x6e, 0x74,
	0x73, 0x54, 0x6f, 0x44, 0x65, 0x6c, 0x61, 0x79, 0x12, 0x50, 0x0a, 0x05, 0x64, 0x65, 0x6c, 0x61,
	0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x42, 0x3a, 0x82, 0xa6, 0x1d, 0x36, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e,
	0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67,
	0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x44, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x05, 0x64, 0x65, 0x6c, 0x61, 0x79, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01,
	0x22, 0x86, 0x02, 0x0a, 0x0b, 0x54, 0x69, 0x6d, 0x65, 0x72, 0x52, 0x65, 0x70, 0x65, 0x61, 0x74,
	0x12, 0x38, 0x0a, 0x10, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x5f, 0x74, 0x6f, 0x5f, 0x72, 0x65,
	0x70, 0x65, 0x61, 0x74, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x65, 0x76, 0x65,
	0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x0e, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x73, 0x54, 0x6f, 0x52, 0x65, 0x70, 0x65, 0x61, 0x74, 0x12, 0x50, 0x0a, 0x05, 0x64, 0x65,
	0x6c, 0x61, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x42, 0x3a, 0x82, 0xa6, 0x1d, 0x36, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f,
	0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70,
	0x6b, 0x67, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x44, 0x75, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x05, 0x64, 0x65, 0x6c, 0x61, 0x79, 0x12, 0x65, 0x0a, 0x0f,
	0x72, 0x65, 0x74, 0x65, 0x6e, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x04, 0x42, 0x3c, 0x82, 0xa6, 0x1d, 0x38, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70,
	0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x2e, 0x52, 0x65, 0x74, 0x65, 0x6e, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e,
	0x64, 0x65, 0x78, 0x52, 0x0e, 0x72, 0x65, 0x74, 0x65, 0x6e, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e,
	0x64, 0x65, 0x78, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x22, 0x82, 0x01, 0x0a, 0x13, 0x54, 0x69,
	0x6d, 0x65, 0x72, 0x47, 0x61, 0x72, 0x62, 0x61, 0x67, 0x65, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63,
	0x74, 0x12, 0x65, 0x0a, 0x0f, 0x72, 0x65, 0x74, 0x65, 0x6e, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69,
	0x6e, 0x64, 0x65, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x42, 0x3c, 0x82, 0xa6, 0x1d, 0x38,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63,
	0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f,
	0x70, 0x6b, 0x67, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x52, 0x65, 0x74, 0x65, 0x6e, 0x74,
	0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x52, 0x0e, 0x72, 0x65, 0x74, 0x65, 0x6e, 0x74,
	0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x42, 0x30,
	0x5a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c,
	0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69,
	0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x62, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x70, 0x62,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_eventpb_eventpb_proto_rawDescOnce sync.Once
	file_eventpb_eventpb_proto_rawDescData = file_eventpb_eventpb_proto_rawDesc
)

func file_eventpb_eventpb_proto_rawDescGZIP() []byte {
	file_eventpb_eventpb_proto_rawDescOnce.Do(func() {
		file_eventpb_eventpb_proto_rawDescData = protoimpl.X.CompressGZIP(file_eventpb_eventpb_proto_rawDescData)
	})
	return file_eventpb_eventpb_proto_rawDescData
}

var file_eventpb_eventpb_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_eventpb_eventpb_proto_goTypes = []interface{}{
	(*Event)(nil),                  // 0: eventpb.Event
	(*Init)(nil),                   // 1: eventpb.Init
	(*TimerEvent)(nil),             // 2: eventpb.TimerEvent
	(*TimerDelay)(nil),             // 3: eventpb.TimerDelay
	(*TimerRepeat)(nil),            // 4: eventpb.TimerRepeat
	(*TimerGarbageCollect)(nil),    // 5: eventpb.TimerGarbageCollect
	(*hasherpb.Event)(nil),         // 6: hasherpb.Event
	(*bcbpb.Event)(nil),            // 7: bcbpb.Event
	(*mempoolpb.Event)(nil),        // 8: mempoolpb.Event
	(*availabilitypb.Event)(nil),   // 9: availabilitypb.Event
	(*batchdbpb.Event)(nil),        // 10: batchdbpb.Event
	(*batchfetcherpb.Event)(nil),   // 11: batchfetcherpb.Event
	(*threshcryptopb.Event)(nil),   // 12: threshcryptopb.Event
	(*checkpointpb.Event)(nil),     // 13: checkpointpb.Event
	(*factorypb.Event)(nil),        // 14: factorypb.Event
	(*isspb.Event)(nil),            // 15: isspb.Event
	(*ordererpb.Event)(nil),        // 16: ordererpb.Event
	(*cryptopb.Event)(nil),         // 17: cryptopb.Event
	(*apppb.Event)(nil),            // 18: apppb.Event
	(*transportpb.Event)(nil),      // 19: transportpb.Event
	(*pingpongpb.Event)(nil),       // 20: pingpongpb.Event
	(*wrapperspb.StringValue)(nil), // 21: google.protobuf.StringValue
	(*wrapperspb.UInt64Value)(nil), // 22: google.protobuf.UInt64Value
}
var file_eventpb_eventpb_proto_depIdxs = []int32{
	1,  // 0: eventpb.Event.init:type_name -> eventpb.Init
	2,  // 1: eventpb.Event.timer:type_name -> eventpb.TimerEvent
	6,  // 2: eventpb.Event.hasher:type_name -> hasherpb.Event
	7,  // 3: eventpb.Event.bcb:type_name -> bcbpb.Event
	8,  // 4: eventpb.Event.mempool:type_name -> mempoolpb.Event
	9,  // 5: eventpb.Event.availability:type_name -> availabilitypb.Event
	10, // 6: eventpb.Event.batch_db:type_name -> batchdbpb.Event
	11, // 7: eventpb.Event.batch_fetcher:type_name -> batchfetcherpb.Event
	12, // 8: eventpb.Event.thresh_crypto:type_name -> threshcryptopb.Event
	13, // 9: eventpb.Event.checkpoint:type_name -> checkpointpb.Event
	14, // 10: eventpb.Event.factory:type_name -> factorypb.Event
	15, // 11: eventpb.Event.iss:type_name -> isspb.Event
	16, // 12: eventpb.Event.orderer:type_name -> ordererpb.Event
	17, // 13: eventpb.Event.crypto:type_name -> cryptopb.Event
	18, // 14: eventpb.Event.app:type_name -> apppb.Event
	19, // 15: eventpb.Event.transport:type_name -> transportpb.Event
	20, // 16: eventpb.Event.ping_pong:type_name -> pingpongpb.Event
	21, // 17: eventpb.Event.testingString:type_name -> google.protobuf.StringValue
	22, // 18: eventpb.Event.testingUint:type_name -> google.protobuf.UInt64Value
	0,  // 19: eventpb.Event.next:type_name -> eventpb.Event
	3,  // 20: eventpb.TimerEvent.delay:type_name -> eventpb.TimerDelay
	4,  // 21: eventpb.TimerEvent.repeat:type_name -> eventpb.TimerRepeat
	5,  // 22: eventpb.TimerEvent.garbage_collect:type_name -> eventpb.TimerGarbageCollect
	0,  // 23: eventpb.TimerDelay.events_to_delay:type_name -> eventpb.Event
	0,  // 24: eventpb.TimerRepeat.events_to_repeat:type_name -> eventpb.Event
	25, // [25:25] is the sub-list for method output_type
	25, // [25:25] is the sub-list for method input_type
	25, // [25:25] is the sub-list for extension type_name
	25, // [25:25] is the sub-list for extension extendee
	0,  // [0:25] is the sub-list for field type_name
}

func init() { file_eventpb_eventpb_proto_init() }
func file_eventpb_eventpb_proto_init() {
	if File_eventpb_eventpb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_eventpb_eventpb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_eventpb_eventpb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Init); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_eventpb_eventpb_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimerEvent); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_eventpb_eventpb_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimerDelay); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_eventpb_eventpb_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimerRepeat); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_eventpb_eventpb_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimerGarbageCollect); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_eventpb_eventpb_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Event_Init)(nil),
		(*Event_Timer)(nil),
		(*Event_Hasher)(nil),
		(*Event_Bcb)(nil),
		(*Event_Mempool)(nil),
		(*Event_Availability)(nil),
		(*Event_BatchDb)(nil),
		(*Event_BatchFetcher)(nil),
		(*Event_ThreshCrypto)(nil),
		(*Event_Checkpoint)(nil),
		(*Event_Factory)(nil),
		(*Event_Iss)(nil),
		(*Event_Orderer)(nil),
		(*Event_Crypto)(nil),
		(*Event_App)(nil),
		(*Event_Transport)(nil),
		(*Event_PingPong)(nil),
		(*Event_TestingString)(nil),
		(*Event_TestingUint)(nil),
	}
	file_eventpb_eventpb_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*TimerEvent_Delay)(nil),
		(*TimerEvent_Repeat)(nil),
		(*TimerEvent_GarbageCollect)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_eventpb_eventpb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_eventpb_eventpb_proto_goTypes,
		DependencyIndexes: file_eventpb_eventpb_proto_depIdxs,
		MessageInfos:      file_eventpb_eventpb_proto_msgTypes,
	}.Build()
	File_eventpb_eventpb_proto = out.File
	file_eventpb_eventpb_proto_rawDesc = nil
	file_eventpb_eventpb_proto_goTypes = nil
	file_eventpb_eventpb_proto_depIdxs = nil
}
