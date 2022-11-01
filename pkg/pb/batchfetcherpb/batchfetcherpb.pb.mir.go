package batchfetcherpb

import (
	commonpb "github.com/filecoin-project/mir/pkg/pb/commonpb"
)

type Event_Type = isEvent_Type

type Event_TypeWrapper[Ev any] interface {
	Event_Type
	Unwrap() *Ev
}

func (p *Event_NewOrderedBatch) Unwrap() *NewOrderedBatch {
	return p.NewOrderedBatch
}

func (p *Event_ClientProgress) Unwrap() *commonpb.ClientProgress {
	return p.ClientProgress
}
