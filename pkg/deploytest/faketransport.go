/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// TODO: This is the original old code with very few modifications.
//       Go through all of it, comment what is to be kept and delete what is not needed.

package deploytest

import (
	"context"
	"fmt"
	"sync"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	t "github.com/filecoin-project/mir/pkg/types"
)

type FakeLink struct {
	FakeTransport *FakeTransport
	Source        t.NodeID
	DoneC         chan struct{}
}

func (fl *FakeLink) ApplyEvents(
	ctx context.Context,
	eventList *events.EventList,
) error {
	iter := eventList.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {

		switch e := event.Type.(type) {
		case *eventpb.Event_Init:
			// no actions on init
		case *eventpb.Event_SendMessage:
			for _, destID := range e.SendMessage.Destinations {
				if t.NodeID(destID) == fl.Source {
					// Send message to myself bypassing the network.
					receivedEvent := events.MessageReceived(
						t.ModuleID(e.SendMessage.Msg.DestModule),
						fl.Source,
						e.SendMessage.Msg,
					)
					eventsOut := fl.FakeTransport.NodeSinks[fl.Source]
					go func() {
						select {
						case eventsOut <- events.ListOf(receivedEvent):
						case <-ctx.Done():
						}
					}()
				} else {
					// Send message to another node.
					if err := fl.Send(t.NodeID(destID), e.SendMessage.Msg); err != nil { // nolint
						// TODO: Handle sending errors (and remove "nolint" comment above).
					}
				}
			}
		default:
			return fmt.Errorf("unexpected type of Net event: %T", event.Type)
		}
	}

	return nil
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (fl *FakeLink) ImplementsModule() {}

func (fl *FakeLink) Send(dest t.NodeID, msg *messagepb.Message) error {
	fl.FakeTransport.Send(fl.Source, dest, msg)
	return nil
}

func (fl *FakeLink) EventsOut() <-chan *events.EventList {
	return fl.FakeTransport.NodeSinks[fl.Source]
}

type FakeTransport struct {
	// Buffers is source x dest
	Buffers   map[t.NodeID]map[t.NodeID]chan *events.EventList
	NodeSinks map[t.NodeID]chan *events.EventList
	WaitGroup sync.WaitGroup
}

func NewFakeTransport(nodeIDs []t.NodeID) *FakeTransport {
	buffers := make(map[t.NodeID]map[t.NodeID]chan *events.EventList)
	nodeSinks := make(map[t.NodeID]chan *events.EventList)
	for _, sourceID := range nodeIDs {
		buffers[sourceID] = make(map[t.NodeID]chan *events.EventList)
		for _, destID := range nodeIDs {
			if sourceID == destID {
				continue
			}
			buffers[sourceID][destID] = make(chan *events.EventList, 10000)
		}
		nodeSinks[sourceID] = make(chan *events.EventList)
	}

	return &FakeTransport{
		Buffers:   buffers,
		NodeSinks: nodeSinks,
	}
}

func (ft *FakeTransport) Send(source, dest t.NodeID, msg *messagepb.Message) {
	select {
	case ft.Buffers[source][dest] <- events.ListOf(
		events.MessageReceived(t.ModuleID(msg.DestModule), source, msg),
	):
	default:
		fmt.Printf("Warning: Dropping message %T from %s to %s\n", msg.Type, source, dest)
	}
}

func (ft *FakeTransport) Link(source t.NodeID) (net.Transport, error) {
	return &FakeLink{
		Source:        source,
		FakeTransport: ft,
		DoneC:         make(chan struct{}),
	}, nil
}

func (ft *FakeTransport) RecvC(dest t.NodeID) <-chan *events.EventList {
	return ft.NodeSinks[dest]
}

func (fl *FakeLink) Start() error {
	return nil
}

func (fl *FakeLink) UpdateConnections(ctx context.Context, m map[t.NodeID]t.NodeAddress) {
	// TODO: implement UpdateConnections in FakeTransport if necessary
	panic("not implemented")
}

func (fl *FakeLink) Connect(ctx context.Context) {
	sourceBuffers := fl.FakeTransport.Buffers[fl.Source]

	for destID, buffer := range sourceBuffers {
		if fl.Source == destID {
			continue
		}

		go func(destID t.NodeID, buffer chan *events.EventList) {
			for {
				select {
				case msg := <-buffer:
					select {
					case fl.FakeTransport.NodeSinks[destID] <- msg:
					case <-ctx.Done():
						return
					case <-fl.DoneC:
						return
					}
				case <-ctx.Done():
					return
				case <-fl.DoneC:
					return
				}
			}
		}(destID, buffer)
	}
}

func (fl *FakeLink) Stop() {
	close(fl.DoneC)
}
