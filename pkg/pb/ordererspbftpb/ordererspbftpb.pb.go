//
//Copyright IBM Corp. All Rights Reserved.
//
//SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.12.4
// source: ordererspbftpb/ordererspbftpb.proto

package ordererspbftpb

import (
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

type Preprepare struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sn      uint64 `protobuf:"varint,1,opt,name=sn,proto3" json:"sn,omitempty"`
	View    uint64 `protobuf:"varint,2,opt,name=view,proto3" json:"view,omitempty"`
	Data    []byte `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	Aborted bool   `protobuf:"varint,4,opt,name=aborted,proto3" json:"aborted,omitempty"`
}

func (x *Preprepare) Reset() {
	*x = Preprepare{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Preprepare) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Preprepare) ProtoMessage() {}

func (x *Preprepare) ProtoReflect() protoreflect.Message {
	mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Preprepare.ProtoReflect.Descriptor instead.
func (*Preprepare) Descriptor() ([]byte, []int) {
	return file_ordererspbftpb_ordererspbftpb_proto_rawDescGZIP(), []int{0}
}

func (x *Preprepare) GetSn() uint64 {
	if x != nil {
		return x.Sn
	}
	return 0
}

func (x *Preprepare) GetView() uint64 {
	if x != nil {
		return x.View
	}
	return 0
}

func (x *Preprepare) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *Preprepare) GetAborted() bool {
	if x != nil {
		return x.Aborted
	}
	return false
}

type Prepare struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sn     uint64 `protobuf:"varint,1,opt,name=sn,proto3" json:"sn,omitempty"`
	View   uint64 `protobuf:"varint,2,opt,name=view,proto3" json:"view,omitempty"`
	Digest []byte `protobuf:"bytes,3,opt,name=digest,proto3" json:"digest,omitempty"`
}

func (x *Prepare) Reset() {
	*x = Prepare{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Prepare) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Prepare) ProtoMessage() {}

func (x *Prepare) ProtoReflect() protoreflect.Message {
	mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Prepare.ProtoReflect.Descriptor instead.
func (*Prepare) Descriptor() ([]byte, []int) {
	return file_ordererspbftpb_ordererspbftpb_proto_rawDescGZIP(), []int{1}
}

func (x *Prepare) GetSn() uint64 {
	if x != nil {
		return x.Sn
	}
	return 0
}

func (x *Prepare) GetView() uint64 {
	if x != nil {
		return x.View
	}
	return 0
}

func (x *Prepare) GetDigest() []byte {
	if x != nil {
		return x.Digest
	}
	return nil
}

type Commit struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sn     uint64 `protobuf:"varint,1,opt,name=sn,proto3" json:"sn,omitempty"`
	View   uint64 `protobuf:"varint,2,opt,name=view,proto3" json:"view,omitempty"`
	Digest []byte `protobuf:"bytes,3,opt,name=digest,proto3" json:"digest,omitempty"`
}

func (x *Commit) Reset() {
	*x = Commit{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Commit) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Commit) ProtoMessage() {}

func (x *Commit) ProtoReflect() protoreflect.Message {
	mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Commit.ProtoReflect.Descriptor instead.
func (*Commit) Descriptor() ([]byte, []int) {
	return file_ordererspbftpb_ordererspbftpb_proto_rawDescGZIP(), []int{2}
}

func (x *Commit) GetSn() uint64 {
	if x != nil {
		return x.Sn
	}
	return 0
}

func (x *Commit) GetView() uint64 {
	if x != nil {
		return x.View
	}
	return 0
}

func (x *Commit) GetDigest() []byte {
	if x != nil {
		return x.Digest
	}
	return nil
}

type Done struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Digests [][]byte `protobuf:"bytes,1,rep,name=digests,proto3" json:"digests,omitempty"`
}

func (x *Done) Reset() {
	*x = Done{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Done) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Done) ProtoMessage() {}

func (x *Done) ProtoReflect() protoreflect.Message {
	mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Done.ProtoReflect.Descriptor instead.
func (*Done) Descriptor() ([]byte, []int) {
	return file_ordererspbftpb_ordererspbftpb_proto_rawDescGZIP(), []int{3}
}

func (x *Done) GetDigests() [][]byte {
	if x != nil {
		return x.Digests
	}
	return nil
}

type CatchUpRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Digest []byte `protobuf:"bytes,1,opt,name=digest,proto3" json:"digest,omitempty"`
	Sn     uint64 `protobuf:"varint,2,opt,name=sn,proto3" json:"sn,omitempty"`
}

func (x *CatchUpRequest) Reset() {
	*x = CatchUpRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CatchUpRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CatchUpRequest) ProtoMessage() {}

func (x *CatchUpRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CatchUpRequest.ProtoReflect.Descriptor instead.
func (*CatchUpRequest) Descriptor() ([]byte, []int) {
	return file_ordererspbftpb_ordererspbftpb_proto_rawDescGZIP(), []int{4}
}

func (x *CatchUpRequest) GetDigest() []byte {
	if x != nil {
		return x.Digest
	}
	return nil
}

func (x *CatchUpRequest) GetSn() uint64 {
	if x != nil {
		return x.Sn
	}
	return 0
}

type ViewChange struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	View uint64       `protobuf:"varint,1,opt,name=view,proto3" json:"view,omitempty"`
	PSet []*PSetEntry `protobuf:"bytes,2,rep,name=p_set,json=pSet,proto3" json:"p_set,omitempty"`
	QSet []*QSetEntry `protobuf:"bytes,3,rep,name=q_set,json=qSet,proto3" json:"q_set,omitempty"`
}

func (x *ViewChange) Reset() {
	*x = ViewChange{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ViewChange) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ViewChange) ProtoMessage() {}

func (x *ViewChange) ProtoReflect() protoreflect.Message {
	mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ViewChange.ProtoReflect.Descriptor instead.
func (*ViewChange) Descriptor() ([]byte, []int) {
	return file_ordererspbftpb_ordererspbftpb_proto_rawDescGZIP(), []int{5}
}

func (x *ViewChange) GetView() uint64 {
	if x != nil {
		return x.View
	}
	return 0
}

func (x *ViewChange) GetPSet() []*PSetEntry {
	if x != nil {
		return x.PSet
	}
	return nil
}

func (x *ViewChange) GetQSet() []*QSetEntry {
	if x != nil {
		return x.QSet
	}
	return nil
}

type SignedViewChange struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ViewChange *ViewChange `protobuf:"bytes,1,opt,name=view_change,json=viewChange,proto3" json:"view_change,omitempty"`
	Signature  []byte      `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (x *SignedViewChange) Reset() {
	*x = SignedViewChange{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignedViewChange) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignedViewChange) ProtoMessage() {}

func (x *SignedViewChange) ProtoReflect() protoreflect.Message {
	mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignedViewChange.ProtoReflect.Descriptor instead.
func (*SignedViewChange) Descriptor() ([]byte, []int) {
	return file_ordererspbftpb_ordererspbftpb_proto_rawDescGZIP(), []int{6}
}

func (x *SignedViewChange) GetViewChange() *ViewChange {
	if x != nil {
		return x.ViewChange
	}
	return nil
}

func (x *SignedViewChange) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

type NewView struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	View              uint64              `protobuf:"varint,1,opt,name=view,proto3" json:"view,omitempty"`
	ViewChangeSenders []string            `protobuf:"bytes,3,rep,name=view_change_senders,json=viewChangeSenders,proto3" json:"view_change_senders,omitempty"`
	SignedViewChanges []*SignedViewChange `protobuf:"bytes,2,rep,name=signed_view_changes,json=signedViewChanges,proto3" json:"signed_view_changes,omitempty"`
	PreprepareSeqNrs  []uint64            `protobuf:"varint,4,rep,packed,name=preprepare_seq_nrs,json=preprepareSeqNrs,proto3" json:"preprepare_seq_nrs,omitempty"`
	Preprepares       []*Preprepare       `protobuf:"bytes,5,rep,name=preprepares,proto3" json:"preprepares,omitempty"`
}

func (x *NewView) Reset() {
	*x = NewView{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NewView) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewView) ProtoMessage() {}

func (x *NewView) ProtoReflect() protoreflect.Message {
	mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewView.ProtoReflect.Descriptor instead.
func (*NewView) Descriptor() ([]byte, []int) {
	return file_ordererspbftpb_ordererspbftpb_proto_rawDescGZIP(), []int{7}
}

func (x *NewView) GetView() uint64 {
	if x != nil {
		return x.View
	}
	return 0
}

func (x *NewView) GetViewChangeSenders() []string {
	if x != nil {
		return x.ViewChangeSenders
	}
	return nil
}

func (x *NewView) GetSignedViewChanges() []*SignedViewChange {
	if x != nil {
		return x.SignedViewChanges
	}
	return nil
}

func (x *NewView) GetPreprepareSeqNrs() []uint64 {
	if x != nil {
		return x.PreprepareSeqNrs
	}
	return nil
}

func (x *NewView) GetPreprepares() []*Preprepare {
	if x != nil {
		return x.Preprepares
	}
	return nil
}

type PSetEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sn     uint64 `protobuf:"varint,1,opt,name=sn,proto3" json:"sn,omitempty"`
	View   uint64 `protobuf:"varint,2,opt,name=view,proto3" json:"view,omitempty"`
	Digest []byte `protobuf:"bytes,3,opt,name=digest,proto3" json:"digest,omitempty"`
}

func (x *PSetEntry) Reset() {
	*x = PSetEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PSetEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PSetEntry) ProtoMessage() {}

func (x *PSetEntry) ProtoReflect() protoreflect.Message {
	mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PSetEntry.ProtoReflect.Descriptor instead.
func (*PSetEntry) Descriptor() ([]byte, []int) {
	return file_ordererspbftpb_ordererspbftpb_proto_rawDescGZIP(), []int{8}
}

func (x *PSetEntry) GetSn() uint64 {
	if x != nil {
		return x.Sn
	}
	return 0
}

func (x *PSetEntry) GetView() uint64 {
	if x != nil {
		return x.View
	}
	return 0
}

func (x *PSetEntry) GetDigest() []byte {
	if x != nil {
		return x.Digest
	}
	return nil
}

type QSetEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sn     uint64 `protobuf:"varint,1,opt,name=sn,proto3" json:"sn,omitempty"`
	View   uint64 `protobuf:"varint,2,opt,name=view,proto3" json:"view,omitempty"`
	Digest []byte `protobuf:"bytes,3,opt,name=digest,proto3" json:"digest,omitempty"`
}

func (x *QSetEntry) Reset() {
	*x = QSetEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QSetEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QSetEntry) ProtoMessage() {}

func (x *QSetEntry) ProtoReflect() protoreflect.Message {
	mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QSetEntry.ProtoReflect.Descriptor instead.
func (*QSetEntry) Descriptor() ([]byte, []int) {
	return file_ordererspbftpb_ordererspbftpb_proto_rawDescGZIP(), []int{9}
}

func (x *QSetEntry) GetSn() uint64 {
	if x != nil {
		return x.Sn
	}
	return 0
}

func (x *QSetEntry) GetView() uint64 {
	if x != nil {
		return x.View
	}
	return 0
}

func (x *QSetEntry) GetDigest() []byte {
	if x != nil {
		return x.Digest
	}
	return nil
}

type PreprepareRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Digest []byte `protobuf:"bytes,1,opt,name=digest,proto3" json:"digest,omitempty"`
	Sn     uint64 `protobuf:"varint,2,opt,name=sn,proto3" json:"sn,omitempty"`
}

func (x *PreprepareRequest) Reset() {
	*x = PreprepareRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PreprepareRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PreprepareRequest) ProtoMessage() {}

func (x *PreprepareRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PreprepareRequest.ProtoReflect.Descriptor instead.
func (*PreprepareRequest) Descriptor() ([]byte, []int) {
	return file_ordererspbftpb_ordererspbftpb_proto_rawDescGZIP(), []int{10}
}

func (x *PreprepareRequest) GetDigest() []byte {
	if x != nil {
		return x.Digest
	}
	return nil
}

func (x *PreprepareRequest) GetSn() uint64 {
	if x != nil {
		return x.Sn
	}
	return 0
}

type ReqWaitReference struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sn   uint64 `protobuf:"varint,1,opt,name=sn,proto3" json:"sn,omitempty"`
	View uint64 `protobuf:"varint,2,opt,name=view,proto3" json:"view,omitempty"`
}

func (x *ReqWaitReference) Reset() {
	*x = ReqWaitReference{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReqWaitReference) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReqWaitReference) ProtoMessage() {}

func (x *ReqWaitReference) ProtoReflect() protoreflect.Message {
	mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReqWaitReference.ProtoReflect.Descriptor instead.
func (*ReqWaitReference) Descriptor() ([]byte, []int) {
	return file_ordererspbftpb_ordererspbftpb_proto_rawDescGZIP(), []int{11}
}

func (x *ReqWaitReference) GetSn() uint64 {
	if x != nil {
		return x.Sn
	}
	return 0
}

func (x *ReqWaitReference) GetView() uint64 {
	if x != nil {
		return x.View
	}
	return 0
}

type VCSNTimeout struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	View         uint64 `protobuf:"varint,1,opt,name=view,proto3" json:"view,omitempty"`
	NumCommitted uint64 `protobuf:"varint,2,opt,name=numCommitted,proto3" json:"numCommitted,omitempty"`
}

func (x *VCSNTimeout) Reset() {
	*x = VCSNTimeout{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VCSNTimeout) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VCSNTimeout) ProtoMessage() {}

func (x *VCSNTimeout) ProtoReflect() protoreflect.Message {
	mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VCSNTimeout.ProtoReflect.Descriptor instead.
func (*VCSNTimeout) Descriptor() ([]byte, []int) {
	return file_ordererspbftpb_ordererspbftpb_proto_rawDescGZIP(), []int{12}
}

func (x *VCSNTimeout) GetView() uint64 {
	if x != nil {
		return x.View
	}
	return 0
}

func (x *VCSNTimeout) GetNumCommitted() uint64 {
	if x != nil {
		return x.NumCommitted
	}
	return 0
}

type Status struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Status) Reset() {
	*x = Status{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[13]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Status) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Status) ProtoMessage() {}

func (x *Status) ProtoReflect() protoreflect.Message {
	mi := &file_ordererspbftpb_ordererspbftpb_proto_msgTypes[13]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Status.ProtoReflect.Descriptor instead.
func (*Status) Descriptor() ([]byte, []int) {
	return file_ordererspbftpb_ordererspbftpb_proto_rawDescGZIP(), []int{13}
}

var File_ordererspbftpb_ordererspbftpb_proto protoreflect.FileDescriptor

var file_ordererspbftpb_ordererspbftpb_proto_rawDesc = []byte{
	0x0a, 0x23, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x72, 0x73, 0x70, 0x62, 0x66, 0x74, 0x70, 0x62,
	0x2f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x72, 0x73, 0x70, 0x62, 0x66, 0x74, 0x70, 0x62, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x72, 0x73, 0x70,
	0x62, 0x66, 0x74, 0x70, 0x62, 0x22, 0x5e, 0x0a, 0x0a, 0x50, 0x72, 0x65, 0x70, 0x72, 0x65, 0x70,
	0x61, 0x72, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x73, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x02, 0x73, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x76, 0x69, 0x65, 0x77, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x04, 0x76, 0x69, 0x65, 0x77, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x18, 0x0a, 0x07, 0x61,
	0x62, 0x6f, 0x72, 0x74, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x61, 0x62,
	0x6f, 0x72, 0x74, 0x65, 0x64, 0x22, 0x45, 0x0a, 0x07, 0x50, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65,
	0x12, 0x0e, 0x0a, 0x02, 0x73, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x02, 0x73, 0x6e,
	0x12, 0x12, 0x0a, 0x04, 0x76, 0x69, 0x65, 0x77, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x04,
	0x76, 0x69, 0x65, 0x77, 0x12, 0x16, 0x0a, 0x06, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74, 0x22, 0x44, 0x0a, 0x06,
	0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x73, 0x6e, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x02, 0x73, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x76, 0x69, 0x65, 0x77, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x04, 0x52, 0x04, 0x76, 0x69, 0x65, 0x77, 0x12, 0x16, 0x0a, 0x06, 0x64, 0x69,
	0x67, 0x65, 0x73, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x64, 0x69, 0x67, 0x65,
	0x73, 0x74, 0x22, 0x20, 0x0a, 0x04, 0x44, 0x6f, 0x6e, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x64, 0x69,
	0x67, 0x65, 0x73, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x07, 0x64, 0x69, 0x67,
	0x65, 0x73, 0x74, 0x73, 0x22, 0x38, 0x0a, 0x0e, 0x43, 0x61, 0x74, 0x63, 0x68, 0x55, 0x70, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74, 0x12, 0x0e,
	0x0a, 0x02, 0x73, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x02, 0x73, 0x6e, 0x22, 0x80,
	0x01, 0x0a, 0x0a, 0x56, 0x69, 0x65, 0x77, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x12, 0x0a,
	0x04, 0x76, 0x69, 0x65, 0x77, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x04, 0x76, 0x69, 0x65,
	0x77, 0x12, 0x2e, 0x0a, 0x05, 0x70, 0x5f, 0x73, 0x65, 0x74, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x19, 0x2e, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x72, 0x73, 0x70, 0x62, 0x66, 0x74, 0x70,
	0x62, 0x2e, 0x50, 0x53, 0x65, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x70, 0x53, 0x65,
	0x74, 0x12, 0x2e, 0x0a, 0x05, 0x71, 0x5f, 0x73, 0x65, 0x74, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x19, 0x2e, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x72, 0x73, 0x70, 0x62, 0x66, 0x74, 0x70,
	0x62, 0x2e, 0x51, 0x53, 0x65, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x71, 0x53, 0x65,
	0x74, 0x22, 0x6d, 0x0a, 0x10, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x56, 0x69, 0x65, 0x77, 0x43,
	0x68, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x3b, 0x0a, 0x0b, 0x76, 0x69, 0x65, 0x77, 0x5f, 0x63, 0x68,
	0x61, 0x6e, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6f, 0x72, 0x64,
	0x65, 0x72, 0x65, 0x72, 0x73, 0x70, 0x62, 0x66, 0x74, 0x70, 0x62, 0x2e, 0x56, 0x69, 0x65, 0x77,
	0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x52, 0x0a, 0x76, 0x69, 0x65, 0x77, 0x43, 0x68, 0x61, 0x6e,
	0x67, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x22, 0x8b, 0x02, 0x0a, 0x07, 0x4e, 0x65, 0x77, 0x56, 0x69, 0x65, 0x77, 0x12, 0x12, 0x0a, 0x04,
	0x76, 0x69, 0x65, 0x77, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x04, 0x76, 0x69, 0x65, 0x77,
	0x12, 0x2e, 0x0a, 0x13, 0x76, 0x69, 0x65, 0x77, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x5f,
	0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x11, 0x76,
	0x69, 0x65, 0x77, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x53, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x73,
	0x12, 0x50, 0x0a, 0x13, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x5f, 0x76, 0x69, 0x65, 0x77, 0x5f,
	0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e,
	0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x72, 0x73, 0x70, 0x62, 0x66, 0x74, 0x70, 0x62, 0x2e, 0x53,
	0x69, 0x67, 0x6e, 0x65, 0x64, 0x56, 0x69, 0x65, 0x77, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x52,
	0x11, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x56, 0x69, 0x65, 0x77, 0x43, 0x68, 0x61, 0x6e, 0x67,
	0x65, 0x73, 0x12, 0x2c, 0x0a, 0x12, 0x70, 0x72, 0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65,
	0x5f, 0x73, 0x65, 0x71, 0x5f, 0x6e, 0x72, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x04, 0x52, 0x10,
	0x70, 0x72, 0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x53, 0x65, 0x71, 0x4e, 0x72, 0x73,
	0x12, 0x3c, 0x0a, 0x0b, 0x70, 0x72, 0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x73, 0x18,
	0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x72, 0x73,
	0x70, 0x62, 0x66, 0x74, 0x70, 0x62, 0x2e, 0x50, 0x72, 0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72,
	0x65, 0x52, 0x0b, 0x70, 0x72, 0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x73, 0x22, 0x47,
	0x0a, 0x09, 0x50, 0x53, 0x65, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x73,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x02, 0x73, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x76,
	0x69, 0x65, 0x77, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x04, 0x76, 0x69, 0x65, 0x77, 0x12,
	0x16, 0x0a, 0x06, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x06, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74, 0x22, 0x47, 0x0a, 0x09, 0x51, 0x53, 0x65, 0x74, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x73, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x02, 0x73, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x76, 0x69, 0x65, 0x77, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x04, 0x76, 0x69, 0x65, 0x77, 0x12, 0x16, 0x0a, 0x06, 0x64, 0x69, 0x67, 0x65,
	0x73, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74,
	0x22, 0x3b, 0x0a, 0x11, 0x50, 0x72, 0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a,
	0x02, 0x73, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x02, 0x73, 0x6e, 0x22, 0x36, 0x0a,
	0x10, 0x52, 0x65, 0x71, 0x57, 0x61, 0x69, 0x74, 0x52, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63,
	0x65, 0x12, 0x0e, 0x0a, 0x02, 0x73, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x02, 0x73,
	0x6e, 0x12, 0x12, 0x0a, 0x04, 0x76, 0x69, 0x65, 0x77, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x04, 0x76, 0x69, 0x65, 0x77, 0x22, 0x45, 0x0a, 0x0b, 0x56, 0x43, 0x53, 0x4e, 0x54, 0x69, 0x6d,
	0x65, 0x6f, 0x75, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x76, 0x69, 0x65, 0x77, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x04, 0x76, 0x69, 0x65, 0x77, 0x12, 0x22, 0x0a, 0x0c, 0x6e, 0x75, 0x6d, 0x43,
	0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0c,
	0x6e, 0x75, 0x6d, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x64, 0x22, 0x08, 0x0a, 0x06,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x37, 0x5a, 0x35, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x62,
	0x2f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x72, 0x73, 0x70, 0x62, 0x66, 0x74, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ordererspbftpb_ordererspbftpb_proto_rawDescOnce sync.Once
	file_ordererspbftpb_ordererspbftpb_proto_rawDescData = file_ordererspbftpb_ordererspbftpb_proto_rawDesc
)

func file_ordererspbftpb_ordererspbftpb_proto_rawDescGZIP() []byte {
	file_ordererspbftpb_ordererspbftpb_proto_rawDescOnce.Do(func() {
		file_ordererspbftpb_ordererspbftpb_proto_rawDescData = protoimpl.X.CompressGZIP(file_ordererspbftpb_ordererspbftpb_proto_rawDescData)
	})
	return file_ordererspbftpb_ordererspbftpb_proto_rawDescData
}

var file_ordererspbftpb_ordererspbftpb_proto_msgTypes = make([]protoimpl.MessageInfo, 14)
var file_ordererspbftpb_ordererspbftpb_proto_goTypes = []interface{}{
	(*Preprepare)(nil),        // 0: ordererspbftpb.Preprepare
	(*Prepare)(nil),           // 1: ordererspbftpb.Prepare
	(*Commit)(nil),            // 2: ordererspbftpb.Commit
	(*Done)(nil),              // 3: ordererspbftpb.Done
	(*CatchUpRequest)(nil),    // 4: ordererspbftpb.CatchUpRequest
	(*ViewChange)(nil),        // 5: ordererspbftpb.ViewChange
	(*SignedViewChange)(nil),  // 6: ordererspbftpb.SignedViewChange
	(*NewView)(nil),           // 7: ordererspbftpb.NewView
	(*PSetEntry)(nil),         // 8: ordererspbftpb.PSetEntry
	(*QSetEntry)(nil),         // 9: ordererspbftpb.QSetEntry
	(*PreprepareRequest)(nil), // 10: ordererspbftpb.PreprepareRequest
	(*ReqWaitReference)(nil),  // 11: ordererspbftpb.ReqWaitReference
	(*VCSNTimeout)(nil),       // 12: ordererspbftpb.VCSNTimeout
	(*Status)(nil),            // 13: ordererspbftpb.Status
}
var file_ordererspbftpb_ordererspbftpb_proto_depIdxs = []int32{
	8, // 0: ordererspbftpb.ViewChange.p_set:type_name -> ordererspbftpb.PSetEntry
	9, // 1: ordererspbftpb.ViewChange.q_set:type_name -> ordererspbftpb.QSetEntry
	5, // 2: ordererspbftpb.SignedViewChange.view_change:type_name -> ordererspbftpb.ViewChange
	6, // 3: ordererspbftpb.NewView.signed_view_changes:type_name -> ordererspbftpb.SignedViewChange
	0, // 4: ordererspbftpb.NewView.preprepares:type_name -> ordererspbftpb.Preprepare
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_ordererspbftpb_ordererspbftpb_proto_init() }
func file_ordererspbftpb_ordererspbftpb_proto_init() {
	if File_ordererspbftpb_ordererspbftpb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ordererspbftpb_ordererspbftpb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Preprepare); i {
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
		file_ordererspbftpb_ordererspbftpb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Prepare); i {
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
		file_ordererspbftpb_ordererspbftpb_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Commit); i {
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
		file_ordererspbftpb_ordererspbftpb_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Done); i {
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
		file_ordererspbftpb_ordererspbftpb_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CatchUpRequest); i {
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
		file_ordererspbftpb_ordererspbftpb_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ViewChange); i {
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
		file_ordererspbftpb_ordererspbftpb_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignedViewChange); i {
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
		file_ordererspbftpb_ordererspbftpb_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NewView); i {
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
		file_ordererspbftpb_ordererspbftpb_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PSetEntry); i {
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
		file_ordererspbftpb_ordererspbftpb_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QSetEntry); i {
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
		file_ordererspbftpb_ordererspbftpb_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PreprepareRequest); i {
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
		file_ordererspbftpb_ordererspbftpb_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReqWaitReference); i {
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
		file_ordererspbftpb_ordererspbftpb_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VCSNTimeout); i {
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
		file_ordererspbftpb_ordererspbftpb_proto_msgTypes[13].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Status); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_ordererspbftpb_ordererspbftpb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   14,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ordererspbftpb_ordererspbftpb_proto_goTypes,
		DependencyIndexes: file_ordererspbftpb_ordererspbftpb_proto_depIdxs,
		MessageInfos:      file_ordererspbftpb_ordererspbftpb_proto_msgTypes,
	}.Build()
	File_ordererspbftpb_ordererspbftpb_proto = out.File
	file_ordererspbftpb_ordererspbftpb_proto_rawDesc = nil
	file_ordererspbftpb_ordererspbftpb_proto_goTypes = nil
	file_ordererspbftpb_ordererspbftpb_proto_depIdxs = nil
}
