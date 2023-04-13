package hasherpb

import (
	reflect "reflect"
)

func (*Event) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*Event_Request)(nil)),
		reflect.TypeOf((*Event_Result)(nil)),
		reflect.TypeOf((*Event_RequestOne)(nil)),
		reflect.TypeOf((*Event_ResultOne)(nil)),
	}
}

func (*HashOrigin) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*HashOrigin_ContextStore)(nil)),
		reflect.TypeOf((*HashOrigin_Request)(nil)),
		reflect.TypeOf((*HashOrigin_Dsl)(nil)),
		reflect.TypeOf((*HashOrigin_Checkpoint)(nil)),
		reflect.TypeOf((*HashOrigin_Sb)(nil)),
	}
}
