package types

import (
	"errors"

	"github.com/dave/jennifer/jen"
)

// Error represents a string field in a message annotated with [(mir.type) = "error"].
type Error struct{}

func (t Error) Same() bool {
	return false
}

func (t Error) PbType() *jen.Statement {
	return jen.Id("string")
}

func (t Error) MirType() *jen.Statement {
	return jen.Id("error")
}

func (t Error) ToMir(code jen.Code) *jen.Statement {
	return jen.Qual(thisPackage, "StringToError").Call(code)
}

func (t Error) ToPb(code jen.Code) *jen.Statement {
	return jen.Qual(thisPackage, "ErrorToString").Call(code)
}

// StringToError is used in the generated code to convert a string to an error.
func StringToError(s string) error {
	if s == "" {
		return nil
	}
	return errors.New(s)
}

// ErrorToString is used in the generated code to convert an error to a string.
func ErrorToString(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	if s == "" {
		panic("error.Error() must not return an empty string")
	}
	return s
}
