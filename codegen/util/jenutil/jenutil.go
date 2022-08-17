package jenutil

import (
	"fmt"
	"reflect"

	"github.com/dave/jennifer/jen"
)

func QualFromType(tp reflect.Type) *jen.Statement {
	switch tp.Kind() {

	case reflect.Array:
		return jen.Index(jen.Lit(tp.Len())).Add(QualFromType(tp.Elem()))

	case reflect.Pointer:
		return jen.Op("*").Add(QualFromType(tp.Elem()))

	case reflect.Slice:
		return jen.Index().Add(QualFromType(tp.Elem()))

	case reflect.Map:
		return jen.Map(QualFromType(tp.Key())).Add(QualFromType(tp.Elem()))

	case reflect.Interface, reflect.Struct, reflect.UnsafePointer:
		if tp.Name() == "" {
			panic("non-defined types are not supported")
		}
		return jen.Qual(tp.PkgPath(), tp.Name())

	case reflect.Chan, reflect.Func:
		switch tp.ChanDir() {
		case reflect.RecvDir:
			return jen.Op("<-").Chan().Add(QualFromType(tp.Elem()))
		case reflect.SendDir:
			return jen.Chan().Op("<-").Add(QualFromType(tp.Elem()))
		case reflect.BothDir:
			return jen.Chan().Add(QualFromType(tp.Elem()))
		default:
			panic(fmt.Errorf("unexpected ChanDir: %v", tp.ChanDir()))
		}

	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32,
		reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String:
		if tp.PkgPath() != "" {
			return jen.Qual(tp.PkgPath(), tp.Name())
		}
		return jen.Id(tp.Name())

	default:
		panic(fmt.Errorf("unknown go type kind: %v", tp.Kind()))
	}
}
