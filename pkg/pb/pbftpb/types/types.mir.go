package pbftpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types2 "github.com/filecoin-project/mir/codegen/model/types"
	types1 "github.com/filecoin-project/mir/pkg/orderers/types"
	pbftpb "github.com/filecoin-project/mir/pkg/pb/pbftpb"
	types "github.com/filecoin-project/mir/pkg/trantor/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Message struct {
	Type Message_Type
}

type Message_Type interface {
	mirreflect.GeneratedType
	isMessage_Type()
	Pb() pbftpb.Message_Type
}

type Message_TypeWrapper[T any] interface {
	Message_Type
	Unwrap() *T
}

func Message_TypeFromPb(pb pbftpb.Message_Type) Message_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *pbftpb.Message_Preprepare:
		return &Message_Preprepare{Preprepare: PreprepareFromPb(pb.Preprepare)}
	case *pbftpb.Message_Prepare:
		return &Message_Prepare{Prepare: PrepareFromPb(pb.Prepare)}
	case *pbftpb.Message_Commit:
		return &Message_Commit{Commit: CommitFromPb(pb.Commit)}
	case *pbftpb.Message_Done:
		return &Message_Done{Done: DoneFromPb(pb.Done)}
	case *pbftpb.Message_CatchUpRequest:
		return &Message_CatchUpRequest{CatchUpRequest: CatchUpRequestFromPb(pb.CatchUpRequest)}
	case *pbftpb.Message_CatchUpResponse:
		return &Message_CatchUpResponse{CatchUpResponse: CatchUpResponseFromPb(pb.CatchUpResponse)}
	case *pbftpb.Message_SignedViewChange:
		return &Message_SignedViewChange{SignedViewChange: SignedViewChangeFromPb(pb.SignedViewChange)}
	case *pbftpb.Message_PreprepareRequest:
		return &Message_PreprepareRequest{PreprepareRequest: PreprepareRequestFromPb(pb.PreprepareRequest)}
	case *pbftpb.Message_MissingPreprepare:
		return &Message_MissingPreprepare{MissingPreprepare: MissingPreprepareFromPb(pb.MissingPreprepare)}
	case *pbftpb.Message_NewView:
		return &Message_NewView{NewView: NewViewFromPb(pb.NewView)}
	}
	return nil
}

type Message_Preprepare struct {
	Preprepare *Preprepare
}

func (*Message_Preprepare) isMessage_Type() {}

func (w *Message_Preprepare) Unwrap() *Preprepare {
	return w.Preprepare
}

func (w *Message_Preprepare) Pb() pbftpb.Message_Type {
	if w == nil {
		return nil
	}
	if w.Preprepare == nil {
		return &pbftpb.Message_Preprepare{}
	}
	return &pbftpb.Message_Preprepare{Preprepare: (w.Preprepare).Pb()}
}

func (*Message_Preprepare) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Message_Preprepare]()}
}

type Message_Prepare struct {
	Prepare *Prepare
}

func (*Message_Prepare) isMessage_Type() {}

func (w *Message_Prepare) Unwrap() *Prepare {
	return w.Prepare
}

func (w *Message_Prepare) Pb() pbftpb.Message_Type {
	if w == nil {
		return nil
	}
	if w.Prepare == nil {
		return &pbftpb.Message_Prepare{}
	}
	return &pbftpb.Message_Prepare{Prepare: (w.Prepare).Pb()}
}

func (*Message_Prepare) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Message_Prepare]()}
}

type Message_Commit struct {
	Commit *Commit
}

func (*Message_Commit) isMessage_Type() {}

func (w *Message_Commit) Unwrap() *Commit {
	return w.Commit
}

func (w *Message_Commit) Pb() pbftpb.Message_Type {
	if w == nil {
		return nil
	}
	if w.Commit == nil {
		return &pbftpb.Message_Commit{}
	}
	return &pbftpb.Message_Commit{Commit: (w.Commit).Pb()}
}

func (*Message_Commit) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Message_Commit]()}
}

type Message_Done struct {
	Done *Done
}

func (*Message_Done) isMessage_Type() {}

func (w *Message_Done) Unwrap() *Done {
	return w.Done
}

func (w *Message_Done) Pb() pbftpb.Message_Type {
	if w == nil {
		return nil
	}
	if w.Done == nil {
		return &pbftpb.Message_Done{}
	}
	return &pbftpb.Message_Done{Done: (w.Done).Pb()}
}

func (*Message_Done) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Message_Done]()}
}

type Message_CatchUpRequest struct {
	CatchUpRequest *CatchUpRequest
}

func (*Message_CatchUpRequest) isMessage_Type() {}

func (w *Message_CatchUpRequest) Unwrap() *CatchUpRequest {
	return w.CatchUpRequest
}

func (w *Message_CatchUpRequest) Pb() pbftpb.Message_Type {
	if w == nil {
		return nil
	}
	if w.CatchUpRequest == nil {
		return &pbftpb.Message_CatchUpRequest{}
	}
	return &pbftpb.Message_CatchUpRequest{CatchUpRequest: (w.CatchUpRequest).Pb()}
}

func (*Message_CatchUpRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Message_CatchUpRequest]()}
}

type Message_CatchUpResponse struct {
	CatchUpResponse *CatchUpResponse
}

func (*Message_CatchUpResponse) isMessage_Type() {}

func (w *Message_CatchUpResponse) Unwrap() *CatchUpResponse {
	return w.CatchUpResponse
}

func (w *Message_CatchUpResponse) Pb() pbftpb.Message_Type {
	if w == nil {
		return nil
	}
	if w.CatchUpResponse == nil {
		return &pbftpb.Message_CatchUpResponse{}
	}
	return &pbftpb.Message_CatchUpResponse{CatchUpResponse: (w.CatchUpResponse).Pb()}
}

func (*Message_CatchUpResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Message_CatchUpResponse]()}
}

type Message_SignedViewChange struct {
	SignedViewChange *SignedViewChange
}

func (*Message_SignedViewChange) isMessage_Type() {}

func (w *Message_SignedViewChange) Unwrap() *SignedViewChange {
	return w.SignedViewChange
}

func (w *Message_SignedViewChange) Pb() pbftpb.Message_Type {
	if w == nil {
		return nil
	}
	if w.SignedViewChange == nil {
		return &pbftpb.Message_SignedViewChange{}
	}
	return &pbftpb.Message_SignedViewChange{SignedViewChange: (w.SignedViewChange).Pb()}
}

func (*Message_SignedViewChange) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Message_SignedViewChange]()}
}

type Message_PreprepareRequest struct {
	PreprepareRequest *PreprepareRequest
}

func (*Message_PreprepareRequest) isMessage_Type() {}

func (w *Message_PreprepareRequest) Unwrap() *PreprepareRequest {
	return w.PreprepareRequest
}

func (w *Message_PreprepareRequest) Pb() pbftpb.Message_Type {
	if w == nil {
		return nil
	}
	if w.PreprepareRequest == nil {
		return &pbftpb.Message_PreprepareRequest{}
	}
	return &pbftpb.Message_PreprepareRequest{PreprepareRequest: (w.PreprepareRequest).Pb()}
}

func (*Message_PreprepareRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Message_PreprepareRequest]()}
}

type Message_MissingPreprepare struct {
	MissingPreprepare *MissingPreprepare
}

func (*Message_MissingPreprepare) isMessage_Type() {}

func (w *Message_MissingPreprepare) Unwrap() *MissingPreprepare {
	return w.MissingPreprepare
}

func (w *Message_MissingPreprepare) Pb() pbftpb.Message_Type {
	if w == nil {
		return nil
	}
	if w.MissingPreprepare == nil {
		return &pbftpb.Message_MissingPreprepare{}
	}
	return &pbftpb.Message_MissingPreprepare{MissingPreprepare: (w.MissingPreprepare).Pb()}
}

func (*Message_MissingPreprepare) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Message_MissingPreprepare]()}
}

type Message_NewView struct {
	NewView *NewView
}

func (*Message_NewView) isMessage_Type() {}

func (w *Message_NewView) Unwrap() *NewView {
	return w.NewView
}

func (w *Message_NewView) Pb() pbftpb.Message_Type {
	if w == nil {
		return nil
	}
	if w.NewView == nil {
		return &pbftpb.Message_NewView{}
	}
	return &pbftpb.Message_NewView{NewView: (w.NewView).Pb()}
}

func (*Message_NewView) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Message_NewView]()}
}

func MessageFromPb(pb *pbftpb.Message) *Message {
	if pb == nil {
		return nil
	}
	return &Message{
		Type: Message_TypeFromPb(pb.Type),
	}
}

func (m *Message) Pb() *pbftpb.Message {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.Message{}
	{
		if m.Type != nil {
			pbMessage.Type = (m.Type).Pb()
		}
	}

	return pbMessage
}

func (*Message) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Message]()}
}

type Preprepare struct {
	Sn      types.SeqNr
	View    types1.ViewNr
	Data    []uint8
	Aborted bool
}

func PreprepareFromPb(pb *pbftpb.Preprepare) *Preprepare {
	if pb == nil {
		return nil
	}
	return &Preprepare{
		Sn:      (types.SeqNr)(pb.Sn),
		View:    (types1.ViewNr)(pb.View),
		Data:    pb.Data,
		Aborted: pb.Aborted,
	}
}

func (m *Preprepare) Pb() *pbftpb.Preprepare {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.Preprepare{}
	{
		pbMessage.Sn = (uint64)(m.Sn)
		pbMessage.View = (uint64)(m.View)
		pbMessage.Data = m.Data
		pbMessage.Aborted = m.Aborted
	}

	return pbMessage
}

func (*Preprepare) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Preprepare]()}
}

type Prepare struct {
	Sn     types.SeqNr
	View   types1.ViewNr
	Digest []uint8
}

func PrepareFromPb(pb *pbftpb.Prepare) *Prepare {
	if pb == nil {
		return nil
	}
	return &Prepare{
		Sn:     (types.SeqNr)(pb.Sn),
		View:   (types1.ViewNr)(pb.View),
		Digest: pb.Digest,
	}
}

func (m *Prepare) Pb() *pbftpb.Prepare {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.Prepare{}
	{
		pbMessage.Sn = (uint64)(m.Sn)
		pbMessage.View = (uint64)(m.View)
		pbMessage.Digest = m.Digest
	}

	return pbMessage
}

func (*Prepare) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Prepare]()}
}

type Commit struct {
	Sn     types.SeqNr
	View   types1.ViewNr
	Digest []uint8
}

func CommitFromPb(pb *pbftpb.Commit) *Commit {
	if pb == nil {
		return nil
	}
	return &Commit{
		Sn:     (types.SeqNr)(pb.Sn),
		View:   (types1.ViewNr)(pb.View),
		Digest: pb.Digest,
	}
}

func (m *Commit) Pb() *pbftpb.Commit {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.Commit{}
	{
		pbMessage.Sn = (uint64)(m.Sn)
		pbMessage.View = (uint64)(m.View)
		pbMessage.Digest = m.Digest
	}

	return pbMessage
}

func (*Commit) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Commit]()}
}

type Done struct {
	Digests [][]uint8
}

func DoneFromPb(pb *pbftpb.Done) *Done {
	if pb == nil {
		return nil
	}
	return &Done{
		Digests: pb.Digests,
	}
}

func (m *Done) Pb() *pbftpb.Done {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.Done{}
	{
		pbMessage.Digests = m.Digests
	}

	return pbMessage
}

func (*Done) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Done]()}
}

type CatchUpRequest struct {
	Digest []uint8
	Sn     types.SeqNr
}

func CatchUpRequestFromPb(pb *pbftpb.CatchUpRequest) *CatchUpRequest {
	if pb == nil {
		return nil
	}
	return &CatchUpRequest{
		Digest: pb.Digest,
		Sn:     (types.SeqNr)(pb.Sn),
	}
}

func (m *CatchUpRequest) Pb() *pbftpb.CatchUpRequest {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.CatchUpRequest{}
	{
		pbMessage.Digest = m.Digest
		pbMessage.Sn = (uint64)(m.Sn)
	}

	return pbMessage
}

func (*CatchUpRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.CatchUpRequest]()}
}

type CatchUpResponse struct {
	Resp *Preprepare
}

func CatchUpResponseFromPb(pb *pbftpb.CatchUpResponse) *CatchUpResponse {
	if pb == nil {
		return nil
	}
	return &CatchUpResponse{
		Resp: PreprepareFromPb(pb.Resp),
	}
}

func (m *CatchUpResponse) Pb() *pbftpb.CatchUpResponse {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.CatchUpResponse{}
	{
		if m.Resp != nil {
			pbMessage.Resp = (m.Resp).Pb()
		}
	}

	return pbMessage
}

func (*CatchUpResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.CatchUpResponse]()}
}

type SignedViewChange struct {
	ViewChange *ViewChange
	Signature  []uint8
}

func SignedViewChangeFromPb(pb *pbftpb.SignedViewChange) *SignedViewChange {
	if pb == nil {
		return nil
	}
	return &SignedViewChange{
		ViewChange: ViewChangeFromPb(pb.ViewChange),
		Signature:  pb.Signature,
	}
}

func (m *SignedViewChange) Pb() *pbftpb.SignedViewChange {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.SignedViewChange{}
	{
		if m.ViewChange != nil {
			pbMessage.ViewChange = (m.ViewChange).Pb()
		}
		pbMessage.Signature = m.Signature
	}

	return pbMessage
}

func (*SignedViewChange) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.SignedViewChange]()}
}

type PreprepareRequest struct {
	Digest []uint8
	Sn     types.SeqNr
}

func PreprepareRequestFromPb(pb *pbftpb.PreprepareRequest) *PreprepareRequest {
	if pb == nil {
		return nil
	}
	return &PreprepareRequest{
		Digest: pb.Digest,
		Sn:     (types.SeqNr)(pb.Sn),
	}
}

func (m *PreprepareRequest) Pb() *pbftpb.PreprepareRequest {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.PreprepareRequest{}
	{
		pbMessage.Digest = m.Digest
		pbMessage.Sn = (uint64)(m.Sn)
	}

	return pbMessage
}

func (*PreprepareRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.PreprepareRequest]()}
}

type MissingPreprepare struct {
	Preprepare *Preprepare
}

func MissingPreprepareFromPb(pb *pbftpb.MissingPreprepare) *MissingPreprepare {
	if pb == nil {
		return nil
	}
	return &MissingPreprepare{
		Preprepare: PreprepareFromPb(pb.Preprepare),
	}
}

func (m *MissingPreprepare) Pb() *pbftpb.MissingPreprepare {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.MissingPreprepare{}
	{
		if m.Preprepare != nil {
			pbMessage.Preprepare = (m.Preprepare).Pb()
		}
	}

	return pbMessage
}

func (*MissingPreprepare) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.MissingPreprepare]()}
}

type NewView struct {
	View              types1.ViewNr
	ViewChangeSenders []string
	SignedViewChanges []*SignedViewChange
	PreprepareSeqNrs  []types.SeqNr
	Preprepares       []*Preprepare
}

func NewViewFromPb(pb *pbftpb.NewView) *NewView {
	if pb == nil {
		return nil
	}
	return &NewView{
		View:              (types1.ViewNr)(pb.View),
		ViewChangeSenders: pb.ViewChangeSenders,
		SignedViewChanges: types2.ConvertSlice(pb.SignedViewChanges, func(t *pbftpb.SignedViewChange) *SignedViewChange {
			return SignedViewChangeFromPb(t)
		}),
		PreprepareSeqNrs: types2.ConvertSlice(pb.PreprepareSeqNrs, func(t uint64) types.SeqNr {
			return (types.SeqNr)(t)
		}),
		Preprepares: types2.ConvertSlice(pb.Preprepares, func(t *pbftpb.Preprepare) *Preprepare {
			return PreprepareFromPb(t)
		}),
	}
}

func (m *NewView) Pb() *pbftpb.NewView {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.NewView{}
	{
		pbMessage.View = (uint64)(m.View)
		pbMessage.ViewChangeSenders = m.ViewChangeSenders
		pbMessage.SignedViewChanges = types2.ConvertSlice(m.SignedViewChanges, func(t *SignedViewChange) *pbftpb.SignedViewChange {
			return (t).Pb()
		})
		pbMessage.PreprepareSeqNrs = types2.ConvertSlice(m.PreprepareSeqNrs, func(t types.SeqNr) uint64 {
			return (uint64)(t)
		})
		pbMessage.Preprepares = types2.ConvertSlice(m.Preprepares, func(t *Preprepare) *pbftpb.Preprepare {
			return (t).Pb()
		})
	}

	return pbMessage
}

func (*NewView) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.NewView]()}
}

type ViewChange struct {
	View types1.ViewNr
	PSet []*PSetEntry
	QSet []*QSetEntry
}

func ViewChangeFromPb(pb *pbftpb.ViewChange) *ViewChange {
	if pb == nil {
		return nil
	}
	return &ViewChange{
		View: (types1.ViewNr)(pb.View),
		PSet: types2.ConvertSlice(pb.PSet, func(t *pbftpb.PSetEntry) *PSetEntry {
			return PSetEntryFromPb(t)
		}),
		QSet: types2.ConvertSlice(pb.QSet, func(t *pbftpb.QSetEntry) *QSetEntry {
			return QSetEntryFromPb(t)
		}),
	}
}

func (m *ViewChange) Pb() *pbftpb.ViewChange {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.ViewChange{}
	{
		pbMessage.View = (uint64)(m.View)
		pbMessage.PSet = types2.ConvertSlice(m.PSet, func(t *PSetEntry) *pbftpb.PSetEntry {
			return (t).Pb()
		})
		pbMessage.QSet = types2.ConvertSlice(m.QSet, func(t *QSetEntry) *pbftpb.QSetEntry {
			return (t).Pb()
		})
	}

	return pbMessage
}

func (*ViewChange) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.ViewChange]()}
}

type PSetEntry struct {
	Sn     types.SeqNr
	View   types1.ViewNr
	Digest []uint8
}

func PSetEntryFromPb(pb *pbftpb.PSetEntry) *PSetEntry {
	if pb == nil {
		return nil
	}
	return &PSetEntry{
		Sn:     (types.SeqNr)(pb.Sn),
		View:   (types1.ViewNr)(pb.View),
		Digest: pb.Digest,
	}
}

func (m *PSetEntry) Pb() *pbftpb.PSetEntry {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.PSetEntry{}
	{
		pbMessage.Sn = (uint64)(m.Sn)
		pbMessage.View = (uint64)(m.View)
		pbMessage.Digest = m.Digest
	}

	return pbMessage
}

func (*PSetEntry) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.PSetEntry]()}
}

type QSetEntry struct {
	Sn     types.SeqNr
	View   types1.ViewNr
	Digest []uint8
}

func QSetEntryFromPb(pb *pbftpb.QSetEntry) *QSetEntry {
	if pb == nil {
		return nil
	}
	return &QSetEntry{
		Sn:     (types.SeqNr)(pb.Sn),
		View:   (types1.ViewNr)(pb.View),
		Digest: pb.Digest,
	}
}

func (m *QSetEntry) Pb() *pbftpb.QSetEntry {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.QSetEntry{}
	{
		pbMessage.Sn = (uint64)(m.Sn)
		pbMessage.View = (uint64)(m.View)
		pbMessage.Digest = m.Digest
	}

	return pbMessage
}

func (*QSetEntry) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.QSetEntry]()}
}

type Event struct {
	Type Event_Type
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() pbftpb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb pbftpb.Event_Type) Event_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *pbftpb.Event_ProposeTimeout:
		return &Event_ProposeTimeout{ProposeTimeout: ProposeTimeoutFromPb(pb.ProposeTimeout)}
	case *pbftpb.Event_ViewChangeSnTimeout:
		return &Event_ViewChangeSnTimeout{ViewChangeSnTimeout: ViewChangeSNTimeoutFromPb(pb.ViewChangeSnTimeout)}
	case *pbftpb.Event_ViewChangeSegTimeout:
		return &Event_ViewChangeSegTimeout{ViewChangeSegTimeout: ViewChangeSegTimeoutFromPb(pb.ViewChangeSegTimeout)}
	}
	return nil
}

type Event_ProposeTimeout struct {
	ProposeTimeout *ProposeTimeout
}

func (*Event_ProposeTimeout) isEvent_Type() {}

func (w *Event_ProposeTimeout) Unwrap() *ProposeTimeout {
	return w.ProposeTimeout
}

func (w *Event_ProposeTimeout) Pb() pbftpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.ProposeTimeout == nil {
		return &pbftpb.Event_ProposeTimeout{}
	}
	return &pbftpb.Event_ProposeTimeout{ProposeTimeout: (w.ProposeTimeout).Pb()}
}

func (*Event_ProposeTimeout) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Event_ProposeTimeout]()}
}

type Event_ViewChangeSnTimeout struct {
	ViewChangeSnTimeout *ViewChangeSNTimeout
}

func (*Event_ViewChangeSnTimeout) isEvent_Type() {}

func (w *Event_ViewChangeSnTimeout) Unwrap() *ViewChangeSNTimeout {
	return w.ViewChangeSnTimeout
}

func (w *Event_ViewChangeSnTimeout) Pb() pbftpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.ViewChangeSnTimeout == nil {
		return &pbftpb.Event_ViewChangeSnTimeout{}
	}
	return &pbftpb.Event_ViewChangeSnTimeout{ViewChangeSnTimeout: (w.ViewChangeSnTimeout).Pb()}
}

func (*Event_ViewChangeSnTimeout) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Event_ViewChangeSnTimeout]()}
}

type Event_ViewChangeSegTimeout struct {
	ViewChangeSegTimeout *ViewChangeSegTimeout
}

func (*Event_ViewChangeSegTimeout) isEvent_Type() {}

func (w *Event_ViewChangeSegTimeout) Unwrap() *ViewChangeSegTimeout {
	return w.ViewChangeSegTimeout
}

func (w *Event_ViewChangeSegTimeout) Pb() pbftpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.ViewChangeSegTimeout == nil {
		return &pbftpb.Event_ViewChangeSegTimeout{}
	}
	return &pbftpb.Event_ViewChangeSegTimeout{ViewChangeSegTimeout: (w.ViewChangeSegTimeout).Pb()}
}

func (*Event_ViewChangeSegTimeout) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Event_ViewChangeSegTimeout]()}
}

func EventFromPb(pb *pbftpb.Event) *Event {
	if pb == nil {
		return nil
	}
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *pbftpb.Event {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.Event{}
	{
		if m.Type != nil {
			pbMessage.Type = (m.Type).Pb()
		}
	}

	return pbMessage
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Event]()}
}

type ProposeTimeout struct {
	ProposeTimeout uint64
}

func ProposeTimeoutFromPb(pb *pbftpb.ProposeTimeout) *ProposeTimeout {
	if pb == nil {
		return nil
	}
	return &ProposeTimeout{
		ProposeTimeout: pb.ProposeTimeout,
	}
}

func (m *ProposeTimeout) Pb() *pbftpb.ProposeTimeout {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.ProposeTimeout{}
	{
		pbMessage.ProposeTimeout = m.ProposeTimeout
	}

	return pbMessage
}

func (*ProposeTimeout) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.ProposeTimeout]()}
}

type ViewChangeSNTimeout struct {
	View         types1.ViewNr
	NumCommitted uint64
}

func ViewChangeSNTimeoutFromPb(pb *pbftpb.ViewChangeSNTimeout) *ViewChangeSNTimeout {
	if pb == nil {
		return nil
	}
	return &ViewChangeSNTimeout{
		View:         (types1.ViewNr)(pb.View),
		NumCommitted: pb.NumCommitted,
	}
}

func (m *ViewChangeSNTimeout) Pb() *pbftpb.ViewChangeSNTimeout {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.ViewChangeSNTimeout{}
	{
		pbMessage.View = (uint64)(m.View)
		pbMessage.NumCommitted = m.NumCommitted
	}

	return pbMessage
}

func (*ViewChangeSNTimeout) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.ViewChangeSNTimeout]()}
}

type ViewChangeSegTimeout struct {
	ViewChangeSegTimeout uint64
}

func ViewChangeSegTimeoutFromPb(pb *pbftpb.ViewChangeSegTimeout) *ViewChangeSegTimeout {
	if pb == nil {
		return nil
	}
	return &ViewChangeSegTimeout{
		ViewChangeSegTimeout: pb.ViewChangeSegTimeout,
	}
}

func (m *ViewChangeSegTimeout) Pb() *pbftpb.ViewChangeSegTimeout {
	if m == nil {
		return nil
	}
	pbMessage := &pbftpb.ViewChangeSegTimeout{}
	{
		pbMessage.ViewChangeSegTimeout = m.ViewChangeSegTimeout
	}

	return pbMessage
}

func (*ViewChangeSegTimeout) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.ViewChangeSegTimeout]()}
}
