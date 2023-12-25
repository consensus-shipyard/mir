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

	"github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	messagepbtypes "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	transportpbtypes "github.com/filecoin-project/mir/pkg/pb/transportpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/util/libp2p"
)

type FakeLink struct {
	FakeTransport *FakeTransport
	Source        stdtypes.NodeID
	DoneC         chan struct{}
	wg            sync.WaitGroup
}

func (fl *FakeLink) ApplyEvents(
	ctx context.Context,
	eventList *stdtypes.EventList,
) error {
	iter := eventList.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {

		// Ignore Init event.
		_, ok := event.(*stdevents.Init)
		if ok {
			return nil
		}

		// We only support proto events.
		pbevent, ok := event.(*eventpb.Event)
		if !ok {
			return es.Errorf("Fake transport only supports proto events, received %T", event)
		}

		switch e := pbevent.Type.(type) {
		case *eventpb.Event_Transport:
			switch e := transportpbtypes.EventFromPb(e.Transport).Type.(type) {
			case *transportpbtypes.Event_SendMessage:
				for _, destID := range e.SendMessage.Destinations {
					if destID == fl.Source {
						// Send message to myself bypassing the network.

						receivedEvent := transportpbevents.MessageReceived(
							e.SendMessage.Msg.DestModule,
							fl.Source,
							e.SendMessage.Msg,
						)
						eventsOut := fl.FakeTransport.NodeSinks[fl.Source]
						go func() {
							select {
							case eventsOut <- stdtypes.ListOf(receivedEvent.Pb()):
							case <-ctx.Done():
							}
						}()
					} else {
						// Send message to another node.
						if err := fl.Send(destID, e.SendMessage.Msg.Pb()); err != nil {
							fl.FakeTransport.logger.Log(logging.LevelWarn, "failed to send a message", "err", err)
						}
					}
				}
			default:
				return es.Errorf("unexpected transport event type: %T", e)
			}
		default:
			return es.Errorf("unexpected type of Net event: %T", pbevent.Type)
		}
	}

	return nil
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (fl *FakeLink) ImplementsModule() {}

func (fl *FakeLink) Send(dest stdtypes.NodeID, msg *messagepb.Message) error {
	fl.FakeTransport.Send(fl.Source, dest, msg)
	return nil
}

func (fl *FakeLink) EventsOut() <-chan *stdtypes.EventList {
	return fl.FakeTransport.NodeSinks[fl.Source]
}

var _ LocalTransportLayer = &FakeTransport{}

type FakeTransport struct {
	// Buffers is source x dest
	Buffers       map[stdtypes.NodeID]map[stdtypes.NodeID]chan *stdtypes.EventList
	NodeSinks     map[stdtypes.NodeID]chan *stdtypes.EventList
	logger        logging.Logger
	nodeIDsWeight map[stdtypes.NodeID]types.VoteWeight
}

func NewFakeTransport(nodeIDsWeight map[stdtypes.NodeID]types.VoteWeight) *FakeTransport {
	buffers := make(map[stdtypes.NodeID]map[stdtypes.NodeID]chan *stdtypes.EventList)
	nodeSinks := make(map[stdtypes.NodeID]chan *stdtypes.EventList)
	for sourceID := range nodeIDsWeight {
		buffers[sourceID] = make(map[stdtypes.NodeID]chan *stdtypes.EventList)
		for destID := range nodeIDsWeight {
			if sourceID == destID {
				continue
			}
			buffers[sourceID][destID] = make(chan *stdtypes.EventList, 10000)
		}
		nodeSinks[sourceID] = make(chan *stdtypes.EventList)
	}

	return &FakeTransport{
		Buffers:       buffers,
		NodeSinks:     nodeSinks,
		logger:        logging.ConsoleErrorLogger,
		nodeIDsWeight: nodeIDsWeight,
	}
}

func (ft *FakeTransport) Send(source, dest stdtypes.NodeID, msg *messagepb.Message) {
	select {
	case ft.Buffers[source][dest] <- stdtypes.ListOf(
		transportpbevents.MessageReceived(stdtypes.ModuleID(msg.DestModule), source, messagepbtypes.MessageFromPb(msg)).Pb(),
	):
	default:
		fmt.Printf("Warning: Dropping message %T from %s to %s\n", msg.Type, source, dest)
	}
}

func (ft *FakeTransport) Link(source stdtypes.NodeID) (net.Transport, error) {
	return &FakeLink{
		Source:        source,
		FakeTransport: ft,
		DoneC:         make(chan struct{}),
	}, nil
}

func (ft *FakeTransport) Membership() *trantorpbtypes.Membership {
	membership := &trantorpbtypes.Membership{make(map[stdtypes.NodeID]*trantorpbtypes.NodeIdentity)} // nolint:govet

	// Dummy addresses. Never actually used.
	for nID := range ft.Buffers {
		membership.Nodes[nID] = &trantorpbtypes.NodeIdentity{ // nolint:govet
			nID,
			libp2p.NewDummyHostAddr(0,
				0).String(),
			nil,
			ft.nodeIDsWeight[nID],
		}
	}

	return membership
}

func (ft *FakeTransport) Close() {}

func (fl *FakeLink) CloseOldConnections(_ *trantorpbtypes.Membership) {}

func (ft *FakeTransport) RecvC(dest stdtypes.NodeID) <-chan *stdtypes.EventList {
	return ft.NodeSinks[dest]
}

func (fl *FakeLink) Start() error {
	return nil
}

func (fl *FakeLink) Connect(_ *trantorpbtypes.Membership) {
	sourceBuffers := fl.FakeTransport.Buffers[fl.Source]

	fl.wg.Add(len(sourceBuffers))

	for destID, buffer := range sourceBuffers {
		if fl.Source == destID {
			fl.wg.Done()
			continue
		}
		go func(destID stdtypes.NodeID, buffer chan *stdtypes.EventList) {
			defer fl.wg.Done()
			for {
				select {
				case msg := <-buffer:
					select {
					case fl.FakeTransport.NodeSinks[destID] <- msg:
					case <-fl.DoneC:
						return
					}
				case <-fl.DoneC:
					return
				}
			}
		}(destID, buffer)
	}
}

// WaitFor returns immediately.
// It does not need to wait for anything, since the Connect() function already waits for all the connections.
func (fl *FakeLink) WaitFor(_ int) error {
	return nil
}

func (fl *FakeLink) Stop() {
	close(fl.DoneC)
	fl.wg.Wait()
}
