package batchdbpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types2 "github.com/filecoin-project/mir/codegen/model/types"
	types "github.com/filecoin-project/mir/pkg/availability/multisigcollector/types"
	batchdbpb "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb"
	types5 "github.com/filecoin-project/mir/pkg/pb/contextstorepb/types"
	types6 "github.com/filecoin-project/mir/pkg/pb/dslpb/types"
	trantorpb "github.com/filecoin-project/mir/pkg/pb/trantorpb"
	types1 "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	types3 "github.com/filecoin-project/mir/pkg/trantor/types"
	types4 "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Event struct {
	Type Event_Type
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() batchdbpb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb batchdbpb.Event_Type) Event_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *batchdbpb.Event_Lookup:
		return &Event_Lookup{Lookup: LookupBatchFromPb(pb.Lookup)}
	case *batchdbpb.Event_LookupResponse:
		return &Event_LookupResponse{LookupResponse: LookupBatchResponseFromPb(pb.LookupResponse)}
	case *batchdbpb.Event_Store:
		return &Event_Store{Store: StoreBatchFromPb(pb.Store)}
	case *batchdbpb.Event_Stored:
		return &Event_Stored{Stored: BatchStoredFromPb(pb.Stored)}
	}
	return nil
}

type Event_Lookup struct {
	Lookup *LookupBatch
}

func (*Event_Lookup) isEvent_Type() {}

func (w *Event_Lookup) Unwrap() *LookupBatch {
	return w.Lookup
}

func (w *Event_Lookup) Pb() batchdbpb.Event_Type {
	if w == nil {
		return nil
	}
	return &batchdbpb.Event_Lookup{Lookup: (w.Lookup).Pb()}
}

func (*Event_Lookup) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.Event_Lookup]()}
}

type Event_LookupResponse struct {
	LookupResponse *LookupBatchResponse
}

func (*Event_LookupResponse) isEvent_Type() {}

func (w *Event_LookupResponse) Unwrap() *LookupBatchResponse {
	return w.LookupResponse
}

func (w *Event_LookupResponse) Pb() batchdbpb.Event_Type {
	if w == nil {
		return nil
	}
	return &batchdbpb.Event_LookupResponse{LookupResponse: (w.LookupResponse).Pb()}
}

func (*Event_LookupResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.Event_LookupResponse]()}
}

type Event_Store struct {
	Store *StoreBatch
}

func (*Event_Store) isEvent_Type() {}

func (w *Event_Store) Unwrap() *StoreBatch {
	return w.Store
}

func (w *Event_Store) Pb() batchdbpb.Event_Type {
	if w == nil {
		return nil
	}
	return &batchdbpb.Event_Store{Store: (w.Store).Pb()}
}

func (*Event_Store) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.Event_Store]()}
}

type Event_Stored struct {
	Stored *BatchStored
}

func (*Event_Stored) isEvent_Type() {}

func (w *Event_Stored) Unwrap() *BatchStored {
	return w.Stored
}

func (w *Event_Stored) Pb() batchdbpb.Event_Type {
	if w == nil {
		return nil
	}
	return &batchdbpb.Event_Stored{Stored: (w.Stored).Pb()}
}

func (*Event_Stored) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.Event_Stored]()}
}

func EventFromPb(pb *batchdbpb.Event) *Event {
	if pb == nil {
		return nil
	}
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *batchdbpb.Event {
	if m == nil {
		return nil
	}
	return &batchdbpb.Event{
		Type: (m.Type).Pb(),
	}
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.Event]()}
}

type LookupBatch struct {
	BatchId types.BatchID
	Origin  *LookupBatchOrigin
}

func LookupBatchFromPb(pb *batchdbpb.LookupBatch) *LookupBatch {
	if pb == nil {
		return nil
	}
	return &LookupBatch{
		BatchId: (types.BatchID)(pb.BatchId),
		Origin:  LookupBatchOriginFromPb(pb.Origin),
	}
}

func (m *LookupBatch) Pb() *batchdbpb.LookupBatch {
	if m == nil {
		return nil
	}
	return &batchdbpb.LookupBatch{
		BatchId: ([]uint8)(m.BatchId),
		Origin:  (m.Origin).Pb(),
	}
}

func (*LookupBatch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.LookupBatch]()}
}

type LookupBatchResponse struct {
	Found  bool
	Txs    []*types1.Transaction
	Origin *LookupBatchOrigin
}

func LookupBatchResponseFromPb(pb *batchdbpb.LookupBatchResponse) *LookupBatchResponse {
	if pb == nil {
		return nil
	}
	return &LookupBatchResponse{
		Found: pb.Found,
		Txs: types2.ConvertSlice(pb.Txs, func(t *trantorpb.Transaction) *types1.Transaction {
			return types1.TransactionFromPb(t)
		}),
		Origin: LookupBatchOriginFromPb(pb.Origin),
	}
}

func (m *LookupBatchResponse) Pb() *batchdbpb.LookupBatchResponse {
	if m == nil {
		return nil
	}
	return &batchdbpb.LookupBatchResponse{
		Found: m.Found,
		Txs: types2.ConvertSlice(m.Txs, func(t *types1.Transaction) *trantorpb.Transaction {
			return (t).Pb()
		}),
		Origin: (m.Origin).Pb(),
	}
}

func (*LookupBatchResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.LookupBatchResponse]()}
}

type StoreBatch struct {
	BatchId  types.BatchID
	TxIds    []types3.TxID
	Txs      []*types1.Transaction
	Metadata []uint8
	Origin   *StoreBatchOrigin
}

func StoreBatchFromPb(pb *batchdbpb.StoreBatch) *StoreBatch {
	if pb == nil {
		return nil
	}
	return &StoreBatch{
		BatchId: (types.BatchID)(pb.BatchId),
		TxIds: types2.ConvertSlice(pb.TxIds, func(t []uint8) types3.TxID {
			return (types3.TxID)(t)
		}),
		Txs: types2.ConvertSlice(pb.Txs, func(t *trantorpb.Transaction) *types1.Transaction {
			return types1.TransactionFromPb(t)
		}),
		Metadata: pb.Metadata,
		Origin:   StoreBatchOriginFromPb(pb.Origin),
	}
}

func (m *StoreBatch) Pb() *batchdbpb.StoreBatch {
	if m == nil {
		return nil
	}
	return &batchdbpb.StoreBatch{
		BatchId: ([]uint8)(m.BatchId),
		TxIds: types2.ConvertSlice(m.TxIds, func(t types3.TxID) []uint8 {
			return ([]uint8)(t)
		}),
		Txs: types2.ConvertSlice(m.Txs, func(t *types1.Transaction) *trantorpb.Transaction {
			return (t).Pb()
		}),
		Metadata: m.Metadata,
		Origin:   (m.Origin).Pb(),
	}
}

func (*StoreBatch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.StoreBatch]()}
}

type BatchStored struct {
	Origin *StoreBatchOrigin
}

func BatchStoredFromPb(pb *batchdbpb.BatchStored) *BatchStored {
	if pb == nil {
		return nil
	}
	return &BatchStored{
		Origin: StoreBatchOriginFromPb(pb.Origin),
	}
}

func (m *BatchStored) Pb() *batchdbpb.BatchStored {
	if m == nil {
		return nil
	}
	return &batchdbpb.BatchStored{
		Origin: (m.Origin).Pb(),
	}
}

func (*BatchStored) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.BatchStored]()}
}

type LookupBatchOrigin struct {
	Module types4.ModuleID
	Type   LookupBatchOrigin_Type
}

type LookupBatchOrigin_Type interface {
	mirreflect.GeneratedType
	isLookupBatchOrigin_Type()
	Pb() batchdbpb.LookupBatchOrigin_Type
}

type LookupBatchOrigin_TypeWrapper[T any] interface {
	LookupBatchOrigin_Type
	Unwrap() *T
}

func LookupBatchOrigin_TypeFromPb(pb batchdbpb.LookupBatchOrigin_Type) LookupBatchOrigin_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *batchdbpb.LookupBatchOrigin_ContextStore:
		return &LookupBatchOrigin_ContextStore{ContextStore: types5.OriginFromPb(pb.ContextStore)}
	case *batchdbpb.LookupBatchOrigin_Dsl:
		return &LookupBatchOrigin_Dsl{Dsl: types6.OriginFromPb(pb.Dsl)}
	}
	return nil
}

type LookupBatchOrigin_ContextStore struct {
	ContextStore *types5.Origin
}

func (*LookupBatchOrigin_ContextStore) isLookupBatchOrigin_Type() {}

func (w *LookupBatchOrigin_ContextStore) Unwrap() *types5.Origin {
	return w.ContextStore
}

func (w *LookupBatchOrigin_ContextStore) Pb() batchdbpb.LookupBatchOrigin_Type {
	if w == nil {
		return nil
	}
	return &batchdbpb.LookupBatchOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*LookupBatchOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.LookupBatchOrigin_ContextStore]()}
}

type LookupBatchOrigin_Dsl struct {
	Dsl *types6.Origin
}

func (*LookupBatchOrigin_Dsl) isLookupBatchOrigin_Type() {}

func (w *LookupBatchOrigin_Dsl) Unwrap() *types6.Origin {
	return w.Dsl
}

func (w *LookupBatchOrigin_Dsl) Pb() batchdbpb.LookupBatchOrigin_Type {
	if w == nil {
		return nil
	}
	return &batchdbpb.LookupBatchOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*LookupBatchOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.LookupBatchOrigin_Dsl]()}
}

func LookupBatchOriginFromPb(pb *batchdbpb.LookupBatchOrigin) *LookupBatchOrigin {
	if pb == nil {
		return nil
	}
	return &LookupBatchOrigin{
		Module: (types4.ModuleID)(pb.Module),
		Type:   LookupBatchOrigin_TypeFromPb(pb.Type),
	}
}

func (m *LookupBatchOrigin) Pb() *batchdbpb.LookupBatchOrigin {
	if m == nil {
		return nil
	}
	return &batchdbpb.LookupBatchOrigin{
		Module: (string)(m.Module),
		Type:   (m.Type).Pb(),
	}
}

func (*LookupBatchOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.LookupBatchOrigin]()}
}

type StoreBatchOrigin struct {
	Module types4.ModuleID
	Type   StoreBatchOrigin_Type
}

type StoreBatchOrigin_Type interface {
	mirreflect.GeneratedType
	isStoreBatchOrigin_Type()
	Pb() batchdbpb.StoreBatchOrigin_Type
}

type StoreBatchOrigin_TypeWrapper[T any] interface {
	StoreBatchOrigin_Type
	Unwrap() *T
}

func StoreBatchOrigin_TypeFromPb(pb batchdbpb.StoreBatchOrigin_Type) StoreBatchOrigin_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *batchdbpb.StoreBatchOrigin_ContextStore:
		return &StoreBatchOrigin_ContextStore{ContextStore: types5.OriginFromPb(pb.ContextStore)}
	case *batchdbpb.StoreBatchOrigin_Dsl:
		return &StoreBatchOrigin_Dsl{Dsl: types6.OriginFromPb(pb.Dsl)}
	}
	return nil
}

type StoreBatchOrigin_ContextStore struct {
	ContextStore *types5.Origin
}

func (*StoreBatchOrigin_ContextStore) isStoreBatchOrigin_Type() {}

func (w *StoreBatchOrigin_ContextStore) Unwrap() *types5.Origin {
	return w.ContextStore
}

func (w *StoreBatchOrigin_ContextStore) Pb() batchdbpb.StoreBatchOrigin_Type {
	if w == nil {
		return nil
	}
	return &batchdbpb.StoreBatchOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*StoreBatchOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.StoreBatchOrigin_ContextStore]()}
}

type StoreBatchOrigin_Dsl struct {
	Dsl *types6.Origin
}

func (*StoreBatchOrigin_Dsl) isStoreBatchOrigin_Type() {}

func (w *StoreBatchOrigin_Dsl) Unwrap() *types6.Origin {
	return w.Dsl
}

func (w *StoreBatchOrigin_Dsl) Pb() batchdbpb.StoreBatchOrigin_Type {
	if w == nil {
		return nil
	}
	return &batchdbpb.StoreBatchOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*StoreBatchOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.StoreBatchOrigin_Dsl]()}
}

func StoreBatchOriginFromPb(pb *batchdbpb.StoreBatchOrigin) *StoreBatchOrigin {
	if pb == nil {
		return nil
	}
	return &StoreBatchOrigin{
		Module: (types4.ModuleID)(pb.Module),
		Type:   StoreBatchOrigin_TypeFromPb(pb.Type),
	}
}

func (m *StoreBatchOrigin) Pb() *batchdbpb.StoreBatchOrigin {
	if m == nil {
		return nil
	}
	return &batchdbpb.StoreBatchOrigin{
		Module: (string)(m.Module),
		Type:   (m.Type).Pb(),
	}
}

func (*StoreBatchOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*batchdbpb.StoreBatchOrigin]()}
}
