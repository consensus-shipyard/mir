package types

import (
	"fmt"
	"reflect"

	"github.com/dave/jennifer/jen"

	"github.com/filecoin-project/mir/codegen/util/jenutil"
)

type Oneof struct {
	Name    string
	Parent  *Message
	Options []*OneofOption
}

func (t *Oneof) Same() bool {
	return t.Parent.Same()
}

func (t *Oneof) PbExportedInterfaceName() string {
	return fmt.Sprintf("%v_%v", t.Parent.Name(), t.Name)
}

func (t *Oneof) PbNativeInterfaceName() string {
	return fmt.Sprintf("is%v_%v", t.Parent.Name(), t.Name)
}

func (t *Oneof) MirInterfaceName() string {
	return fmt.Sprintf("%v_%v", t.Parent.Name(), t.Name)
}

func (t *Oneof) MirWrapperInterfaceName() string {
	return t.MirInterfaceName() + "Wrapper"
}

func (t *Oneof) MirWrapperInterface() *jen.Statement {
	return jen.Qual(t.Parent.MirPkgPath(), t.MirWrapperInterfaceName())
}

func (t *Oneof) PbMethodName() string {
	return "is" + t.PbExportedInterfaceName()
}

func (t *Oneof) MirMethodName() string {
	return t.PbMethodName()
}

func (t *Oneof) PbType() *jen.Statement {
	return jen.Qual(t.Parent.PbPkgPath(), t.PbExportedInterfaceName())
}

func (t *Oneof) MirType() *jen.Statement {
	return jen.Qual(t.Parent.MirPkgPath(), t.MirInterfaceName())
}

func (t *Oneof) ToMir(code jen.Code) *jen.Statement {
	return jen.Qual(t.Parent.MirPkgPath(), t.MirInterfaceName()+"FromPb").Call(code)
}

func (t *Oneof) ToPb(code jen.Code) *jen.Statement {
	return jen.Parens(code).Dot("Pb").Call()
}

type OneofOption struct {
	PbWrapperReflect reflect.Type
	WrapperName      string
	Field            *Field
}

// Name returns the name of the oneof option.
// Same as opt.Field.Name.
func (opt *OneofOption) Name() string {
	return opt.Field.Name
}

func (opt *OneofOption) PbWrapperType() *jen.Statement {
	return jen.Op("*").Add(jenutil.QualFromType(opt.PbWrapperReflect.Elem()))
}

func (opt *OneofOption) NewPbWrapperType() *jen.Statement {
	return jen.Op("&").Add(jenutil.QualFromType(opt.PbWrapperReflect.Elem()))
}

func (opt *OneofOption) MirWrapperStructType() *jen.Statement {
	return jen.Qual(PackagePath(opt.PbWrapperReflect.Elem().PkgPath()), opt.WrapperName)
}

func (opt *OneofOption) MirWrapperType() *jen.Statement {
	return jen.Op("*").Add(opt.MirWrapperStructType())
}

func (opt *OneofOption) NewMirWrapperType() *jen.Statement {
	return jen.Op("&").Add(opt.MirWrapperStructType())
}

func (opt *OneofOption) ConstructMirWrapperType(underlying jen.Code) *jen.Statement {
	return opt.NewMirWrapperType().Values(
		jen.Line().Id(opt.Field.Name).Op(":").Add(underlying),
		jen.Line(),
	)
}
