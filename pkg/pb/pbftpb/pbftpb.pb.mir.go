package pbftpb

import (
	reflect "reflect"
)

func (*Message) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*Message_Preprepare)(nil)),
		reflect.TypeOf((*Message_Prepare)(nil)),
		reflect.TypeOf((*Message_Commit)(nil)),
		reflect.TypeOf((*Message_Done)(nil)),
		reflect.TypeOf((*Message_CatchUpRequest)(nil)),
		reflect.TypeOf((*Message_CatchUpResponse)(nil)),
		reflect.TypeOf((*Message_SignedViewChange)(nil)),
		reflect.TypeOf((*Message_PreprepareRequest)(nil)),
		reflect.TypeOf((*Message_MissingPreprepare)(nil)),
		reflect.TypeOf((*Message_NewView)(nil)),
	}
}

func (*Event) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*Event_ProposeTimeout)(nil)),
		reflect.TypeOf((*Event_ViewChangeSnTimeout)(nil)),
		reflect.TypeOf((*Event_ViewChangeSegTimeout)(nil)),
	}
}

func (*HashOrigin) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*HashOrigin_Preprepare)(nil)),
		reflect.TypeOf((*HashOrigin_MissingPreprepare)(nil)),
		reflect.TypeOf((*HashOrigin_NewView)(nil)),
		reflect.TypeOf((*HashOrigin_EmptyPreprepares)(nil)),
		reflect.TypeOf((*HashOrigin_CatchUpResponse)(nil)),
	}
}
