// Code generated by Mir codegen. DO NOT EDIT.

package applicationpb

import (
	reflect "reflect"
)

func (*Event) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*Event_NewHead)(nil)),
		reflect.TypeOf((*Event_VerifyBlockRequest)(nil)),
		reflect.TypeOf((*Event_VerifyBlockResponse)(nil)),
		reflect.TypeOf((*Event_PayloadRequest)(nil)),
		reflect.TypeOf((*Event_PayloadResponse)(nil)),
		reflect.TypeOf((*Event_ForkUpdate)(nil)),
	}
}
