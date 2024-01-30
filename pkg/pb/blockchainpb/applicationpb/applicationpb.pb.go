// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.4
// source: blockchainpb/applicationpb/applicationpb.proto

package applicationpb

import (
	blockchainpb "github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	payloadpb "github.com/filecoin-project/mir/pkg/pb/blockchainpb/payloadpb"
	statepb "github.com/filecoin-project/mir/pkg/pb/blockchainpb/statepb"
	_ "github.com/filecoin-project/mir/pkg/pb/mir"
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
	//
	//	*Event_NewHead
	//	*Event_VerifyBlockRequest
	//	*Event_VerifyBlockResponse
	//	*Event_PayloadRequest
	//	*Event_PayloadResponse
	//	*Event_ForkUpdate
	//	*Event_MessageInput
	Type isEvent_Type `protobuf_oneof:"type"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[0]
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
	return file_blockchainpb_applicationpb_applicationpb_proto_rawDescGZIP(), []int{0}
}

func (m *Event) GetType() isEvent_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Event) GetNewHead() *NewHead {
	if x, ok := x.GetType().(*Event_NewHead); ok {
		return x.NewHead
	}
	return nil
}

func (x *Event) GetVerifyBlockRequest() *VerifyBlocksRequest {
	if x, ok := x.GetType().(*Event_VerifyBlockRequest); ok {
		return x.VerifyBlockRequest
	}
	return nil
}

func (x *Event) GetVerifyBlockResponse() *VerifyBlocksResponse {
	if x, ok := x.GetType().(*Event_VerifyBlockResponse); ok {
		return x.VerifyBlockResponse
	}
	return nil
}

func (x *Event) GetPayloadRequest() *PayloadRequest {
	if x, ok := x.GetType().(*Event_PayloadRequest); ok {
		return x.PayloadRequest
	}
	return nil
}

func (x *Event) GetPayloadResponse() *PayloadResponse {
	if x, ok := x.GetType().(*Event_PayloadResponse); ok {
		return x.PayloadResponse
	}
	return nil
}

func (x *Event) GetForkUpdate() *ForkUpdate {
	if x, ok := x.GetType().(*Event_ForkUpdate); ok {
		return x.ForkUpdate
	}
	return nil
}

func (x *Event) GetMessageInput() *MessageInput {
	if x, ok := x.GetType().(*Event_MessageInput); ok {
		return x.MessageInput
	}
	return nil
}

type isEvent_Type interface {
	isEvent_Type()
}

type Event_NewHead struct {
	// Application-application events
	NewHead *NewHead `protobuf:"bytes,10,opt,name=new_head,json=newHead,proto3,oneof"`
}

type Event_VerifyBlockRequest struct {
	VerifyBlockRequest *VerifyBlocksRequest `protobuf:"bytes,11,opt,name=verify_block_request,json=verifyBlockRequest,proto3,oneof"`
}

type Event_VerifyBlockResponse struct {
	VerifyBlockResponse *VerifyBlocksResponse `protobuf:"bytes,12,opt,name=verify_block_response,json=verifyBlockResponse,proto3,oneof"`
}

type Event_PayloadRequest struct {
	// Transaction management application events
	PayloadRequest *PayloadRequest `protobuf:"bytes,20,opt,name=payload_request,json=payloadRequest,proto3,oneof"`
}

type Event_PayloadResponse struct {
	PayloadResponse *PayloadResponse `protobuf:"bytes,21,opt,name=payload_response,json=payloadResponse,proto3,oneof"`
}

type Event_ForkUpdate struct {
	ForkUpdate *ForkUpdate `protobuf:"bytes,22,opt,name=fork_update,json=forkUpdate,proto3,oneof"`
}

type Event_MessageInput struct {
	// Message input
	MessageInput *MessageInput `protobuf:"bytes,30,opt,name=message_input,json=messageInput,proto3,oneof"`
}

func (*Event_NewHead) isEvent_Type() {}

func (*Event_VerifyBlockRequest) isEvent_Type() {}

func (*Event_VerifyBlockResponse) isEvent_Type() {}

func (*Event_PayloadRequest) isEvent_Type() {}

func (*Event_PayloadResponse) isEvent_Type() {}

func (*Event_ForkUpdate) isEvent_Type() {}

func (*Event_MessageInput) isEvent_Type() {}

// Application-application events
type NewHead struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HeadId uint64 `protobuf:"varint,1,opt,name=head_id,json=headId,proto3" json:"head_id,omitempty"`
}

func (x *NewHead) Reset() {
	*x = NewHead{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NewHead) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewHead) ProtoMessage() {}

func (x *NewHead) ProtoReflect() protoreflect.Message {
	mi := &file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewHead.ProtoReflect.Descriptor instead.
func (*NewHead) Descriptor() ([]byte, []int) {
	return file_blockchainpb_applicationpb_applicationpb_proto_rawDescGZIP(), []int{1}
}

func (x *NewHead) GetHeadId() uint64 {
	if x != nil {
		return x.HeadId
	}
	return 0
}

type VerifyBlocksRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CheckpointState        *statepb.State        `protobuf:"bytes,1,opt,name=checkpoint_state,json=checkpointState,proto3" json:"checkpoint_state,omitempty"`
	ChainCheckpointToStart []*blockchainpb.Block `protobuf:"bytes,2,rep,name=chain_checkpoint_to_start,json=chainCheckpointToStart,proto3" json:"chain_checkpoint_to_start,omitempty"`
	ChainToVerify          []*blockchainpb.Block `protobuf:"bytes,3,rep,name=chain_to_verify,json=chainToVerify,proto3" json:"chain_to_verify,omitempty"`
}

func (x *VerifyBlocksRequest) Reset() {
	*x = VerifyBlocksRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VerifyBlocksRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VerifyBlocksRequest) ProtoMessage() {}

func (x *VerifyBlocksRequest) ProtoReflect() protoreflect.Message {
	mi := &file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VerifyBlocksRequest.ProtoReflect.Descriptor instead.
func (*VerifyBlocksRequest) Descriptor() ([]byte, []int) {
	return file_blockchainpb_applicationpb_applicationpb_proto_rawDescGZIP(), []int{2}
}

func (x *VerifyBlocksRequest) GetCheckpointState() *statepb.State {
	if x != nil {
		return x.CheckpointState
	}
	return nil
}

func (x *VerifyBlocksRequest) GetChainCheckpointToStart() []*blockchainpb.Block {
	if x != nil {
		return x.ChainCheckpointToStart
	}
	return nil
}

func (x *VerifyBlocksRequest) GetChainToVerify() []*blockchainpb.Block {
	if x != nil {
		return x.ChainToVerify
	}
	return nil
}

type VerifyBlocksResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	VerifiedBlocks []*blockchainpb.Block `protobuf:"bytes,1,rep,name=verified_blocks,json=verifiedBlocks,proto3" json:"verified_blocks,omitempty"`
}

func (x *VerifyBlocksResponse) Reset() {
	*x = VerifyBlocksResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VerifyBlocksResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VerifyBlocksResponse) ProtoMessage() {}

func (x *VerifyBlocksResponse) ProtoReflect() protoreflect.Message {
	mi := &file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VerifyBlocksResponse.ProtoReflect.Descriptor instead.
func (*VerifyBlocksResponse) Descriptor() ([]byte, []int) {
	return file_blockchainpb_applicationpb_applicationpb_proto_rawDescGZIP(), []int{3}
}

func (x *VerifyBlocksResponse) GetVerifiedBlocks() []*blockchainpb.Block {
	if x != nil {
		return x.VerifiedBlocks
	}
	return nil
}

type ForkUpdate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RemovedChain         []*blockchainpb.Block `protobuf:"bytes,1,rep,name=removed_chain,json=removedChain,proto3" json:"removed_chain,omitempty"`
	AddedChain           []*blockchainpb.Block `protobuf:"bytes,2,rep,name=added_chain,json=addedChain,proto3" json:"added_chain,omitempty"`
	CheckpointToForkRoot []*blockchainpb.Block `protobuf:"bytes,3,rep,name=checkpoint_to_fork_root,json=checkpointToForkRoot,proto3" json:"checkpoint_to_fork_root,omitempty"`
	CheckpointState      *statepb.State        `protobuf:"bytes,4,opt,name=checkpoint_state,json=checkpointState,proto3" json:"checkpoint_state,omitempty"`
}

func (x *ForkUpdate) Reset() {
	*x = ForkUpdate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ForkUpdate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ForkUpdate) ProtoMessage() {}

func (x *ForkUpdate) ProtoReflect() protoreflect.Message {
	mi := &file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ForkUpdate.ProtoReflect.Descriptor instead.
func (*ForkUpdate) Descriptor() ([]byte, []int) {
	return file_blockchainpb_applicationpb_applicationpb_proto_rawDescGZIP(), []int{4}
}

func (x *ForkUpdate) GetRemovedChain() []*blockchainpb.Block {
	if x != nil {
		return x.RemovedChain
	}
	return nil
}

func (x *ForkUpdate) GetAddedChain() []*blockchainpb.Block {
	if x != nil {
		return x.AddedChain
	}
	return nil
}

func (x *ForkUpdate) GetCheckpointToForkRoot() []*blockchainpb.Block {
	if x != nil {
		return x.CheckpointToForkRoot
	}
	return nil
}

func (x *ForkUpdate) GetCheckpointState() *statepb.State {
	if x != nil {
		return x.CheckpointState
	}
	return nil
}

// Transaction management application events
type PayloadRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HeadId uint64 `protobuf:"varint,1,opt,name=head_id,json=headId,proto3" json:"head_id,omitempty"`
}

func (x *PayloadRequest) Reset() {
	*x = PayloadRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PayloadRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PayloadRequest) ProtoMessage() {}

func (x *PayloadRequest) ProtoReflect() protoreflect.Message {
	mi := &file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PayloadRequest.ProtoReflect.Descriptor instead.
func (*PayloadRequest) Descriptor() ([]byte, []int) {
	return file_blockchainpb_applicationpb_applicationpb_proto_rawDescGZIP(), []int{5}
}

func (x *PayloadRequest) GetHeadId() uint64 {
	if x != nil {
		return x.HeadId
	}
	return 0
}

type PayloadResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HeadId  uint64             `protobuf:"varint,1,opt,name=head_id,json=headId,proto3" json:"head_id,omitempty"`
	Payload *payloadpb.Payload `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (x *PayloadResponse) Reset() {
	*x = PayloadResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PayloadResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PayloadResponse) ProtoMessage() {}

func (x *PayloadResponse) ProtoReflect() protoreflect.Message {
	mi := &file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PayloadResponse.ProtoReflect.Descriptor instead.
func (*PayloadResponse) Descriptor() ([]byte, []int) {
	return file_blockchainpb_applicationpb_applicationpb_proto_rawDescGZIP(), []int{6}
}

func (x *PayloadResponse) GetHeadId() uint64 {
	if x != nil {
		return x.HeadId
	}
	return 0
}

func (x *PayloadResponse) GetPayload() *payloadpb.Payload {
	if x != nil {
		return x.Payload
	}
	return nil
}

type MessageInput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Text string `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
}

func (x *MessageInput) Reset() {
	*x = MessageInput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageInput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageInput) ProtoMessage() {}

func (x *MessageInput) ProtoReflect() protoreflect.Message {
	mi := &file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageInput.ProtoReflect.Descriptor instead.
func (*MessageInput) Descriptor() ([]byte, []int) {
	return file_blockchainpb_applicationpb_applicationpb_proto_rawDescGZIP(), []int{7}
}

func (x *MessageInput) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

var File_blockchainpb_applicationpb_applicationpb_proto protoreflect.FileDescriptor

var file_blockchainpb_applicationpb_applicationpb_proto_rawDesc = []byte{
	0x0a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x70, 0x62, 0x2f, 0x61,
	0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x70, 0x62, 0x2f, 0x61, 0x70, 0x70,
	0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x0d, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x70, 0x62, 0x1a,
	0x1c, 0x6d, 0x69, 0x72, 0x2f, 0x63, 0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e, 0x5f, 0x65, 0x78, 0x74,
	0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x26, 0x62,
	0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x70, 0x62, 0x2f, 0x70, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x70, 0x62, 0x2f, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x70, 0x62, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x22, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69,
	0x6e, 0x70, 0x62, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x70, 0x62, 0x2f, 0x73, 0x74, 0x61, 0x74,
	0x65, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x63, 0x68, 0x61, 0x69, 0x6e, 0x70, 0x62, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61,
	0x69, 0x6e, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x9c, 0x04, 0x0a, 0x05, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x12, 0x33, 0x0a, 0x08, 0x6e, 0x65, 0x77, 0x5f, 0x68, 0x65, 0x61, 0x64,
	0x18, 0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x70, 0x62, 0x2e, 0x4e, 0x65, 0x77, 0x48, 0x65, 0x61, 0x64, 0x48, 0x00,
	0x52, 0x07, 0x6e, 0x65, 0x77, 0x48, 0x65, 0x61, 0x64, 0x12, 0x56, 0x0a, 0x14, 0x76, 0x65, 0x72,
	0x69, 0x66, 0x79, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x70, 0x62, 0x2e, 0x56, 0x65, 0x72, 0x69, 0x66, 0x79, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x12, 0x76,
	0x65, 0x72, 0x69, 0x66, 0x79, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x59, 0x0a, 0x15, 0x76, 0x65, 0x72, 0x69, 0x66, 0x79, 0x5f, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x23, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x70, 0x62,
	0x2e, 0x56, 0x65, 0x72, 0x69, 0x66, 0x79, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x48, 0x00, 0x52, 0x13, 0x76, 0x65, 0x72, 0x69, 0x66, 0x79, 0x42,
	0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x48, 0x0a, 0x0f,
	0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18,
	0x14, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x70, 0x62, 0x2e, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x0e, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x4b, 0x0a, 0x10, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61,
	0x64, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x15, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1e, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x70, 0x62,
	0x2e, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x48, 0x00, 0x52, 0x0f, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x3c, 0x0a, 0x0b, 0x66, 0x6f, 0x72, 0x6b, 0x5f, 0x75, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x18, 0x16, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x70, 0x62, 0x2e, 0x46, 0x6f, 0x72, 0x6b, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x48, 0x00, 0x52, 0x0a, 0x66, 0x6f, 0x72, 0x6b, 0x55, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x12, 0x42, 0x0a, 0x0d, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x69, 0x6e, 0x70,
	0x75, 0x74, 0x18, 0x1e, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x49, 0x6e, 0x70, 0x75, 0x74, 0x48, 0x00, 0x52, 0x0c, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x49, 0x6e, 0x70, 0x75, 0x74, 0x3a, 0x04, 0x90, 0xa6, 0x1d, 0x01, 0x42, 0x0c, 0x0a, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x12, 0x04, 0x80, 0xa6, 0x1d, 0x01, 0x22, 0x28, 0x0a, 0x07, 0x4e, 0x65, 0x77,
	0x48, 0x65, 0x61, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x68, 0x65, 0x61, 0x64, 0x49, 0x64, 0x3a, 0x04, 0x98,
	0xa6, 0x1d, 0x01, 0x22, 0xe3, 0x01, 0x0a, 0x13, 0x56, 0x65, 0x72, 0x69, 0x66, 0x79, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x39, 0x0a, 0x10, 0x63,
	0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x70, 0x62, 0x2e,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x0f, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x4e, 0x0a, 0x19, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f,
	0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x5f, 0x74, 0x6f, 0x5f, 0x73, 0x74,
	0x61, 0x72, 0x74, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x70, 0x62, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x16,
	0x63, 0x68, 0x61, 0x69, 0x6e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x54,
	0x6f, 0x53, 0x74, 0x61, 0x72, 0x74, 0x12, 0x3b, 0x0a, 0x0f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f,
	0x74, 0x6f, 0x5f, 0x76, 0x65, 0x72, 0x69, 0x66, 0x79, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x13, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x70, 0x62, 0x2e, 0x42,
	0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x0d, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x54, 0x6f, 0x56, 0x65, 0x72,
	0x69, 0x66, 0x79, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x22, 0x5a, 0x0a, 0x14, 0x56, 0x65, 0x72,
	0x69, 0x66, 0x79, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x3c, 0x0a, 0x0f, 0x76, 0x65, 0x72, 0x69, 0x66, 0x69, 0x65, 0x64, 0x5f, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x70, 0x62, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52,
	0x0e, 0x76, 0x65, 0x72, 0x69, 0x66, 0x69, 0x65, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x3a,
	0x04, 0x98, 0xa6, 0x1d, 0x01, 0x22, 0x89, 0x02, 0x0a, 0x0a, 0x46, 0x6f, 0x72, 0x6b, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x12, 0x38, 0x0a, 0x0d, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x64, 0x5f,
	0x63, 0x68, 0x61, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x70, 0x62, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b,
	0x52, 0x0c, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x64, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x12, 0x34,
	0x0a, 0x0b, 0x61, 0x64, 0x64, 0x65, 0x64, 0x5f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e,
	0x70, 0x62, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x0a, 0x61, 0x64, 0x64, 0x65, 0x64, 0x43,
	0x68, 0x61, 0x69, 0x6e, 0x12, 0x4a, 0x0a, 0x17, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69,
	0x6e, 0x74, 0x5f, 0x74, 0x6f, 0x5f, 0x66, 0x6f, 0x72, 0x6b, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x18,
	0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61,
	0x69, 0x6e, 0x70, 0x62, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x14, 0x63, 0x68, 0x65, 0x63,
	0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x54, 0x6f, 0x46, 0x6f, 0x72, 0x6b, 0x52, 0x6f, 0x6f, 0x74,
	0x12, 0x39, 0x0a, 0x10, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x5f, 0x73,
	0x74, 0x61, 0x74, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x73, 0x74, 0x61,
	0x74, 0x65, 0x70, 0x62, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x0f, 0x63, 0x68, 0x65, 0x63,
	0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x3a, 0x04, 0x98, 0xa6, 0x1d,
	0x01, 0x22, 0x2f, 0x0a, 0x0e, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x5f, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x68, 0x65, 0x61, 0x64, 0x49, 0x64, 0x3a, 0x04, 0x98, 0xa6,
	0x1d, 0x01, 0x22, 0x5e, 0x0a, 0x0f, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x68, 0x65, 0x61, 0x64, 0x49, 0x64, 0x12, 0x2c,
	0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x12, 0x2e, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x70, 0x62, 0x2e, 0x50, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x3a, 0x04, 0x98, 0xa6,
	0x1d, 0x01, 0x22, 0x28, 0x0a, 0x0c, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x6e, 0x70,
	0x75, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x74, 0x65, 0x78, 0x74, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x42, 0x43, 0x5a, 0x41,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63,
	0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f,
	0x70, 0x6b, 0x67, 0x2f, 0x70, 0x62, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69,
	0x6e, 0x70, 0x62, 0x2f, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x70,
	0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_blockchainpb_applicationpb_applicationpb_proto_rawDescOnce sync.Once
	file_blockchainpb_applicationpb_applicationpb_proto_rawDescData = file_blockchainpb_applicationpb_applicationpb_proto_rawDesc
)

func file_blockchainpb_applicationpb_applicationpb_proto_rawDescGZIP() []byte {
	file_blockchainpb_applicationpb_applicationpb_proto_rawDescOnce.Do(func() {
		file_blockchainpb_applicationpb_applicationpb_proto_rawDescData = protoimpl.X.CompressGZIP(file_blockchainpb_applicationpb_applicationpb_proto_rawDescData)
	})
	return file_blockchainpb_applicationpb_applicationpb_proto_rawDescData
}

var file_blockchainpb_applicationpb_applicationpb_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_blockchainpb_applicationpb_applicationpb_proto_goTypes = []interface{}{
	(*Event)(nil),                // 0: applicationpb.Event
	(*NewHead)(nil),              // 1: applicationpb.NewHead
	(*VerifyBlocksRequest)(nil),  // 2: applicationpb.VerifyBlocksRequest
	(*VerifyBlocksResponse)(nil), // 3: applicationpb.VerifyBlocksResponse
	(*ForkUpdate)(nil),           // 4: applicationpb.ForkUpdate
	(*PayloadRequest)(nil),       // 5: applicationpb.PayloadRequest
	(*PayloadResponse)(nil),      // 6: applicationpb.PayloadResponse
	(*MessageInput)(nil),         // 7: applicationpb.MessageInput
	(*statepb.State)(nil),        // 8: statepb.State
	(*blockchainpb.Block)(nil),   // 9: blockchainpb.Block
	(*payloadpb.Payload)(nil),    // 10: payloadpb.Payload
}
var file_blockchainpb_applicationpb_applicationpb_proto_depIdxs = []int32{
	1,  // 0: applicationpb.Event.new_head:type_name -> applicationpb.NewHead
	2,  // 1: applicationpb.Event.verify_block_request:type_name -> applicationpb.VerifyBlocksRequest
	3,  // 2: applicationpb.Event.verify_block_response:type_name -> applicationpb.VerifyBlocksResponse
	5,  // 3: applicationpb.Event.payload_request:type_name -> applicationpb.PayloadRequest
	6,  // 4: applicationpb.Event.payload_response:type_name -> applicationpb.PayloadResponse
	4,  // 5: applicationpb.Event.fork_update:type_name -> applicationpb.ForkUpdate
	7,  // 6: applicationpb.Event.message_input:type_name -> applicationpb.MessageInput
	8,  // 7: applicationpb.VerifyBlocksRequest.checkpoint_state:type_name -> statepb.State
	9,  // 8: applicationpb.VerifyBlocksRequest.chain_checkpoint_to_start:type_name -> blockchainpb.Block
	9,  // 9: applicationpb.VerifyBlocksRequest.chain_to_verify:type_name -> blockchainpb.Block
	9,  // 10: applicationpb.VerifyBlocksResponse.verified_blocks:type_name -> blockchainpb.Block
	9,  // 11: applicationpb.ForkUpdate.removed_chain:type_name -> blockchainpb.Block
	9,  // 12: applicationpb.ForkUpdate.added_chain:type_name -> blockchainpb.Block
	9,  // 13: applicationpb.ForkUpdate.checkpoint_to_fork_root:type_name -> blockchainpb.Block
	8,  // 14: applicationpb.ForkUpdate.checkpoint_state:type_name -> statepb.State
	10, // 15: applicationpb.PayloadResponse.payload:type_name -> payloadpb.Payload
	16, // [16:16] is the sub-list for method output_type
	16, // [16:16] is the sub-list for method input_type
	16, // [16:16] is the sub-list for extension type_name
	16, // [16:16] is the sub-list for extension extendee
	0,  // [0:16] is the sub-list for field type_name
}

func init() { file_blockchainpb_applicationpb_applicationpb_proto_init() }
func file_blockchainpb_applicationpb_applicationpb_proto_init() {
	if File_blockchainpb_applicationpb_applicationpb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NewHead); i {
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
		file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VerifyBlocksRequest); i {
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
		file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VerifyBlocksResponse); i {
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
		file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ForkUpdate); i {
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
		file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PayloadRequest); i {
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
		file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PayloadResponse); i {
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
		file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageInput); i {
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
	file_blockchainpb_applicationpb_applicationpb_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Event_NewHead)(nil),
		(*Event_VerifyBlockRequest)(nil),
		(*Event_VerifyBlockResponse)(nil),
		(*Event_PayloadRequest)(nil),
		(*Event_PayloadResponse)(nil),
		(*Event_ForkUpdate)(nil),
		(*Event_MessageInput)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_blockchainpb_applicationpb_applicationpb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_blockchainpb_applicationpb_applicationpb_proto_goTypes,
		DependencyIndexes: file_blockchainpb_applicationpb_applicationpb_proto_depIdxs,
		MessageInfos:      file_blockchainpb_applicationpb_applicationpb_proto_msgTypes,
	}.Build()
	File_blockchainpb_applicationpb_applicationpb_proto = out.File
	file_blockchainpb_applicationpb_applicationpb_proto_rawDesc = nil
	file_blockchainpb_applicationpb_applicationpb_proto_goTypes = nil
	file_blockchainpb_applicationpb_applicationpb_proto_depIdxs = nil
}
