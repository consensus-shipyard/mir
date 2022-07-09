package crypto

import (
	"fmt"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
	"hash"
)

type HashImpl interface {
	New() hash.Hash
}

type Hasher struct {
	hashImpl HashImpl
}

func NewHasher(hashImpl HashImpl) *Hasher {
	return &Hasher{
		hashImpl: hashImpl,
	}
}

func (hasher *Hasher) ApplyEvents(eventsIn *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsConcurrently(eventsIn, hasher.ApplyEvent)
}

func (hasher *Hasher) ApplyEvent(event *eventpb.Event) (*events.EventList, error) {
	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
		// no actions on init
		return events.EmptyList(), nil
	case *eventpb.Event_HashRequest:
		// HashRequest is the only event understood by the hasher module.

		// Create a slice for the resulting digests containing one element for each data item to be hashed.
		digests := make([][]byte, len(e.HashRequest.Data))

		// Hash each data item contained in the event
		for i, data := range e.HashRequest.Data {

			// One data item consists of potentially multiple byte slices.
			// Add each of them to the hash function.
			h := hasher.hashImpl.New()
			for _, d := range data.Data {
				h.Write(d)
			}

			// Save resulting digest in the result slice
			digests[i] = h.Sum(nil)
		}

		// Return all computed digests in one common event.
		return events.ListOf(
			events.HashResult(t.ModuleID(e.HashRequest.Origin.Module), digests, e.HashRequest.Origin),
		), nil
	default:
		// Complain about all other incoming event types.
		return nil, fmt.Errorf("unexpected type of Hash event: %T", event.Type)
	}
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (hasher *Hasher) ImplementsModule() {}
