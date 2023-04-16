package pbftpb

type Message_Type = isMessage_Type

type Message_TypeWrapper[T any] interface {
	Message_Type
	Unwrap() *T
}

func (w *Message_Preprepare) Unwrap() *Preprepare {
	return w.Preprepare
}

func (w *Message_Prepare) Unwrap() *Prepare {
	return w.Prepare
}

func (w *Message_Commit) Unwrap() *Commit {
	return w.Commit
}

func (w *Message_Done) Unwrap() *Done {
	return w.Done
}

func (w *Message_CatchUpRequest) Unwrap() *CatchUpRequest {
	return w.CatchUpRequest
}

func (w *Message_CatchUpResponse) Unwrap() *CatchUpResponse {
	return w.CatchUpResponse
}

func (w *Message_SignedViewChange) Unwrap() *SignedViewChange {
	return w.SignedViewChange
}

func (w *Message_PreprepareRequest) Unwrap() *PreprepareRequest {
	return w.PreprepareRequest
}

func (w *Message_MissingPreprepare) Unwrap() *MissingPreprepare {
	return w.MissingPreprepare
}

func (w *Message_NewView) Unwrap() *NewView {
	return w.NewView
}

type Event_Type = isEvent_Type

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func (w *Event_ProposeTimeout) Unwrap() uint64 {
	return w.ProposeTimeout
}

func (w *Event_ViewChangeSnTimeout) Unwrap() *VCSNTimeout {
	return w.ViewChangeSnTimeout
}

func (w *Event_ViewChangeSegTimeout) Unwrap() uint64 {
	return w.ViewChangeSegTimeout
}

type HashOrigin_Type = isHashOrigin_Type

type HashOrigin_TypeWrapper[T any] interface {
	HashOrigin_Type
	Unwrap() *T
}

func (w *HashOrigin_Preprepare) Unwrap() *Preprepare {
	return w.Preprepare
}

func (w *HashOrigin_MissingPreprepare) Unwrap() *Preprepare {
	return w.MissingPreprepare
}

func (w *HashOrigin_NewView) Unwrap() *NewView {
	return w.NewView
}

func (w *HashOrigin_EmptyPreprepares) Unwrap() uint64 {
	return w.EmptyPreprepares
}

func (w *HashOrigin_CatchUpResponse) Unwrap() *Preprepare {
	return w.CatchUpResponse
}
