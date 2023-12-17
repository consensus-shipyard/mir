package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/bcbpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
)

type controlModule struct {
	eventsOut chan *stdtypes.EventList
	isLeader  bool
}

func newControlModule(isLeader bool) modules.ActiveModule {
	return &controlModule{
		eventsOut: make(chan *stdtypes.EventList),
		isLeader:  isLeader,
	}
}

func (m *controlModule) ImplementsModule() {}

func (m *controlModule) ApplyEvents(_ context.Context, events *stdtypes.EventList) error {
	iter := events.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {

		// We only support proto events.
		pbevent, ok := event.(*eventpb.Event)
		if !ok {
			return es.Errorf("The bcb control module only supports proto events, received %T", event)
		}

		switch pbevent.Type.(type) {

		case *eventpb.Event_Init:
			if m.isLeader {
				go func() {
					err := m.readMessageFromConsole()
					if err != nil {
						panic(err)
					}
				}()
			} else {
				fmt.Println("Waiting for the message...")
			}

		case *eventpb.Event_Bcb:
			bcbEvent := pbevent.Type.(*eventpb.Event_Bcb).Bcb
			switch bcbEvent.Type.(type) {

			case *bcbpb.Event_Deliver:
				deliverEvent := bcbEvent.Type.(*bcbpb.Event_Deliver).Deliver
				fmt.Println("Leader says: ", string(deliverEvent.Data))

			default:
				return es.Errorf("unknown bcb event type: %T", bcbEvent.Type)
			}

		default:
			return es.Errorf("unknown event type: %T", pbevent.Type)
		}
	}

	return nil
}

func (m *controlModule) EventsOut() <-chan *stdtypes.EventList {
	return m.eventsOut
}

func (m *controlModule) readMessageFromConsole() error {
	// Read the user input
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Type in a message and press Enter: ")
	scanner.Scan()
	if scanner.Err() != nil {
		return es.Errorf("error reading from console: %w", scanner.Err())
	}

	m.eventsOut <- stdtypes.ListOf(&eventpb.Event{
		DestModule: "bcb",
		Type: &eventpb.Event_Bcb{
			Bcb: &bcbpb.Event{
				Type: &bcbpb.Event_Request{
					Request: &bcbpb.BroadcastRequest{
						Data: []byte(scanner.Text()),
					},
				},
			},
		},
	})

	return nil
}
