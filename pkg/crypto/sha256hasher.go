package crypto

import (
	"crypto"
	"fmt"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/statuspb"
)

type SHA256Hasher struct{}

func (hasher *SHA256Hasher) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpb.Event_HashRequest:
		// HashRequest is the only event understood by the hasher module.

		// Create a slice for the resulting digests containing one element for each data item to be hashed.
		digests := make([][]byte, len(e.HashRequest.Data))

		// Hash each data item contained in the event
		for i, data := range e.HashRequest.Data {

			// One data item consists of potentially multiple byte slices.
			// Add each of them to the hash function.
			h := crypto.SHA256.New()
			for _, d := range data.Data {
				h.Write(d)
			}

			// Save resulting digest in the result slice
			digests[i] = h.Sum(nil)
		}

		// Return all computed digests in one common event.
		return (&events.EventList{}).PushBack(
			events.HashResult(e.HashRequest.Origin.Module, digests, e.HashRequest.Origin),
		), nil
	default:
		// Complain about all other incoming event types.
		return nil, fmt.Errorf("unexpected type of Hash event: %T", event.Type)
	}
}

func (hasher *SHA256Hasher) Status() (s *statuspb.ProtocolStatus, err error) {
	//TODO implement me
	panic("implement me")
}
