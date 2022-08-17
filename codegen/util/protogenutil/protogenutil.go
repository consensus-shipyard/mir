package protogenutil

import (
	"reflect"

	"google.golang.org/protobuf/compiler/protogen"
)

func GoIdentByType(tp reflect.Type) protogen.GoIdent {
	return protogen.GoIdent{
		GoName:       tp.Name(),
		GoImportPath: protogen.GoImportPath(tp.PkgPath()),
	}
}
