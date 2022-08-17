package mempoolpb

import (
	reflect "reflect"
)

func (*Event) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*Event_RequestBatch)(nil)),
		reflect.TypeOf((*Event_NewBatch)(nil)),
		reflect.TypeOf((*Event_RequestTransactions)(nil)),
		reflect.TypeOf((*Event_TransactionsResponse)(nil)),
		reflect.TypeOf((*Event_RequestTransactionIds)(nil)),
		reflect.TypeOf((*Event_TransactionIdsResponse)(nil)),
		reflect.TypeOf((*Event_RequestBatchId)(nil)),
		reflect.TypeOf((*Event_BatchIdResponse)(nil)),
	}
}

func (*RequestBatchOrigin) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*RequestBatchOrigin_ContextStore)(nil)),
		reflect.TypeOf((*RequestBatchOrigin_Dsl)(nil)),
	}
}

func (*RequestTransactionsOrigin) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*RequestTransactionsOrigin_ContextStore)(nil)),
		reflect.TypeOf((*RequestTransactionsOrigin_Dsl)(nil)),
	}
}

func (*RequestTransactionIDsOrigin) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*RequestTransactionIDsOrigin_ContextStore)(nil)),
		reflect.TypeOf((*RequestTransactionIDsOrigin_Dsl)(nil)),
	}
}

func (*RequestBatchIDOrigin) ReflectTypeOptions() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf((*RequestBatchIDOrigin_ContextStore)(nil)),
		reflect.TypeOf((*RequestBatchIDOrigin_Dsl)(nil)),
	}
}
