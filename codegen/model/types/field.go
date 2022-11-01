package types

import (
	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/filecoin-project/mir/codegen/util/astutil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

// Field represents a field in a protobuf message.
// Note: oneofs are not considered as fields in the protobuf data module, but are considered as fields here.
// The reason for that is that oneofs are mapped to fields in the generated Go code.
type Field struct {
	// The name of the field.
	Name string

	// The information about the type of the field.
	Type Type

	// The type to which the field belongs.
	Parent *Message

	// The protobuf descriptor of the field.
	// The descriptor can be either protoreflect.FieldDescriptor or protoreflect.OneofDescriptor.
	ProtoDesc protoreflect.Descriptor
}

// LowercaseName returns the lowercase name of the field.
func (f *Field) LowercaseName() string {
	return astutil.ToUnexported(f.Name)
}

// FuncParamPbType returns the field lowercase name followed by its pb type.
func (f *Field) FuncParamPbType() jen.Code {
	return jen.Id(f.LowercaseName()).Add(f.Type.PbType())
}

// FuncParamMirType returns the field lowercase name followed by its mir type.
func (f *Field) FuncParamMirType() jen.Code {
	return jen.Id(f.LowercaseName()).Add(f.Type.MirType())
}

// Fields is a list of fields of a protobuf message.
type Fields []*Field

// FuncParamsPbTypes returns a list of field lowercase names followed by their pb types.
func (fs Fields) FuncParamsPbTypes() []jen.Code {
	return sliceutil.Transform(fs, func(i int, f *Field) jen.Code { return f.FuncParamPbType() })
}

// FuncParamsMirTypes returns a list of field lowercase names followed by their mir types.
func (fs Fields) FuncParamsMirTypes() []jen.Code {
	return sliceutil.Transform(fs, func(i int, f *Field) jen.Code { return f.FuncParamMirType() })
}

// FuncParamsIDs returns a list of fields lowercase names as identifiers, without the types.
func (fs Fields) FuncParamsIDs() []jen.Code {
	return sliceutil.Transform(fs, func(i int, f *Field) jen.Code { return jen.Id(f.LowercaseName()) })
}

// ByName returns the field with the given name (or nil if there is no such field).
func (fs Fields) ByName(name string) *Field {
	for _, f := range fs {
		if f.Name == name {
			return f
		}
	}
	return nil
}
