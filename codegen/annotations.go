package codegen

import (
	"github.com/filecoin-project/mir/pkg/pb/mir"
	"github.com/filecoin-project/mir/pkg/pb/net"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

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
