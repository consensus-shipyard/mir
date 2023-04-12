package factorymodulepb

import (
	reflect "reflect"
)

func (*Factory) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*Factory_NewModule)(nil)),
		reflect.TypeOf((*Factory_GarbageCollect)(nil)),
	}
}

func (*GeneratorParams) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*GeneratorParams_MultisigCollector)(nil)),
		reflect.TypeOf((*GeneratorParams_Checkpoint)(nil)),
		reflect.TypeOf((*GeneratorParams_EchoTestModule)(nil)),
		reflect.TypeOf((*GeneratorParams_PbftModule)(nil)),
	}
}
