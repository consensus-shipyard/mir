package types

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/filecoin-project/mir/codegen"
	"github.com/filecoin-project/mir/codegen/util/astutil"
)

// Message contains the information needed to generate code for a protobuf message.
type Message struct {
	shouldGenerateMirType bool
	mirPkgPath            string

	pbStructType         jen.Code
	mirStructType        jen.Code
	protoDesc            protoreflect.MessageDescriptor
	pbGoStructPtrReflect reflect.Type
}

func (m *Message) Name() string {
	return m.pbGoStructPtrReflect.Elem().Name()
}

func (m *Message) PbPkgPath() string {
	return m.pbGoStructPtrReflect.Elem().PkgPath()
}

func (m *Message) MirPkgPath() string {
	// Return the cached value if present.
	if m.mirPkgPath != "" {
		return m.mirPkgPath
	}

	m.mirPkgPath = PackagePath(m.PbPkgPath())
	return m.mirPkgPath
}

func (m *Message) Same() bool {
	return !m.ShouldGenerateMirType()
}

// PbReflectType returns the reflect.Type corresponding to the pointer
// type to the protoc-generated struct for this message.
func (m *Message) PbReflectType() reflect.Type {
	return m.pbGoStructPtrReflect
}

func (m *Message) PbType() *jen.Statement {
	return jen.Op("*").Add(m.pbStructType)
}

func (m *Message) NewPbType() *jen.Statement {
	return jen.Op("&").Add(m.pbStructType)
}

func (m *Message) NewMirType() *jen.Statement {
	return jen.Op("&").Add(m.mirStructType)
}

func (m *Message) MirType() *jen.Statement {
	return jen.Op("*").Add(m.mirStructType)
}

func (m *Message) ToMir(code jen.Code) *jen.Statement {
	if m.Same() {
		return jen.Add(code)
	}
	return jen.Qual(m.MirPkgPath(), m.Name()+"FromPb").Call(code)
}

func (m *Message) ToPb(code jen.Code) *jen.Statement {
	if m.Same() {
		return jen.Add(code)
	}
	return jen.Parens(code).Dot("Pb").Call()
}

// LowercaseName returns the name of the message in lowercase.
func (m *Message) LowercaseName() string {
	return astutil.ToUnexported(m.Name())
}

// ProtoDesc returns the proto descriptor of the message.
func (m *Message) ProtoDesc() protoreflect.MessageDescriptor {
	return m.protoDesc
}

func (m *Message) IsMirEvent() bool {
	return codegen.IsMirEvent(m.protoDesc)
}

func (m *Message) IsMirMessage() bool {
	return codegen.IsNetMessage(m.protoDesc)
}

func (m *Message) IsMirStruct() bool {
	return codegen.IsMirStruct(m.protoDesc)
}

// ShouldGenerateMirType returns true if Mir should generate a struct for the message type.
func (m *Message) ShouldGenerateMirType() bool {
	return m.shouldGenerateMirType
}

func getProtoNameOfField(field reflect.StructField) (protoName protoreflect.Name, err error) {
	protobufTag, ok := field.Tag.Lookup("protobuf")
	if !ok {
		return "", fmt.Errorf("field %v has no protobuf tag", field.Name)
	}

	for _, tagPart := range strings.Split(protobufTag, ",") {
		if strings.HasPrefix(tagPart, "name=") {
			return protoreflect.Name(strings.TrimPrefix(tagPart, "name=")), nil
		}
	}

	return "", fmt.Errorf("proto name of field %v is not specified in the tag", field.Name)
}
