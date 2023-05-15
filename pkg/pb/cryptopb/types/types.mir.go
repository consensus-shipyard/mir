package cryptopbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types1 "github.com/filecoin-project/mir/codegen/model/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/contextstorepb/types"
	cryptopb "github.com/filecoin-project/mir/pkg/pb/cryptopb"
	types3 "github.com/filecoin-project/mir/pkg/pb/dslpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Event struct {
	Type Event_Type
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() cryptopb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb cryptopb.Event_Type) Event_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *cryptopb.Event_SignRequest:
		return &Event_SignRequest{SignRequest: SignRequestFromPb(pb.SignRequest)}
	case *cryptopb.Event_SignResult:
		return &Event_SignResult{SignResult: SignResultFromPb(pb.SignResult)}
	case *cryptopb.Event_VerifySig:
		return &Event_VerifySig{VerifySig: VerifySigFromPb(pb.VerifySig)}
	case *cryptopb.Event_SigVerified:
		return &Event_SigVerified{SigVerified: SigVerifiedFromPb(pb.SigVerified)}
	case *cryptopb.Event_VerifySigs:
		return &Event_VerifySigs{VerifySigs: VerifySigsFromPb(pb.VerifySigs)}
	case *cryptopb.Event_SigsVerified:
		return &Event_SigsVerified{SigsVerified: SigsVerifiedFromPb(pb.SigsVerified)}
	}
	return nil
}

type Event_SignRequest struct {
	SignRequest *SignRequest
}

func (*Event_SignRequest) isEvent_Type() {}

func (w *Event_SignRequest) Unwrap() *SignRequest {
	return w.SignRequest
}

func (w *Event_SignRequest) Pb() cryptopb.Event_Type {
	if w == nil {
		return nil
	}
	if w.SignRequest == nil {
		return &cryptopb.Event_SignRequest{}
	}
	return &cryptopb.Event_SignRequest{SignRequest: (w.SignRequest).Pb()}
}

func (*Event_SignRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.Event_SignRequest]()}
}

type Event_SignResult struct {
	SignResult *SignResult
}

func (*Event_SignResult) isEvent_Type() {}

func (w *Event_SignResult) Unwrap() *SignResult {
	return w.SignResult
}

func (w *Event_SignResult) Pb() cryptopb.Event_Type {
	if w == nil {
		return nil
	}
	if w.SignResult == nil {
		return &cryptopb.Event_SignResult{}
	}
	return &cryptopb.Event_SignResult{SignResult: (w.SignResult).Pb()}
}

func (*Event_SignResult) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.Event_SignResult]()}
}

type Event_VerifySig struct {
	VerifySig *VerifySig
}

func (*Event_VerifySig) isEvent_Type() {}

func (w *Event_VerifySig) Unwrap() *VerifySig {
	return w.VerifySig
}

func (w *Event_VerifySig) Pb() cryptopb.Event_Type {
	if w == nil {
		return nil
	}
	if w.VerifySig == nil {
		return &cryptopb.Event_VerifySig{}
	}
	return &cryptopb.Event_VerifySig{VerifySig: (w.VerifySig).Pb()}
}

func (*Event_VerifySig) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.Event_VerifySig]()}
}

type Event_SigVerified struct {
	SigVerified *SigVerified
}

func (*Event_SigVerified) isEvent_Type() {}

func (w *Event_SigVerified) Unwrap() *SigVerified {
	return w.SigVerified
}

func (w *Event_SigVerified) Pb() cryptopb.Event_Type {
	if w == nil {
		return nil
	}
	if w.SigVerified == nil {
		return &cryptopb.Event_SigVerified{}
	}
	return &cryptopb.Event_SigVerified{SigVerified: (w.SigVerified).Pb()}
}

func (*Event_SigVerified) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.Event_SigVerified]()}
}

type Event_VerifySigs struct {
	VerifySigs *VerifySigs
}

func (*Event_VerifySigs) isEvent_Type() {}

func (w *Event_VerifySigs) Unwrap() *VerifySigs {
	return w.VerifySigs
}

func (w *Event_VerifySigs) Pb() cryptopb.Event_Type {
	if w == nil {
		return nil
	}
	if w.VerifySigs == nil {
		return &cryptopb.Event_VerifySigs{}
	}
	return &cryptopb.Event_VerifySigs{VerifySigs: (w.VerifySigs).Pb()}
}

func (*Event_VerifySigs) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.Event_VerifySigs]()}
}

type Event_SigsVerified struct {
	SigsVerified *SigsVerified
}

func (*Event_SigsVerified) isEvent_Type() {}

func (w *Event_SigsVerified) Unwrap() *SigsVerified {
	return w.SigsVerified
}

func (w *Event_SigsVerified) Pb() cryptopb.Event_Type {
	if w == nil {
		return nil
	}
	if w.SigsVerified == nil {
		return &cryptopb.Event_SigsVerified{}
	}
	return &cryptopb.Event_SigsVerified{SigsVerified: (w.SigsVerified).Pb()}
}

func (*Event_SigsVerified) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.Event_SigsVerified]()}
}

func EventFromPb(pb *cryptopb.Event) *Event {
	if pb == nil {
		return nil
	}
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *cryptopb.Event {
	if m == nil {
		return nil
	}
	pbMessage := &cryptopb.Event{}
	{
		if m.Type != nil {
			pbMessage.Type = (m.Type).Pb()
		}
	}

	return pbMessage
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.Event]()}
}

type SignRequest struct {
	Data   *SignedData
	Origin *SignOrigin
}

func SignRequestFromPb(pb *cryptopb.SignRequest) *SignRequest {
	if pb == nil {
		return nil
	}
	return &SignRequest{
		Data:   SignedDataFromPb(pb.Data),
		Origin: SignOriginFromPb(pb.Origin),
	}
}

func (m *SignRequest) Pb() *cryptopb.SignRequest {
	if m == nil {
		return nil
	}
	pbMessage := &cryptopb.SignRequest{}
	{
		if m.Data != nil {
			pbMessage.Data = (m.Data).Pb()
		}
		if m.Origin != nil {
			pbMessage.Origin = (m.Origin).Pb()
		}
	}

	return pbMessage
}

func (*SignRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.SignRequest]()}
}

type SignResult struct {
	Signature []uint8
	Origin    *SignOrigin
}

func SignResultFromPb(pb *cryptopb.SignResult) *SignResult {
	if pb == nil {
		return nil
	}
	return &SignResult{
		Signature: pb.Signature,
		Origin:    SignOriginFromPb(pb.Origin),
	}
}

func (m *SignResult) Pb() *cryptopb.SignResult {
	if m == nil {
		return nil
	}
	pbMessage := &cryptopb.SignResult{}
	{
		pbMessage.Signature = m.Signature
		if m.Origin != nil {
			pbMessage.Origin = (m.Origin).Pb()
		}
	}

	return pbMessage
}

func (*SignResult) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.SignResult]()}
}

type VerifySig struct {
	Data      *SignedData
	Signature []uint8
	Origin    *SigVerOrigin
	NodeId    types.NodeID
}

func VerifySigFromPb(pb *cryptopb.VerifySig) *VerifySig {
	if pb == nil {
		return nil
	}
	return &VerifySig{
		Data:      SignedDataFromPb(pb.Data),
		Signature: pb.Signature,
		Origin:    SigVerOriginFromPb(pb.Origin),
		NodeId:    (types.NodeID)(pb.NodeId),
	}
}

func (m *VerifySig) Pb() *cryptopb.VerifySig {
	if m == nil {
		return nil
	}
	pbMessage := &cryptopb.VerifySig{}
	{
		if m.Data != nil {
			pbMessage.Data = (m.Data).Pb()
		}
		pbMessage.Signature = m.Signature
		if m.Origin != nil {
			pbMessage.Origin = (m.Origin).Pb()
		}
		pbMessage.NodeId = (string)(m.NodeId)
	}

	return pbMessage
}

func (*VerifySig) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.VerifySig]()}
}

type SigVerified struct {
	Origin *SigVerOrigin
	NodeId types.NodeID
	Error  error
}

func SigVerifiedFromPb(pb *cryptopb.SigVerified) *SigVerified {
	if pb == nil {
		return nil
	}
	return &SigVerified{
		Origin: SigVerOriginFromPb(pb.Origin),
		NodeId: (types.NodeID)(pb.NodeId),
		Error:  types1.StringToError(pb.Error),
	}
}

func (m *SigVerified) Pb() *cryptopb.SigVerified {
	if m == nil {
		return nil
	}
	pbMessage := &cryptopb.SigVerified{}
	{
		if m.Origin != nil {
			pbMessage.Origin = (m.Origin).Pb()
		}
		pbMessage.NodeId = (string)(m.NodeId)
		pbMessage.Error = types1.ErrorToString(m.Error)
	}

	return pbMessage
}

func (*SigVerified) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.SigVerified]()}
}

type VerifySigs struct {
	Data       []*SignedData
	Signatures [][]uint8
	Origin     *SigVerOrigin
	NodeIds    []types.NodeID
}

func VerifySigsFromPb(pb *cryptopb.VerifySigs) *VerifySigs {
	if pb == nil {
		return nil
	}
	return &VerifySigs{
		Data: types1.ConvertSlice(pb.Data, func(t *cryptopb.SignedData) *SignedData {
			return SignedDataFromPb(t)
		}),
		Signatures: pb.Signatures,
		Origin:     SigVerOriginFromPb(pb.Origin),
		NodeIds: types1.ConvertSlice(pb.NodeIds, func(t string) types.NodeID {
			return (types.NodeID)(t)
		}),
	}
}

func (m *VerifySigs) Pb() *cryptopb.VerifySigs {
	if m == nil {
		return nil
	}
	pbMessage := &cryptopb.VerifySigs{}
	{
		pbMessage.Data = types1.ConvertSlice(m.Data, func(t *SignedData) *cryptopb.SignedData {
			return (t).Pb()
		})
		pbMessage.Signatures = m.Signatures
		if m.Origin != nil {
			pbMessage.Origin = (m.Origin).Pb()
		}
		pbMessage.NodeIds = types1.ConvertSlice(m.NodeIds, func(t types.NodeID) string {
			return (string)(t)
		})
	}

	return pbMessage
}

func (*VerifySigs) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.VerifySigs]()}
}

type SigsVerified struct {
	Origin  *SigVerOrigin
	NodeIds []types.NodeID
	Errors  []error
	AllOk   bool
}

func SigsVerifiedFromPb(pb *cryptopb.SigsVerified) *SigsVerified {
	if pb == nil {
		return nil
	}
	return &SigsVerified{
		Origin: SigVerOriginFromPb(pb.Origin),
		NodeIds: types1.ConvertSlice(pb.NodeIds, func(t string) types.NodeID {
			return (types.NodeID)(t)
		}),
		Errors: types1.ConvertSlice(pb.Errors, func(t string) error {
			return types1.StringToError(t)
		}),
		AllOk: pb.AllOk,
	}
}

func (m *SigsVerified) Pb() *cryptopb.SigsVerified {
	if m == nil {
		return nil
	}
	pbMessage := &cryptopb.SigsVerified{}
	{
		if m.Origin != nil {
			pbMessage.Origin = (m.Origin).Pb()
		}
		pbMessage.NodeIds = types1.ConvertSlice(m.NodeIds, func(t types.NodeID) string {
			return (string)(t)
		})
		pbMessage.Errors = types1.ConvertSlice(m.Errors, func(t error) string {
			return types1.ErrorToString(t)
		})
		pbMessage.AllOk = m.AllOk
	}

	return pbMessage
}

func (*SigsVerified) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.SigsVerified]()}
}

type SignOrigin struct {
	Module types.ModuleID
	Type   SignOrigin_Type
}

type SignOrigin_Type interface {
	mirreflect.GeneratedType
	isSignOrigin_Type()
	Pb() cryptopb.SignOrigin_Type
}

type SignOrigin_TypeWrapper[T any] interface {
	SignOrigin_Type
	Unwrap() *T
}

func SignOrigin_TypeFromPb(pb cryptopb.SignOrigin_Type) SignOrigin_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *cryptopb.SignOrigin_ContextStore:
		return &SignOrigin_ContextStore{ContextStore: types2.OriginFromPb(pb.ContextStore)}
	case *cryptopb.SignOrigin_Dsl:
		return &SignOrigin_Dsl{Dsl: types3.OriginFromPb(pb.Dsl)}
	}
	return nil
}

type SignOrigin_ContextStore struct {
	ContextStore *types2.Origin
}

func (*SignOrigin_ContextStore) isSignOrigin_Type() {}

func (w *SignOrigin_ContextStore) Unwrap() *types2.Origin {
	return w.ContextStore
}

func (w *SignOrigin_ContextStore) Pb() cryptopb.SignOrigin_Type {
	if w == nil {
		return nil
	}
	if w.ContextStore == nil {
		return &cryptopb.SignOrigin_ContextStore{}
	}
	return &cryptopb.SignOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*SignOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.SignOrigin_ContextStore]()}
}

type SignOrigin_Dsl struct {
	Dsl *types3.Origin
}

func (*SignOrigin_Dsl) isSignOrigin_Type() {}

func (w *SignOrigin_Dsl) Unwrap() *types3.Origin {
	return w.Dsl
}

func (w *SignOrigin_Dsl) Pb() cryptopb.SignOrigin_Type {
	if w == nil {
		return nil
	}
	if w.Dsl == nil {
		return &cryptopb.SignOrigin_Dsl{}
	}
	return &cryptopb.SignOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*SignOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.SignOrigin_Dsl]()}
}

func SignOriginFromPb(pb *cryptopb.SignOrigin) *SignOrigin {
	if pb == nil {
		return nil
	}
	return &SignOrigin{
		Module: (types.ModuleID)(pb.Module),
		Type:   SignOrigin_TypeFromPb(pb.Type),
	}
}

func (m *SignOrigin) Pb() *cryptopb.SignOrigin {
	if m == nil {
		return nil
	}
	pbMessage := &cryptopb.SignOrigin{}
	{
		pbMessage.Module = (string)(m.Module)
		if m.Type != nil {
			pbMessage.Type = (m.Type).Pb()
		}
	}

	return pbMessage
}

func (*SignOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.SignOrigin]()}
}

type SigVerOrigin struct {
	Module types.ModuleID
	Type   SigVerOrigin_Type
}

type SigVerOrigin_Type interface {
	mirreflect.GeneratedType
	isSigVerOrigin_Type()
	Pb() cryptopb.SigVerOrigin_Type
}

type SigVerOrigin_TypeWrapper[T any] interface {
	SigVerOrigin_Type
	Unwrap() *T
}

func SigVerOrigin_TypeFromPb(pb cryptopb.SigVerOrigin_Type) SigVerOrigin_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *cryptopb.SigVerOrigin_ContextStore:
		return &SigVerOrigin_ContextStore{ContextStore: types2.OriginFromPb(pb.ContextStore)}
	case *cryptopb.SigVerOrigin_Dsl:
		return &SigVerOrigin_Dsl{Dsl: types3.OriginFromPb(pb.Dsl)}
	}
	return nil
}

type SigVerOrigin_ContextStore struct {
	ContextStore *types2.Origin
}

func (*SigVerOrigin_ContextStore) isSigVerOrigin_Type() {}

func (w *SigVerOrigin_ContextStore) Unwrap() *types2.Origin {
	return w.ContextStore
}

func (w *SigVerOrigin_ContextStore) Pb() cryptopb.SigVerOrigin_Type {
	if w == nil {
		return nil
	}
	if w.ContextStore == nil {
		return &cryptopb.SigVerOrigin_ContextStore{}
	}
	return &cryptopb.SigVerOrigin_ContextStore{ContextStore: (w.ContextStore).Pb()}
}

func (*SigVerOrigin_ContextStore) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.SigVerOrigin_ContextStore]()}
}

type SigVerOrigin_Dsl struct {
	Dsl *types3.Origin
}

func (*SigVerOrigin_Dsl) isSigVerOrigin_Type() {}

func (w *SigVerOrigin_Dsl) Unwrap() *types3.Origin {
	return w.Dsl
}

func (w *SigVerOrigin_Dsl) Pb() cryptopb.SigVerOrigin_Type {
	if w == nil {
		return nil
	}
	if w.Dsl == nil {
		return &cryptopb.SigVerOrigin_Dsl{}
	}
	return &cryptopb.SigVerOrigin_Dsl{Dsl: (w.Dsl).Pb()}
}

func (*SigVerOrigin_Dsl) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.SigVerOrigin_Dsl]()}
}

func SigVerOriginFromPb(pb *cryptopb.SigVerOrigin) *SigVerOrigin {
	if pb == nil {
		return nil
	}
	return &SigVerOrigin{
		Module: (types.ModuleID)(pb.Module),
		Type:   SigVerOrigin_TypeFromPb(pb.Type),
	}
}

func (m *SigVerOrigin) Pb() *cryptopb.SigVerOrigin {
	if m == nil {
		return nil
	}
	pbMessage := &cryptopb.SigVerOrigin{}
	{
		pbMessage.Module = (string)(m.Module)
		if m.Type != nil {
			pbMessage.Type = (m.Type).Pb()
		}
	}

	return pbMessage
}

func (*SigVerOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.SigVerOrigin]()}
}

type SignedData struct {
	Data [][]uint8
}

func SignedDataFromPb(pb *cryptopb.SignedData) *SignedData {
	if pb == nil {
		return nil
	}
	return &SignedData{
		Data: pb.Data,
	}
}

func (m *SignedData) Pb() *cryptopb.SignedData {
	if m == nil {
		return nil
	}
	pbMessage := &cryptopb.SignedData{}
	{
		pbMessage.Data = m.Data
	}

	return pbMessage
}

func (*SignedData) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*cryptopb.SignedData]()}
}
