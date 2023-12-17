// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.14.0
// source: pingpongpb/pingpongpb.proto

package pingpongpb

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"

	_ "github.com/filecoin-project/mir/pkg/pb/mir"
	_ "github.com/filecoin-project/mir/pkg/pb/net"
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
	//	*Event_PingTime
	Type isEvent_Type `protobuf_oneof:"type"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pingpongpb_pingpongpb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_pingpongpb_pingpongpb_proto_msgTypes[0]
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
	return file_pingpongpb_pingpongpb_proto_rawDescGZIP(), []int{0}
}

func (m *Event) GetType() isEvent_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Event) GetPingTime() *PingTime {
	if x, ok := x.GetType().(*Event_PingTime); ok {
		return x.PingTime
	}
	return nil
}

type isEvent_Type interface {
	isEvent_Type()
}

type Event_PingTime struct {
	PingTime *PingTime `protobuf:"bytes,1,opt,name=ping_time,json=pingTime,proto3,oneof"`
}

func (*Event_PingTime) isEvent_Type() {}

type PingTime struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *PingTime) Reset() {
	*x = PingTime{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pingpongpb_pingpongpb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingTime) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingTime) ProtoMessage() {}

func (x *PingTime) ProtoReflect() protoreflect.Message {
	mi := &file_pingpongpb_pingpongpb_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingTime.ProtoReflect.Descriptor instead.
func (*PingTime) Descriptor() ([]byte, []int) {
	return file_pingpongpb_pingpongpb_proto_rawDescGZIP(), []int{1}
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//	*Message_Ping
	//	*Message_Pong
	Type isMessage_Type `protobuf_oneof:"type"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pingpongpb_pingpongpb_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_pingpongpb_pingpongpb_proto_msgTypes[2]
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
	return file_pingpongpb_pingpongpb_proto_rawDescGZIP(), []int{2}
}

func (m *Message) GetType() isMessage_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Message) GetPing() *Ping {
	if x, ok := x.GetType().(*Message_Ping); ok {
		return x.Ping
	}
	return nil
}

func (x *Message) GetPong() *Pong {
	if x, ok := x.GetType().(*Message_Pong); ok {
		return x.Pong
	}
	return nil
}

type isMessage_Type interface {
	isMessage_Type()
}

type Message_Ping struct {
	Ping *Ping `protobuf:"bytes,1,opt,name=ping,proto3,oneof"`
}

type Message_Pong struct {
	Pong *Pong `protobuf:"bytes,2,opt,name=pong,proto3,oneof"`
}

func (*Message_Ping) isMessage_Type() {}

func (*Message_Pong) isMessage_Type() {}

type Ping struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SeqNr uint64 `protobuf:"varint,1,opt,name=seq_nr,json=seqNr,proto3" json:"seq_nr,omitempty"`
}

func (x *Ping) Reset() {
	*x = Ping{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pingpongpb_pingpongpb_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Ping) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Ping) ProtoMessage() {}

func (x *Ping) ProtoReflect() protoreflect.Message {
	mi := &file_pingpongpb_pingpongpb_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Ping.ProtoReflect.Descriptor instead.
func (*Ping) Descriptor() ([]byte, []int) {
	return file_pingpongpb_pingpongpb_proto_rawDescGZIP(), []int{3}
}

func (x *Ping) GetSeqNr() uint64 {
	if x != nil {
		return x.SeqNr
	}
	return 0
}

type Pong struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SeqNr uint64 `protobuf:"varint,1,opt,name=seq_nr,json=seqNr,proto3" json:"seq_nr,omitempty"`
}

func (x *Pong) Reset() {
	*x = Pong{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pingpongpb_pingpongpb_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Pong) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Pong) ProtoMessage() {}

func (x *Pong) ProtoReflect() protoreflect.Message {
	mi := &file_pingpongpb_pingpongpb_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Pong.ProtoReflect.Descriptor instead.
func (*Pong) Descriptor() ([]byte, []int) {
	return file_pingpongpb_pingpongpb_proto_rawDescGZIP(), []int{4}
}

func (x *Pong) GetSeqNr() uint64 {
	if x != nil {
		return x.SeqNr
	}
	return 0
}

var File_pingpongpb_pingpongpb_proto protoreflect.FileDescriptor

var file_pingpongpb_pingpongpb_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x70, 0x69, 0x6e, 0x67, 0x70, 0x6f, 0x6e, 0x67, 0x70, 0x62, 0x2f, 0x70, 0x69, 0x6e,
	0x67, 0x70, 0x6f, 0x6e, 0x67, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x70,
	0x69, 0x6e, 0x67, 0x70, 0x6f, 0x6e, 0x67, 0x70, 0x62, 0x1a, 0x1c, 0x6d, 0x69, 0x72, 0x2f, 0x63,
	0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e, 0x5f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6e, 0x65, 0x74, 0x2f, 0x63, 0x6f, 0x64,
	0x65, 0x67, 0x65, 0x6e, 0x5f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x50, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x33,
	0x0a, 0x09, 0x70, 0x69, 0x6e, 0x67, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x14, 0x2e, 0x70, 0x69, 0x6e, 0x67, 0x70, 0x6f, 0x6e, 0x67, 0x70, 0x62, 0x2e, 0x50,
	0x69, 0x6e, 0x67, 0x54, 0x69, 0x6d, 0x65, 0x48, 0x00, 0x52, 0x08, 0x70, 0x69, 0x6e, 0x67, 0x54,
	0x69, 0x6d, 0x65, 0x3a, 0x04, 0x90, 0xa6, 0x1d, 0x01, 0x42, 0x0c, 0x0a, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x12, 0x04, 0x80, 0xa6, 0x1d, 0x01, 0x22, 0x10, 0x0a, 0x08, 0x50, 0x69, 0x6e, 0x67, 0x54,
	0x69, 0x6d, 0x65, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x22, 0x6d, 0x0a, 0x07, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x12, 0x26, 0x0a, 0x04, 0x70, 0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x10, 0x2e, 0x70, 0x69, 0x6e, 0x67, 0x70, 0x6f, 0x6e, 0x67, 0x70, 0x62, 0x2e,
	0x50, 0x69, 0x6e, 0x67, 0x48, 0x00, 0x52, 0x04, 0x70, 0x69, 0x6e, 0x67, 0x12, 0x26, 0x0a, 0x04,
	0x70, 0x6f, 0x6e, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x70, 0x69, 0x6e,
	0x67, 0x70, 0x6f, 0x6e, 0x67, 0x70, 0x62, 0x2e, 0x50, 0x6f, 0x6e, 0x67, 0x48, 0x00, 0x52, 0x04,
	0x70, 0x6f, 0x6e, 0x67, 0x3a, 0x04, 0xc8, 0xe4, 0x1d, 0x01, 0x42, 0x0c, 0x0a, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x12, 0x04, 0xc8, 0xe4, 0x1d, 0x01, 0x22, 0x23, 0x0a, 0x04, 0x50, 0x69, 0x6e, 0x67,
	0x12, 0x15, 0x0a, 0x06, 0x73, 0x65, 0x71, 0x5f, 0x6e, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x05, 0x73, 0x65, 0x71, 0x4e, 0x72, 0x3a, 0x04, 0xd0, 0xe4, 0x1d, 0x01, 0x22, 0x23, 0x0a,
	0x04, 0x50, 0x6f, 0x6e, 0x67, 0x12, 0x15, 0x0a, 0x06, 0x73, 0x65, 0x71, 0x5f, 0x6e, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x73, 0x65, 0x71, 0x4e, 0x72, 0x3a, 0x04, 0xd0, 0xe4,
	0x1d, 0x01, 0x42, 0x33, 0x5a, 0x31, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x62, 0x2f, 0x70, 0x69, 0x6e,
	0x67, 0x70, 0x6f, 0x6e, 0x67, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pingpongpb_pingpongpb_proto_rawDescOnce sync.Once
	file_pingpongpb_pingpongpb_proto_rawDescData = file_pingpongpb_pingpongpb_proto_rawDesc
)

func file_pingpongpb_pingpongpb_proto_rawDescGZIP() []byte {
	file_pingpongpb_pingpongpb_proto_rawDescOnce.Do(func() {
		file_pingpongpb_pingpongpb_proto_rawDescData = protoimpl.X.CompressGZIP(file_pingpongpb_pingpongpb_proto_rawDescData)
	})
	return file_pingpongpb_pingpongpb_proto_rawDescData
}

var file_pingpongpb_pingpongpb_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_pingpongpb_pingpongpb_proto_goTypes = []interface{}{
	(*Event)(nil),    // 0: pingpongpb.Event
	(*PingTime)(nil), // 1: pingpongpb.PingTime
	(*Message)(nil),  // 2: pingpongpb.Message
	(*Ping)(nil),     // 3: pingpongpb.Ping
	(*Pong)(nil),     // 4: pingpongpb.Pong
}
var file_pingpongpb_pingpongpb_proto_depIdxs = []int32{
	1, // 0: pingpongpb.Event.ping_time:type_name -> pingpongpb.PingTime
	3, // 1: pingpongpb.Message.ping:type_name -> pingpongpb.Ping
	4, // 2: pingpongpb.Message.pong:type_name -> pingpongpb.Pong
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_pingpongpb_pingpongpb_proto_init() }
func file_pingpongpb_pingpongpb_proto_init() {
	if File_pingpongpb_pingpongpb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pingpongpb_pingpongpb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_pingpongpb_pingpongpb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingTime); i {
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
		file_pingpongpb_pingpongpb_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
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
		file_pingpongpb_pingpongpb_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Ping); i {
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
		file_pingpongpb_pingpongpb_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Pong); i {
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
	file_pingpongpb_pingpongpb_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Event_PingTime)(nil),
	}
	file_pingpongpb_pingpongpb_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*Message_Ping)(nil),
		(*Message_Pong)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pingpongpb_pingpongpb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pingpongpb_pingpongpb_proto_goTypes,
		DependencyIndexes: file_pingpongpb_pingpongpb_proto_depIdxs,
		MessageInfos:      file_pingpongpb_pingpongpb_proto_msgTypes,
	}.Build()
	File_pingpongpb_pingpongpb_proto = out.File
	file_pingpongpb_pingpongpb_proto_rawDesc = nil
	file_pingpongpb_pingpongpb_proto_goTypes = nil
	file_pingpongpb_pingpongpb_proto_depIdxs = nil
}
