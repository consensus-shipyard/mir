// TODO: This code should eventually be generated based on events.go.

package pingpongevents

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/net/grpc"
	t "github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

func UponGRPCMessage[M any](m dsl.Module, handler func(msg *M, from t.NodeID) error) {
	dsl.UponEvent(m, func(msgRecEvent *grpc.MessageReceived) error {
		msg, err := ParseMessage(msgRecEvent.Data())
		if err != nil {
			return es.Errorf("failed parsing message: %w", err)
		}

		message, ok := msg.(*M)
		if !ok {
			// TODO: Define a dsl package level error saying "unsupported type" or something similar and return it here.
			// In the implementation of the DSL module, when looping through a list of handlers,
			// make sure that at least one of the handlers does NOT return it.
			// This will prevent the currently existing following problem:
			// Each concrete message type has a separate handler for the MessageReceived event. For example, say there
			// are two MessageReceived handlers, one checking for message type A and one checking for message type B
			// (the line above). If a MessageReceived event containing message type B arrives,
			// the A-handler does nothing (returns nil here) and only the B-handler actually executes
			// the given (sub-)handler function. The problem is that if a MessageReceived event
			// containing message type C arrives, none of the handlers will do anything and what is probably an error
			// will go unnoticed.
			// Ideally, there should be a possibility to define a default handler for any event type that will only
			// be called if all other handlers return "unsupported type". The default implementation of this default
			// handler could just return the "unsupported type" error. If the user chooses to tolerate unsupported types
			// more gracefully (as would make sense if messages received from the network are concerned), they will
			// always be able to override this default implementation.
			// This concept already exists on the top level of event handling as UponOtherEvent.
			// The idea is to apply it also to each top-level event type.
			return nil
		}

		return handler(message, msgRecEvent.SourceNode)
	})
}
