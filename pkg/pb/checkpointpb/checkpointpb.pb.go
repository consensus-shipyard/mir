// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.14.0
// source: checkpointpb/checkpointpb.proto

package checkpointpb

import (
	commonpb "github.com/filecoin-project/mir/pkg/pb/commonpb"
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

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//	*Message_Checkpoint
	Type isMessage_Type `protobuf_oneof:"type"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_checkpointpb_checkpointpb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_checkpointpb_checkpointpb_proto_msgTypes[0]
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
	return file_checkpointpb_checkpointpb_proto_rawDescGZIP(), []int{0}
}

func (m *Message) GetType() isMessage_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Message) GetCheckpoint() *Checkpoint {
	if x, ok := x.GetType().(*Message_Checkpoint); ok {
		return x.Checkpoint
	}
	return nil
}

type isMessage_Type interface {
	isMessage_Type()
}

type Message_Checkpoint struct {
	Checkpoint *Checkpoint `protobuf:"bytes,1,opt,name=checkpoint,proto3,oneof"`
}

func (*Message_Checkpoint) isMessage_Type() {}

type Checkpoint struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Epoch        uint64 `protobuf:"varint,1,opt,name=epoch,proto3" json:"epoch,omitempty"`
	Sn           uint64 `protobuf:"varint,2,opt,name=sn,proto3" json:"sn,omitempty"`
	SnapshotHash []byte `protobuf:"bytes,3,opt,name=snapshotHash,proto3" json:"snapshotHash,omitempty"`
	Signature    []byte `protobuf:"bytes,4,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (x *Checkpoint) Reset() {
	*x = Checkpoint{}
	if protoimpl.UnsafeEnabled {
		mi := &file_checkpointpb_checkpointpb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Checkpoint) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Checkpoint) ProtoMessage() {}

func (x *Checkpoint) ProtoReflect() protoreflect.Message {
	mi := &file_checkpointpb_checkpointpb_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Checkpoint.ProtoReflect.Descriptor instead.
func (*Checkpoint) Descriptor() ([]byte, []int) {
	return file_checkpointpb_checkpointpb_proto_rawDescGZIP(), []int{1}
}

func (x *Checkpoint) GetEpoch() uint64 {
	if x != nil {
		return x.Epoch
	}
	return 0
}

func (x *Checkpoint) GetSn() uint64 {
	if x != nil {
		return x.Sn
	}
	return 0
}

func (x *Checkpoint) GetSnapshotHash() []byte {
	if x != nil {
		return x.SnapshotHash
	}
	return nil
}

func (x *Checkpoint) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

type Event struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//	*Event_EpochConfig
	//	*Event_StableCheckpoint
	//	*Event_EpochProgress
	Type isEvent_Type `protobuf_oneof:"type"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_checkpointpb_checkpointpb_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_checkpointpb_checkpointpb_proto_msgTypes[2]
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
	return file_checkpointpb_checkpointpb_proto_rawDescGZIP(), []int{2}
}

func (m *Event) GetType() isEvent_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Event) GetEpochConfig() *commonpb.EpochConfig {
	if x, ok := x.GetType().(*Event_EpochConfig); ok {
		return x.EpochConfig
	}
	return nil
}

func (x *Event) GetStableCheckpoint() *StableCheckpoint {
	if x, ok := x.GetType().(*Event_StableCheckpoint); ok {
		return x.StableCheckpoint
	}
	return nil
}

func (x *Event) GetEpochProgress() *EpochProgress {
	if x, ok := x.GetType().(*Event_EpochProgress); ok {
		return x.EpochProgress
	}
	return nil
}

type isEvent_Type interface {
	isEvent_Type()
}

type Event_EpochConfig struct {
	EpochConfig *commonpb.EpochConfig `protobuf:"bytes,1,opt,name=epoch_config,json=epochConfig,proto3,oneof"`
}

type Event_StableCheckpoint struct {
	StableCheckpoint *StableCheckpoint `protobuf:"bytes,2,opt,name=stable_checkpoint,json=stableCheckpoint,proto3,oneof"`
}

type Event_EpochProgress struct {
	EpochProgress *EpochProgress `protobuf:"bytes,3,opt,name=epoch_progress,json=epochProgress,proto3,oneof"`
}

func (*Event_EpochConfig) isEvent_Type() {}

func (*Event_StableCheckpoint) isEvent_Type() {}

func (*Event_EpochProgress) isEvent_Type() {}

type StableCheckpoint struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sn       uint64                  `protobuf:"varint,1,opt,name=sn,proto3" json:"sn,omitempty"`
	Snapshot *commonpb.StateSnapshot `protobuf:"bytes,2,opt,name=snapshot,proto3" json:"snapshot,omitempty"`
	Cert     map[string][]byte       `protobuf:"bytes,3,rep,name=cert,proto3" json:"cert,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *StableCheckpoint) Reset() {
	*x = StableCheckpoint{}
	if protoimpl.UnsafeEnabled {
		mi := &file_checkpointpb_checkpointpb_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StableCheckpoint) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StableCheckpoint) ProtoMessage() {}

func (x *StableCheckpoint) ProtoReflect() protoreflect.Message {
	mi := &file_checkpointpb_checkpointpb_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StableCheckpoint.ProtoReflect.Descriptor instead.
func (*StableCheckpoint) Descriptor() ([]byte, []int) {
	return file_checkpointpb_checkpointpb_proto_rawDescGZIP(), []int{3}
}

func (x *StableCheckpoint) GetSn() uint64 {
	if x != nil {
		return x.Sn
	}
	return 0
}

func (x *StableCheckpoint) GetSnapshot() *commonpb.StateSnapshot {
	if x != nil {
		return x.Snapshot
	}
	return nil
}

func (x *StableCheckpoint) GetCert() map[string][]byte {
	if x != nil {
		return x.Cert
	}
	return nil
}

type EpochProgress struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeId string `protobuf:"bytes,1,opt,name=node_id,json=nodeId,proto3" json:"node_id,omitempty"`
	Epoch  uint64 `protobuf:"varint,2,opt,name=epoch,proto3" json:"epoch,omitempty"`
}

func (x *EpochProgress) Reset() {
	*x = EpochProgress{}
	if protoimpl.UnsafeEnabled {
		mi := &file_checkpointpb_checkpointpb_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EpochProgress) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EpochProgress) ProtoMessage() {}

func (x *EpochProgress) ProtoReflect() protoreflect.Message {
	mi := &file_checkpointpb_checkpointpb_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EpochProgress.ProtoReflect.Descriptor instead.
func (*EpochProgress) Descriptor() ([]byte, []int) {
	return file_checkpointpb_checkpointpb_proto_rawDescGZIP(), []int{4}
}

func (x *EpochProgress) GetNodeId() string {
	if x != nil {
		return x.NodeId
	}
	return ""
}

func (x *EpochProgress) GetEpoch() uint64 {
	if x != nil {
		return x.Epoch
	}
	return 0
}

type HashOrigin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *HashOrigin) Reset() {
	*x = HashOrigin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_checkpointpb_checkpointpb_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HashOrigin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HashOrigin) ProtoMessage() {}

func (x *HashOrigin) ProtoReflect() protoreflect.Message {
	mi := &file_checkpointpb_checkpointpb_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HashOrigin.ProtoReflect.Descriptor instead.
func (*HashOrigin) Descriptor() ([]byte, []int) {
	return file_checkpointpb_checkpointpb_proto_rawDescGZIP(), []int{5}
}

type SignOrigin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SignOrigin) Reset() {
	*x = SignOrigin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_checkpointpb_checkpointpb_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignOrigin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignOrigin) ProtoMessage() {}

func (x *SignOrigin) ProtoReflect() protoreflect.Message {
	mi := &file_checkpointpb_checkpointpb_proto_msgTypes[6]
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
	return file_checkpointpb_checkpointpb_proto_rawDescGZIP(), []int{6}
}

type SigVerOrigin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SigVerOrigin) Reset() {
	*x = SigVerOrigin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_checkpointpb_checkpointpb_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SigVerOrigin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SigVerOrigin) ProtoMessage() {}

func (x *SigVerOrigin) ProtoReflect() protoreflect.Message {
	mi := &file_checkpointpb_checkpointpb_proto_msgTypes[7]
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
	return file_checkpointpb_checkpointpb_proto_rawDescGZIP(), []int{7}
}

type InstanceParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Membership       *commonpb.Membership  `protobuf:"bytes,1,opt,name=membership,proto3" json:"membership,omitempty"`
	ResendPeriod     uint64                `protobuf:"varint,2,opt,name=resend_period,json=resendPeriod,proto3" json:"resend_period,omitempty"`
	LeaderPolicyData []byte                `protobuf:"bytes,3,opt,name=leader_policy_data,json=leaderPolicyData,proto3" json:"leader_policy_data,omitempty"`
	EpochConfig      *commonpb.EpochConfig `protobuf:"bytes,4,opt,name=epoch_config,json=epochConfig,proto3" json:"epoch_config,omitempty"`
}

func (x *InstanceParams) Reset() {
	*x = InstanceParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_checkpointpb_checkpointpb_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InstanceParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InstanceParams) ProtoMessage() {}

func (x *InstanceParams) ProtoReflect() protoreflect.Message {
	mi := &file_checkpointpb_checkpointpb_proto_msgTypes[8]
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
	return file_checkpointpb_checkpointpb_proto_rawDescGZIP(), []int{8}
}

func (x *InstanceParams) GetMembership() *commonpb.Membership {
	if x != nil {
		return x.Membership
	}
	return nil
}

func (x *InstanceParams) GetResendPeriod() uint64 {
	if x != nil {
		return x.ResendPeriod
	}
	return 0
}

func (x *InstanceParams) GetLeaderPolicyData() []byte {
	if x != nil {
		return x.LeaderPolicyData
	}
	return nil
}

func (x *InstanceParams) GetEpochConfig() *commonpb.EpochConfig {
	if x != nil {
		return x.EpochConfig
	}
	return nil
}

var File_checkpointpb_checkpointpb_proto protoreflect.FileDescriptor

var file_checkpointpb_checkpointpb_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x70, 0x62, 0x2f, 0x63,
	0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x0c, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x70, 0x62, 0x1a,
	0x17, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x70, 0x62, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6d, 0x69, 0x72, 0x2f, 0x63, 0x6f,
	0x64, 0x65, 0x67, 0x65, 0x6e, 0x5f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6e, 0x65, 0x74, 0x2f, 0x63, 0x6f, 0x64, 0x65,
	0x67, 0x65, 0x6e, 0x5f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x59, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12,
	0x3a, 0x0a, 0x0a, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74,
	0x70, 0x62, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x48, 0x00, 0x52,
	0x0a, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x3a, 0x04, 0xc8, 0xe4, 0x1d,
	0x01, 0x42, 0x0c, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x04, 0xc8, 0xe4, 0x1d, 0x01, 0x22,
	0xf6, 0x01, 0x0a, 0x0a, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x53,
	0x0a, 0x05, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x42, 0x3d, 0x82,
	0xa6, 0x1d, 0x39, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69,
	0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d,
	0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x74, 0x6f, 0x72, 0x2f, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x2e, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x4e, 0x72, 0x52, 0x05, 0x65, 0x70,
	0x6f, 0x63, 0x68, 0x12, 0x4b, 0x0a, 0x02, 0x73, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x42,
	0x3b, 0x82, 0xa6, 0x1d, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x74, 0x6f, 0x72,
	0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x53, 0x65, 0x71, 0x4e, 0x72, 0x52, 0x02, 0x73, 0x6e,
	0x12, 0x22, 0x0a, 0x0c, 0x73, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x48, 0x61, 0x73, 0x68,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x73, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74,
	0x48, 0x61, 0x73, 0x68, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x3a, 0x04, 0xd0, 0xe4, 0x1d, 0x01, 0x22, 0xec, 0x01, 0x0a, 0x05, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x12, 0x3a, 0x0a, 0x0c, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x5f, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x70, 0x62, 0x2e, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x48,
	0x00, 0x52, 0x0b, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x4d,
	0x0a, 0x11, 0x73, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f,
	0x69, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x63, 0x68, 0x65, 0x63,
	0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x53, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x43,
	0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x10, 0x73, 0x74, 0x61,
	0x62, 0x6c, 0x65, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x44, 0x0a,
	0x0e, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x5f, 0x70, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69,
	0x6e, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x50, 0x72, 0x6f, 0x67, 0x72, 0x65,
	0x73, 0x73, 0x48, 0x00, 0x52, 0x0d, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x50, 0x72, 0x6f, 0x67, 0x72,
	0x65, 0x73, 0x73, 0x3a, 0x04, 0x90, 0xa6, 0x1d, 0x01, 0x42, 0x0c, 0x0a, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x12, 0x04, 0x80, 0xa6, 0x1d, 0x01, 0x22, 0xcb, 0x02, 0x0a, 0x10, 0x53, 0x74, 0x61, 0x62,
	0x6c, 0x65, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x4b, 0x0a, 0x02,
	0x73, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x42, 0x3b, 0x82, 0xa6, 0x1d, 0x37, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69,
	0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b,
	0x67, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x74, 0x6f, 0x72, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e,
	0x53, 0x65, 0x71, 0x4e, 0x72, 0x52, 0x02, 0x73, 0x6e, 0x12, 0x33, 0x0a, 0x08, 0x73, 0x6e, 0x61,
	0x70, 0x73, 0x68, 0x6f, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x70, 0x62, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x53, 0x6e, 0x61, 0x70,
	0x73, 0x68, 0x6f, 0x74, 0x52, 0x08, 0x73, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x12, 0x72,
	0x0a, 0x04, 0x63, 0x65, 0x72, 0x74, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x63,
	0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x53, 0x74, 0x61, 0x62,
	0x6c, 0x65, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x2e, 0x43, 0x65, 0x72,
	0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x42, 0x34, 0xaa, 0xa6, 0x1d, 0x30, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d,
	0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f,
	0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x49, 0x44, 0x52, 0x04, 0x63, 0x65,
	0x72, 0x74, 0x1a, 0x37, 0x0a, 0x09, 0x43, 0x65, 0x72, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x3a, 0x08, 0x98, 0xa6, 0x1d,
	0x01, 0xd0, 0xe4, 0x1d, 0x01, 0x22, 0xb9, 0x01, 0x0a, 0x0d, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x50,
	0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x12, 0x4d, 0x0a, 0x07, 0x6e, 0x6f, 0x64, 0x65, 0x5f,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x34, 0x82, 0xa6, 0x1d, 0x30, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69,
	0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b,
	0x67, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x49, 0x44, 0x52, 0x06,
	0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x12, 0x53, 0x0a, 0x05, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x04, 0x42, 0x3d, 0x82, 0xa6, 0x1d, 0x39, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70,
	0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74,
	0x72, 0x61, 0x6e, 0x74, 0x6f, 0x72, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x45, 0x70, 0x6f,
	0x63, 0x68, 0x4e, 0x72, 0x52, 0x05, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x3a, 0x04, 0x98, 0xa6, 0x1d,
	0x01, 0x22, 0x12, 0x0a, 0x0a, 0x48, 0x61, 0x73, 0x68, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x3a,
	0x04, 0x80, 0xa6, 0x1d, 0x01, 0x22, 0x12, 0x0a, 0x0a, 0x53, 0x69, 0x67, 0x6e, 0x4f, 0x72, 0x69,
	0x67, 0x69, 0x6e, 0x3a, 0x04, 0x80, 0xa6, 0x1d, 0x01, 0x22, 0x14, 0x0a, 0x0c, 0x53, 0x69, 0x67,
	0x56, 0x65, 0x72, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x3a, 0x04, 0x80, 0xa6, 0x1d, 0x01, 0x22,
	0x97, 0x02, 0x0a, 0x0e, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x50, 0x61, 0x72, 0x61,
	0x6d, 0x73, 0x12, 0x34, 0x0a, 0x0a, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x70,
	0x62, 0x2e, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x52, 0x0a, 0x6d, 0x65,
	0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x12, 0x61, 0x0a, 0x0d, 0x72, 0x65, 0x73, 0x65,
	0x6e, 0x64, 0x5f, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x42,
	0x3c, 0x82, 0xa6, 0x1d, 0x38, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x72, 0x2f, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0c, 0x72,
	0x65, 0x73, 0x65, 0x6e, 0x64, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x12, 0x2c, 0x0a, 0x12, 0x6c,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x5f, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x5f, 0x64, 0x61, 0x74,
	0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x10, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x50,
	0x6f, 0x6c, 0x69, 0x63, 0x79, 0x44, 0x61, 0x74, 0x61, 0x12, 0x38, 0x0a, 0x0c, 0x65, 0x70, 0x6f,
	0x63, 0x68, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x15, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x70, 0x62, 0x2e, 0x45, 0x70, 0x6f, 0x63, 0x68,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x0b, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x3a, 0x04, 0x80, 0xa6, 0x1d, 0x01, 0x42, 0x35, 0x5a, 0x33, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e,
	0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67,
	0x2f, 0x70, 0x62, 0x2f, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x70, 0x62,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_checkpointpb_checkpointpb_proto_rawDescOnce sync.Once
	file_checkpointpb_checkpointpb_proto_rawDescData = file_checkpointpb_checkpointpb_proto_rawDesc
)

func file_checkpointpb_checkpointpb_proto_rawDescGZIP() []byte {
	file_checkpointpb_checkpointpb_proto_rawDescOnce.Do(func() {
		file_checkpointpb_checkpointpb_proto_rawDescData = protoimpl.X.CompressGZIP(file_checkpointpb_checkpointpb_proto_rawDescData)
	})
	return file_checkpointpb_checkpointpb_proto_rawDescData
}

var file_checkpointpb_checkpointpb_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_checkpointpb_checkpointpb_proto_goTypes = []interface{}{
	(*Message)(nil),                // 0: checkpointpb.Message
	(*Checkpoint)(nil),             // 1: checkpointpb.Checkpoint
	(*Event)(nil),                  // 2: checkpointpb.Event
	(*StableCheckpoint)(nil),       // 3: checkpointpb.StableCheckpoint
	(*EpochProgress)(nil),          // 4: checkpointpb.EpochProgress
	(*HashOrigin)(nil),             // 5: checkpointpb.HashOrigin
	(*SignOrigin)(nil),             // 6: checkpointpb.SignOrigin
	(*SigVerOrigin)(nil),           // 7: checkpointpb.SigVerOrigin
	(*InstanceParams)(nil),         // 8: checkpointpb.InstanceParams
	nil,                            // 9: checkpointpb.StableCheckpoint.CertEntry
	(*commonpb.EpochConfig)(nil),   // 10: commonpb.EpochConfig
	(*commonpb.StateSnapshot)(nil), // 11: commonpb.StateSnapshot
	(*commonpb.Membership)(nil),    // 12: commonpb.Membership
}
var file_checkpointpb_checkpointpb_proto_depIdxs = []int32{
	1,  // 0: checkpointpb.Message.checkpoint:type_name -> checkpointpb.Checkpoint
	10, // 1: checkpointpb.Event.epoch_config:type_name -> commonpb.EpochConfig
	3,  // 2: checkpointpb.Event.stable_checkpoint:type_name -> checkpointpb.StableCheckpoint
	4,  // 3: checkpointpb.Event.epoch_progress:type_name -> checkpointpb.EpochProgress
	11, // 4: checkpointpb.StableCheckpoint.snapshot:type_name -> commonpb.StateSnapshot
	9,  // 5: checkpointpb.StableCheckpoint.cert:type_name -> checkpointpb.StableCheckpoint.CertEntry
	12, // 6: checkpointpb.InstanceParams.membership:type_name -> commonpb.Membership
	10, // 7: checkpointpb.InstanceParams.epoch_config:type_name -> commonpb.EpochConfig
	8,  // [8:8] is the sub-list for method output_type
	8,  // [8:8] is the sub-list for method input_type
	8,  // [8:8] is the sub-list for extension type_name
	8,  // [8:8] is the sub-list for extension extendee
	0,  // [0:8] is the sub-list for field type_name
}

func init() { file_checkpointpb_checkpointpb_proto_init() }
func file_checkpointpb_checkpointpb_proto_init() {
	if File_checkpointpb_checkpointpb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_checkpointpb_checkpointpb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_checkpointpb_checkpointpb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Checkpoint); i {
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
		file_checkpointpb_checkpointpb_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
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
		file_checkpointpb_checkpointpb_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StableCheckpoint); i {
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
		file_checkpointpb_checkpointpb_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EpochProgress); i {
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
		file_checkpointpb_checkpointpb_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HashOrigin); i {
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
		file_checkpointpb_checkpointpb_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
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
		file_checkpointpb_checkpointpb_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
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
		file_checkpointpb_checkpointpb_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
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
	file_checkpointpb_checkpointpb_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Message_Checkpoint)(nil),
	}
	file_checkpointpb_checkpointpb_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*Event_EpochConfig)(nil),
		(*Event_StableCheckpoint)(nil),
		(*Event_EpochProgress)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_checkpointpb_checkpointpb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_checkpointpb_checkpointpb_proto_goTypes,
		DependencyIndexes: file_checkpointpb_checkpointpb_proto_depIdxs,
		MessageInfos:      file_checkpointpb_checkpointpb_proto_msgTypes,
	}.Build()
	File_checkpointpb_checkpointpb_proto = out.File
	file_checkpointpb_checkpointpb_proto_rawDesc = nil
	file_checkpointpb_checkpointpb_proto_goTypes = nil
	file_checkpointpb_checkpointpb_proto_depIdxs = nil
}
