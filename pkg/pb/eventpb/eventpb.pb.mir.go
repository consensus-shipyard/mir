package eventpb

import (
	reflect "reflect"
)

func (*Event) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*Event_Init)(nil)),
		reflect.TypeOf((*Event_Timer)(nil)),
		reflect.TypeOf((*Event_Hasher)(nil)),
		reflect.TypeOf((*Event_Bcb)(nil)),
		reflect.TypeOf((*Event_Mempool)(nil)),
		reflect.TypeOf((*Event_Availability)(nil)),
		reflect.TypeOf((*Event_BatchDb)(nil)),
		reflect.TypeOf((*Event_BatchFetcher)(nil)),
		reflect.TypeOf((*Event_ThreshCrypto)(nil)),
		reflect.TypeOf((*Event_Checkpoint)(nil)),
		reflect.TypeOf((*Event_Factory)(nil)),
		reflect.TypeOf((*Event_Iss)(nil)),
		reflect.TypeOf((*Event_Orderer)(nil)),
		reflect.TypeOf((*Event_Crypto)(nil)),
		reflect.TypeOf((*Event_App)(nil)),
		reflect.TypeOf((*Event_Transport)(nil)),
		reflect.TypeOf((*Event_PingPong)(nil)),
		reflect.TypeOf((*Event_TestingString)(nil)),
		reflect.TypeOf((*Event_TestingUint)(nil)),
	}
}

func (*TimerEvent) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*TimerEvent_Delay)(nil)),
		reflect.TypeOf((*TimerEvent_Repeat)(nil)),
		reflect.TypeOf((*TimerEvent_GarbageCollect)(nil)),
	}
}
