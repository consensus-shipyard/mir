package transportpb

import (
	reflect "reflect"
)

func (*Event) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*Event_SendMessage)(nil)),
		reflect.TypeOf((*Event_MessageReceived)(nil)),
	}
}
