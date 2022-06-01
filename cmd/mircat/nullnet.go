package main

import (
	"context"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/statuspb"
)

// nullNet represents a Net module that simply drops all messages and never delivers any.
// It is used for instantiating a debugger node, to which messages are not fed from the network,
// but from a pre-recorded event log.
type nullNet struct {
}

func (nn *nullNet) Run(
	ctx context.Context,
	eventsIn <-chan *events.EventList,
	eventsOut chan<- *events.EventList,
	interceptor eventlog.Interceptor,
) error {
	<-ctx.Done()
	return nil
}

func (nn *nullNet) Status() (s *statuspb.ProtocolStatus, err error) {
	return nil, nil
}
