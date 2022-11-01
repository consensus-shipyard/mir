package codegen

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/filecoin-project/mir/pkg/pb/mir"
	"github.com/filecoin-project/mir/pkg/pb/net"
)

// ShouldGenerateMirType returns true if the message is marked by one of the standard Mir annotations.
// Namely, one of the following:
//
//	option (mir.struct) = true;
//	option (mir.event_root) = true;
//	option (mir.event_class) = true;
//	option (mir.event) = true;
//	option (net.message_root) = true;
//	option (net.message_class) = true;
//	option (net.message) = true;
//
// Among these, (mir.struct) is the only one that has no special meaning.
func ShouldGenerateMirType(protoDesc protoreflect.MessageDescriptor) bool {
	return IsMirStruct(protoDesc) ||
		IsMirEventRoot(protoDesc) || IsMirEventClass(protoDesc) || IsMirEvent(protoDesc) ||
		IsNetMessageRoot(protoDesc) || IsNetMessageClass(protoDesc) || IsNetMessage(protoDesc)
}

func IsMirStruct(protoDesc protoreflect.MessageDescriptor) bool {
	return proto.GetExtension(protoDesc.Options().(*descriptorpb.MessageOptions), mir.E_Struct).(bool)
}

func IsMirEventClass(protoDesc protoreflect.MessageDescriptor) bool {
	return proto.GetExtension(protoDesc.Options().(*descriptorpb.MessageOptions), mir.E_EventClass).(bool)
}

func IsMirEventRoot(protoDesc protoreflect.MessageDescriptor) bool {
	return proto.GetExtension(protoDesc.Options().(*descriptorpb.MessageOptions), mir.E_EventRoot).(bool)
}

func IsMirEvent(protoDesc protoreflect.MessageDescriptor) bool {
	return proto.GetExtension(protoDesc.Options().(*descriptorpb.MessageOptions), mir.E_Event).(bool)
}

func IsNetMessageRoot(protoDesc protoreflect.MessageDescriptor) bool {
	return proto.GetExtension(protoDesc.Options().(*descriptorpb.MessageOptions), net.E_MessageRoot).(bool)
}

func IsNetMessageClass(protoDesc protoreflect.MessageDescriptor) bool {
	return proto.GetExtension(protoDesc.Options().(*descriptorpb.MessageOptions), net.E_MessageClass).(bool)
}

func IsNetMessage(protoDesc protoreflect.MessageDescriptor) bool {
	return proto.GetExtension(protoDesc.Options().(*descriptorpb.MessageOptions), net.E_Message).(bool)
}
