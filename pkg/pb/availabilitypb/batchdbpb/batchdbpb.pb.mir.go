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

func (*LookupBatchOrigin) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*LookupBatchOrigin_ContextStore)(nil)),
		reflect.TypeOf((*LookupBatchOrigin_Dsl)(nil)),
	}
}

func (*StoreBatchOrigin) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*StoreBatchOrigin_ContextStore)(nil)),
		reflect.TypeOf((*StoreBatchOrigin_Dsl)(nil)),
	}
}
