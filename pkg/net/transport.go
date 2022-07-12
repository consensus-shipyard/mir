package net

import (
	"context"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type Transport interface {
	Start() error
	Stop()
	ImplementsModule()
	EventsOut() <-chan *events.EventList
	ApplyEvents(ctx context.Context, eventList *events.EventList) error
	Send(dest t.NodeID, msg *messagepb.Message) error
	Connect(ctx context.Context)
}
