// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.12.4
// source: accountabilitypb/accountabilitypb.proto

package accountabilitypb

import (
	_ "github.com/filecoin-project/mir/pkg/pb/mir"
	_ "github.com/filecoin-project/mir/pkg/pb/net"
	trantorpb "github.com/filecoin-project/mir/pkg/pb/trantorpb"
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
	//	*Event_Predecided
	//	*Event_Decided
	//	*Event_Poms
	Type isEvent_Type `protobuf_oneof:"type"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[0]
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
	return file_accountabilitypb_accountabilitypb_proto_rawDescGZIP(), []int{0}
}

func (m *Event) GetType() isEvent_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Event) GetPredecided() *Predecided {
	if x, ok := x.GetType().(*Event_Predecided); ok {
		return x.Predecided
	}
	return nil
}

func (x *Event) GetDecided() *Decided {
	if x, ok := x.GetType().(*Event_Decided); ok {
		return x.Decided
	}
	return nil
}

func (x *Event) GetPoms() *PoMs {
	if x, ok := x.GetType().(*Event_Poms); ok {
		return x.Poms
	}
	return nil
}

type isEvent_Type interface {
	isEvent_Type()
}

type Event_Predecided struct {
	Predecided *Predecided `protobuf:"bytes,1,opt,name=predecided,proto3,oneof"`
}

type Event_Decided struct {
	Decided *Decided `protobuf:"bytes,2,opt,name=decided,proto3,oneof"`
}

type Event_Poms struct {
	Poms *PoMs `protobuf:"bytes,3,opt,name=poms,proto3,oneof"`
}

func (*Event_Predecided) isEvent_Type() {}

func (*Event_Decided) isEvent_Type() {}

func (*Event_Poms) isEvent_Type() {}

type Predecided struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *Predecided) Reset() {
	*x = Predecided{}
	if protoimpl.UnsafeEnabled {
		mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Predecided) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Predecided) ProtoMessage() {}

func (x *Predecided) ProtoReflect() protoreflect.Message {
	mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Predecided.ProtoReflect.Descriptor instead.
func (*Predecided) Descriptor() ([]byte, []int) {
	return file_accountabilitypb_accountabilitypb_proto_rawDescGZIP(), []int{1}
}

func (x *Predecided) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type Decided struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *Decided) Reset() {
	*x = Decided{}
	if protoimpl.UnsafeEnabled {
		mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Decided) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Decided) ProtoMessage() {}

func (x *Decided) ProtoReflect() protoreflect.Message {
	mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Decided.ProtoReflect.Descriptor instead.
func (*Decided) Descriptor() ([]byte, []int) {
	return file_accountabilitypb_accountabilitypb_proto_rawDescGZIP(), []int{2}
}

func (x *Decided) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type PoM struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeId           string             `protobuf:"bytes,1,opt,name=node_id,json=nodeId,proto3" json:"node_id,omitempty"`
	ConflictingMsg_1 *SignedPredecision `protobuf:"bytes,2,opt,name=conflicting_msg_1,json=conflictingMsg1,proto3" json:"conflicting_msg_1,omitempty"`
	ConflictingMsg_2 *SignedPredecision `protobuf:"bytes,3,opt,name=conflicting_msg_2,json=conflictingMsg2,proto3" json:"conflicting_msg_2,omitempty"`
}

func (x *PoM) Reset() {
	*x = PoM{}
	if protoimpl.UnsafeEnabled {
		mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PoM) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PoM) ProtoMessage() {}

func (x *PoM) ProtoReflect() protoreflect.Message {
	mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PoM.ProtoReflect.Descriptor instead.
func (*PoM) Descriptor() ([]byte, []int) {
	return file_accountabilitypb_accountabilitypb_proto_rawDescGZIP(), []int{3}
}

func (x *PoM) GetNodeId() string {
	if x != nil {
		return x.NodeId
	}
	return ""
}

func (x *PoM) GetConflictingMsg_1() *SignedPredecision {
	if x != nil {
		return x.ConflictingMsg_1
	}
	return nil
}

func (x *PoM) GetConflictingMsg_2() *SignedPredecision {
	if x != nil {
		return x.ConflictingMsg_2
	}
	return nil
}

type LightCertificate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *LightCertificate) Reset() {
	*x = LightCertificate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LightCertificate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LightCertificate) ProtoMessage() {}

func (x *LightCertificate) ProtoReflect() protoreflect.Message {
	mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LightCertificate.ProtoReflect.Descriptor instead.
func (*LightCertificate) Descriptor() ([]byte, []int) {
	return file_accountabilitypb_accountabilitypb_proto_rawDescGZIP(), []int{4}
}

func (x *LightCertificate) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type PoMs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Poms []*PoM `protobuf:"bytes,1,rep,name=poms,proto3" json:"poms,omitempty"`
}

func (x *PoMs) Reset() {
	*x = PoMs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PoMs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PoMs) ProtoMessage() {}

func (x *PoMs) ProtoReflect() protoreflect.Message {
	mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PoMs.ProtoReflect.Descriptor instead.
func (*PoMs) Descriptor() ([]byte, []int) {
	return file_accountabilitypb_accountabilitypb_proto_rawDescGZIP(), []int{5}
}

func (x *PoMs) GetPoms() []*PoM {
	if x != nil {
		return x.Poms
	}
	return nil
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//	*Message_SignedPredecision
	//	*Message_Certificate
	//	*Message_Poms
	//	*Message_LightCertificate
	Type isMessage_Type `protobuf_oneof:"type"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[6]
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
	return file_accountabilitypb_accountabilitypb_proto_rawDescGZIP(), []int{6}
}

func (m *Message) GetType() isMessage_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Message) GetSignedPredecision() *SignedPredecision {
	if x, ok := x.GetType().(*Message_SignedPredecision); ok {
		return x.SignedPredecision
	}
	return nil
}

func (x *Message) GetCertificate() *FullCertificate {
	if x, ok := x.GetType().(*Message_Certificate); ok {
		return x.Certificate
	}
	return nil
}

func (x *Message) GetPoms() *PoMs {
	if x, ok := x.GetType().(*Message_Poms); ok {
		return x.Poms
	}
	return nil
}

func (x *Message) GetLightCertificate() *LightCertificate {
	if x, ok := x.GetType().(*Message_LightCertificate); ok {
		return x.LightCertificate
	}
	return nil
}

type isMessage_Type interface {
	isMessage_Type()
}

type Message_SignedPredecision struct {
	SignedPredecision *SignedPredecision `protobuf:"bytes,1,opt,name=signed_predecision,json=signedPredecision,proto3,oneof"`
}

type Message_Certificate struct {
	Certificate *FullCertificate `protobuf:"bytes,2,opt,name=certificate,proto3,oneof"`
}

type Message_Poms struct {
	Poms *PoMs `protobuf:"bytes,3,opt,name=poms,proto3,oneof"`
}

type Message_LightCertificate struct {
	LightCertificate *LightCertificate `protobuf:"bytes,4,opt,name=light_certificate,json=lightCertificate,proto3,oneof"`
}

func (*Message_SignedPredecision) isMessage_Type() {}

func (*Message_Certificate) isMessage_Type() {}

func (*Message_Poms) isMessage_Type() {}

func (*Message_LightCertificate) isMessage_Type() {}

type SignedPredecision struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Predecision []byte `protobuf:"bytes,1,opt,name=predecision,proto3" json:"predecision,omitempty"`
	Signature   []byte `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (x *SignedPredecision) Reset() {
	*x = SignedPredecision{}
	if protoimpl.UnsafeEnabled {
		mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignedPredecision) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignedPredecision) ProtoMessage() {}

func (x *SignedPredecision) ProtoReflect() protoreflect.Message {
	mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignedPredecision.ProtoReflect.Descriptor instead.
func (*SignedPredecision) Descriptor() ([]byte, []int) {
	return file_accountabilitypb_accountabilitypb_proto_rawDescGZIP(), []int{7}
}

func (x *SignedPredecision) GetPredecision() []byte {
	if x != nil {
		return x.Predecision
	}
	return nil
}

func (x *SignedPredecision) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

type FullCertificate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Certificate map[string]*SignedPredecision `protobuf:"bytes,1,rep,name=certificate,proto3" json:"certificate,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *FullCertificate) Reset() {
	*x = FullCertificate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FullCertificate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FullCertificate) ProtoMessage() {}

func (x *FullCertificate) ProtoReflect() protoreflect.Message {
	mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FullCertificate.ProtoReflect.Descriptor instead.
func (*FullCertificate) Descriptor() ([]byte, []int) {
	return file_accountabilitypb_accountabilitypb_proto_rawDescGZIP(), []int{8}
}

func (x *FullCertificate) GetCertificate() map[string]*SignedPredecision {
	if x != nil {
		return x.Certificate
	}
	return nil
}

type InstanceParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Membership *trantorpb.Membership `protobuf:"bytes,3,opt,name=membership,proto3" json:"membership,omitempty"`
}

func (x *InstanceParams) Reset() {
	*x = InstanceParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InstanceParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InstanceParams) ProtoMessage() {}

func (x *InstanceParams) ProtoReflect() protoreflect.Message {
	mi := &file_accountabilitypb_accountabilitypb_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InstanceParams.ProtoReflect.Descriptor instead.
func (*InstanceParams) Descriptor() ([]byte, []int) {
	return file_accountabilitypb_accountabilitypb_proto_rawDescGZIP(), []int{9}
}

func (x *InstanceParams) GetMembership() *trantorpb.Membership {
	if x != nil {
		return x.Membership
	}
	return nil
}

var File_accountabilitypb_accountabilitypb_proto protoreflect.FileDescriptor

var file_accountabilitypb_accountabilitypb_proto_rawDesc = []byte{
	0x0a, 0x27, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79,
	0x70, 0x62, 0x2f, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74,
	0x79, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10, 0x61, 0x63, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x70, 0x62, 0x1a, 0x19, 0x74, 0x72, 0x61,
	0x6e, 0x74, 0x6f, 0x72, 0x70, 0x62, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x74, 0x6f, 0x72, 0x70, 0x62,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6e, 0x65, 0x74, 0x2f, 0x63, 0x6f, 0x64, 0x65,
	0x67, 0x65, 0x6e, 0x5f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6d, 0x69, 0x72, 0x2f, 0x63, 0x6f, 0x64, 0x65, 0x67, 0x65,
	0x6e, 0x5f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0xc0, 0x01, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x3e, 0x0a, 0x0a,
	0x70, 0x72, 0x65, 0x64, 0x65, 0x63, 0x69, 0x64, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1c, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74,
	0x79, 0x70, 0x62, 0x2e, 0x50, 0x72, 0x65, 0x64, 0x65, 0x63, 0x69, 0x64, 0x65, 0x64, 0x48, 0x00,
	0x52, 0x0a, 0x70, 0x72, 0x65, 0x64, 0x65, 0x63, 0x69, 0x64, 0x65, 0x64, 0x12, 0x35, 0x0a, 0x07,
	0x64, 0x65, 0x63, 0x69, 0x64, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e,
	0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x70, 0x62,
	0x2e, 0x44, 0x65, 0x63, 0x69, 0x64, 0x65, 0x64, 0x48, 0x00, 0x52, 0x07, 0x64, 0x65, 0x63, 0x69,
	0x64, 0x65, 0x64, 0x12, 0x2c, 0x0a, 0x04, 0x70, 0x6f, 0x6d, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x16, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x61, 0x62, 0x69, 0x6c, 0x69,
	0x74, 0x79, 0x70, 0x62, 0x2e, 0x50, 0x6f, 0x4d, 0x73, 0x48, 0x00, 0x52, 0x04, 0x70, 0x6f, 0x6d,
	0x73, 0x3a, 0x04, 0x90, 0xa6, 0x1d, 0x01, 0x42, 0x0c, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12,
	0x04, 0x80, 0xa6, 0x1d, 0x01, 0x22, 0x26, 0x0a, 0x0a, 0x50, 0x72, 0x65, 0x64, 0x65, 0x63, 0x69,
	0x64, 0x65, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x22, 0x23, 0x0a,
	0x07, 0x44, 0x65, 0x63, 0x69, 0x64, 0x65, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x3a, 0x04, 0x98, 0xa6,
	0x1d, 0x01, 0x22, 0xfc, 0x01, 0x0a, 0x03, 0x50, 0x6f, 0x4d, 0x12, 0x4d, 0x0a, 0x07, 0x6e, 0x6f,
	0x64, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x34, 0x82, 0xa6, 0x1d,
	0x30, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65,
	0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72,
	0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x49,
	0x44, 0x52, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x12, 0x4f, 0x0a, 0x11, 0x63, 0x6f, 0x6e,
	0x66, 0x6c, 0x69, 0x63, 0x74, 0x69, 0x6e, 0x67, 0x5f, 0x6d, 0x73, 0x67, 0x5f, 0x31, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x61, 0x62,
	0x69, 0x6c, 0x69, 0x74, 0x79, 0x70, 0x62, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x50, 0x72,
	0x65, 0x64, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0f, 0x63, 0x6f, 0x6e, 0x66, 0x6c,
	0x69, 0x63, 0x74, 0x69, 0x6e, 0x67, 0x4d, 0x73, 0x67, 0x31, 0x12, 0x4f, 0x0a, 0x11, 0x63, 0x6f,
	0x6e, 0x66, 0x6c, 0x69, 0x63, 0x74, 0x69, 0x6e, 0x67, 0x5f, 0x6d, 0x73, 0x67, 0x5f, 0x32, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x61,
	0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x70, 0x62, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x50,
	0x72, 0x65, 0x64, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0f, 0x63, 0x6f, 0x6e, 0x66,
	0x6c, 0x69, 0x63, 0x74, 0x69, 0x6e, 0x67, 0x4d, 0x73, 0x67, 0x32, 0x3a, 0x04, 0x80, 0xa6, 0x1d,
	0x01, 0x22, 0x2c, 0x0a, 0x10, 0x4c, 0x69, 0x67, 0x68, 0x74, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x3a, 0x04, 0xd0, 0xe4, 0x1d, 0x01, 0x22,
	0x3b, 0x0a, 0x04, 0x50, 0x6f, 0x4d, 0x73, 0x12, 0x29, 0x0a, 0x04, 0x70, 0x6f, 0x6d, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x61,
	0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x70, 0x62, 0x2e, 0x50, 0x6f, 0x4d, 0x52, 0x04, 0x70, 0x6f,
	0x6d, 0x73, 0x3a, 0x08, 0x98, 0xa6, 0x1d, 0x01, 0xd0, 0xe4, 0x1d, 0x01, 0x22, 0xbb, 0x02, 0x0a,
	0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x54, 0x0a, 0x12, 0x73, 0x69, 0x67, 0x6e,
	0x65, 0x64, 0x5f, 0x70, 0x72, 0x65, 0x64, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x61, 0x62,
	0x69, 0x6c, 0x69, 0x74, 0x79, 0x70, 0x62, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x50, 0x72,
	0x65, 0x64, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x11, 0x73, 0x69, 0x67,
	0x6e, 0x65, 0x64, 0x50, 0x72, 0x65, 0x64, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x45,
	0x0a, 0x0b, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x61, 0x62, 0x69,
	0x6c, 0x69, 0x74, 0x79, 0x70, 0x62, 0x2e, 0x46, 0x75, 0x6c, 0x6c, 0x43, 0x65, 0x72, 0x74, 0x69,
	0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x48, 0x00, 0x52, 0x0b, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x65, 0x12, 0x2c, 0x0a, 0x04, 0x70, 0x6f, 0x6d, 0x73, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x61, 0x62, 0x69,
	0x6c, 0x69, 0x74, 0x79, 0x70, 0x62, 0x2e, 0x50, 0x6f, 0x4d, 0x73, 0x48, 0x00, 0x52, 0x04, 0x70,
	0x6f, 0x6d, 0x73, 0x12, 0x51, 0x0a, 0x11, 0x6c, 0x69, 0x67, 0x68, 0x74, 0x5f, 0x63, 0x65, 0x72,
	0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22,
	0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x70,
	0x62, 0x2e, 0x4c, 0x69, 0x67, 0x68, 0x74, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x65, 0x48, 0x00, 0x52, 0x10, 0x6c, 0x69, 0x67, 0x68, 0x74, 0x43, 0x65, 0x72, 0x74, 0x69,
	0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x3a, 0x04, 0xc8, 0xe4, 0x1d, 0x01, 0x42, 0x0c, 0x0a, 0x04,
	0x74, 0x79, 0x70, 0x65, 0x12, 0x04, 0xc8, 0xe4, 0x1d, 0x01, 0x22, 0x5d, 0x0a, 0x11, 0x53, 0x69,
	0x67, 0x6e, 0x65, 0x64, 0x50, 0x72, 0x65, 0x64, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x12,
	0x20, 0x0a, 0x0b, 0x70, 0x72, 0x65, 0x64, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x70, 0x72, 0x65, 0x64, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f,
	0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x3a,
	0x08, 0xd0, 0xe4, 0x1d, 0x01, 0x80, 0xa6, 0x1d, 0x01, 0x22, 0x89, 0x02, 0x0a, 0x0f, 0x46, 0x75,
	0x6c, 0x6c, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x12, 0x8a, 0x01,
	0x0a, 0x0b, 0x63, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x32, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x61, 0x62, 0x69,
	0x6c, 0x69, 0x74, 0x79, 0x70, 0x62, 0x2e, 0x46, 0x75, 0x6c, 0x6c, 0x43, 0x65, 0x72, 0x74, 0x69,
	0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x2e, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x42, 0x34, 0xaa, 0xa6, 0x1d, 0x30, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e,
	0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67,
	0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x49, 0x44, 0x52, 0x0b, 0x63,
	0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x1a, 0x63, 0x0a, 0x10, 0x43, 0x65,
	0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x39, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x23, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79,
	0x70, 0x62, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x50, 0x72, 0x65, 0x64, 0x65, 0x63, 0x69,
	0x73, 0x69, 0x6f, 0x6e, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x3a,
	0x04, 0xd0, 0xe4, 0x1d, 0x01, 0x22, 0x4d, 0x0a, 0x0e, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63,
	0x65, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x12, 0x35, 0x0a, 0x0a, 0x6d, 0x65, 0x6d, 0x62, 0x65,
	0x72, 0x73, 0x68, 0x69, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x74, 0x72,
	0x61, 0x6e, 0x74, 0x6f, 0x72, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68,
	0x69, 0x70, 0x52, 0x0a, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x3a, 0x04,
	0x80, 0xa6, 0x1d, 0x01, 0x42, 0x39, 0x5a, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x62, 0x2f, 0x61,
	0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_accountabilitypb_accountabilitypb_proto_rawDescOnce sync.Once
	file_accountabilitypb_accountabilitypb_proto_rawDescData = file_accountabilitypb_accountabilitypb_proto_rawDesc
)

func file_accountabilitypb_accountabilitypb_proto_rawDescGZIP() []byte {
	file_accountabilitypb_accountabilitypb_proto_rawDescOnce.Do(func() {
		file_accountabilitypb_accountabilitypb_proto_rawDescData = protoimpl.X.CompressGZIP(file_accountabilitypb_accountabilitypb_proto_rawDescData)
	})
	return file_accountabilitypb_accountabilitypb_proto_rawDescData
}

var file_accountabilitypb_accountabilitypb_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_accountabilitypb_accountabilitypb_proto_goTypes = []interface{}{
	(*Event)(nil),                // 0: accountabilitypb.Event
	(*Predecided)(nil),           // 1: accountabilitypb.Predecided
	(*Decided)(nil),              // 2: accountabilitypb.Decided
	(*PoM)(nil),                  // 3: accountabilitypb.PoM
	(*LightCertificate)(nil),     // 4: accountabilitypb.LightCertificate
	(*PoMs)(nil),                 // 5: accountabilitypb.PoMs
	(*Message)(nil),              // 6: accountabilitypb.Message
	(*SignedPredecision)(nil),    // 7: accountabilitypb.SignedPredecision
	(*FullCertificate)(nil),      // 8: accountabilitypb.FullCertificate
	(*InstanceParams)(nil),       // 9: accountabilitypb.InstanceParams
	nil,                          // 10: accountabilitypb.FullCertificate.CertificateEntry
	(*trantorpb.Membership)(nil), // 11: trantorpb.Membership
}
var file_accountabilitypb_accountabilitypb_proto_depIdxs = []int32{
	1,  // 0: accountabilitypb.Event.predecided:type_name -> accountabilitypb.Predecided
	2,  // 1: accountabilitypb.Event.decided:type_name -> accountabilitypb.Decided
	5,  // 2: accountabilitypb.Event.poms:type_name -> accountabilitypb.PoMs
	7,  // 3: accountabilitypb.PoM.conflicting_msg_1:type_name -> accountabilitypb.SignedPredecision
	7,  // 4: accountabilitypb.PoM.conflicting_msg_2:type_name -> accountabilitypb.SignedPredecision
	3,  // 5: accountabilitypb.PoMs.poms:type_name -> accountabilitypb.PoM
	7,  // 6: accountabilitypb.Message.signed_predecision:type_name -> accountabilitypb.SignedPredecision
	8,  // 7: accountabilitypb.Message.certificate:type_name -> accountabilitypb.FullCertificate
	5,  // 8: accountabilitypb.Message.poms:type_name -> accountabilitypb.PoMs
	4,  // 9: accountabilitypb.Message.light_certificate:type_name -> accountabilitypb.LightCertificate
	10, // 10: accountabilitypb.FullCertificate.certificate:type_name -> accountabilitypb.FullCertificate.CertificateEntry
	11, // 11: accountabilitypb.InstanceParams.membership:type_name -> trantorpb.Membership
	7,  // 12: accountabilitypb.FullCertificate.CertificateEntry.value:type_name -> accountabilitypb.SignedPredecision
	13, // [13:13] is the sub-list for method output_type
	13, // [13:13] is the sub-list for method input_type
	13, // [13:13] is the sub-list for extension type_name
	13, // [13:13] is the sub-list for extension extendee
	0,  // [0:13] is the sub-list for field type_name
}

func init() { file_accountabilitypb_accountabilitypb_proto_init() }
func file_accountabilitypb_accountabilitypb_proto_init() {
	if File_accountabilitypb_accountabilitypb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_accountabilitypb_accountabilitypb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_accountabilitypb_accountabilitypb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Predecided); i {
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
		file_accountabilitypb_accountabilitypb_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Decided); i {
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
		file_accountabilitypb_accountabilitypb_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PoM); i {
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
		file_accountabilitypb_accountabilitypb_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LightCertificate); i {
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
		file_accountabilitypb_accountabilitypb_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PoMs); i {
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
		file_accountabilitypb_accountabilitypb_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
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
		file_accountabilitypb_accountabilitypb_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignedPredecision); i {
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
		file_accountabilitypb_accountabilitypb_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FullCertificate); i {
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
		file_accountabilitypb_accountabilitypb_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InstanceParams); i {
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
	file_accountabilitypb_accountabilitypb_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Event_Predecided)(nil),
		(*Event_Decided)(nil),
		(*Event_Poms)(nil),
	}
	file_accountabilitypb_accountabilitypb_proto_msgTypes[6].OneofWrappers = []interface{}{
		(*Message_SignedPredecision)(nil),
		(*Message_Certificate)(nil),
		(*Message_Poms)(nil),
		(*Message_LightCertificate)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_accountabilitypb_accountabilitypb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_accountabilitypb_accountabilitypb_proto_goTypes,
		DependencyIndexes: file_accountabilitypb_accountabilitypb_proto_depIdxs,
		MessageInfos:      file_accountabilitypb_accountabilitypb_proto_msgTypes,
	}.Build()
	File_accountabilitypb_accountabilitypb_proto = out.File
	file_accountabilitypb_accountabilitypb_proto_rawDesc = nil
	file_accountabilitypb_accountabilitypb_proto_goTypes = nil
	file_accountabilitypb_accountabilitypb_proto_depIdxs = nil
}
