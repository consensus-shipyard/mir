package isspb

import (
	reflect "reflect"
)

func (*ISSMessage) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*ISSMessage_StableCheckpoint)(nil)),
	}
}

func (*Event) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*Event_PushCheckpoint)(nil)),
		reflect.TypeOf((*Event_SbDeliver)(nil)),
		reflect.TypeOf((*Event_DeliverCert)(nil)),
		reflect.TypeOf((*Event_NewConfig)(nil)),
	}
}
