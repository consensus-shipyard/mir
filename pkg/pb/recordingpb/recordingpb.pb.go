//
//Copyright IBM Corp. All Rights Reserved.
//
//SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.12.4
// source: recordingpb/recordingpb.proto

package recordingpb

import (
	eventpb "github.com/filecoin-project/mir/pkg/pb/eventpb"
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

type Entry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeId string           `protobuf:"bytes,1,opt,name=node_id,json=nodeId,proto3" json:"node_id,omitempty"`
	Time   int64            `protobuf:"varint,2,opt,name=time,proto3" json:"time,omitempty"`
	Events []*eventpb.Event `protobuf:"bytes,3,rep,name=events,proto3" json:"events,omitempty"`
}

func (x *Entry) Reset() {
	*x = Entry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_recordingpb_recordingpb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Entry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Entry) ProtoMessage() {}

func (x *Entry) ProtoReflect() protoreflect.Message {
	mi := &file_recordingpb_recordingpb_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Entry.ProtoReflect.Descriptor instead.
func (*Entry) Descriptor() ([]byte, []int) {
	return file_recordingpb_recordingpb_proto_rawDescGZIP(), []int{0}
}

func (x *Entry) GetNodeId() string {
	if x != nil {
		return x.NodeId
	}
	return ""
}

func (x *Entry) GetTime() int64 {
	if x != nil {
		return x.Time
	}
	return 0
}

func (x *Entry) GetEvents() []*eventpb.Event {
	if x != nil {
		return x.Events
	}
	return nil
}

var File_recordingpb_recordingpb_proto protoreflect.FileDescriptor

var file_recordingpb_recordingpb_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x62, 0x2f, 0x72, 0x65,
	0x63, 0x6f, 0x72, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x0b, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x62, 0x1a, 0x15, 0x65, 0x76,
	0x65, 0x6e, 0x74, 0x70, 0x62, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x5c, 0x0a, 0x05, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x17, 0x0a, 0x07,
	0x6e, 0x6f, 0x64, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6e,
	0x6f, 0x64, 0x65, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x26, 0x0a, 0x06, 0x65, 0x76, 0x65,
	0x6e, 0x74, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x70, 0x62, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74,
	0x73, 0x42, 0x34, 0x5a, 0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x62, 0x2f, 0x72, 0x65, 0x63, 0x6f,
	0x72, 0x64, 0x69, 0x6e, 0x67, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_recordingpb_recordingpb_proto_rawDescOnce sync.Once
	file_recordingpb_recordingpb_proto_rawDescData = file_recordingpb_recordingpb_proto_rawDesc
)

func file_recordingpb_recordingpb_proto_rawDescGZIP() []byte {
	file_recordingpb_recordingpb_proto_rawDescOnce.Do(func() {
		file_recordingpb_recordingpb_proto_rawDescData = protoimpl.X.CompressGZIP(file_recordingpb_recordingpb_proto_rawDescData)
	})
	return file_recordingpb_recordingpb_proto_rawDescData
}

var file_recordingpb_recordingpb_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_recordingpb_recordingpb_proto_goTypes = []interface{}{
	(*Entry)(nil),         // 0: recordingpb.Entry
	(*eventpb.Event)(nil), // 1: eventpb.Event
}
var file_recordingpb_recordingpb_proto_depIdxs = []int32{
	1, // 0: recordingpb.Entry.events:type_name -> eventpb.Event
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_recordingpb_recordingpb_proto_init() }
func file_recordingpb_recordingpb_proto_init() {
	if File_recordingpb_recordingpb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_recordingpb_recordingpb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Entry); i {
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
			RawDescriptor: file_recordingpb_recordingpb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_recordingpb_recordingpb_proto_goTypes,
		DependencyIndexes: file_recordingpb_recordingpb_proto_depIdxs,
		MessageInfos:      file_recordingpb_recordingpb_proto_msgTypes,
	}.Build()
	File_recordingpb_recordingpb_proto = out.File
	file_recordingpb_recordingpb_proto_rawDesc = nil
	file_recordingpb_recordingpb_proto_goTypes = nil
	file_recordingpb_recordingpb_proto_depIdxs = nil
}
