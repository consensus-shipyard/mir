package availabilitypb

import (
	reflect "reflect"
)

func (*Event) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*Event_RequestCert)(nil)),
		reflect.TypeOf((*Event_NewCert)(nil)),
		reflect.TypeOf((*Event_VerifyCert)(nil)),
		reflect.TypeOf((*Event_CertVerified)(nil)),
		reflect.TypeOf((*Event_RequestTransactions)(nil)),
		reflect.TypeOf((*Event_ProvideTransactions)(nil)),
	}
}

func (*RequestCertOrigin) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*RequestCertOrigin_ContextStore)(nil)),
		reflect.TypeOf((*RequestCertOrigin_Dsl)(nil)),
	}
}

func (*VerifyCertOrigin) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*VerifyCertOrigin_ContextStore)(nil)),
		reflect.TypeOf((*VerifyCertOrigin_Dsl)(nil)),
	}
}
