package types

import (
	"reflect"

	"github.com/dave/jennifer/jen"
)

var thisPackage = reflect.TypeOf(Slice{}).PkgPath()

// Slice is used to represent repeated fields.
type Slice struct {
	Underlying Type
}

func (s Slice) Same() bool {
	return s.Underlying.Same()
}

func (s Slice) PbType() *jen.Statement {
	return jen.Index().Add(s.Underlying.PbType())
}

func (s Slice) MirType() *jen.Statement {
	return jen.Index().Add(s.Underlying.MirType())
}

func (s Slice) ToMir(code jen.Code) *jen.Statement {
	if s.Same() {
		return jen.Add(code)
	}

	return jen.Qual(thisPackage, "ConvertSlice").Call(code,
		jen.Func().Params(jen.Id("t").Add(s.Underlying.PbType())).Add(s.Underlying.MirType()).Block(
			jen.Return(s.Underlying.ToMir(jen.Id("t"))),
		))
}

func (s Slice) ToPb(code jen.Code) *jen.Statement {
	if s.Same() {
		return jen.Add(code)
	}

	return jen.Qual(thisPackage, "ConvertSlice").Call(code,
		jen.Func().Params(jen.Id("t").Add(s.Underlying.MirType())).Add(s.Underlying.PbType()).Block(
			jen.Return(s.Underlying.ToPb(jen.Id("t"))),
		))
}

// ConvertSlice is used by the generated code.
func ConvertSlice[T, R any](ts []T, f func(t T) R) []R {
	rs := make([]R, len(ts))
	for i, t := range ts {
		rs[i] = f(t)
	}
	return rs
}
