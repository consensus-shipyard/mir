// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.12.4
// source: ordererpb/pprepvalidatorpb/pprepvalidatorpb.proto

package pprepvalidatorpb

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"

	contextstorepb "github.com/filecoin-project/mir/pkg/pb/contextstorepb"
	dslpb "github.com/filecoin-project/mir/pkg/pb/dslpb"
	_ "github.com/filecoin-project/mir/pkg/pb/mir"
	pbftpb "github.com/filecoin-project/mir/pkg/pb/pbftpb"
	trantorpb "github.com/filecoin-project/mir/pkg/pb/trantorpb"
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
	//	*Event_ValidatePreprepare
	//	*Event_PreprepareValidated
	Type isEvent_Type `protobuf_oneof:"type"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes[0]
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
	return file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_rawDescGZIP(), []int{0}
}

func (m *Event) GetType() isEvent_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Event) GetValidatePreprepare() *ValidatePreprepare {
	if x, ok := x.GetType().(*Event_ValidatePreprepare); ok {
		return x.ValidatePreprepare
	}
	return nil
}

func (x *Event) GetPreprepareValidated() *PreprepareValidated {
	if x, ok := x.GetType().(*Event_PreprepareValidated); ok {
		return x.PreprepareValidated
	}
	return nil
}

type isEvent_Type interface {
	isEvent_Type()
}

type Event_ValidatePreprepare struct {
	ValidatePreprepare *ValidatePreprepare `protobuf:"bytes,1,opt,name=validate_preprepare,json=validatePreprepare,proto3,oneof"`
}

type Event_PreprepareValidated struct {
	PreprepareValidated *PreprepareValidated `protobuf:"bytes,2,opt,name=preprepare_validated,json=preprepareValidated,proto3,oneof"`
}

func (*Event_ValidatePreprepare) isEvent_Type() {}

func (*Event_PreprepareValidated) isEvent_Type() {}

type ValidatePreprepare struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Preprepare *pbftpb.Preprepare        `protobuf:"bytes,1,opt,name=preprepare,proto3" json:"preprepare,omitempty"`
	Origin     *ValidatePreprepareOrigin `protobuf:"bytes,2,opt,name=origin,proto3" json:"origin,omitempty"`
}

func (x *ValidatePreprepare) Reset() {
	*x = ValidatePreprepare{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValidatePreprepare) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidatePreprepare) ProtoMessage() {}

func (x *ValidatePreprepare) ProtoReflect() protoreflect.Message {
	mi := &file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidatePreprepare.ProtoReflect.Descriptor instead.
func (*ValidatePreprepare) Descriptor() ([]byte, []int) {
	return file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_rawDescGZIP(), []int{1}
}

func (x *ValidatePreprepare) GetPreprepare() *pbftpb.Preprepare {
	if x != nil {
		return x.Preprepare
	}
	return nil
}

func (x *ValidatePreprepare) GetOrigin() *ValidatePreprepareOrigin {
	if x != nil {
		return x.Origin
	}
	return nil
}

type PreprepareValidated struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error  string                    `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	Origin *ValidatePreprepareOrigin `protobuf:"bytes,2,opt,name=origin,proto3" json:"origin,omitempty"`
}

func (x *PreprepareValidated) Reset() {
	*x = PreprepareValidated{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PreprepareValidated) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PreprepareValidated) ProtoMessage() {}

func (x *PreprepareValidated) ProtoReflect() protoreflect.Message {
	mi := &file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PreprepareValidated.ProtoReflect.Descriptor instead.
func (*PreprepareValidated) Descriptor() ([]byte, []int) {
	return file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_rawDescGZIP(), []int{2}
}

func (x *PreprepareValidated) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

func (x *PreprepareValidated) GetOrigin() *ValidatePreprepareOrigin {
	if x != nil {
		return x.Origin
	}
	return nil
}

type ValidatePreprepareOrigin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Module string `protobuf:"bytes,1,opt,name=module,proto3" json:"module,omitempty"`
	// Types that are assignable to Type:
	//	*ValidatePreprepareOrigin_ContextStore
	//	*ValidatePreprepareOrigin_Dsl
	Type isValidatePreprepareOrigin_Type `protobuf_oneof:"type"`
}

func (x *ValidatePreprepareOrigin) Reset() {
	*x = ValidatePreprepareOrigin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValidatePreprepareOrigin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidatePreprepareOrigin) ProtoMessage() {}

func (x *ValidatePreprepareOrigin) ProtoReflect() protoreflect.Message {
	mi := &file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidatePreprepareOrigin.ProtoReflect.Descriptor instead.
func (*ValidatePreprepareOrigin) Descriptor() ([]byte, []int) {
	return file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_rawDescGZIP(), []int{3}
}

func (x *ValidatePreprepareOrigin) GetModule() string {
	if x != nil {
		return x.Module
	}
	return ""
}

func (m *ValidatePreprepareOrigin) GetType() isValidatePreprepareOrigin_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *ValidatePreprepareOrigin) GetContextStore() *contextstorepb.Origin {
	if x, ok := x.GetType().(*ValidatePreprepareOrigin_ContextStore); ok {
		return x.ContextStore
	}
	return nil
}

func (x *ValidatePreprepareOrigin) GetDsl() *dslpb.Origin {
	if x, ok := x.GetType().(*ValidatePreprepareOrigin_Dsl); ok {
		return x.Dsl
	}
	return nil
}

type isValidatePreprepareOrigin_Type interface {
	isValidatePreprepareOrigin_Type()
}

type ValidatePreprepareOrigin_ContextStore struct {
	ContextStore *contextstorepb.Origin `protobuf:"bytes,2,opt,name=context_store,json=contextStore,proto3,oneof"`
}

type ValidatePreprepareOrigin_Dsl struct {
	Dsl *dslpb.Origin `protobuf:"bytes,3,opt,name=dsl,proto3,oneof"`
}

func (*ValidatePreprepareOrigin_ContextStore) isValidatePreprepareOrigin_Type() {}

func (*ValidatePreprepareOrigin_Dsl) isValidatePreprepareOrigin_Type() {}

type PPrepValidatorChkp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Membership *trantorpb.Membership `protobuf:"bytes,2,opt,name=membership,proto3" json:"membership,omitempty"`
}

func (x *PPrepValidatorChkp) Reset() {
	*x = PPrepValidatorChkp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PPrepValidatorChkp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PPrepValidatorChkp) ProtoMessage() {}

func (x *PPrepValidatorChkp) ProtoReflect() protoreflect.Message {
	mi := &file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PPrepValidatorChkp.ProtoReflect.Descriptor instead.
func (*PPrepValidatorChkp) Descriptor() ([]byte, []int) {
	return file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_rawDescGZIP(), []int{4}
}

func (x *PPrepValidatorChkp) GetMembership() *trantorpb.Membership {
	if x != nil {
		return x.Membership
	}
	return nil
}

var File_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto protoreflect.FileDescriptor

var file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_rawDesc = []byte{
	0x0a, 0x31, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x72, 0x70, 0x62, 0x2f, 0x70, 0x70, 0x72, 0x65,
	0x70, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x70, 0x62, 0x2f, 0x70, 0x70, 0x72,
	0x65, 0x70, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x70, 0x62, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x10, 0x70, 0x70, 0x72, 0x65, 0x70, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x6f, 0x72, 0x70, 0x62, 0x1a, 0x23, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x73, 0x74,
	0x6f, 0x72, 0x65, 0x70, 0x62, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x73, 0x74, 0x6f,
	0x72, 0x65, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x11, 0x64, 0x73, 0x6c, 0x70,
	0x62, 0x2f, 0x64, 0x73, 0x6c, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x74,
	0x72, 0x61, 0x6e, 0x74, 0x6f, 0x72, 0x70, 0x62, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x74, 0x6f, 0x72,
	0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x13, 0x70, 0x62, 0x66, 0x74, 0x70, 0x62,
	0x2f, 0x70, 0x62, 0x66, 0x74, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6d,
	0x69, 0x72, 0x2f, 0x63, 0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e, 0x5f, 0x65, 0x78, 0x74, 0x65, 0x6e,
	0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd0, 0x01, 0x0a, 0x05,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x57, 0x0a, 0x13, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x65, 0x5f, 0x70, 0x72, 0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x24, 0x2e, 0x70, 0x70, 0x72, 0x65, 0x70, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x6f, 0x72, 0x70, 0x62, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x50, 0x72,
	0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x48, 0x00, 0x52, 0x12, 0x76, 0x61, 0x6c, 0x69,
	0x64, 0x61, 0x74, 0x65, 0x50, 0x72, 0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x12, 0x5a,
	0x0a, 0x14, 0x70, 0x72, 0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x5f, 0x76, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x70,
	0x70, 0x72, 0x65, 0x70, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x70, 0x62, 0x2e,
	0x50, 0x72, 0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x64, 0x48, 0x00, 0x52, 0x13, 0x70, 0x72, 0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72,
	0x65, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x64, 0x3a, 0x04, 0x90, 0xa6, 0x1d, 0x01,
	0x42, 0x0c, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x04, 0x80, 0xa6, 0x1d, 0x01, 0x22, 0x98,
	0x01, 0x0a, 0x12, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x50, 0x72, 0x65, 0x70, 0x72,
	0x65, 0x70, 0x61, 0x72, 0x65, 0x12, 0x32, 0x0a, 0x0a, 0x70, 0x72, 0x65, 0x70, 0x72, 0x65, 0x70,
	0x61, 0x72, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x70, 0x62, 0x66, 0x74,
	0x70, 0x62, 0x2e, 0x50, 0x72, 0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x52, 0x0a, 0x70,
	0x72, 0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x12, 0x48, 0x0a, 0x06, 0x6f, 0x72, 0x69,
	0x67, 0x69, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x70, 0x70, 0x72, 0x65,
	0x70, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x70, 0x62, 0x2e, 0x56, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x65, 0x50, 0x72, 0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x4f,
	0x72, 0x69, 0x67, 0x69, 0x6e, 0x42, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x52, 0x06, 0x6f, 0x72, 0x69,
	0x67, 0x69, 0x6e, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x22, 0x86, 0x01, 0x0a, 0x13, 0x50, 0x72,
	0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x64, 0x12, 0x1f, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x42, 0x09, 0x82, 0xa6, 0x1d, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x05, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x12, 0x48, 0x0a, 0x06, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x70, 0x70, 0x72, 0x65, 0x70, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x6f, 0x72, 0x70, 0x62, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x50, 0x72,
	0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x42, 0x04,
	0xa0, 0xa6, 0x1d, 0x01, 0x52, 0x06, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x3a, 0x04, 0x98, 0xa6,
	0x1d, 0x01, 0x22, 0xda, 0x01, 0x0a, 0x18, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x50,
	0x72, 0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x12,
	0x4e, 0x0a, 0x06, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42,
	0x36, 0x82, 0xa6, 0x1d, 0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x4d,
	0x6f, 0x64, 0x75, 0x6c, 0x65, 0x49, 0x44, 0x52, 0x06, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x12,
	0x3d, 0x0a, 0x0d, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x73, 0x74, 0x6f, 0x72, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74,
	0x73, 0x74, 0x6f, 0x72, 0x65, 0x70, 0x62, 0x2e, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x48, 0x00,
	0x52, 0x0c, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x12, 0x21,
	0x0a, 0x03, 0x64, 0x73, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x64, 0x73,
	0x6c, 0x70, 0x62, 0x2e, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x48, 0x00, 0x52, 0x03, 0x64, 0x73,
	0x6c, 0x3a, 0x04, 0x80, 0xa6, 0x1d, 0x01, 0x42, 0x06, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x22,
	0x51, 0x0a, 0x12, 0x50, 0x50, 0x72, 0x65, 0x70, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f,
	0x72, 0x43, 0x68, 0x6b, 0x70, 0x12, 0x35, 0x0a, 0x0a, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73,
	0x68, 0x69, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x74, 0x72, 0x61, 0x6e,
	0x74, 0x6f, 0x72, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70,
	0x52, 0x0a, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x3a, 0x04, 0x80, 0xa6,
	0x1d, 0x01, 0x42, 0x43, 0x5a, 0x41, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x62, 0x2f, 0x6f, 0x72, 0x64,
	0x65, 0x72, 0x65, 0x72, 0x70, 0x62, 0x2f, 0x70, 0x70, 0x72, 0x65, 0x70, 0x76, 0x61, 0x6c, 0x69,
	0x64, 0x61, 0x74, 0x6f, 0x72, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_rawDescOnce sync.Once
	file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_rawDescData = file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_rawDesc
)

func file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_rawDescGZIP() []byte {
	file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_rawDescOnce.Do(func() {
		file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_rawDescData = protoimpl.X.CompressGZIP(file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_rawDescData)
	})
	return file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_rawDescData
}

var file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_goTypes = []interface{}{
	(*Event)(nil),                    // 0: pprepvalidatorpb.Event
	(*ValidatePreprepare)(nil),       // 1: pprepvalidatorpb.ValidatePreprepare
	(*PreprepareValidated)(nil),      // 2: pprepvalidatorpb.PreprepareValidated
	(*ValidatePreprepareOrigin)(nil), // 3: pprepvalidatorpb.ValidatePreprepareOrigin
	(*PPrepValidatorChkp)(nil),       // 4: pprepvalidatorpb.PPrepValidatorChkp
	(*pbftpb.Preprepare)(nil),        // 5: pbftpb.Preprepare
	(*contextstorepb.Origin)(nil),    // 6: contextstorepb.Origin
	(*dslpb.Origin)(nil),             // 7: dslpb.Origin
	(*trantorpb.Membership)(nil),     // 8: trantorpb.Membership
}
var file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_depIdxs = []int32{
	1, // 0: pprepvalidatorpb.Event.validate_preprepare:type_name -> pprepvalidatorpb.ValidatePreprepare
	2, // 1: pprepvalidatorpb.Event.preprepare_validated:type_name -> pprepvalidatorpb.PreprepareValidated
	5, // 2: pprepvalidatorpb.ValidatePreprepare.preprepare:type_name -> pbftpb.Preprepare
	3, // 3: pprepvalidatorpb.ValidatePreprepare.origin:type_name -> pprepvalidatorpb.ValidatePreprepareOrigin
	3, // 4: pprepvalidatorpb.PreprepareValidated.origin:type_name -> pprepvalidatorpb.ValidatePreprepareOrigin
	6, // 5: pprepvalidatorpb.ValidatePreprepareOrigin.context_store:type_name -> contextstorepb.Origin
	7, // 6: pprepvalidatorpb.ValidatePreprepareOrigin.dsl:type_name -> dslpb.Origin
	8, // 7: pprepvalidatorpb.PPrepValidatorChkp.membership:type_name -> trantorpb.Membership
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_init() }
func file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_init() {
	if File_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValidatePreprepare); i {
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
		file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PreprepareValidated); i {
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
		file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValidatePreprepareOrigin); i {
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
		file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PPrepValidatorChkp); i {
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
	file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Event_ValidatePreprepare)(nil),
		(*Event_PreprepareValidated)(nil),
	}
	file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes[3].OneofWrappers = []interface{}{
		(*ValidatePreprepareOrigin_ContextStore)(nil),
		(*ValidatePreprepareOrigin_Dsl)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_goTypes,
		DependencyIndexes: file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_depIdxs,
		MessageInfos:      file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_msgTypes,
	}.Build()
	File_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto = out.File
	file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_rawDesc = nil
	file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_goTypes = nil
	file_ordererpb_pprepvalidatorpb_pprepvalidatorpb_proto_depIdxs = nil
}
