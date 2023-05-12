package mempoolpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types1 "github.com/filecoin-project/mir/codegen/model/types"
	types3 "github.com/filecoin-project/mir/pkg/availability/multisigcollector/types"
	types5 "github.com/filecoin-project/mir/pkg/pb/contextstorepb/types"
	types6 "github.com/filecoin-project/mir/pkg/pb/dslpb/types"
	mempoolpb "github.com/filecoin-project/mir/pkg/pb/mempoolpb"
	trantorpb "github.com/filecoin-project/mir/pkg/pb/trantorpb"
	types "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	types2 "github.com/filecoin-project/mir/pkg/trantor/types"
	types4 "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Event struct {
	Type Event_Type
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() mempoolpb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb mempoolpb.Event_Type) Event_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *mempoolpb.Event_RequestBatch:
		return &Event_RequestBatch{RequestBatch: RequestBatchFromPb(pb.RequestBatch)}
	case *mempoolpb.Event_NewBatch:
		return &Event_NewBatch{NewBatch: NewBatchFromPb(pb.NewBatch)}
	case *mempoolpb.Event_RequestTransactions:
		return &Event_RequestTransactions{RequestTransactions: RequestTransactionsFromPb(pb.RequestTransactions)}
	case *mempoolpb.Event_TransactionsResponse:
		return &Event_TransactionsResponse{TransactionsResponse: TransactionsResponseFromPb(pb.TransactionsResponse)}
	case *mempoolpb.Event_RequestTransactionIds:
		return &Event_RequestTransactionIds{RequestTransactionIds: RequestTransactionIDsFromPb(pb.RequestTransactionIds)}
	case *mempoolpb.Event_TransactionIdsResponse:
		return &Event_TransactionIdsResponse{TransactionIdsResponse: TransactionIDsResponseFromPb(pb.TransactionIdsResponse)}
	case *mempoolpb.Event_RequestBatchId:
		return &Event_RequestBatchId{RequestBatchId: RequestBatchIDFromPb(pb.RequestBatchId)}
	case *mempoolpb.Event_BatchIdResponse:
		return &Event_BatchIdResponse{BatchIdResponse: BatchIDResponseFromPb(pb.BatchIdResponse)}
	case *mempoolpb.Event_NewTransactions:
		return &Event_NewTransactions{NewTransactions: NewTransactionsFromPb(pb.NewTransactions)}
	}
	return nil
}

type Event_RequestBatch struct {
	RequestBatch *RequestBatch
}

func (*Event_RequestBatch) isEvent_Type() {}

func (w *Event_RequestBatch) Unwrap() *RequestBatch {
	return w.RequestBatch
}

func (w *Event_RequestBatch) Pb() mempoolpb.Event_Type {
	if w == nil {
		return nil
	}
	return &mempoolpb.Event_RequestBatch{RequestBatch: (w.RequestBatch).Pb()}
}

func (*Event_RequestBatch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.Event_RequestBatch]()}
}

type Event_NewBatch struct {
	NewBatch *NewBatch
}

func (*Event_NewBatch) isEvent_Type() {}

func (w *Event_NewBatch) Unwrap() *NewBatch {
	return w.NewBatch
}

func (w *Event_NewBatch) Pb() mempoolpb.Event_Type {
	if w == nil {
		return nil
	}
	return &mempoolpb.Event_NewBatch{NewBatch: (w.NewBatch).Pb()}
}

func (*Event_NewBatch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.Event_NewBatch]()}
}

type Event_RequestTransactions struct {
	RequestTransactions *RequestTransactions
}

func (*Event_RequestTransactions) isEvent_Type() {}

func (w *Event_RequestTransactions) Unwrap() *RequestTransactions {
	return w.RequestTransactions
}

func (w *Event_RequestTransactions) Pb() mempoolpb.Event_Type {
	if w == nil {
		return nil
	}
	return &mempoolpb.Event_RequestTransactions{RequestTransactions: (w.RequestTransactions).Pb()}
}

func (*Event_RequestTransactions) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.Event_RequestTransactions]()}
}

type Event_TransactionsResponse struct {
	TransactionsResponse *TransactionsResponse
}

func (*Event_TransactionsResponse) isEvent_Type() {}

func (w *Event_TransactionsResponse) Unwrap() *TransactionsResponse {
	return w.TransactionsResponse
}

func (w *Event_TransactionsResponse) Pb() mempoolpb.Event_Type {
	if w == nil {
		return nil
	}
	return &mempoolpb.Event_TransactionsResponse{TransactionsResponse: (w.TransactionsResponse).Pb()}
}

func (*Event_TransactionsResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.Event_TransactionsResponse]()}
}

type Event_RequestTransactionIds struct {
	RequestTransactionIds *RequestTransactionIDs
}

func (*Event_RequestTransactionIds) isEvent_Type() {}

func (w *Event_RequestTransactionIds) Unwrap() *RequestTransactionIDs {
	return w.RequestTransactionIds
}

func (w *Event_RequestTransactionIds) Pb() mempoolpb.Event_Type {
	if w == nil {
		return nil
	}
	return &mempoolpb.Event_RequestTransactionIds{RequestTransactionIds: (w.RequestTransactionIds).Pb()}
}

func (*Event_RequestTransactionIds) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.Event_RequestTransactionIds]()}
}

type Event_TransactionIdsResponse struct {
	TransactionIdsResponse *TransactionIDsResponse
}

func (*Event_TransactionIdsResponse) isEvent_Type() {}

func (w *Event_TransactionIdsResponse) Unwrap() *TransactionIDsResponse {
	return w.TransactionIdsResponse
}

func (w *Event_TransactionIdsResponse) Pb() mempoolpb.Event_Type {
	if w == nil {
		return nil
	}
	return &mempoolpb.Event_TransactionIdsResponse{TransactionIdsResponse: (w.TransactionIdsResponse).Pb()}
}

func (*Event_TransactionIdsResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.Event_TransactionIdsResponse]()}
}

type Event_RequestBatchId struct {
	RequestBatchId *RequestBatchID
}

func (*Event_RequestBatchId) isEvent_Type() {}

func (w *Event_RequestBatchId) Unwrap() *RequestBatchID {
	return w.RequestBatchId
}

func (w *Event_RequestBatchId) Pb() mempoolpb.Event_Type {
	if w == nil {
		return nil
	}
	return &mempoolpb.Event_RequestBatchId{RequestBatchId: (w.RequestBatchId).Pb()}
}

func (*Event_RequestBatchId) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.Event_RequestBatchId]()}
}

type Event_BatchIdResponse struct {
	BatchIdResponse *BatchIDResponse
}

func (*Event_BatchIdResponse) isEvent_Type() {}

func (w *Event_BatchIdResponse) Unwrap() *BatchIDResponse {
	return w.BatchIdResponse
}

func (w *Event_BatchIdResponse) Pb() mempoolpb.Event_Type {
	if w == nil {
		return nil
	}
	return &mempoolpb.Event_BatchIdResponse{BatchIdResponse: (w.BatchIdResponse).Pb()}
}

func (*Event_BatchIdResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.Event_BatchIdResponse]()}
}

type Event_NewTransactions struct {
	NewTransactions *NewTransactions
}

func (*Event_NewTransactions) isEvent_Type() {}

func (w *Event_NewTransactions) Unwrap() *NewTransactions {
	return w.NewTransactions
}

func (w *Event_NewTransactions) Pb() mempoolpb.Event_Type {
	if w == nil {
		return nil
	}
	return &mempoolpb.Event_NewTransactions{NewTransactions: (w.NewTransactions).Pb()}
}

func (*Event_NewTransactions) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.Event_NewTransactions]()}
}

func EventFromPb(pb *mempoolpb.Event) *Event {
	if pb == nil {
		return nil
	}
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *mempoolpb.Event {
	if m == nil {
		return nil
	}
	return &mempoolpb.Event{
		Type: (m.Type).Pb(),
	}
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.Event]()}
}

type NewTransactions struct {
	Transactions []*types.Transaction
}

func NewTransactionsFromPb(pb *mempoolpb.NewTransactions) *NewTransactions {
	if pb == nil {
		return nil
	}
	return &NewTransactions{
		Transactions: types1.ConvertSlice(pb.Transactions, func(t *trantorpb.Transaction) *types.Transaction {
			return types.TransactionFromPb(t)
		}),
	}
}

func (m *NewTransactions) Pb() *mempoolpb.NewTransactions {
	if m == nil {
		return nil
	}
	return &mempoolpb.NewTransactions{
		Transactions: types1.ConvertSlice(m.Transactions, func(t *types.Transaction) *trantorpb.Transaction {
			return (t).Pb()
		}),
	}
}

func (*NewTransactions) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.NewTransactions]()}
}

type RequestBatch struct {
	Origin *RequestBatchOrigin
}

func RequestBatchFromPb(pb *mempoolpb.RequestBatch) *RequestBatch {
	if pb == nil {
		return nil
	}
	return &RequestBatch{
		Origin: RequestBatchOriginFromPb(pb.Origin),
	}
}

func (m *RequestBatch) Pb() *mempoolpb.RequestBatch {
	if m == nil {
		return nil
	}
	return &mempoolpb.RequestBatch{
		Origin: (m.Origin).Pb(),
	}
}

func (*RequestBatch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.RequestBatch]()}
}

type NewBatch struct {
	TxIds  []types2.TxID
	Txs    []*types.Transaction
	Origin *RequestBatchOrigin
}

func NewBatchFromPb(pb *mempoolpb.NewBatch) *NewBatch {
	if pb == nil {
		return nil
	}
	return &NewBatch{
		TxIds: types1.ConvertSlice(pb.TxIds, func(t []uint8) types2.TxID {
			return (types2.TxID)(t)
		}),
		Txs: types1.ConvertSlice(pb.Txs, func(t *trantorpb.Transaction) *types.Transaction {
			return types.TransactionFromPb(t)
		}),
		Origin: RequestBatchOriginFromPb(pb.Origin),
	}
}

func (m *NewBatch) Pb() *mempoolpb.NewBatch {
	if m == nil {
		return nil
	}
	return &mempoolpb.NewBatch{
		TxIds: types1.ConvertSlice(m.TxIds, func(t types2.TxID) []uint8 {
			return ([]uint8)(t)
		}),
		Txs: types1.ConvertSlice(m.Txs, func(t *types.Transaction) *trantorpb.Transaction {
			return (t).Pb()
		}),
		Origin: (m.Origin).Pb(),
	}
}

func (*NewBatch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.NewBatch]()}
}

type RequestTransactions struct {
	TxIds  []types2.TxID
	Origin *RequestTransactionsOrigin
}

func RequestTransactionsFromPb(pb *mempoolpb.RequestTransactions) *RequestTransactions {
	if pb == nil {
		return nil
	}
	return &RequestTransactions{
		TxIds: types1.ConvertSlice(pb.TxIds, func(t []uint8) types2.TxID {
			return (types2.TxID)(t)
		}),
		Origin: RequestTransactionsOriginFromPb(pb.Origin),
	}
}

func (m *RequestTransactions) Pb() *mempoolpb.RequestTransactions {
	if m == nil {
		return nil
	}
	return &mempoolpb.RequestTransactions{
		TxIds: types1.ConvertSlice(m.TxIds, func(t types2.TxID) []uint8 {
			return ([]uint8)(t)
		}),
		Origin: (m.Origin).Pb(),
	}
}

func (*RequestTransactions) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.RequestTransactions]()}
}

type TransactionsResponse struct {
	Present []bool
	Txs     []*types.Transaction
	Origin  *RequestTransactionsOrigin
}

func TransactionsResponseFromPb(pb *mempoolpb.TransactionsResponse) *TransactionsResponse {
	if pb == nil {
		return nil
	}
	return &TransactionsResponse{
		Present: pb.Present,
		Txs: types1.ConvertSlice(pb.Txs, func(t *trantorpb.Transaction) *types.Transaction {
			return types.TransactionFromPb(t)
		}),
		Origin: RequestTransactionsOriginFromPb(pb.Origin),
	}
}

func (m *TransactionsResponse) Pb() *mempoolpb.TransactionsResponse {
	if m == nil {
		return nil
	}
	return &mempoolpb.TransactionsResponse{
		Present: m.Present,
		Txs: types1.ConvertSlice(m.Txs, func(t *types.Transaction) *trantorpb.Transaction {
			return (t).Pb()
		}),
		Origin: (m.Origin).Pb(),
	}
}

func (*TransactionsResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.TransactionsResponse]()}
}

type RequestTransactionIDs struct {
	Txs    []*types.Transaction
	Origin *RequestTransactionIDsOrigin
}

func RequestTransactionIDsFromPb(pb *mempoolpb.RequestTransactionIDs) *RequestTransactionIDs {
	if pb == nil {
		return nil
	}
	return &RequestTransactionIDs{
		Txs: types1.ConvertSlice(pb.Txs, func(t *trantorpb.Transaction) *types.Transaction {
			return types.TransactionFromPb(t)
		}),
		Origin: RequestTransactionIDsOriginFromPb(pb.Origin),
	}
}

func (m *RequestTransactionIDs) Pb() *mempoolpb.RequestTransactionIDs {
	if m == nil {
		return nil
	}
	return &mempoolpb.RequestTransactionIDs{
		Txs: types1.ConvertSlice(m.Txs, func(t *types.Transaction) *trantorpb.Transaction {
			return (t).Pb()
		}),
		Origin: (m.Origin).Pb(),
	}
}

func (*RequestTransactionIDs) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.RequestTransactionIDs]()}
}

type TransactionIDsResponse struct {
	TxIds  []types2.TxID
	Origin *RequestTransactionIDsOrigin
}

func TransactionIDsResponseFromPb(pb *mempoolpb.TransactionIDsResponse) *TransactionIDsResponse {
	if pb == nil {
		return nil
	}
	return &TransactionIDsResponse{
		TxIds: types1.ConvertSlice(pb.TxIds, func(t []uint8) types2.TxID {
			return (types2.TxID)(t)
		}),
		Origin: RequestTransactionIDsOriginFromPb(pb.Origin),
	}
}

func (m *TransactionIDsResponse) Pb() *mempoolpb.TransactionIDsResponse {
	if m == nil {
		return nil
	}
	return &mempoolpb.TransactionIDsResponse{
		TxIds: types1.ConvertSlice(m.TxIds, func(t types2.TxID) []uint8 {
			return ([]uint8)(t)
		}),
		Origin: (m.Origin).Pb(),
	}
}

func (*TransactionIDsResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.TransactionIDsResponse]()}
}

type RequestBatchID struct {
	TxIds  []types2.TxID
	Origin *RequestBatchIDOrigin
}

func RequestBatchIDFromPb(pb *mempoolpb.RequestBatchID) *RequestBatchID {
	if pb == nil {
		return nil
	}
	return &RequestBatchID{
		TxIds: types1.ConvertSlice(pb.TxIds, func(t []uint8) types2.TxID {
			return (types2.TxID)(t)
		}),
		Origin: RequestBatchIDOriginFromPb(pb.Origin),
	}
}

func (m *RequestBatchID) Pb() *mempoolpb.RequestBatchID {
	if m == nil {
		return nil
	}
	return &mempoolpb.RequestBatchID{
		TxIds: types1.ConvertSlice(m.TxIds, func(t types2.TxID) []uint8 {
			return ([]uint8)(t)
		}),
		Origin: (m.Origin).Pb(),
	}
}

func (*RequestBatchID) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.RequestBatchID]()}
}

type BatchIDResponse struct {
	BatchId types3.BatchID
	Origin  *RequestBatchIDOrigin
}

func BatchIDResponseFromPb(pb *mempoolpb.BatchIDResponse) *BatchIDResponse {
	if pb == nil {
		return nil
	}
	return &BatchIDResponse{
		BatchId: (types3.BatchID)(pb.BatchId),
		Origin:  RequestBatchIDOriginFromPb(pb.Origin),
	}
}

func (m *BatchIDResponse) Pb() *mempoolpb.BatchIDResponse {
	if m == nil {
		return nil
	}
	return &mempoolpb.BatchIDResponse{
		BatchId: ([]uint8)(m.BatchId),
		Origin:  (m.Origin).Pb(),
	}
}

func (*BatchIDResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.BatchIDResponse]()}
}

type RequestBatchOrigin struct {
	Module types4.ModuleID
	Type   RequestBatchOrigin_Type
}

type RequestBatchOrigin_Type interface {
	mirreflect.GeneratedType
	isRequestBatchOrigin_Type()
	Pb() mempoolpb.RequestBatchOrigin_Type
}

type RequestBatchOrigin_TypeWrapper[T any] interface {
	RequestBatchOrigin_Type
	Unwrap() *T
}

func RequestBatchOrigin_TypeFromPb(pb mempoolpb.RequestBatchOrigin_Type) RequestBatchOrigin_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *mempoolpb.RequestBatchOrigin_ContextStore:
		return &RequestBatchOrigin_ContextStore{ContextStore: types5.OriginFromPb(pb.ContextStore)}
	case *mempoolpb.RequestBatchOrigin_Dsl:
		return &RequestBatchOrigin_Dsl{Dsl: types6.OriginFromPb(pb.Dsl)}
	}
	return nil
}

type RequestBatchOrigin_ContextStore struct {
	ContextStore *types5.Origin
}

func (*RequestBatchOrigin_ContextStore) isRequestBatchOrigin_Type() {}

func (w *RequestBatchOrigin_ContextStore) Unwrap() *types5.Origin {
	return w.ContextStore
}

func (w *RequestBatchOrigin_ContextStore) Pb() mempoolpb.RequestBatchOrigin_Type {
	if w == nil {
		return nil
	}
	return &mempoolpb.RequestBatchOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*RequestBatchOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.RequestBatchOrigin_ContextStore]()}
}

type RequestBatchOrigin_Dsl struct {
	Dsl *types6.Origin
}

func (*RequestBatchOrigin_Dsl) isRequestBatchOrigin_Type() {}

func (w *RequestBatchOrigin_Dsl) Unwrap() *types6.Origin {
	return w.Dsl
}

func (w *RequestBatchOrigin_Dsl) Pb() mempoolpb.RequestBatchOrigin_Type {
	if w == nil {
		return nil
	}
	return &mempoolpb.RequestBatchOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*RequestBatchOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.RequestBatchOrigin_Dsl]()}
}

func RequestBatchOriginFromPb(pb *mempoolpb.RequestBatchOrigin) *RequestBatchOrigin {
	if pb == nil {
		return nil
	}
	return &RequestBatchOrigin{
		Module: (types4.ModuleID)(pb.Module),
		Type:   RequestBatchOrigin_TypeFromPb(pb.Type),
	}
}

func (m *RequestBatchOrigin) Pb() *mempoolpb.RequestBatchOrigin {
	if m == nil {
		return nil
	}
	return &mempoolpb.RequestBatchOrigin{
		Module: (string)(m.Module),
		Type:   (m.Type).Pb(),
	}
}

func (*RequestBatchOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.RequestBatchOrigin]()}
}

type RequestTransactionsOrigin struct {
	Module types4.ModuleID
	Type   RequestTransactionsOrigin_Type
}

type RequestTransactionsOrigin_Type interface {
	mirreflect.GeneratedType
	isRequestTransactionsOrigin_Type()
	Pb() mempoolpb.RequestTransactionsOrigin_Type
}

type RequestTransactionsOrigin_TypeWrapper[T any] interface {
	RequestTransactionsOrigin_Type
	Unwrap() *T
}

func RequestTransactionsOrigin_TypeFromPb(pb mempoolpb.RequestTransactionsOrigin_Type) RequestTransactionsOrigin_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *mempoolpb.RequestTransactionsOrigin_ContextStore:
		return &RequestTransactionsOrigin_ContextStore{ContextStore: types5.OriginFromPb(pb.ContextStore)}
	case *mempoolpb.RequestTransactionsOrigin_Dsl:
		return &RequestTransactionsOrigin_Dsl{Dsl: types6.OriginFromPb(pb.Dsl)}
	}
	return nil
}

type RequestTransactionsOrigin_ContextStore struct {
	ContextStore *types5.Origin
}

func (*RequestTransactionsOrigin_ContextStore) isRequestTransactionsOrigin_Type() {}

func (w *RequestTransactionsOrigin_ContextStore) Unwrap() *types5.Origin {
	return w.ContextStore
}

func (w *RequestTransactionsOrigin_ContextStore) Pb() mempoolpb.RequestTransactionsOrigin_Type {
	if w == nil {
		return nil
	}
	return &mempoolpb.RequestTransactionsOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*RequestTransactionsOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.RequestTransactionsOrigin_ContextStore]()}
}

type RequestTransactionsOrigin_Dsl struct {
	Dsl *types6.Origin
}

func (*RequestTransactionsOrigin_Dsl) isRequestTransactionsOrigin_Type() {}

func (w *RequestTransactionsOrigin_Dsl) Unwrap() *types6.Origin {
	return w.Dsl
}

func (w *RequestTransactionsOrigin_Dsl) Pb() mempoolpb.RequestTransactionsOrigin_Type {
	if w == nil {
		return nil
	}
	return &mempoolpb.RequestTransactionsOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*RequestTransactionsOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.RequestTransactionsOrigin_Dsl]()}
}

func RequestTransactionsOriginFromPb(pb *mempoolpb.RequestTransactionsOrigin) *RequestTransactionsOrigin {
	if pb == nil {
		return nil
	}
	return &RequestTransactionsOrigin{
		Module: (types4.ModuleID)(pb.Module),
		Type:   RequestTransactionsOrigin_TypeFromPb(pb.Type),
	}
}

func (m *RequestTransactionsOrigin) Pb() *mempoolpb.RequestTransactionsOrigin {
	if m == nil {
		return nil
	}
	return &mempoolpb.RequestTransactionsOrigin{
		Module: (string)(m.Module),
		Type:   (m.Type).Pb(),
	}
}

func (*RequestTransactionsOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.RequestTransactionsOrigin]()}
}

type RequestTransactionIDsOrigin struct {
	Module types4.ModuleID
	Type   RequestTransactionIDsOrigin_Type
}

type RequestTransactionIDsOrigin_Type interface {
	mirreflect.GeneratedType
	isRequestTransactionIDsOrigin_Type()
	Pb() mempoolpb.RequestTransactionIDsOrigin_Type
}

type RequestTransactionIDsOrigin_TypeWrapper[T any] interface {
	RequestTransactionIDsOrigin_Type
	Unwrap() *T
}

func RequestTransactionIDsOrigin_TypeFromPb(pb mempoolpb.RequestTransactionIDsOrigin_Type) RequestTransactionIDsOrigin_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *mempoolpb.RequestTransactionIDsOrigin_ContextStore:
		return &RequestTransactionIDsOrigin_ContextStore{ContextStore: types5.OriginFromPb(pb.ContextStore)}
	case *mempoolpb.RequestTransactionIDsOrigin_Dsl:
		return &RequestTransactionIDsOrigin_Dsl{Dsl: types6.OriginFromPb(pb.Dsl)}
	}
	return nil
}

type RequestTransactionIDsOrigin_ContextStore struct {
	ContextStore *types5.Origin
}

func (*RequestTransactionIDsOrigin_ContextStore) isRequestTransactionIDsOrigin_Type() {}

func (w *RequestTransactionIDsOrigin_ContextStore) Unwrap() *types5.Origin {
	return w.ContextStore
}

func (w *RequestTransactionIDsOrigin_ContextStore) Pb() mempoolpb.RequestTransactionIDsOrigin_Type {
	if w == nil {
		return nil
	}
	return &mempoolpb.RequestTransactionIDsOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*RequestTransactionIDsOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.RequestTransactionIDsOrigin_ContextStore]()}
}

type RequestTransactionIDsOrigin_Dsl struct {
	Dsl *types6.Origin
}

func (*RequestTransactionIDsOrigin_Dsl) isRequestTransactionIDsOrigin_Type() {}

func (w *RequestTransactionIDsOrigin_Dsl) Unwrap() *types6.Origin {
	return w.Dsl
}

func (w *RequestTransactionIDsOrigin_Dsl) Pb() mempoolpb.RequestTransactionIDsOrigin_Type {
	if w == nil {
		return nil
	}
	return &mempoolpb.RequestTransactionIDsOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*RequestTransactionIDsOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.RequestTransactionIDsOrigin_Dsl]()}
}

func RequestTransactionIDsOriginFromPb(pb *mempoolpb.RequestTransactionIDsOrigin) *RequestTransactionIDsOrigin {
	if pb == nil {
		return nil
	}
	return &RequestTransactionIDsOrigin{
		Module: (types4.ModuleID)(pb.Module),
		Type:   RequestTransactionIDsOrigin_TypeFromPb(pb.Type),
	}
}

func (m *RequestTransactionIDsOrigin) Pb() *mempoolpb.RequestTransactionIDsOrigin {
	if m == nil {
		return nil
	}
	return &mempoolpb.RequestTransactionIDsOrigin{
		Module: (string)(m.Module),
		Type:   (m.Type).Pb(),
	}
}

func (*RequestTransactionIDsOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.RequestTransactionIDsOrigin]()}
}

type RequestBatchIDOrigin struct {
	Module types4.ModuleID
	Type   RequestBatchIDOrigin_Type
}

type RequestBatchIDOrigin_Type interface {
	mirreflect.GeneratedType
	isRequestBatchIDOrigin_Type()
	Pb() mempoolpb.RequestBatchIDOrigin_Type
}

type RequestBatchIDOrigin_TypeWrapper[T any] interface {
	RequestBatchIDOrigin_Type
	Unwrap() *T
}

func RequestBatchIDOrigin_TypeFromPb(pb mempoolpb.RequestBatchIDOrigin_Type) RequestBatchIDOrigin_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *mempoolpb.RequestBatchIDOrigin_ContextStore:
		return &RequestBatchIDOrigin_ContextStore{ContextStore: types5.OriginFromPb(pb.ContextStore)}
	case *mempoolpb.RequestBatchIDOrigin_Dsl:
		return &RequestBatchIDOrigin_Dsl{Dsl: types6.OriginFromPb(pb.Dsl)}
	}
	return nil
}

type RequestBatchIDOrigin_ContextStore struct {
	ContextStore *types5.Origin
}

func (*RequestBatchIDOrigin_ContextStore) isRequestBatchIDOrigin_Type() {}

func (w *RequestBatchIDOrigin_ContextStore) Unwrap() *types5.Origin {
	return w.ContextStore
}

func (w *RequestBatchIDOrigin_ContextStore) Pb() mempoolpb.RequestBatchIDOrigin_Type {
	if w == nil {
		return nil
	}
	return &mempoolpb.RequestBatchIDOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*RequestBatchIDOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.RequestBatchIDOrigin_ContextStore]()}
}

type RequestBatchIDOrigin_Dsl struct {
	Dsl *types6.Origin
}

func (*RequestBatchIDOrigin_Dsl) isRequestBatchIDOrigin_Type() {}

func (w *RequestBatchIDOrigin_Dsl) Unwrap() *types6.Origin {
	return w.Dsl
}

func (w *RequestBatchIDOrigin_Dsl) Pb() mempoolpb.RequestBatchIDOrigin_Type {
	if w == nil {
		return nil
	}
	return &mempoolpb.RequestBatchIDOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*RequestBatchIDOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.RequestBatchIDOrigin_Dsl]()}
}

func RequestBatchIDOriginFromPb(pb *mempoolpb.RequestBatchIDOrigin) *RequestBatchIDOrigin {
	if pb == nil {
		return nil
	}
	return &RequestBatchIDOrigin{
		Module: (types4.ModuleID)(pb.Module),
		Type:   RequestBatchIDOrigin_TypeFromPb(pb.Type),
	}
}

func (m *RequestBatchIDOrigin) Pb() *mempoolpb.RequestBatchIDOrigin {
	if m == nil {
		return nil
	}
	return &mempoolpb.RequestBatchIDOrigin{
		Module: (string)(m.Module),
		Type:   (m.Type).Pb(),
	}
}

func (*RequestBatchIDOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*mempoolpb.RequestBatchIDOrigin]()}
}
