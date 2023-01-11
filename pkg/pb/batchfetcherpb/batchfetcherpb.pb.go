// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.12.4
// source: batchfetcherpb/batchfetcherpb.proto

package batchfetcherpb

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"

	commonpb "github.com/filecoin-project/mir/pkg/pb/commonpb"
	_ "github.com/filecoin-project/mir/pkg/pb/mir"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
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
	//	*Event_NewOrderedBatch
	//	*Event_ClientProgress
	Type isEvent_Type `protobuf_oneof:"Type"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_batchfetcherpb_batchfetcherpb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_batchfetcherpb_batchfetcherpb_proto_msgTypes[0]
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
	return file_batchfetcherpb_batchfetcherpb_proto_rawDescGZIP(), []int{0}
}

func (m *Event) GetType() isEvent_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Event) GetNewOrderedBatch() *NewOrderedBatch {
	if x, ok := x.GetType().(*Event_NewOrderedBatch); ok {
		return x.NewOrderedBatch
	}
	return nil
}

func (x *Event) GetClientProgress() *commonpb.ClientProgress {
	if x, ok := x.GetType().(*Event_ClientProgress); ok {
		return x.ClientProgress
	}
	return nil
}

type isEvent_Type interface {
	isEvent_Type()
}

type Event_NewOrderedBatch struct {
	NewOrderedBatch *NewOrderedBatch `protobuf:"bytes,1,opt,name=new_ordered_batch,json=newOrderedBatch,proto3,oneof"`
}

type Event_ClientProgress struct {
	ClientProgress *commonpb.ClientProgress `protobuf:"bytes,2,opt,name=client_progress,json=clientProgress,proto3,oneof"`
}

func (*Event_NewOrderedBatch) isEvent_Type() {}

func (*Event_ClientProgress) isEvent_Type() {}

type NewOrderedBatch struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Txs []*requestpb.Request `protobuf:"bytes,1,rep,name=txs,proto3" json:"txs,omitempty"`
}

func (x *NewOrderedBatch) Reset() {
	*x = NewOrderedBatch{}
	if protoimpl.UnsafeEnabled {
		mi := &file_batchfetcherpb_batchfetcherpb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NewOrderedBatch) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewOrderedBatch) ProtoMessage() {}

func (x *NewOrderedBatch) ProtoReflect() protoreflect.Message {
	mi := &file_batchfetcherpb_batchfetcherpb_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewOrderedBatch.ProtoReflect.Descriptor instead.
func (*NewOrderedBatch) Descriptor() ([]byte, []int) {
	return file_batchfetcherpb_batchfetcherpb_proto_rawDescGZIP(), []int{1}
}

func (x *NewOrderedBatch) GetTxs() []*requestpb.Request {
	if x != nil {
		return x.Txs
	}
	return nil
}

var File_batchfetcherpb_batchfetcherpb_proto protoreflect.FileDescriptor

var file_batchfetcherpb_batchfetcherpb_proto_rawDesc = []byte{
	0x0a, 0x23, 0x62, 0x61, 0x74, 0x63, 0x68, 0x66, 0x65, 0x74, 0x63, 0x68, 0x65, 0x72, 0x70, 0x62,
	0x2f, 0x62, 0x61, 0x74, 0x63, 0x68, 0x66, 0x65, 0x74, 0x63, 0x68, 0x65, 0x72, 0x70, 0x62, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x62, 0x61, 0x74, 0x63, 0x68, 0x66, 0x65, 0x74, 0x63,
	0x68, 0x65, 0x72, 0x70, 0x62, 0x1a, 0x19, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x70, 0x62,
	0x2f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x17, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x70, 0x62, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6d, 0x69, 0x72, 0x2f, 0x63,
	0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e, 0x5f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xaf, 0x01, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x12, 0x4d, 0x0a, 0x11, 0x6e, 0x65, 0x77, 0x5f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x64,
	0x5f, 0x62, 0x61, 0x74, 0x63, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x62,
	0x61, 0x74, 0x63, 0x68, 0x66, 0x65, 0x74, 0x63, 0x68, 0x65, 0x72, 0x70, 0x62, 0x2e, 0x4e, 0x65,
	0x77, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x64, 0x42, 0x61, 0x74, 0x63, 0x68, 0x48, 0x00, 0x52,
	0x0f, 0x6e, 0x65, 0x77, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x64, 0x42, 0x61, 0x74, 0x63, 0x68,
	0x12, 0x43, 0x0a, 0x0f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x70, 0x72, 0x6f, 0x67, 0x72,
	0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x70, 0x62, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x67, 0x72,
	0x65, 0x73, 0x73, 0x48, 0x00, 0x52, 0x0e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x50, 0x72, 0x6f,
	0x67, 0x72, 0x65, 0x73, 0x73, 0x3a, 0x04, 0x90, 0xa6, 0x1d, 0x01, 0x42, 0x0c, 0x0a, 0x04, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x04, 0x80, 0xa6, 0x1d, 0x01, 0x22, 0x3d, 0x0a, 0x0f, 0x4e, 0x65, 0x77,
	0x4f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x64, 0x42, 0x61, 0x74, 0x63, 0x68, 0x12, 0x24, 0x0a, 0x03,
	0x74, 0x78, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x52, 0x03, 0x74,
	0x78, 0x73, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x42, 0x37, 0x5a, 0x35, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d,
	0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f,
	0x70, 0x62, 0x2f, 0x62, 0x61, 0x74, 0x63, 0x68, 0x66, 0x65, 0x74, 0x63, 0x68, 0x65, 0x72, 0x70,
	0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_batchfetcherpb_batchfetcherpb_proto_rawDescOnce sync.Once
	file_batchfetcherpb_batchfetcherpb_proto_rawDescData = file_batchfetcherpb_batchfetcherpb_proto_rawDesc
)

func file_batchfetcherpb_batchfetcherpb_proto_rawDescGZIP() []byte {
	file_batchfetcherpb_batchfetcherpb_proto_rawDescOnce.Do(func() {
		file_batchfetcherpb_batchfetcherpb_proto_rawDescData = protoimpl.X.CompressGZIP(file_batchfetcherpb_batchfetcherpb_proto_rawDescData)
	})
	return file_batchfetcherpb_batchfetcherpb_proto_rawDescData
}

var file_batchfetcherpb_batchfetcherpb_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_batchfetcherpb_batchfetcherpb_proto_goTypes = []interface{}{
	(*Event)(nil),                   // 0: batchfetcherpb.Event
	(*NewOrderedBatch)(nil),         // 1: batchfetcherpb.NewOrderedBatch
	(*commonpb.ClientProgress)(nil), // 2: commonpb.ClientProgress
	(*requestpb.Request)(nil),       // 3: requestpb.Request
}
var file_batchfetcherpb_batchfetcherpb_proto_depIdxs = []int32{
	1, // 0: batchfetcherpb.Event.new_ordered_batch:type_name -> batchfetcherpb.NewOrderedBatch
	2, // 1: batchfetcherpb.Event.client_progress:type_name -> commonpb.ClientProgress
	3, // 2: batchfetcherpb.NewOrderedBatch.txs:type_name -> requestpb.Request
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_batchfetcherpb_batchfetcherpb_proto_init() }
func file_batchfetcherpb_batchfetcherpb_proto_init() {
	if File_batchfetcherpb_batchfetcherpb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_batchfetcherpb_batchfetcherpb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_batchfetcherpb_batchfetcherpb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NewOrderedBatch); i {
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
	file_batchfetcherpb_batchfetcherpb_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Event_NewOrderedBatch)(nil),
		(*Event_ClientProgress)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_batchfetcherpb_batchfetcherpb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_batchfetcherpb_batchfetcherpb_proto_goTypes,
		DependencyIndexes: file_batchfetcherpb_batchfetcherpb_proto_depIdxs,
		MessageInfos:      file_batchfetcherpb_batchfetcherpb_proto_msgTypes,
	}.Build()
	File_batchfetcherpb_batchfetcherpb_proto = out.File
	file_batchfetcherpb_batchfetcherpb_proto_rawDesc = nil
	file_batchfetcherpb_batchfetcherpb_proto_goTypes = nil
	file_batchfetcherpb_batchfetcherpb_proto_depIdxs = nil
}
