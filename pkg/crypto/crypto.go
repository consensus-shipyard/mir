// Package crypto provides an implementation of the MirModule module.
// It supports RSA and ECDSA signatures.
package crypto

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type MirModule struct {
	crypto Crypto
}

func New(crypto Crypto) *MirModule {
	return &MirModule{crypto: crypto}
}

func (c *MirModule) ApplyEvents(eventsIn *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsConcurrently(eventsIn, c.ApplyEvent)
}

func (c *MirModule) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
		// no actions on init
		return events.EmptyList(), nil
	case *eventpb.Event_SignRequest:
		// Compute a signature over the provided data and produce a SignResult event.

		signature, err := c.crypto.Sign(e.SignRequest.Data)
		if err != nil {
			return nil, err
		}
		return events.ListOf(
			events.SignResult(t.ModuleID(e.SignRequest.Origin.Module), signature, e.SignRequest.Origin),
		), nil

	case *eventpb.Event_VerifyNodeSigs:
		// Verify a batch of node signatures

		// Convenience variables
		verifyEvent := e.VerifyNodeSigs
		results := make([]bool, len(verifyEvent.Data))
		errors := make([]string, len(verifyEvent.Data))
		allOK := true

		// Verify each signature.
		for i, data := range verifyEvent.Data {
			err := c.crypto.Verify(data.Data, verifyEvent.Signatures[i], t.NodeID(verifyEvent.NodeIds[i]))
			if err == nil {
				results[i] = true
				errors[i] = ""
			} else {
				results[i] = false
				errors[i] = err.Error()
				allOK = false
			}
		}

		// Return result event
		return events.ListOf(events.NodeSigsVerified(
			t.ModuleID(verifyEvent.Origin.Module),
			results,
			errors,
			t.NodeIDSlice(verifyEvent.NodeIds),
			verifyEvent.Origin,
			allOK,
		)), nil

	default:
		// Complain about all other incoming event types.
		return nil, fmt.Errorf("unexpected type of MirModule event: %T", event.Type)
	}
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (c *MirModule) ImplementsModule() {}
