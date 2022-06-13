// Package crypto provides an implementation of the Crypto module.
// It supports RSA and ECDSA signatures.
package crypto

import (
	"fmt"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/statuspb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type Crypto struct {
	modules.Module
	impl Impl
}

func New(impl Impl) *Crypto {
	return &Crypto{impl: impl}
}

func (c *Crypto) ApplyEvents(eventsIn *events.EventList) (*events.EventList, error) {

	// TODO: Parallelize this.

	eventsOut := &events.EventList{}

	iter := eventsIn.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {
		evts, err := c.ApplyEvent(event)
		if err != nil {
			return nil, err
		}
		eventsOut.PushBackList(evts)
	}

	return eventsOut, nil
}

func (c *Crypto) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpb.Event_VerifyRequestSig:
		// Verify client request signature.
		// The signature is only computed (and verified) over the digest of a request.
		// The other fields can safely be ignored.

		// Convenience variable
		reqRef := e.VerifyRequestSig.RequestRef

		// Verify signature.
		err := c.impl.VerifyClientSig(
			[][]byte{reqRef.Digest},
			e.VerifyRequestSig.Signature,
			t.ClientID(reqRef.ClientId))

		// Create result event, depending on verification outcome.
		if err != nil {
			return (&events.EventList{}).PushBack(
				events.RequestSigVerified("clientTracker", reqRef, false, err.Error()),
			), nil
		}
		return (&events.EventList{}).PushBack(
			events.RequestSigVerified("clientTracker", reqRef, true, ""),
		), nil
	case *eventpb.Event_SignRequest:
		// Compute a signature over the provided data and produce a SignResult event.

		signature, err := c.impl.Sign(e.SignRequest.Data)
		if err != nil {
			return nil, err
		}
		return (&events.EventList{}).PushBack(
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
			err := c.impl.VerifyNodeSig(data.Data, verifyEvent.Signatures[i], t.NodeID(verifyEvent.NodeIds[i]))
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
		return (&events.EventList{}).PushBack(events.NodeSigsVerified(
			t.ModuleID(verifyEvent.Origin.Module),
			results,
			errors,
			t.NodeIDSlice(verifyEvent.NodeIds),
			verifyEvent.Origin,
			allOK,
		)), nil

	default:
		// Complain about all other incoming event types.
		return nil, fmt.Errorf("unexpected type of Crypto event: %T", event.Type)
	}
}

func (c *Crypto) Status() (s *statuspb.ProtocolStatus, err error) {
	//TODO implement me
	panic("implement me")
}
