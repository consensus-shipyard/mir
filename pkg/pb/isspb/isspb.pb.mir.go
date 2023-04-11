package isspb

import (
	reflect "reflect"
)

func (*ISSMessage) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*ISSMessage_StableCheckpoint)(nil)),
		reflect.TypeOf((*ISSMessage_RetransmitRequests)(nil)),
	}
}

func (*ISSEvent) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*ISSEvent_PersistCheckpoint)(nil)),
		reflect.TypeOf((*ISSEvent_PersistStableCheckpoint)(nil)),
		reflect.TypeOf((*ISSEvent_PushCheckpoint)(nil)),
		reflect.TypeOf((*ISSEvent_SbDeliver)(nil)),
	}
}
