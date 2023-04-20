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
	return &pbftpb.Message_NewView{NewView: (w.NewView).Pb()}
}

func (*Message_NewView) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Message_NewView]()}
}

func MessageFromPb(pb *pbftpb.Message) *Message {
	return &Message{
		Type: Message_TypeFromPb(pb.Type),
	}
}

func (m *Message) Pb() *pbftpb.Message {
	return &pbftpb.Message{
		Type: (m.Type).Pb(),
	}
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
	return &Preprepare{
		Sn:      (types.SeqNr)(pb.Sn),
		View:    (types1.ViewNr)(pb.View),
		Data:    pb.Data,
		Aborted: pb.Aborted,
	}
}

func (m *Preprepare) Pb() *pbftpb.Preprepare {
	return &pbftpb.Preprepare{
		Sn:      (uint64)(m.Sn),
		View:    (uint64)(m.View),
		Data:    m.Data,
		Aborted: m.Aborted,
	}
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
	return &Prepare{
		Sn:     (types.SeqNr)(pb.Sn),
		View:   (types1.ViewNr)(pb.View),
		Digest: pb.Digest,
	}
}

func (m *Prepare) Pb() *pbftpb.Prepare {
	return &pbftpb.Prepare{
		Sn:     (uint64)(m.Sn),
		View:   (uint64)(m.View),
		Digest: m.Digest,
	}
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
	return &Commit{
		Sn:     (types.SeqNr)(pb.Sn),
		View:   (types1.ViewNr)(pb.View),
		Digest: pb.Digest,
	}
}

func (m *Commit) Pb() *pbftpb.Commit {
	return &pbftpb.Commit{
		Sn:     (uint64)(m.Sn),
		View:   (uint64)(m.View),
		Digest: m.Digest,
	}
}

func (*Commit) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Commit]()}
}

type Done struct {
	Digests [][]uint8
}

func DoneFromPb(pb *pbftpb.Done) *Done {
	return &Done{
		Digests: pb.Digests,
	}
}

func (m *Done) Pb() *pbftpb.Done {
	return &pbftpb.Done{
		Digests: m.Digests,
	}
}

func (*Done) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Done]()}
}

type CatchUpRequest struct {
	Digest []uint8
	Sn     types.SeqNr
}

func CatchUpRequestFromPb(pb *pbftpb.CatchUpRequest) *CatchUpRequest {
	return &CatchUpRequest{
		Digest: pb.Digest,
		Sn:     (types.SeqNr)(pb.Sn),
	}
}

func (m *CatchUpRequest) Pb() *pbftpb.CatchUpRequest {
	return &pbftpb.CatchUpRequest{
		Digest: m.Digest,
		Sn:     (uint64)(m.Sn),
	}
}

func (*CatchUpRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.CatchUpRequest]()}
}

type CatchUpResponse struct {
	Resp *Preprepare
}

func CatchUpResponseFromPb(pb *pbftpb.CatchUpResponse) *CatchUpResponse {
	return &CatchUpResponse{
		Resp: PreprepareFromPb(pb.Resp),
	}
}

func (m *CatchUpResponse) Pb() *pbftpb.CatchUpResponse {
	return &pbftpb.CatchUpResponse{
		Resp: (m.Resp).Pb(),
	}
}

func (*CatchUpResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.CatchUpResponse]()}
}

type SignedViewChange struct {
	ViewChange *ViewChange
	Signature  []uint8
}

func SignedViewChangeFromPb(pb *pbftpb.SignedViewChange) *SignedViewChange {
	return &SignedViewChange{
		ViewChange: ViewChangeFromPb(pb.ViewChange),
		Signature:  pb.Signature,
	}
}

func (m *SignedViewChange) Pb() *pbftpb.SignedViewChange {
	return &pbftpb.SignedViewChange{
		ViewChange: (m.ViewChange).Pb(),
		Signature:  m.Signature,
	}
}

func (*SignedViewChange) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.SignedViewChange]()}
}

type PreprepareRequest struct {
	Digest []uint8
	Sn     types.SeqNr
}

func PreprepareRequestFromPb(pb *pbftpb.PreprepareRequest) *PreprepareRequest {
	return &PreprepareRequest{
		Digest: pb.Digest,
		Sn:     (types.SeqNr)(pb.Sn),
	}
}

func (m *PreprepareRequest) Pb() *pbftpb.PreprepareRequest {
	return &pbftpb.PreprepareRequest{
		Digest: m.Digest,
		Sn:     (uint64)(m.Sn),
	}
}

func (*PreprepareRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.PreprepareRequest]()}
}

type MissingPreprepare struct {
	Preprepare *Preprepare
}

func MissingPreprepareFromPb(pb *pbftpb.MissingPreprepare) *MissingPreprepare {
	return &MissingPreprepare{
		Preprepare: PreprepareFromPb(pb.Preprepare),
	}
}

func (m *MissingPreprepare) Pb() *pbftpb.MissingPreprepare {
	return &pbftpb.MissingPreprepare{
		Preprepare: (m.Preprepare).Pb(),
	}
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
	return &pbftpb.NewView{
		View:              (uint64)(m.View),
		ViewChangeSenders: m.ViewChangeSenders,
		SignedViewChanges: types2.ConvertSlice(m.SignedViewChanges, func(t *SignedViewChange) *pbftpb.SignedViewChange {
			return (t).Pb()
		}),
		PreprepareSeqNrs: types2.ConvertSlice(m.PreprepareSeqNrs, func(t types.SeqNr) uint64 {
			return (uint64)(t)
		}),
		Preprepares: types2.ConvertSlice(m.Preprepares, func(t *Preprepare) *pbftpb.Preprepare {
			return (t).Pb()
		}),
	}
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
	return &pbftpb.ViewChange{
		View: (uint64)(m.View),
		PSet: types2.ConvertSlice(m.PSet, func(t *PSetEntry) *pbftpb.PSetEntry {
			return (t).Pb()
		}),
		QSet: types2.ConvertSlice(m.QSet, func(t *QSetEntry) *pbftpb.QSetEntry {
			return (t).Pb()
		}),
	}
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
	return &PSetEntry{
		Sn:     (types.SeqNr)(pb.Sn),
		View:   (types1.ViewNr)(pb.View),
		Digest: pb.Digest,
	}
}

func (m *PSetEntry) Pb() *pbftpb.PSetEntry {
	return &pbftpb.PSetEntry{
		Sn:     (uint64)(m.Sn),
		View:   (uint64)(m.View),
		Digest: m.Digest,
	}
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
	return &QSetEntry{
		Sn:     (types.SeqNr)(pb.Sn),
		View:   (types1.ViewNr)(pb.View),
		Digest: pb.Digest,
	}
}

func (m *QSetEntry) Pb() *pbftpb.QSetEntry {
	return &pbftpb.QSetEntry{
		Sn:     (uint64)(m.Sn),
		View:   (uint64)(m.View),
		Digest: m.Digest,
	}
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
	switch pb := pb.(type) {
	case *pbftpb.Event_ProposeTimeout:
		return &Event_ProposeTimeout{ProposeTimeout: pb.ProposeTimeout}
	case *pbftpb.Event_ViewChangeSnTimeout:
		return &Event_ViewChangeSnTimeout{ViewChangeSnTimeout: VCSNTimeoutFromPb(pb.ViewChangeSnTimeout)}
	case *pbftpb.Event_ViewChangeSegTimeout:
		return &Event_ViewChangeSegTimeout{ViewChangeSegTimeout: pb.ViewChangeSegTimeout}
	}
	return nil
}

type Event_ProposeTimeout struct {
	ProposeTimeout uint64
}

func (*Event_ProposeTimeout) isEvent_Type() {}

func (w *Event_ProposeTimeout) Unwrap() uint64 {
	return w.ProposeTimeout
}

func (w *Event_ProposeTimeout) Pb() pbftpb.Event_Type {
	return &pbftpb.Event_ProposeTimeout{ProposeTimeout: w.ProposeTimeout}
}

func (*Event_ProposeTimeout) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Event_ProposeTimeout]()}
}

type Event_ViewChangeSnTimeout struct {
	ViewChangeSnTimeout *VCSNTimeout
}

func (*Event_ViewChangeSnTimeout) isEvent_Type() {}

func (w *Event_ViewChangeSnTimeout) Unwrap() *VCSNTimeout {
	return w.ViewChangeSnTimeout
}

func (w *Event_ViewChangeSnTimeout) Pb() pbftpb.Event_Type {
	return &pbftpb.Event_ViewChangeSnTimeout{ViewChangeSnTimeout: (w.ViewChangeSnTimeout).Pb()}
}

func (*Event_ViewChangeSnTimeout) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Event_ViewChangeSnTimeout]()}
}

type Event_ViewChangeSegTimeout struct {
	ViewChangeSegTimeout uint64
}

func (*Event_ViewChangeSegTimeout) isEvent_Type() {}

func (w *Event_ViewChangeSegTimeout) Unwrap() uint64 {
	return w.ViewChangeSegTimeout
}

func (w *Event_ViewChangeSegTimeout) Pb() pbftpb.Event_Type {
	return &pbftpb.Event_ViewChangeSegTimeout{ViewChangeSegTimeout: w.ViewChangeSegTimeout}
}

func (*Event_ViewChangeSegTimeout) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Event_ViewChangeSegTimeout]()}
}

func EventFromPb(pb *pbftpb.Event) *Event {
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *pbftpb.Event {
	return &pbftpb.Event{
		Type: (m.Type).Pb(),
	}
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.Event]()}
}

type VCSNTimeout struct {
	View         types1.ViewNr
	NumCommitted uint64
}

func VCSNTimeoutFromPb(pb *pbftpb.VCSNTimeout) *VCSNTimeout {
	return &VCSNTimeout{
		View:         (types1.ViewNr)(pb.View),
		NumCommitted: pb.NumCommitted,
	}
}

func (m *VCSNTimeout) Pb() *pbftpb.VCSNTimeout {
	return &pbftpb.VCSNTimeout{
		View:         (uint64)(m.View),
		NumCommitted: m.NumCommitted,
	}
}

func (*VCSNTimeout) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.VCSNTimeout]()}
}

type HashOrigin struct {
	Type HashOrigin_Type
}

type HashOrigin_Type interface {
	mirreflect.GeneratedType
	isHashOrigin_Type()
	Pb() pbftpb.HashOrigin_Type
}

type HashOrigin_TypeWrapper[T any] interface {
	HashOrigin_Type
	Unwrap() *T
}

func HashOrigin_TypeFromPb(pb pbftpb.HashOrigin_Type) HashOrigin_Type {
	switch pb := pb.(type) {
	case *pbftpb.HashOrigin_Preprepare:
		return &HashOrigin_Preprepare{Preprepare: PreprepareFromPb(pb.Preprepare)}
	case *pbftpb.HashOrigin_MissingPreprepare:
		return &HashOrigin_MissingPreprepare{MissingPreprepare: PreprepareFromPb(pb.MissingPreprepare)}
	case *pbftpb.HashOrigin_NewView:
		return &HashOrigin_NewView{NewView: NewViewFromPb(pb.NewView)}
	case *pbftpb.HashOrigin_EmptyPreprepares:
		return &HashOrigin_EmptyPreprepares{EmptyPreprepares: pb.EmptyPreprepares}
	case *pbftpb.HashOrigin_CatchUpResponse:
		return &HashOrigin_CatchUpResponse{CatchUpResponse: PreprepareFromPb(pb.CatchUpResponse)}
	}
	return nil
}

type HashOrigin_Preprepare struct {
	Preprepare *Preprepare
}

func (*HashOrigin_Preprepare) isHashOrigin_Type() {}

func (w *HashOrigin_Preprepare) Unwrap() *Preprepare {
	return w.Preprepare
}

func (w *HashOrigin_Preprepare) Pb() pbftpb.HashOrigin_Type {
	return &pbftpb.HashOrigin_Preprepare{Preprepare: (w.Preprepare).Pb()}
}

func (*HashOrigin_Preprepare) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.HashOrigin_Preprepare]()}
}

type HashOrigin_MissingPreprepare struct {
	MissingPreprepare *Preprepare
}

func (*HashOrigin_MissingPreprepare) isHashOrigin_Type() {}

func (w *HashOrigin_MissingPreprepare) Unwrap() *Preprepare {
	return w.MissingPreprepare
}

func (w *HashOrigin_MissingPreprepare) Pb() pbftpb.HashOrigin_Type {
	return &pbftpb.HashOrigin_MissingPreprepare{MissingPreprepare: (w.MissingPreprepare).Pb()}
}

func (*HashOrigin_MissingPreprepare) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.HashOrigin_MissingPreprepare]()}
}

type HashOrigin_NewView struct {
	NewView *NewView
}

func (*HashOrigin_NewView) isHashOrigin_Type() {}

func (w *HashOrigin_NewView) Unwrap() *NewView {
	return w.NewView
}

func (w *HashOrigin_NewView) Pb() pbftpb.HashOrigin_Type {
	return &pbftpb.HashOrigin_NewView{NewView: (w.NewView).Pb()}
}

func (*HashOrigin_NewView) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.HashOrigin_NewView]()}
}

type HashOrigin_EmptyPreprepares struct {
	EmptyPreprepares uint64
}

func (*HashOrigin_EmptyPreprepares) isHashOrigin_Type() {}

func (w *HashOrigin_EmptyPreprepares) Unwrap() uint64 {
	return w.EmptyPreprepares
}

func (w *HashOrigin_EmptyPreprepares) Pb() pbftpb.HashOrigin_Type {
	return &pbftpb.HashOrigin_EmptyPreprepares{EmptyPreprepares: w.EmptyPreprepares}
}

func (*HashOrigin_EmptyPreprepares) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.HashOrigin_EmptyPreprepares]()}
}

type HashOrigin_CatchUpResponse struct {
	CatchUpResponse *Preprepare
}

func (*HashOrigin_CatchUpResponse) isHashOrigin_Type() {}

func (w *HashOrigin_CatchUpResponse) Unwrap() *Preprepare {
	return w.CatchUpResponse
}

func (w *HashOrigin_CatchUpResponse) Pb() pbftpb.HashOrigin_Type {
	return &pbftpb.HashOrigin_CatchUpResponse{CatchUpResponse: (w.CatchUpResponse).Pb()}
}

func (*HashOrigin_CatchUpResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.HashOrigin_CatchUpResponse]()}
}

func HashOriginFromPb(pb *pbftpb.HashOrigin) *HashOrigin {
	return &HashOrigin{
		Type: HashOrigin_TypeFromPb(pb.Type),
	}
}

func (m *HashOrigin) Pb() *pbftpb.HashOrigin {
	return &pbftpb.HashOrigin{
		Type: (m.Type).Pb(),
	}
}

func (*HashOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*pbftpb.HashOrigin]()}
}
