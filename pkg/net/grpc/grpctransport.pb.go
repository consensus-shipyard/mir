//
//Copyright IBM Corp. All Rights Reserved.
//
//SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.14.0
// source: net/grpc/grpctransport.proto

package grpc

import (
	messagepb "github.com/filecoin-project/mir/pkg/pb/messagepb"
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

type GrpcMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sender []byte `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	// Types that are assignable to Type:
	//	*GrpcMessage_PbMsg
	//	*GrpcMessage_RawMsg
	Type isGrpcMessage_Type `protobuf_oneof:"type"`
}

func (x *GrpcMessage) Reset() {
	*x = GrpcMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_net_grpc_grpctransport_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GrpcMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GrpcMessage) ProtoMessage() {}

func (x *GrpcMessage) ProtoReflect() protoreflect.Message {
	mi := &file_net_grpc_grpctransport_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GrpcMessage.ProtoReflect.Descriptor instead.
func (*GrpcMessage) Descriptor() ([]byte, []int) {
	return file_net_grpc_grpctransport_proto_rawDescGZIP(), []int{0}
}

func (x *GrpcMessage) GetSender() []byte {
	if x != nil {
		return x.Sender
	}
	return nil
}

func (m *GrpcMessage) GetType() isGrpcMessage_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *GrpcMessage) GetPbMsg() *messagepb.Message {
	if x, ok := x.GetType().(*GrpcMessage_PbMsg); ok {
		return x.PbMsg
	}
	return nil
}

func (x *GrpcMessage) GetRawMsg() *RawMessage {
	if x, ok := x.GetType().(*GrpcMessage_RawMsg); ok {
		return x.RawMsg
	}
	return nil
}

type isGrpcMessage_Type interface {
	isGrpcMessage_Type()
}

type GrpcMessage_PbMsg struct {
	PbMsg *messagepb.Message `protobuf:"bytes,2,opt,name=pb_msg,json=pbMsg,proto3,oneof"`
}

type GrpcMessage_RawMsg struct {
	RawMsg *RawMessage `protobuf:"bytes,3,opt,name=raw_msg,json=rawMsg,proto3,oneof"`
}

func (*GrpcMessage_PbMsg) isGrpcMessage_Type() {}

func (*GrpcMessage_RawMsg) isGrpcMessage_Type() {}

type RawMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DestModule string `protobuf:"bytes,1,opt,name=dest_module,json=destModule,proto3" json:"dest_module,omitempty"`
	Data       []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *RawMessage) Reset() {
	*x = RawMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_net_grpc_grpctransport_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RawMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RawMessage) ProtoMessage() {}

func (x *RawMessage) ProtoReflect() protoreflect.Message {
	mi := &file_net_grpc_grpctransport_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RawMessage.ProtoReflect.Descriptor instead.
func (*RawMessage) Descriptor() ([]byte, []int) {
	return file_net_grpc_grpctransport_proto_rawDescGZIP(), []int{1}
}

func (x *RawMessage) GetDestModule() string {
	if x != nil {
		return x.DestModule
	}
	return ""
}

func (x *RawMessage) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type ByeBye struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ByeBye) Reset() {
	*x = ByeBye{}
	if protoimpl.UnsafeEnabled {
		mi := &file_net_grpc_grpctransport_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ByeBye) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ByeBye) ProtoMessage() {}

func (x *ByeBye) ProtoReflect() protoreflect.Message {
	mi := &file_net_grpc_grpctransport_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ByeBye.ProtoReflect.Descriptor instead.
func (*ByeBye) Descriptor() ([]byte, []int) {
	return file_net_grpc_grpctransport_proto_rawDescGZIP(), []int{2}
}

var File_net_grpc_grpctransport_proto protoreflect.FileDescriptor

var file_net_grpc_grpctransport_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x6e, 0x65, 0x74, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d,
	0x67, 0x72, 0x70, 0x63, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x1a, 0x19, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x70, 0x62, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x90, 0x01, 0x0a, 0x0b, 0x47, 0x72, 0x70,
	0x63, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x6e, 0x64,
	0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72,
	0x12, 0x2b, 0x0a, 0x06, 0x70, 0x62, 0x5f, 0x6d, 0x73, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x12, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x48, 0x00, 0x52, 0x05, 0x70, 0x62, 0x4d, 0x73, 0x67, 0x12, 0x34, 0x0a,
	0x07, 0x72, 0x61, 0x77, 0x5f, 0x6d, 0x73, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19,
	0x2e, 0x67, 0x72, 0x70, 0x63, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x52,
	0x61, 0x77, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48, 0x00, 0x52, 0x06, 0x72, 0x61, 0x77,
	0x4d, 0x73, 0x67, 0x42, 0x06, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x22, 0x41, 0x0a, 0x0a, 0x52,
	0x61, 0x77, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x64, 0x65, 0x73,
	0x74, 0x5f, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a,
	0x64, 0x65, 0x73, 0x74, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61,
	0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x08,
	0x0a, 0x06, 0x42, 0x79, 0x65, 0x42, 0x79, 0x65, 0x32, 0x4e, 0x0a, 0x0d, 0x47, 0x72, 0x70, 0x63,
	0x54, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x3d, 0x0a, 0x06, 0x4c, 0x69, 0x73,
	0x74, 0x65, 0x6e, 0x12, 0x1a, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70,
	0x6f, 0x72, 0x74, 0x2e, 0x47, 0x72, 0x70, 0x63, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a,
	0x15, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e,
	0x42, 0x79, 0x65, 0x42, 0x79, 0x65, 0x28, 0x01, 0x42, 0x2e, 0x5a, 0x2c, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d,
	0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f,
	0x6e, 0x65, 0x74, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_net_grpc_grpctransport_proto_rawDescOnce sync.Once
	file_net_grpc_grpctransport_proto_rawDescData = file_net_grpc_grpctransport_proto_rawDesc
)

func file_net_grpc_grpctransport_proto_rawDescGZIP() []byte {
	file_net_grpc_grpctransport_proto_rawDescOnce.Do(func() {
		file_net_grpc_grpctransport_proto_rawDescData = protoimpl.X.CompressGZIP(file_net_grpc_grpctransport_proto_rawDescData)
	})
	return file_net_grpc_grpctransport_proto_rawDescData
}

var file_net_grpc_grpctransport_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_net_grpc_grpctransport_proto_goTypes = []interface{}{
	(*GrpcMessage)(nil),       // 0: grpctransport.GrpcMessage
	(*RawMessage)(nil),        // 1: grpctransport.RawMessage
	(*ByeBye)(nil),            // 2: grpctransport.ByeBye
	(*messagepb.Message)(nil), // 3: messagepb.Message
}
var file_net_grpc_grpctransport_proto_depIdxs = []int32{
	3, // 0: grpctransport.GrpcMessage.pb_msg:type_name -> messagepb.Message
	1, // 1: grpctransport.GrpcMessage.raw_msg:type_name -> grpctransport.RawMessage
	0, // 2: grpctransport.GrpcTransport.Listen:input_type -> grpctransport.GrpcMessage
	2, // 3: grpctransport.GrpcTransport.Listen:output_type -> grpctransport.ByeBye
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_net_grpc_grpctransport_proto_init() }
func file_net_grpc_grpctransport_proto_init() {
	if File_net_grpc_grpctransport_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_net_grpc_grpctransport_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GrpcMessage); i {
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
		file_net_grpc_grpctransport_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RawMessage); i {
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
		file_net_grpc_grpctransport_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ByeBye); i {
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
	file_net_grpc_grpctransport_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*GrpcMessage_PbMsg)(nil),
		(*GrpcMessage_RawMsg)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_net_grpc_grpctransport_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_net_grpc_grpctransport_proto_goTypes,
		DependencyIndexes: file_net_grpc_grpctransport_proto_depIdxs,
		MessageInfos:      file_net_grpc_grpctransport_proto_msgTypes,
	}.Build()
	File_net_grpc_grpctransport_proto = out.File
	file_net_grpc_grpctransport_proto_rawDesc = nil
	file_net_grpc_grpctransport_proto_goTypes = nil
	file_net_grpc_grpctransport_proto_depIdxs = nil
}
