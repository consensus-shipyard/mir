// Package crypto provides an implementation of the MirModule module.
// It supports RSA and ECDSA signatures.
package crypto

import (
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/cryptopb"
	cryptopbevents "github.com/filecoin-project/mir/pkg/pb/cryptopb/events"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type MirModule struct {
	crypto Crypto
}

func New(crypto Crypto) *MirModule {
	return &MirModule{crypto: crypto}
}

func (c *MirModule) ApplyEvents(eventsIn *stdtypes.EventList) (*stdtypes.EventList, error) {
	return modules.ApplyEventsConcurrently(eventsIn, c.ApplyEvent)
}

func (c *MirModule) ApplyEvent(event stdtypes.Event) (*stdtypes.EventList, error) {

	// We only support proto events.
	pbevent, ok := event.(*eventpb.Event)
	if !ok {
		return nil, es.Errorf("The crypto module only supports proto events, received %T", event)
	}

	switch e := pbevent.Type.(type) {
	case *eventpb.Event_Init:
		// no actions on init
		return stdtypes.EmptyList(), nil
	case *eventpb.Event_Crypto:
		switch e := e.Crypto.Type.(type) {
		case *cryptopb.Event_SignRequest:
			// Compute a signature over the provided data and produce a SignResult event.

			signature, err := c.crypto.Sign(e.SignRequest.Data.Data)
			if err != nil {
				return nil, err
			}
			return stdtypes.ListOf(
				cryptopbevents.SignResult(
					t.ModuleID(e.SignRequest.Origin.Module),
					signature,
					cryptopbtypes.SignOriginFromPb(e.SignRequest.Origin),
				).Pb()), nil

		case *cryptopb.Event_VerifySigs:
			// Verify a batch of node signatures

			// Convenience variables
			verifyEvent := e.VerifySigs
			errors := make([]error, len(verifyEvent.Data))
			allOK := true

			// Verify each signature.
			for i, data := range verifyEvent.Data {
				errors[i] = c.crypto.Verify(data.Data, verifyEvent.Signatures[i], t.NodeID(verifyEvent.NodeIds[i]))
				if errors[i] != nil {
					allOK = false
				}
			}

			// Return result event
			return stdtypes.ListOf(cryptopbevents.SigsVerified(
				t.ModuleID(verifyEvent.Origin.Module),
				cryptopbtypes.SigVerOriginFromPb(verifyEvent.Origin),
				t.NodeIDSlice(verifyEvent.NodeIds),
				errors,
				allOK,
			).Pb()), nil

		case *cryptopb.Event_VerifySig:
			err := c.crypto.Verify(
				e.VerifySig.Data.Data,
				e.VerifySig.Signature,
				t.NodeID(e.VerifySig.NodeId),
			)

			return stdtypes.ListOf(cryptopbevents.SigVerified(
				t.ModuleID(e.VerifySig.Origin.Module),
				cryptopbtypes.SigVerOriginFromPb(e.VerifySig.Origin),
				t.NodeID(e.VerifySig.NodeId),
				err,
			).Pb()), nil

		default:
			return nil, es.Errorf("unexpected type of crypto event: %T", e)
		}
	default:
		// Complain about all other incoming event types.
		return nil, es.Errorf("unexpected type of MirModule event: %T", pbevent.Type)
	}
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (c *MirModule) ImplementsModule() {}
