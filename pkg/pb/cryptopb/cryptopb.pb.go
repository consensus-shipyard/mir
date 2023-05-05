// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.12.4
// source: cryptopb/cryptopb.proto

package cryptopb

import (
	checkpointpb "github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	contextstorepb "github.com/filecoin-project/mir/pkg/pb/contextstorepb"
	dslpb "github.com/filecoin-project/mir/pkg/pb/dslpb"
	_ "github.com/filecoin-project/mir/pkg/pb/mir"
	ordererpb "github.com/filecoin-project/mir/pkg/pb/ordererpb"
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
	//	*Event_SignRequest
	//	*Event_SignResult
	//	*Event_VerifySig
	//	*Event_SigVerified
	//	*Event_VerifySigs
	//	*Event_SigsVerified
	Type isEvent_Type `protobuf_oneof:"type"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptopb_cryptopb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_cryptopb_cryptopb_proto_msgTypes[0]
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
	return file_cryptopb_cryptopb_proto_rawDescGZIP(), []int{0}
}

func (m *Event) GetType() isEvent_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Event) GetSignRequest() *SignRequest {
	if x, ok := x.GetType().(*Event_SignRequest); ok {
		return x.SignRequest
	}
	return nil
}

func (x *Event) GetSignResult() *SignResult {
	if x, ok := x.GetType().(*Event_SignResult); ok {
		return x.SignResult
	}
	return nil
}

func (x *Event) GetVerifySig() *VerifySig {
	if x, ok := x.GetType().(*Event_VerifySig); ok {
		return x.VerifySig
	}
	return nil
}

func (x *Event) GetSigVerified() *SigVerified {
	if x, ok := x.GetType().(*Event_SigVerified); ok {
		return x.SigVerified
	}
	return nil
}

func (x *Event) GetVerifySigs() *VerifySigs {
	if x, ok := x.GetType().(*Event_VerifySigs); ok {
		return x.VerifySigs
	}
	return nil
}

func (x *Event) GetSigsVerified() *SigsVerified {
	if x, ok := x.GetType().(*Event_SigsVerified); ok {
		return x.SigsVerified
	}
	return nil
}

type isEvent_Type interface {
	isEvent_Type()
}

type Event_SignRequest struct {
	SignRequest *SignRequest `protobuf:"bytes,1,opt,name=sign_request,json=signRequest,proto3,oneof"`
}

type Event_SignResult struct {
	SignResult *SignResult `protobuf:"bytes,2,opt,name=sign_result,json=signResult,proto3,oneof"`
}

type Event_VerifySig struct {
	VerifySig *VerifySig `protobuf:"bytes,3,opt,name=verify_sig,json=verifySig,proto3,oneof"`
}

type Event_SigVerified struct {
	SigVerified *SigVerified `protobuf:"bytes,4,opt,name=sig_verified,json=sigVerified,proto3,oneof"`
}

type Event_VerifySigs struct {
	VerifySigs *VerifySigs `protobuf:"bytes,5,opt,name=verify_sigs,json=verifySigs,proto3,oneof"`
}

type Event_SigsVerified struct {
	SigsVerified *SigsVerified `protobuf:"bytes,6,opt,name=sigs_verified,json=sigsVerified,proto3,oneof"`
}

func (*Event_SignRequest) isEvent_Type() {}

func (*Event_SignResult) isEvent_Type() {}

func (*Event_VerifySig) isEvent_Type() {}

func (*Event_SigVerified) isEvent_Type() {}

func (*Event_VerifySigs) isEvent_Type() {}

func (*Event_SigsVerified) isEvent_Type() {}

type SignRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data   *SignedData `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	Origin *SignOrigin `protobuf:"bytes,2,opt,name=origin,proto3" json:"origin,omitempty"`
}

func (x *SignRequest) Reset() {
	*x = SignRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptopb_cryptopb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignRequest) ProtoMessage() {}

func (x *SignRequest) ProtoReflect() protoreflect.Message {
	mi := &file_cryptopb_cryptopb_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignRequest.ProtoReflect.Descriptor instead.
func (*SignRequest) Descriptor() ([]byte, []int) {
	return file_cryptopb_cryptopb_proto_rawDescGZIP(), []int{1}
}

func (x *SignRequest) GetData() *SignedData {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *SignRequest) GetOrigin() *SignOrigin {
	if x != nil {
		return x.Origin
	}
	return nil
}

type SignResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Signature []byte      `protobuf:"bytes,1,opt,name=signature,proto3" json:"signature,omitempty"`
	Origin    *SignOrigin `protobuf:"bytes,2,opt,name=origin,proto3" json:"origin,omitempty"`
}

func (x *SignResult) Reset() {
	*x = SignResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptopb_cryptopb_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignResult) ProtoMessage() {}

func (x *SignResult) ProtoReflect() protoreflect.Message {
	mi := &file_cryptopb_cryptopb_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignResult.ProtoReflect.Descriptor instead.
func (*SignResult) Descriptor() ([]byte, []int) {
	return file_cryptopb_cryptopb_proto_rawDescGZIP(), []int{2}
}

func (x *SignResult) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *SignResult) GetOrigin() *SignOrigin {
	if x != nil {
		return x.Origin
	}
	return nil
}

type VerifySig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data      *SignedData   `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	Signature []byte        `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
	Origin    *SigVerOrigin `protobuf:"bytes,3,opt,name=origin,proto3" json:"origin,omitempty"`
	NodeId    string        `protobuf:"bytes,4,opt,name=node_id,json=nodeId,proto3" json:"node_id,omitempty"`
}

func (x *VerifySig) Reset() {
	*x = VerifySig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptopb_cryptopb_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VerifySig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VerifySig) ProtoMessage() {}

func (x *VerifySig) ProtoReflect() protoreflect.Message {
	mi := &file_cryptopb_cryptopb_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VerifySig.ProtoReflect.Descriptor instead.
func (*VerifySig) Descriptor() ([]byte, []int) {
	return file_cryptopb_cryptopb_proto_rawDescGZIP(), []int{3}
}

func (x *VerifySig) GetData() *SignedData {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *VerifySig) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *VerifySig) GetOrigin() *SigVerOrigin {
	if x != nil {
		return x.Origin
	}
	return nil
}

func (x *VerifySig) GetNodeId() string {
	if x != nil {
		return x.NodeId
	}
	return ""
}

type SigVerified struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Origin *SigVerOrigin `protobuf:"bytes,1,opt,name=origin,proto3" json:"origin,omitempty"`
	NodeId string        `protobuf:"bytes,2,opt,name=node_id,json=nodeId,proto3" json:"node_id,omitempty"`
	Error  string        `protobuf:"bytes,3,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *SigVerified) Reset() {
	*x = SigVerified{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptopb_cryptopb_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SigVerified) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SigVerified) ProtoMessage() {}

func (x *SigVerified) ProtoReflect() protoreflect.Message {
	mi := &file_cryptopb_cryptopb_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SigVerified.ProtoReflect.Descriptor instead.
func (*SigVerified) Descriptor() ([]byte, []int) {
	return file_cryptopb_cryptopb_proto_rawDescGZIP(), []int{4}
}

func (x *SigVerified) GetOrigin() *SigVerOrigin {
	if x != nil {
		return x.Origin
	}
	return nil
}

func (x *SigVerified) GetNodeId() string {
	if x != nil {
		return x.NodeId
	}
	return ""
}

func (x *SigVerified) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type VerifySigs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data       []*SignedData `protobuf:"bytes,1,rep,name=data,proto3" json:"data,omitempty"`
	Signatures [][]byte      `protobuf:"bytes,2,rep,name=signatures,proto3" json:"signatures,omitempty"`
	Origin     *SigVerOrigin `protobuf:"bytes,3,opt,name=origin,proto3" json:"origin,omitempty"`
	NodeIds    []string      `protobuf:"bytes,4,rep,name=node_ids,json=nodeIds,proto3" json:"node_ids,omitempty"`
}

func (x *VerifySigs) Reset() {
	*x = VerifySigs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptopb_cryptopb_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VerifySigs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VerifySigs) ProtoMessage() {}

func (x *VerifySigs) ProtoReflect() protoreflect.Message {
	mi := &file_cryptopb_cryptopb_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VerifySigs.ProtoReflect.Descriptor instead.
func (*VerifySigs) Descriptor() ([]byte, []int) {
	return file_cryptopb_cryptopb_proto_rawDescGZIP(), []int{5}
}

func (x *VerifySigs) GetData() []*SignedData {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *VerifySigs) GetSignatures() [][]byte {
	if x != nil {
		return x.Signatures
	}
	return nil
}

func (x *VerifySigs) GetOrigin() *SigVerOrigin {
	if x != nil {
		return x.Origin
	}
	return nil
}

func (x *VerifySigs) GetNodeIds() []string {
	if x != nil {
		return x.NodeIds
	}
	return nil
}

type SigsVerified struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Origin  *SigVerOrigin `protobuf:"bytes,1,opt,name=origin,proto3" json:"origin,omitempty"`
	NodeIds []string      `protobuf:"bytes,2,rep,name=node_ids,json=nodeIds,proto3" json:"node_ids,omitempty"`
	Errors  []string      `protobuf:"bytes,3,rep,name=errors,proto3" json:"errors,omitempty"`
	AllOk   bool          `protobuf:"varint,4,opt,name=all_ok,json=allOk,proto3" json:"all_ok,omitempty"`
}

func (x *SigsVerified) Reset() {
	*x = SigsVerified{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptopb_cryptopb_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SigsVerified) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SigsVerified) ProtoMessage() {}

func (x *SigsVerified) ProtoReflect() protoreflect.Message {
	mi := &file_cryptopb_cryptopb_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SigsVerified.ProtoReflect.Descriptor instead.
func (*SigsVerified) Descriptor() ([]byte, []int) {
	return file_cryptopb_cryptopb_proto_rawDescGZIP(), []int{6}
}

func (x *SigsVerified) GetOrigin() *SigVerOrigin {
	if x != nil {
		return x.Origin
	}
	return nil
}

func (x *SigsVerified) GetNodeIds() []string {
	if x != nil {
		return x.NodeIds
	}
	return nil
}

func (x *SigsVerified) GetErrors() []string {
	if x != nil {
		return x.Errors
	}
	return nil
}

func (x *SigsVerified) GetAllOk() bool {
	if x != nil {
		return x.AllOk
	}
	return false
}

type SignOrigin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Module string `protobuf:"bytes,1,opt,name=module,proto3" json:"module,omitempty"`
	// Types that are assignable to Type:
	//	*SignOrigin_ContextStore
	//	*SignOrigin_Dsl
	Type isSignOrigin_Type `protobuf_oneof:"type"`
}

func (x *SignOrigin) Reset() {
	*x = SignOrigin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptopb_cryptopb_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignOrigin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignOrigin) ProtoMessage() {}

func (x *SignOrigin) ProtoReflect() protoreflect.Message {
	mi := &file_cryptopb_cryptopb_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignOrigin.ProtoReflect.Descriptor instead.
func (*SignOrigin) Descriptor() ([]byte, []int) {
	return file_cryptopb_cryptopb_proto_rawDescGZIP(), []int{7}
}

func (x *SignOrigin) GetModule() string {
	if x != nil {
		return x.Module
	}
	return ""
}

func (m *SignOrigin) GetType() isSignOrigin_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *SignOrigin) GetContextStore() *contextstorepb.Origin {
	if x, ok := x.GetType().(*SignOrigin_ContextStore); ok {
		return x.ContextStore
	}
	return nil
}

func (x *SignOrigin) GetDsl() *dslpb.Origin {
	if x, ok := x.GetType().(*SignOrigin_Dsl); ok {
		return x.Dsl
	}
	return nil
}

type isSignOrigin_Type interface {
	isSignOrigin_Type()
}

type SignOrigin_ContextStore struct {
	ContextStore *contextstorepb.Origin `protobuf:"bytes,2,opt,name=context_store,json=contextStore,proto3,oneof"`
}

type SignOrigin_Dsl struct {
	Dsl *dslpb.Origin `protobuf:"bytes,4,opt,name=dsl,proto3,oneof"`
}

func (*SignOrigin_ContextStore) isSignOrigin_Type() {}

func (*SignOrigin_Dsl) isSignOrigin_Type() {}

type SigVerOrigin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Module string `protobuf:"bytes,1,opt,name=module,proto3" json:"module,omitempty"`
	// Types that are assignable to Type:
	//	*SigVerOrigin_ContextStore
	//	*SigVerOrigin_Dsl
	Type isSigVerOrigin_Type `protobuf_oneof:"type"`
}

func (x *SigVerOrigin) Reset() {
	*x = SigVerOrigin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptopb_cryptopb_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SigVerOrigin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SigVerOrigin) ProtoMessage() {}

func (x *SigVerOrigin) ProtoReflect() protoreflect.Message {
	mi := &file_cryptopb_cryptopb_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SigVerOrigin.ProtoReflect.Descriptor instead.
func (*SigVerOrigin) Descriptor() ([]byte, []int) {
	return file_cryptopb_cryptopb_proto_rawDescGZIP(), []int{8}
}

func (x *SigVerOrigin) GetModule() string {
	if x != nil {
		return x.Module
	}
	return ""
}

func (m *SigVerOrigin) GetType() isSigVerOrigin_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *SigVerOrigin) GetContextStore() *contextstorepb.Origin {
	if x, ok := x.GetType().(*SigVerOrigin_ContextStore); ok {
		return x.ContextStore
	}
	return nil
}

func (x *SigVerOrigin) GetDsl() *dslpb.Origin {
	if x, ok := x.GetType().(*SigVerOrigin_Dsl); ok {
		return x.Dsl
	}
	return nil
}

type isSigVerOrigin_Type interface {
	isSigVerOrigin_Type()
}

type SigVerOrigin_ContextStore struct {
	ContextStore *contextstorepb.Origin `protobuf:"bytes,2,opt,name=context_store,json=contextStore,proto3,oneof"`
}

type SigVerOrigin_Dsl struct {
	Dsl *dslpb.Origin `protobuf:"bytes,4,opt,name=dsl,proto3,oneof"`
}

func (*SigVerOrigin_ContextStore) isSigVerOrigin_Type() {}

func (*SigVerOrigin_Dsl) isSigVerOrigin_Type() {}

type SignedData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data [][]byte `protobuf:"bytes,1,rep,name=data,proto3" json:"data,omitempty"`
}

func (x *SignedData) Reset() {
	*x = SignedData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cryptopb_cryptopb_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignedData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignedData) ProtoMessage() {}

func (x *SignedData) ProtoReflect() protoreflect.Message {
	mi := &file_cryptopb_cryptopb_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignedData.ProtoReflect.Descriptor instead.
func (*SignedData) Descriptor() ([]byte, []int) {
	return file_cryptopb_cryptopb_proto_rawDescGZIP(), []int{9}
}

func (x *SignedData) GetData() [][]byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_cryptopb_cryptopb_proto protoreflect.FileDescriptor

var file_cryptopb_cryptopb_proto_rawDesc = []byte{
	0x0a, 0x17, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x70, 0x62, 0x2f, 0x63, 0x72, 0x79, 0x70, 0x74,
	0x6f, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x63, 0x72, 0x79, 0x70, 0x74,
	0x6f, 0x70, 0x62, 0x1a, 0x23, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x73, 0x74, 0x6f, 0x72,
	0x65, 0x70, 0x62, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x73, 0x74, 0x6f, 0x72, 0x65,
	0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x11, 0x64, 0x73, 0x6c, 0x70, 0x62, 0x2f,
	0x64, 0x73, 0x6c, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6d, 0x69, 0x72,
	0x2f, 0x63, 0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e, 0x5f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69,
	0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xfa, 0x02, 0x0a, 0x05, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x12, 0x3a, 0x0a, 0x0c, 0x73, 0x69, 0x67, 0x6e, 0x5f, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x63, 0x72, 0x79, 0x70,
	0x74, 0x6f, 0x70, 0x62, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x48, 0x00, 0x52, 0x0b, 0x73, 0x69, 0x67, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x37, 0x0a, 0x0b, 0x73, 0x69, 0x67, 0x6e, 0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x70, 0x62, 0x2e,
	0x53, 0x69, 0x67, 0x6e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x48, 0x00, 0x52, 0x0a, 0x73, 0x69,
	0x67, 0x6e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x34, 0x0a, 0x0a, 0x76, 0x65, 0x72, 0x69,
	0x66, 0x79, 0x5f, 0x73, 0x69, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x63,
	0x72, 0x79, 0x70, 0x74, 0x6f, 0x70, 0x62, 0x2e, 0x56, 0x65, 0x72, 0x69, 0x66, 0x79, 0x53, 0x69,
	0x67, 0x48, 0x00, 0x52, 0x09, 0x76, 0x65, 0x72, 0x69, 0x66, 0x79, 0x53, 0x69, 0x67, 0x12, 0x3a,
	0x0a, 0x0c, 0x73, 0x69, 0x67, 0x5f, 0x76, 0x65, 0x72, 0x69, 0x66, 0x69, 0x65, 0x64, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x70, 0x62, 0x2e,
	0x53, 0x69, 0x67, 0x56, 0x65, 0x72, 0x69, 0x66, 0x69, 0x65, 0x64, 0x48, 0x00, 0x52, 0x0b, 0x73,
	0x69, 0x67, 0x56, 0x65, 0x72, 0x69, 0x66, 0x69, 0x65, 0x64, 0x12, 0x37, 0x0a, 0x0b, 0x76, 0x65,
	0x72, 0x69, 0x66, 0x79, 0x5f, 0x73, 0x69, 0x67, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x14, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x70, 0x62, 0x2e, 0x56, 0x65, 0x72, 0x69, 0x66,
	0x79, 0x53, 0x69, 0x67, 0x73, 0x48, 0x00, 0x52, 0x0a, 0x76, 0x65, 0x72, 0x69, 0x66, 0x79, 0x53,
	0x69, 0x67, 0x73, 0x12, 0x3d, 0x0a, 0x0d, 0x73, 0x69, 0x67, 0x73, 0x5f, 0x76, 0x65, 0x72, 0x69,
	0x66, 0x69, 0x65, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x72, 0x79,
	0x70, 0x74, 0x6f, 0x70, 0x62, 0x2e, 0x53, 0x69, 0x67, 0x73, 0x56, 0x65, 0x72, 0x69, 0x66, 0x69,
	0x65, 0x64, 0x48, 0x00, 0x52, 0x0c, 0x73, 0x69, 0x67, 0x73, 0x56, 0x65, 0x72, 0x69, 0x66, 0x69,
	0x65, 0x64, 0x3a, 0x04, 0x90, 0xa6, 0x1d, 0x01, 0x42, 0x0c, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x12, 0x04, 0x80, 0xa6, 0x1d, 0x01, 0x22, 0x71, 0x0a, 0x0b, 0x53, 0x69, 0x67, 0x6e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x28, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x70, 0x62, 0x2e, 0x53,
	0x69, 0x67, 0x6e, 0x65, 0x64, 0x44, 0x61, 0x74, 0x61, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12,
	0x32, 0x0a, 0x06, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x14, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x70, 0x62, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x4f,
	0x72, 0x69, 0x67, 0x69, 0x6e, 0x42, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x52, 0x06, 0x6f, 0x72, 0x69,
	0x67, 0x69, 0x6e, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x22, 0x64, 0x0a, 0x0a, 0x53, 0x69, 0x67,
	0x6e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x32, 0x0a, 0x06, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x70, 0x62,
	0x2e, 0x53, 0x69, 0x67, 0x6e, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x42, 0x04, 0xa0, 0xa6, 0x1d,
	0x01, 0x52, 0x06, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x22,
	0xde, 0x01, 0x0a, 0x09, 0x56, 0x65, 0x72, 0x69, 0x66, 0x79, 0x53, 0x69, 0x67, 0x12, 0x28, 0x0a,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x63, 0x72,
	0x79, 0x70, 0x74, 0x6f, 0x70, 0x62, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x44, 0x61, 0x74,
	0x61, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x34, 0x0a, 0x06, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x70, 0x62,
	0x2e, 0x53, 0x69, 0x67, 0x56, 0x65, 0x72, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x42, 0x04, 0x98,
	0xa6, 0x1d, 0x01, 0x52, 0x06, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x12, 0x4d, 0x0a, 0x07, 0x6e,
	0x6f, 0x64, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x42, 0x34, 0x82, 0xa6,
	0x1d, 0x30, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c,
	0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69,
	0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x4e, 0x6f, 0x64, 0x65,
	0x49, 0x44, 0x52, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01,
	0x22, 0xb9, 0x01, 0x0a, 0x0b, 0x53, 0x69, 0x67, 0x56, 0x65, 0x72, 0x69, 0x66, 0x69, 0x65, 0x64,
	0x12, 0x34, 0x0a, 0x06, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x16, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x70, 0x62, 0x2e, 0x53, 0x69, 0x67, 0x56,
	0x65, 0x72, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x42, 0x04, 0xa0, 0xa6, 0x1d, 0x01, 0x52, 0x06,
	0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x12, 0x4d, 0x0a, 0x07, 0x6e, 0x6f, 0x64, 0x65, 0x5f, 0x69,
	0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x34, 0x82, 0xa6, 0x1d, 0x30, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e,
	0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67,
	0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x49, 0x44, 0x52, 0x06, 0x6e,
	0x6f, 0x64, 0x65, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x09, 0x82, 0xa6, 0x1d, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x52,
	0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x22, 0xe3, 0x01, 0x0a,
	0x0a, 0x56, 0x65, 0x72, 0x69, 0x66, 0x79, 0x53, 0x69, 0x67, 0x73, 0x12, 0x28, 0x0a, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x63, 0x72, 0x79, 0x70,
	0x74, 0x6f, 0x70, 0x62, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x44, 0x61, 0x74, 0x61, 0x52,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x0a, 0x73, 0x69, 0x67, 0x6e, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x73, 0x12, 0x34, 0x0a, 0x06, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x70, 0x62,
	0x2e, 0x53, 0x69, 0x67, 0x56, 0x65, 0x72, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x42, 0x04, 0x98,
	0xa6, 0x1d, 0x01, 0x52, 0x06, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x12, 0x4f, 0x0a, 0x08, 0x6e,
	0x6f, 0x64, 0x65, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x42, 0x34, 0x82,
	0xa6, 0x1d, 0x30, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69,
	0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d,
	0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x4e, 0x6f, 0x64,
	0x65, 0x49, 0x44, 0x52, 0x07, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x73, 0x3a, 0x04, 0x98, 0xa6,
	0x1d, 0x01, 0x22, 0xd5, 0x01, 0x0a, 0x0c, 0x53, 0x69, 0x67, 0x73, 0x56, 0x65, 0x72, 0x69, 0x66,
	0x69, 0x65, 0x64, 0x12, 0x34, 0x0a, 0x06, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x70, 0x62, 0x2e, 0x53,
	0x69, 0x67, 0x56, 0x65, 0x72, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x42, 0x04, 0xa0, 0xa6, 0x1d,
	0x01, 0x52, 0x06, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x12, 0x4f, 0x0a, 0x08, 0x6e, 0x6f, 0x64,
	0x65, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x42, 0x34, 0x82, 0xa6, 0x1d,
	0x30, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65,
	0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72,
	0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x49,
	0x44, 0x52, 0x07, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x73, 0x12, 0x21, 0x0a, 0x06, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x42, 0x09, 0x82, 0xa6, 0x1d, 0x05,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x06, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x12, 0x15, 0x0a,
	0x06, 0x61, 0x6c, 0x6c, 0x5f, 0x6f, 0x6b, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x61,
	0x6c, 0x6c, 0x4f, 0x6b, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x22, 0xcc, 0x01, 0x0a, 0x0a, 0x53,
	0x69, 0x67, 0x6e, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x12, 0x4e, 0x0a, 0x06, 0x6d, 0x6f, 0x64,
	0x75, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x36, 0x82, 0xa6, 0x1d, 0x32, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f,
	0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70,
	0x6b, 0x67, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x49,
	0x44, 0x52, 0x06, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x12, 0x3d, 0x0a, 0x0d, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x78, 0x74, 0x5f, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x16, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x70,
	0x62, 0x2e, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x48, 0x00, 0x52, 0x0c, 0x63, 0x6f, 0x6e, 0x74,
	0x65, 0x78, 0x74, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x12, 0x21, 0x0a, 0x03, 0x64, 0x73, 0x6c, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x64, 0x73, 0x6c, 0x70, 0x62, 0x2e, 0x4f, 0x72,
	0x69, 0x67, 0x69, 0x6e, 0x48, 0x00, 0x52, 0x03, 0x64, 0x73, 0x6c, 0x3a, 0x04, 0x80, 0xa6, 0x1d,
	0x01, 0x42, 0x06, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x22, 0xce, 0x01, 0x0a, 0x0c, 0x53, 0x69,
	0x67, 0x56, 0x65, 0x72, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x12, 0x4e, 0x0a, 0x06, 0x6d, 0x6f,
	0x64, 0x75, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x36, 0x82, 0xa6, 0x1d, 0x32,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63,
	0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f,
	0x70, 0x6b, 0x67, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65,
	0x49, 0x44, 0x52, 0x06, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x12, 0x3d, 0x0a, 0x0d, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x16, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x73, 0x74, 0x6f, 0x72, 0x65,
	0x70, 0x62, 0x2e, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x48, 0x00, 0x52, 0x0c, 0x63, 0x6f, 0x6e,
	0x74, 0x65, 0x78, 0x74, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x12, 0x21, 0x0a, 0x03, 0x64, 0x73, 0x6c,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x64, 0x73, 0x6c, 0x70, 0x62, 0x2e, 0x4f,
	0x72, 0x69, 0x67, 0x69, 0x6e, 0x48, 0x00, 0x52, 0x03, 0x64, 0x73, 0x6c, 0x3a, 0x04, 0x80, 0xa6,
	0x1d, 0x01, 0x42, 0x06, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x22, 0x26, 0x0a, 0x0a, 0x53, 0x69,
	0x67, 0x6e, 0x65, 0x64, 0x44, 0x61, 0x74, 0x61, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x3a, 0x04, 0x80, 0xa6,
	0x1d, 0x01, 0x42, 0x31, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x62, 0x2f, 0x63, 0x72, 0x79,
	0x70, 0x74, 0x6f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cryptopb_cryptopb_proto_rawDescOnce sync.Once
	file_cryptopb_cryptopb_proto_rawDescData = file_cryptopb_cryptopb_proto_rawDesc
)

func file_cryptopb_cryptopb_proto_rawDescGZIP() []byte {
	file_cryptopb_cryptopb_proto_rawDescOnce.Do(func() {
		file_cryptopb_cryptopb_proto_rawDescData = protoimpl.X.CompressGZIP(file_cryptopb_cryptopb_proto_rawDescData)
	})
	return file_cryptopb_cryptopb_proto_rawDescData
}

var file_cryptopb_cryptopb_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_cryptopb_cryptopb_proto_goTypes = []interface{}{
	(*Event)(nil),                 // 0: cryptopb.Event
	(*SignRequest)(nil),           // 1: cryptopb.SignRequest
	(*SignResult)(nil),            // 2: cryptopb.SignResult
	(*VerifySig)(nil),             // 3: cryptopb.VerifySig
	(*SigVerified)(nil),           // 4: cryptopb.SigVerified
	(*VerifySigs)(nil),            // 5: cryptopb.VerifySigs
	(*SigsVerified)(nil),          // 6: cryptopb.SigsVerified
	(*SignOrigin)(nil),            // 7: cryptopb.SignOrigin
	(*SigVerOrigin)(nil),          // 8: cryptopb.SigVerOrigin
	(*SignedData)(nil),            // 9: cryptopb.SignedData
	(*contextstorepb.Origin)(nil), // 10: contextstorepb.Origin
	(*dslpb.Origin)(nil),          // 11: dslpb.Origin
}
var file_cryptopb_cryptopb_proto_depIdxs = []int32{
	1,  // 0: cryptopb.Event.sign_request:type_name -> cryptopb.SignRequest
	2,  // 1: cryptopb.Event.sign_result:type_name -> cryptopb.SignResult
	3,  // 2: cryptopb.Event.verify_sig:type_name -> cryptopb.VerifySig
	4,  // 3: cryptopb.Event.sig_verified:type_name -> cryptopb.SigVerified
	5,  // 4: cryptopb.Event.verify_sigs:type_name -> cryptopb.VerifySigs
	6,  // 5: cryptopb.Event.sigs_verified:type_name -> cryptopb.SigsVerified
	9,  // 6: cryptopb.SignRequest.data:type_name -> cryptopb.SignedData
	7,  // 7: cryptopb.SignRequest.origin:type_name -> cryptopb.SignOrigin
	7,  // 8: cryptopb.SignResult.origin:type_name -> cryptopb.SignOrigin
	9,  // 9: cryptopb.VerifySig.data:type_name -> cryptopb.SignedData
	8,  // 10: cryptopb.VerifySig.origin:type_name -> cryptopb.SigVerOrigin
	8,  // 11: cryptopb.SigVerified.origin:type_name -> cryptopb.SigVerOrigin
	9,  // 12: cryptopb.VerifySigs.data:type_name -> cryptopb.SignedData
	8,  // 13: cryptopb.VerifySigs.origin:type_name -> cryptopb.SigVerOrigin
	8,  // 14: cryptopb.SigsVerified.origin:type_name -> cryptopb.SigVerOrigin
	10, // 15: cryptopb.SignOrigin.context_store:type_name -> contextstorepb.Origin
	11, // 16: cryptopb.SignOrigin.dsl:type_name -> dslpb.Origin
	10, // 17: cryptopb.SigVerOrigin.context_store:type_name -> contextstorepb.Origin
	11, // 18: cryptopb.SigVerOrigin.dsl:type_name -> dslpb.Origin
	19, // [19:19] is the sub-list for method output_type
	19, // [19:19] is the sub-list for method input_type
	19, // [19:19] is the sub-list for extension type_name
	19, // [19:19] is the sub-list for extension extendee
	0,  // [0:19] is the sub-list for field type_name
}

func init() { file_cryptopb_cryptopb_proto_init() }
func file_cryptopb_cryptopb_proto_init() {
	if File_cryptopb_cryptopb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cryptopb_cryptopb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_cryptopb_cryptopb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignRequest); i {
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
		file_cryptopb_cryptopb_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignResult); i {
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
		file_cryptopb_cryptopb_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VerifySig); i {
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
		file_cryptopb_cryptopb_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SigVerified); i {
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
		file_cryptopb_cryptopb_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VerifySigs); i {
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
		file_cryptopb_cryptopb_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SigsVerified); i {
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
		file_cryptopb_cryptopb_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignOrigin); i {
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
		file_cryptopb_cryptopb_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SigVerOrigin); i {
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
		file_cryptopb_cryptopb_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignedData); i {
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
	file_cryptopb_cryptopb_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Event_SignRequest)(nil),
		(*Event_SignResult)(nil),
		(*Event_VerifySig)(nil),
		(*Event_SigVerified)(nil),
		(*Event_VerifySigs)(nil),
		(*Event_SigsVerified)(nil),
	}
	file_cryptopb_cryptopb_proto_msgTypes[7].OneofWrappers = []interface{}{
		(*SignOrigin_ContextStore)(nil),
		(*SignOrigin_Dsl)(nil),
	}
	file_cryptopb_cryptopb_proto_msgTypes[8].OneofWrappers = []interface{}{
		(*SigVerOrigin_ContextStore)(nil),
		(*SigVerOrigin_Dsl)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_cryptopb_cryptopb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_cryptopb_cryptopb_proto_goTypes,
		DependencyIndexes: file_cryptopb_cryptopb_proto_depIdxs,
		MessageInfos:      file_cryptopb_cryptopb_proto_msgTypes,
	}.Build()
	File_cryptopb_cryptopb_proto = out.File
	file_cryptopb_cryptopb_proto_rawDesc = nil
	file_cryptopb_cryptopb_proto_goTypes = nil
	file_cryptopb_cryptopb_proto_depIdxs = nil
}
