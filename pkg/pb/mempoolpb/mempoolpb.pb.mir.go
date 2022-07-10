package mempoolpb

type Event_Type = isEvent_Type

type Event_TypeWrapper[Ev any] interface {
	Event_Type
	Unwrap() *Ev
}

func (p *Event_RequestBatch) Unwrap() *RequestBatch {
	return p.RequestBatch
}

func (p *Event_NewBatch) Unwrap() *NewBatch {
	return p.NewBatch
}

func (p *Event_RequestTransactions) Unwrap() *RequestTransactions {
	return p.RequestTransactions
}

func (p *Event_TransactionsResponse) Unwrap() *TransactionsResponse {
	return p.TransactionsResponse
}

func (p *Event_RequestTransactionIds) Unwrap() *RequestTransactionIDs {
	return p.RequestTransactionIds
}

func (p *Event_TransactionIdsResponse) Unwrap() *TransactionIDsResponse {
	return p.TransactionIdsResponse
}

func (p *Event_RequestBatchId) Unwrap() *RequestBatchID {
	return p.RequestBatchId
}

func (p *Event_BatchIdResponse) Unwrap() *BatchIDResponse {
	return p.BatchIdResponse
}
