//
//Copyright IBM Corp. All Rights Reserved.
//
//SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.12.4
// source: messagepb/messagepb.proto

package messagepb

import (
	mscpb "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb"
	bcbpb "github.com/filecoin-project/mir/pkg/pb/bcbpb"
	checkpointpb "github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	isspb "github.com/filecoin-project/mir/pkg/pb/isspb"
	ordererspb "github.com/filecoin-project/mir/pkg/pb/ordererspb"
	pingpongpb "github.com/filecoin-project/mir/pkg/pb/pingpongpb"
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

	DestModule string `protobuf:"bytes,1,opt,name=dest_module,json=destModule,proto3" json:"dest_module,omitempty"`
	// Types that are assignable to Type:
	//	*Message_Iss
	//	*Message_Bcb
	//	*Message_MultisigCollector
	//	*Message_Pingpong
	//	*Message_Checkpoint
	//	*Message_SbMessage
	Type isMessage_Type `protobuf_oneof:"type"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_messagepb_messagepb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_messagepb_messagepb_proto_msgTypes[0]
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
	return file_messagepb_messagepb_proto_rawDescGZIP(), []int{0}
}

func (x *Message) GetDestModule() string {
	if x != nil {
		return x.DestModule
	}
	return ""
}

func (m *Message) GetType() isMessage_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Message) GetIss() *isspb.ISSMessage {
	if x, ok := x.GetType().(*Message_Iss); ok {
		return x.Iss
	}
	return nil
}

func (x *Message) GetBcb() *bcbpb.Message {
	if x, ok := x.GetType().(*Message_Bcb); ok {
		return x.Bcb
	}
	return nil
}

func (x *Message) GetMultisigCollector() *mscpb.Message {
	if x, ok := x.GetType().(*Message_MultisigCollector); ok {
		return x.MultisigCollector
	}
	return nil
}

func (x *Message) GetPingpong() *pingpongpb.Message {
	if x, ok := x.GetType().(*Message_Pingpong); ok {
		return x.Pingpong
	}
	return nil
}

func (x *Message) GetCheckpoint() *checkpointpb.Message {
	if x, ok := x.GetType().(*Message_Checkpoint); ok {
		return x.Checkpoint
	}
	return nil
}

func (x *Message) GetSbMessage() *ordererspb.SBInstanceMessage {
	if x, ok := x.GetType().(*Message_SbMessage); ok {
		return x.SbMessage
	}
	return nil
}

type isMessage_Type interface {
	isMessage_Type()
}

type Message_Iss struct {
	Iss *isspb.ISSMessage `protobuf:"bytes,2,opt,name=iss,proto3,oneof"`
}

type Message_Bcb struct {
	Bcb *bcbpb.Message `protobuf:"bytes,3,opt,name=bcb,proto3,oneof"`
}

type Message_MultisigCollector struct {
	MultisigCollector *mscpb.Message `protobuf:"bytes,4,opt,name=multisig_collector,json=multisigCollector,proto3,oneof"`
}

type Message_Pingpong struct {
	Pingpong *pingpongpb.Message `protobuf:"bytes,5,opt,name=pingpong,proto3,oneof"`
}

type Message_Checkpoint struct {
	Checkpoint *checkpointpb.Message `protobuf:"bytes,6,opt,name=checkpoint,proto3,oneof"`
}

type Message_SbMessage struct {
	SbMessage *ordererspb.SBInstanceMessage `protobuf:"bytes,7,opt,name=sb_message,json=sbMessage,proto3,oneof"`
}

func (*Message_Iss) isMessage_Type() {}

func (*Message_Bcb) isMessage_Type() {}

func (*Message_MultisigCollector) isMessage_Type() {}

func (*Message_Pingpong) isMessage_Type() {}

func (*Message_Checkpoint) isMessage_Type() {}

func (*Message_SbMessage) isMessage_Type() {}

var File_messagepb_messagepb_proto protoreflect.FileDescriptor

var file_messagepb_messagepb_proto_rawDesc = []byte{
	0x0a, 0x19, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x70, 0x62, 0x2f, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x70, 0x62, 0x1a, 0x11, 0x69, 0x73, 0x73, 0x70, 0x62, 0x2f, 0x69, 0x73,
	0x73, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x11, 0x62, 0x63, 0x62, 0x70, 0x62,
	0x2f, 0x62, 0x63, 0x62, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x61, 0x76,
	0x61, 0x69, 0x6c, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x70, 0x62, 0x2f, 0x6d, 0x73, 0x63,
	0x70, 0x62, 0x2f, 0x6d, 0x73, 0x63, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b,
	0x70, 0x69, 0x6e, 0x67, 0x70, 0x6f, 0x6e, 0x67, 0x70, 0x62, 0x2f, 0x70, 0x69, 0x6e, 0x67, 0x70,
	0x6f, 0x6e, 0x67, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x63, 0x68, 0x65,
	0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x70, 0x62, 0x2f, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70,
	0x6f, 0x69, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x6f, 0x72,
	0x64, 0x65, 0x72, 0x65, 0x72, 0x73, 0x70, 0x62, 0x2f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x72,
	0x73, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xf9, 0x02, 0x0a, 0x07, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x74, 0x5f, 0x6d, 0x6f,
	0x64, 0x75, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x64, 0x65, 0x73, 0x74,
	0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x12, 0x25, 0x0a, 0x03, 0x69, 0x73, 0x73, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x69, 0x73, 0x73, 0x70, 0x62, 0x2e, 0x49, 0x53, 0x53, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48, 0x00, 0x52, 0x03, 0x69, 0x73, 0x73, 0x12, 0x22, 0x0a,
	0x03, 0x62, 0x63, 0x62, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x62, 0x63, 0x62,
	0x70, 0x62, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48, 0x00, 0x52, 0x03, 0x62, 0x63,
	0x62, 0x12, 0x4e, 0x0a, 0x12, 0x6d, 0x75, 0x6c, 0x74, 0x69, 0x73, 0x69, 0x67, 0x5f, 0x63, 0x6f,
	0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e,
	0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x70, 0x62, 0x2e, 0x6d,
	0x73, 0x63, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48, 0x00, 0x52, 0x11,
	0x6d, 0x75, 0x6c, 0x74, 0x69, 0x73, 0x69, 0x67, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x12, 0x31, 0x0a, 0x08, 0x70, 0x69, 0x6e, 0x67, 0x70, 0x6f, 0x6e, 0x67, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x70, 0x69, 0x6e, 0x67, 0x70, 0x6f, 0x6e, 0x67, 0x70, 0x62,
	0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48, 0x00, 0x52, 0x08, 0x70, 0x69, 0x6e, 0x67,
	0x70, 0x6f, 0x6e, 0x67, 0x12, 0x37, 0x0a, 0x0a, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69,
	0x6e, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x63, 0x68, 0x65, 0x63, 0x6b,
	0x70, 0x6f, 0x69, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48,
	0x00, 0x52, 0x0a, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x3e, 0x0a,
	0x0a, 0x73, 0x62, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1d, 0x2e, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x72, 0x73, 0x70, 0x62, 0x2e, 0x53,
	0x42, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x48, 0x00, 0x52, 0x09, 0x73, 0x62, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x42, 0x06, 0x0a,
	0x04, 0x74, 0x79, 0x70, 0x65, 0x42, 0x32, 0x5a, 0x30, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f,
	0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x62, 0x2f,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_messagepb_messagepb_proto_rawDescOnce sync.Once
	file_messagepb_messagepb_proto_rawDescData = file_messagepb_messagepb_proto_rawDesc
)

func file_messagepb_messagepb_proto_rawDescGZIP() []byte {
	file_messagepb_messagepb_proto_rawDescOnce.Do(func() {
		file_messagepb_messagepb_proto_rawDescData = protoimpl.X.CompressGZIP(file_messagepb_messagepb_proto_rawDescData)
	})
	return file_messagepb_messagepb_proto_rawDescData
}

var file_messagepb_messagepb_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_messagepb_messagepb_proto_goTypes = []interface{}{
	(*Message)(nil),                      // 0: messagepb.Message
	(*isspb.ISSMessage)(nil),             // 1: isspb.ISSMessage
	(*bcbpb.Message)(nil),                // 2: bcbpb.Message
	(*mscpb.Message)(nil),                // 3: availabilitypb.mscpb.Message
	(*pingpongpb.Message)(nil),           // 4: pingpongpb.Message
	(*checkpointpb.Message)(nil),         // 5: checkpointpb.Message
	(*ordererspb.SBInstanceMessage)(nil), // 6: ordererspb.SBInstanceMessage
}
var file_messagepb_messagepb_proto_depIdxs = []int32{
	1, // 0: messagepb.Message.iss:type_name -> isspb.ISSMessage
	2, // 1: messagepb.Message.bcb:type_name -> bcbpb.Message
	3, // 2: messagepb.Message.multisig_collector:type_name -> availabilitypb.mscpb.Message
	4, // 3: messagepb.Message.pingpong:type_name -> pingpongpb.Message
	5, // 4: messagepb.Message.checkpoint:type_name -> checkpointpb.Message
	6, // 5: messagepb.Message.sb_message:type_name -> ordererspb.SBInstanceMessage
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_messagepb_messagepb_proto_init() }
func file_messagepb_messagepb_proto_init() {
	if File_messagepb_messagepb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_messagepb_messagepb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
	}
	file_messagepb_messagepb_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Message_Iss)(nil),
		(*Message_Bcb)(nil),
		(*Message_MultisigCollector)(nil),
		(*Message_Pingpong)(nil),
		(*Message_Checkpoint)(nil),
		(*Message_SbMessage)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_messagepb_messagepb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_messagepb_messagepb_proto_goTypes,
		DependencyIndexes: file_messagepb_messagepb_proto_depIdxs,
		MessageInfos:      file_messagepb_messagepb_proto_msgTypes,
	}.Build()
	File_messagepb_messagepb_proto = out.File
	file_messagepb_messagepb_proto_rawDesc = nil
	file_messagepb_messagepb_proto_goTypes = nil
	file_messagepb_messagepb_proto_depIdxs = nil
}
