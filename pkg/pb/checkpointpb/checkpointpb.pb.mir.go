package checkpointpb

import (
	reflect "reflect"
)

func (*Event) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*Event_EpochConfig)(nil)),
		reflect.TypeOf((*Event_StableCheckpoint)(nil)),
		reflect.TypeOf((*Event_EpochProgress)(nil)),
	}
}
