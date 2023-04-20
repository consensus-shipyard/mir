package cryptopb

import (
	reflect "reflect"
)

func (*Event) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*Event_SignRequest)(nil)),
		reflect.TypeOf((*Event_SignResult)(nil)),
		reflect.TypeOf((*Event_VerifySig)(nil)),
		reflect.TypeOf((*Event_SigVerified)(nil)),
		reflect.TypeOf((*Event_VerifySigs)(nil)),
		reflect.TypeOf((*Event_SigsVerified)(nil)),
	}
}

func (*SignOrigin) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*SignOrigin_ContextStore)(nil)),
		reflect.TypeOf((*SignOrigin_Dsl)(nil)),
		reflect.TypeOf((*SignOrigin_Checkpoint)(nil)),
		reflect.TypeOf((*SignOrigin_Sb)(nil)),
	}
}

func (*SigVerOrigin) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*SigVerOrigin_ContextStore)(nil)),
		reflect.TypeOf((*SigVerOrigin_Dsl)(nil)),
		reflect.TypeOf((*SigVerOrigin_Checkpoint)(nil)),
		reflect.TypeOf((*SigVerOrigin_Sb)(nil)),
	}
}
