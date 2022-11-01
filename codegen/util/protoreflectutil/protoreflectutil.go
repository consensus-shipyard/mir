package protoreflectutil

import (
	"reflect"
	"strings"

	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/filecoin-project/mir/pkg/util/reflectutil"
)

func IsProtoMessage(goStructPtr reflect.Type) bool {
	return goStructPtr.Kind() == reflect.Pointer &&
		goStructPtr.Implements(reflectutil.TypeOf[protoreflect.ProtoMessage]())
}

func IsOneofOption(ptrType reflect.Type) bool {
	if ptrType.Kind() != reflect.Pointer || ptrType.Elem().Kind() != reflect.Struct || ptrType.Elem().NumField() != 1 {
		// Oneof wrapper is a pointer to a struct with a single field.
		return false
	}

	tag, ok := ptrType.Elem().Field(0).Tag.Lookup("protobuf")
	if !ok {
		return false
	}

	return slices.Index(strings.Split(tag, ","), "oneof") != -1
}

func DescriptorForType(goStructPtr reflect.Type) (protoreflect.MessageDescriptor, bool) {
	nilMsg, ok := reflect.Zero(goStructPtr).Interface().(protoreflect.ProtoMessage)
	if !ok {
		return nil, false
	}
	return nilMsg.ProtoReflect().Descriptor(), true
}
