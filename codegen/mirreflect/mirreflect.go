package mirreflect

import "reflect"

type GeneratedType interface {
	MirReflect() Type
}

// Type represents runtime information about a Mir-generated type.
type Type interface {
	// PbType returns the protoc-generated type that corresponds to this Mir-generated type.
	PbType() reflect.Type
}

// TypeImpl is an implementation of the Type interface.
// As new functionality is being added to the mirreflect package,
// this type is likely to be removed and replaced by more specific types.
type TypeImpl struct {
	PbType_ reflect.Type // nolint
}

func (t TypeImpl) PbType() reflect.Type {
	return t.PbType_
}
