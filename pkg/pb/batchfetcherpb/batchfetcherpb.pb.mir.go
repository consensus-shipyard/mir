package batchfetcherpb

import (
	reflect "reflect"
)

func (*Event) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*Event_NewOrderedBatch)(nil)),
		reflect.TypeOf((*Event_ClientProgress)(nil)),
	}
}
