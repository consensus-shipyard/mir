package isspb

import (
	reflect "reflect"
)

func (*ISSMessage) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*ISSMessage_StableCheckpoint)(nil)),
	}
}

func (*ISSEvent) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*ISSEvent_PushCheckpoint)(nil)),
		reflect.TypeOf((*ISSEvent_SbDeliver)(nil)),
	}
}
