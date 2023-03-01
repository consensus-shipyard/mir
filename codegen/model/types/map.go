package types

import (
	"reflect"

	"github.com/dave/jennifer/jen"
)

var _thisPackage = reflect.TypeOf(Map{}).PkgPath()

// Map is used to represent map types.
type Map struct {
	KeyType   Type
	ValueType Type
}

func (m Map) Same() bool {
	return m.KeyType.Same() && m.ValueType.Same()
}

func (m Map) PbType() *jen.Statement {
	return jen.Map(m.KeyType.PbType()).Add(m.ValueType.PbType())
}

func (m Map) MirType() *jen.Statement {
	return jen.Map(m.KeyType.MirType()).Add(m.ValueType.MirType())
}

func (m Map) ToMir(code jen.Code) *jen.Statement {
	if m.Same() {
		return jen.Add(code)
	}

	return jen.Qual(_thisPackage, "ConvertMap").Call(code,
		jen.Func().Params(jen.Id("k").Add(m.KeyType.PbType()), jen.Id("v").Add(m.ValueType.PbType())).Params(m.KeyType.MirType(), m.ValueType.MirType()).Block(
			jen.Return(m.KeyType.ToMir(jen.Id("k")), m.ValueType.ToMir(jen.Id("v"))),
		))
}

func (m Map) ToPb(code jen.Code) *jen.Statement {
	if m.Same() {
		return jen.Add(code)
	}

	return jen.Qual(_thisPackage, "ConvertMap").Call(code,
		jen.Func().Params(jen.Id("k").Add(m.KeyType.MirType()), jen.Id("v").Add(m.ValueType.MirType())).Params(m.KeyType.PbType(), m.ValueType.PbType()).Block(
			jen.Return(m.KeyType.ToPb(jen.Id("k")), m.ValueType.ToPb(jen.Id("v"))),
		))
}

// ConvertMap is used by the generated code.
func ConvertMap[K, Rk comparable, V, Rv any](tm map[K]V, f func(tk K, tv V) (Rk, Rv)) map[Rk]Rv {
	rm := make(map[Rk]Rv)
	for k, v := range tm {
		_k, _v := f(k, v)
		rm[_k] = _v
	}
	return rm
}
