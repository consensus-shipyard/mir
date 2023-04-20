package apppb

import (
	reflect "reflect"
)

func (*Event) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*Event_SnapshotRequest)(nil)),
		reflect.TypeOf((*Event_Snapshot)(nil)),
		reflect.TypeOf((*Event_RestoreState)(nil)),
		reflect.TypeOf((*Event_NewEpoch)(nil)),
	}
}
