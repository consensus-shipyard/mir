package batchdbpb

import (
	reflect "reflect"
)

func (*Event) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*Event_Lookup)(nil)),
		reflect.TypeOf((*Event_LookupResponse)(nil)),
		reflect.TypeOf((*Event_Store)(nil)),
		reflect.TypeOf((*Event_Stored)(nil)),
	}
}
