package main

import (
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// nullNet represents a Net module that simply drops all messages and never delivers any.
// It is used for instantiating a debugger node, to which messages are not fed from the network,
// but from a pre-recorded event log.
type nullNet struct {
}

func (dn *nullNet) Send(_ t.NodeID, _ *messagepb.Message) error {
	return nil
}

func (dn *nullNet) ReceiveChan() <-chan modules.ReceivedMessage {
	return nil
}
