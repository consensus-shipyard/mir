package crypto

import (
	"hash"

	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/hasherpb"
	hasherpbevents "github.com/filecoin-project/mir/pkg/pb/hasherpb/events"
	hasherpbtypes "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
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

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (hasher *Hasher) ImplementsModule() {}

func (hasher *Hasher) ApplyEvents(eventsIn *stdtypes.EventList) (*stdtypes.EventList, error) {
	return modules.ApplyEventsConcurrently(eventsIn, hasher.ApplyEvent)
}

func (hasher *Hasher) ApplyEvent(event stdtypes.Event) (*stdtypes.EventList, error) {

	// We only support proto events.
	pbevent, ok := event.(*eventpb.Event)
	if !ok {
		return nil, es.Errorf("The hasher module only supports proto events, received %T", event)
	}

	switch e := pbevent.Type.(type) {
	case *eventpb.Event_Init:
		// no actions on init
		return stdtypes.EmptyList(), nil
	case *eventpb.Event_Hasher:
		switch e := e.Hasher.Type.(type) {
		case *hasherpb.Event_Request:
			// Return all computed digests in one common event.
			return stdtypes.ListOf(hasherpbevents.Result(
				stdtypes.ModuleID(e.Request.Origin.Module),
				hasher.computeDigests(e.Request.Data),
				hasherpbtypes.HashOriginFromPb(e.Request.Origin),
			).Pb()), nil
		case *hasherpb.Event_RequestOne:
			// Return a single computed digests.
			return stdtypes.ListOf(hasherpbevents.ResultOne(
				stdtypes.ModuleID(e.RequestOne.Origin.Module),
				hasher.computeDigests([]*hasherpb.HashData{e.RequestOne.Data})[0],
				hasherpbtypes.HashOriginFromPb(e.RequestOne.Origin),
			).Pb()), nil
		default:
			return nil, es.Errorf("unexpected hasher event type: %T", e)
		}
	default:
		// Complain about all other incoming event types.
		return nil, es.Errorf("unexpected type of Hash event: %T", pbevent.Type)
	}
}

func (hasher *Hasher) computeDigests(allData []*hasherpb.HashData) [][]byte {
	// Create a slice for the resulting digests containing one element for each data item to be hashed.
	digests := make([][]byte, len(allData))

	// Hash each data item contained in the event
	for i, data := range allData {

		// One data item consists of potentially multiple byte slices.
		// Add each of them to the hash function.
		h := hasher.hashImpl.New()
		for _, d := range data.Data {
			h.Write(d)
		}

		// Save resulting digest in the result slice
		digests[i] = h.Sum(nil)
	}

	return digests
}
