package types

import (
	"path"
	"strings"

	"github.com/dave/jennifer/jen"
)

func PackagePath(sourcePackagePath string) string {
	return sourcePackagePath + "/types"
}

func PackageName(sourcePackagePath string) string {
	return sourcePackagePath[strings.LastIndex(sourcePackagePath, "/")+1:] + "types"
}

func OutputDir(sourceDir string) string {
	return path.Join(sourceDir, "types")
}

type Type interface {
	// Same returns whether PbType() and MirType() represent the same type.
	Same() bool
	// PbType is the type in the protoc-generated code.
	PbType() *jen.Statement
	// MirType is the type in the Mir-generated code.
	MirType() *jen.Statement
	// ToMir converts an object of the type in protoc representation to its Mir representation.
	ToMir(code jen.Code) *jen.Statement
	// ToPb converts an object of the type in Mir representation to its protoc representation.
	ToPb(code jen.Code) *jen.Statement
}

// Same is used when the same type is used by protoc-generated code and Mir-generated code.
type Same struct {
	Type jen.Code
}

func (t Same) Same() bool {
	return true
}

func (t Same) PbType() *jen.Statement {
	return jen.Add(t.Type)
}

func (t Same) MirType() *jen.Statement {
	return jen.Add(t.Type)
}

func (t Same) ToMir(code jen.Code) *jen.Statement {
	return jen.Add(code)
}

func (t Same) ToPb(code jen.Code) *jen.Statement {
	return jen.Add(code)
}

// Castable is used when the types used by protoc-generated code and
// Mir-generated code can be directly cast to one another.
type Castable struct {
	PbType_  jen.Code // nolint
	MirType_ jen.Code // nolint
}

func (t Castable) Same() bool {
	return false
}

func (t Castable) PbType() *jen.Statement {
	return jen.Add(t.PbType_)
}

func (t Castable) MirType() *jen.Statement {
	return jen.Add(t.MirType_)
}

func (t Castable) ToMir(code jen.Code) *jen.Statement {
	return jen.Parens(t.MirType_).Call(code)
}

func (t Castable) ToPb(code jen.Code) *jen.Statement {
	return jen.Parens(t.PbType_).Call(code)
}
