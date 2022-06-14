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
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/statuspb"
	"strconv"
	"sync"

	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// unsafeIDtoi calls strconv.Atoi, but doesn't check an error.
// It must be used in trusted deployment only.
func unsafeIDtoi(in interface{}) (out int) {
	switch id := in.(type) {
	case t.NodeID:
		out, _ = strconv.Atoi(string(id))
	case t.ClientID:
		out, _ = strconv.Atoi(string(id))
	}
	return
}

type FakeLink struct {
	modules.Module

	FakeTransport *FakeTransport
	Source        t.NodeID
}

func (fl *FakeLink) ApplyEvents(
	ctx context.Context,
	eventList *events.EventList,
) error {
	iter := eventList.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {

		switch e := event.Type.(type) {
		case *eventpb.Event_SendMessage:
			for _, destID := range e.SendMessage.Destinations {
				if t.NodeID(destID) == fl.Source {
					// Send message to myself bypassing the network.
					receivedEvent := events.MessageReceived(
						t.ModuleID(e.SendMessage.Msg.DestModule),
						fl.Source,
						e.SendMessage.Msg,
					)
					eventsOut := fl.FakeTransport.NodeSinks[unsafeIDtoi(fl.Source)]
					go func() {
						select {
						case eventsOut <- (&events.EventList{}).PushBack(receivedEvent):
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

func (fl *FakeLink) Status() (s *statuspb.ProtocolStatus, err error) {
	//TODO implement me
	panic("implement me")
}

func (fl *FakeLink) Send(dest t.NodeID, msg *messagepb.Message) error {
	fl.FakeTransport.Send(fl.Source, dest, msg)
	return nil
}

func (fl *FakeLink) EventsOut() <-chan *events.EventList {
	return fl.FakeTransport.NodeSinks[unsafeIDtoi(fl.Source)]
}

type FakeTransport struct {
	// Buffers is source x dest
	Buffers   [][]chan *events.EventList
	NodeSinks []chan *events.EventList
	WaitGroup sync.WaitGroup
	DoneC     chan struct{}
}

func NewFakeTransport(nodes int) *FakeTransport {
	buffers := make([][]chan *events.EventList, nodes)
	nodeSinks := make([]chan *events.EventList, nodes)
	for i := 0; i < nodes; i++ {
		buffers[i] = make([]chan *events.EventList, nodes)
		for j := 0; j < nodes; j++ {
			if i == j {
				continue
			}
			buffers[i][j] = make(chan *events.EventList, 10000)
		}
		nodeSinks[i] = make(chan *events.EventList)
	}

	return &FakeTransport{
		Buffers:   buffers,
		NodeSinks: nodeSinks,
		DoneC:     make(chan struct{}),
	}
}

func (ft *FakeTransport) Send(source, dest t.NodeID, msg *messagepb.Message) {
	select {
	case ft.Buffers[unsafeIDtoi(source)][unsafeIDtoi(dest)] <- (&events.EventList{}).PushBack(
		events.MessageReceived(t.ModuleID(msg.DestModule), source, msg),
	):
	default:
		fmt.Printf("Warning: Dropping message %T from %s to %s\n", msg.Type, source, dest)
	}
}

func (ft *FakeTransport) Link(source t.NodeID) *FakeLink {
	return &FakeLink{
		Source:        source,
		FakeTransport: ft,
	}
}

func (ft *FakeTransport) RecvC(dest t.NodeID) <-chan *events.EventList {
	return ft.NodeSinks[unsafeIDtoi(dest)]
}

func (ft *FakeTransport) Start() {
	for i, sourceBuffers := range ft.Buffers {
		for j, buffer := range sourceBuffers {
			if i == j {
				continue
			}

			ft.WaitGroup.Add(1)
			go func(i, j int, buffer chan *events.EventList) {
				// fmt.Printf("Starting drain thread from %d to %d\n", i, j)
				defer ft.WaitGroup.Done()
				for {
					select {
					case msg := <-buffer:
						// fmt.Printf("Sending message from %d to %d\n", i, j)
						select {
						case ft.NodeSinks[j] <- msg:
						case <-ft.DoneC:
							return
						}
					case <-ft.DoneC:
						return
					}
				}
			}(i, j, buffer)
		}
	}
}

func (ft *FakeTransport) Stop() {
	close(ft.DoneC)
	ft.WaitGroup.Wait()
}
