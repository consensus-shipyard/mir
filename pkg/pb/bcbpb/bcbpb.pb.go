// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.14.0
// source: bcbpb/bcbpb.proto

package bcbpb

import (
	_ "github.com/filecoin-project/mir/pkg/pb/mir"
	_ "github.com/filecoin-project/mir/pkg/pb/net"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Event struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//	*Event_Request
	//	*Event_Deliver
	Type isEvent_Type `protobuf_oneof:"type"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bcbpb_bcbpb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_bcbpb_bcbpb_proto_msgTypes[0]
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
	return file_bcbpb_bcbpb_proto_rawDescGZIP(), []int{0}
}

func (m *Event) GetType() isEvent_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Event) GetRequest() *BroadcastRequest {
	if x, ok := x.GetType().(*Event_Request); ok {
		return x.Request
	}
	return nil
}

func (x *Event) GetDeliver() *Deliver {
	if x, ok := x.GetType().(*Event_Deliver); ok {
		return x.Deliver
	}
	return nil
}

type isEvent_Type interface {
	isEvent_Type()
}

type Event_Request struct {
	Request *BroadcastRequest `protobuf:"bytes,1,opt,name=request,proto3,oneof"`
}

type Event_Deliver struct {
	Deliver *Deliver `protobuf:"bytes,2,opt,name=deliver,proto3,oneof"`
}

func (*Event_Request) isEvent_Type() {}

func (*Event_Deliver) isEvent_Type() {}

type BroadcastRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *BroadcastRequest) Reset() {
	*x = BroadcastRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bcbpb_bcbpb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BroadcastRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BroadcastRequest) ProtoMessage() {}

func (x *BroadcastRequest) ProtoReflect() protoreflect.Message {
	mi := &file_bcbpb_bcbpb_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BroadcastRequest.ProtoReflect.Descriptor instead.
func (*BroadcastRequest) Descriptor() ([]byte, []int) {
	return file_bcbpb_bcbpb_proto_rawDescGZIP(), []int{1}
}

func (x *BroadcastRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type Deliver struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *Deliver) Reset() {
	*x = Deliver{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bcbpb_bcbpb_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Deliver) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Deliver) ProtoMessage() {}

func (x *Deliver) ProtoReflect() protoreflect.Message {
	mi := &file_bcbpb_bcbpb_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Deliver.ProtoReflect.Descriptor instead.
func (*Deliver) Descriptor() ([]byte, []int) {
	return file_bcbpb_bcbpb_proto_rawDescGZIP(), []int{2}
}

func (x *Deliver) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//	*Message_StartMessage
	//	*Message_EchoMessage
	//	*Message_FinalMessage
	Type isMessage_Type `protobuf_oneof:"type"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bcbpb_bcbpb_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_bcbpb_bcbpb_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_bcbpb_bcbpb_proto_rawDescGZIP(), []int{3}
}

func (m *Message) GetType() isMessage_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Message) GetStartMessage() *StartMessage {
	if x, ok := x.GetType().(*Message_StartMessage); ok {
		return x.StartMessage
	}
	return nil
}

func (x *Message) GetEchoMessage() *EchoMessage {
	if x, ok := x.GetType().(*Message_EchoMessage); ok {
		return x.EchoMessage
	}
	return nil
}

func (x *Message) GetFinalMessage() *FinalMessage {
	if x, ok := x.GetType().(*Message_FinalMessage); ok {
		return x.FinalMessage
	}
	return nil
}

type isMessage_Type interface {
	isMessage_Type()
}

type Message_StartMessage struct {
	StartMessage *StartMessage `protobuf:"bytes,1,opt,name=start_message,json=startMessage,proto3,oneof"`
}

type Message_EchoMessage struct {
	EchoMessage *EchoMessage `protobuf:"bytes,2,opt,name=echo_message,json=echoMessage,proto3,oneof"`
}

type Message_FinalMessage struct {
	FinalMessage *FinalMessage `protobuf:"bytes,3,opt,name=final_message,json=finalMessage,proto3,oneof"`
}

func (*Message_StartMessage) isMessage_Type() {}

func (*Message_EchoMessage) isMessage_Type() {}

func (*Message_FinalMessage) isMessage_Type() {}

type StartMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *StartMessage) Reset() {
	*x = StartMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bcbpb_bcbpb_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StartMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StartMessage) ProtoMessage() {}

func (x *StartMessage) ProtoReflect() protoreflect.Message {
	mi := &file_bcbpb_bcbpb_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StartMessage.ProtoReflect.Descriptor instead.
func (*StartMessage) Descriptor() ([]byte, []int) {
	return file_bcbpb_bcbpb_proto_rawDescGZIP(), []int{4}
}

func (x *StartMessage) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type EchoMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Signature []byte `protobuf:"bytes,1,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (x *EchoMessage) Reset() {
	*x = EchoMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bcbpb_bcbpb_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EchoMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EchoMessage) ProtoMessage() {}

func (x *EchoMessage) ProtoReflect() protoreflect.Message {
	mi := &file_bcbpb_bcbpb_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EchoMessage.ProtoReflect.Descriptor instead.
func (*EchoMessage) Descriptor() ([]byte, []int) {
	return file_bcbpb_bcbpb_proto_rawDescGZIP(), []int{5}
}

func (x *EchoMessage) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

type FinalMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data       []byte   `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	Signers    []string `protobuf:"bytes,2,rep,name=signers,proto3" json:"signers,omitempty"`
	Signatures [][]byte `protobuf:"bytes,3,rep,name=signatures,proto3" json:"signatures,omitempty"`
}

func (x *FinalMessage) Reset() {
	*x = FinalMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_bcbpb_bcbpb_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FinalMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FinalMessage) ProtoMessage() {}

func (x *FinalMessage) ProtoReflect() protoreflect.Message {
	mi := &file_bcbpb_bcbpb_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FinalMessage.ProtoReflect.Descriptor instead.
func (*FinalMessage) Descriptor() ([]byte, []int) {
	return file_bcbpb_bcbpb_proto_rawDescGZIP(), []int{6}
}

func (x *FinalMessage) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *FinalMessage) GetSigners() []string {
	if x != nil {
		return x.Signers
	}
	return nil
}

func (x *FinalMessage) GetSignatures() [][]byte {
	if x != nil {
		return x.Signatures
	}
	return nil
}

var File_bcbpb_bcbpb_proto protoreflect.FileDescriptor

var file_bcbpb_bcbpb_proto_rawDesc = []byte{
	0x0a, 0x11, 0x62, 0x63, 0x62, 0x70, 0x62, 0x2f, 0x62, 0x63, 0x62, 0x70, 0x62, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x05, 0x62, 0x63, 0x62, 0x70, 0x62, 0x1a, 0x1c, 0x6d, 0x69, 0x72, 0x2f,
	0x63, 0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e, 0x5f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6e, 0x65, 0x74, 0x2f, 0x63, 0x6f,
	0x64, 0x65, 0x67, 0x65, 0x6e, 0x5f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x7c, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12,
	0x33, 0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x17, 0x2e, 0x62, 0x63, 0x62, 0x70, 0x62, 0x2e, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61,
	0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x07, 0x72, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x2a, 0x0a, 0x07, 0x64, 0x65, 0x6c, 0x69, 0x76, 0x65, 0x72, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x62, 0x63, 0x62, 0x70, 0x62, 0x2e, 0x44, 0x65,
	0x6c, 0x69, 0x76, 0x65, 0x72, 0x48, 0x00, 0x52, 0x07, 0x64, 0x65, 0x6c, 0x69, 0x76, 0x65, 0x72,
	0x3a, 0x04, 0x90, 0xa6, 0x1d, 0x01, 0x42, 0x0c, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x04,
	0x80, 0xa6, 0x1d, 0x01, 0x22, 0x2c, 0x0a, 0x10, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61, 0x73,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x3a, 0x04, 0x98, 0xa6,
	0x1d, 0x01, 0x22, 0x23, 0x0a, 0x07, 0x44, 0x65, 0x6c, 0x69, 0x76, 0x65, 0x72, 0x12, 0x12, 0x0a,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74,
	0x61, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x22, 0xce, 0x01, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x12, 0x3a, 0x0a, 0x0d, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x62, 0x63, 0x62,
	0x70, 0x62, 0x2e, 0x53, 0x74, 0x61, 0x72, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48,
	0x00, 0x52, 0x0c, 0x73, 0x74, 0x61, 0x72, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12,
	0x37, 0x0a, 0x0c, 0x65, 0x63, 0x68, 0x6f, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x62, 0x63, 0x62, 0x70, 0x62, 0x2e, 0x45, 0x63,
	0x68, 0x6f, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48, 0x00, 0x52, 0x0b, 0x65, 0x63, 0x68,
	0x6f, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x3a, 0x0a, 0x0d, 0x66, 0x69, 0x6e, 0x61,
	0x6c, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x13, 0x2e, 0x62, 0x63, 0x62, 0x70, 0x62, 0x2e, 0x46, 0x69, 0x6e, 0x61, 0x6c, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x48, 0x00, 0x52, 0x0c, 0x66, 0x69, 0x6e, 0x61, 0x6c, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x3a, 0x04, 0xc8, 0xe4, 0x1d, 0x01, 0x42, 0x0c, 0x0a, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x12, 0x04, 0xc8, 0xe4, 0x1d, 0x01, 0x22, 0x28, 0x0a, 0x0c, 0x53, 0x74, 0x61, 0x72,
	0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x3a, 0x04, 0xd0, 0xe4,
	0x1d, 0x01, 0x22, 0x31, 0x0a, 0x0b, 0x45, 0x63, 0x68, 0x6f, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x3a,
	0x04, 0xd0, 0xe4, 0x1d, 0x01, 0x22, 0x98, 0x01, 0x0a, 0x0c, 0x46, 0x69, 0x6e, 0x61, 0x6c, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x4e, 0x0a, 0x07, 0x73, 0x69,
	0x67, 0x6e, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x42, 0x34, 0x82, 0xa6, 0x1d,
	0x30, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65,
	0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72,
	0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x49,
	0x44, 0x52, 0x07, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x72, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x69,
	0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x0a,
	0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x3a, 0x04, 0xd0, 0xe4, 0x1d, 0x01,
	0x42, 0x2e, 0x5a, 0x2c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66,
	0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f,
	0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x62, 0x2f, 0x62, 0x63, 0x62, 0x70, 0x62,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_bcbpb_bcbpb_proto_rawDescOnce sync.Once
	file_bcbpb_bcbpb_proto_rawDescData = file_bcbpb_bcbpb_proto_rawDesc
)

func file_bcbpb_bcbpb_proto_rawDescGZIP() []byte {
	file_bcbpb_bcbpb_proto_rawDescOnce.Do(func() {
		file_bcbpb_bcbpb_proto_rawDescData = protoimpl.X.CompressGZIP(file_bcbpb_bcbpb_proto_rawDescData)
	})
	return file_bcbpb_bcbpb_proto_rawDescData
}

var file_bcbpb_bcbpb_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_bcbpb_bcbpb_proto_goTypes = []interface{}{
	(*Event)(nil),            // 0: bcbpb.Event
	(*BroadcastRequest)(nil), // 1: bcbpb.BroadcastRequest
	(*Deliver)(nil),          // 2: bcbpb.Deliver
	(*Message)(nil),          // 3: bcbpb.Message
	(*StartMessage)(nil),     // 4: bcbpb.StartMessage
	(*EchoMessage)(nil),      // 5: bcbpb.EchoMessage
	(*FinalMessage)(nil),     // 6: bcbpb.FinalMessage
}
var file_bcbpb_bcbpb_proto_depIdxs = []int32{
	1, // 0: bcbpb.Event.request:type_name -> bcbpb.BroadcastRequest
	2, // 1: bcbpb.Event.deliver:type_name -> bcbpb.Deliver
	4, // 2: bcbpb.Message.start_message:type_name -> bcbpb.StartMessage
	5, // 3: bcbpb.Message.echo_message:type_name -> bcbpb.EchoMessage
	6, // 4: bcbpb.Message.final_message:type_name -> bcbpb.FinalMessage
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_bcbpb_bcbpb_proto_init() }
func file_bcbpb_bcbpb_proto_init() {
	if File_bcbpb_bcbpb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_bcbpb_bcbpb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_bcbpb_bcbpb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BroadcastRequest); i {
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
		file_bcbpb_bcbpb_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Deliver); i {
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
		file_bcbpb_bcbpb_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Message); i {
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
		file_bcbpb_bcbpb_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StartMessage); i {
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
		file_bcbpb_bcbpb_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EchoMessage); i {
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
		file_bcbpb_bcbpb_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FinalMessage); i {
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
	file_bcbpb_bcbpb_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Event_Request)(nil),
		(*Event_Deliver)(nil),
	}
	file_bcbpb_bcbpb_proto_msgTypes[3].OneofWrappers = []interface{}{
		(*Message_StartMessage)(nil),
		(*Message_EchoMessage)(nil),
		(*Message_FinalMessage)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_bcbpb_bcbpb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_bcbpb_bcbpb_proto_goTypes,
		DependencyIndexes: file_bcbpb_bcbpb_proto_depIdxs,
		MessageInfos:      file_bcbpb_bcbpb_proto_msgTypes,
	}.Build()
	File_bcbpb_bcbpb_proto = out.File
	file_bcbpb_bcbpb_proto_rawDesc = nil
	file_bcbpb_bcbpb_proto_goTypes = nil
	file_bcbpb_bcbpb_proto_depIdxs = nil
}
