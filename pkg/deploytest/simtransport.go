// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package deploytest

import (
	"context"
	"fmt"
	"time"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/testsim"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/libp2p"
)

type MessageDelayFn func(from, to t.NodeID) time.Duration

type SimTransport struct {
	*Simulation
	delayFn MessageDelayFn
	nodes   map[t.NodeID]*simTransportModule
}

func NewSimTransport(s *Simulation, nodeIDs []t.NodeID, delayFn MessageDelayFn) *SimTransport {
	st := &SimTransport{
		Simulation: s,
		delayFn:    delayFn,
		nodes:      make(map[t.NodeID]*simTransportModule, len(nodeIDs)),
	}

	for _, id := range nodeIDs {
		st.nodes[id] = newModule(st, id, s.Node(id))
	}

	return st
}

func (st *SimTransport) Link(source t.NodeID) (net.Transport, error) {
	return st.nodes[source], nil
}

func (st *SimTransport) Nodes() map[t.NodeID]t.NodeAddress {
	membership := make(map[t.NodeID]t.NodeAddress)

	// Dummy addresses. Never actually used.
	for nID := range st.nodes {
		membership[nID] = libp2p.NewDummyHostAddr(0, 0)
	}

	return membership
}

func (st *SimTransport) Close() {}

type simTransportModule struct {
	*SimTransport
	*SimNode
	id       t.NodeID
	outChan  chan *events.EventList
	simChan  *testsim.Chan
	stopChan chan struct{}
}

func newModule(t *SimTransport, id t.NodeID, node *SimNode) *simTransportModule {
	return &simTransportModule{
		SimTransport: t,
		SimNode:      node,
		id:           id,
		outChan:      make(chan *events.EventList, 1),
		simChan:      testsim.NewChan(),
		stopChan:     make(chan struct{}),
	}
}

func (m *simTransportModule) ImplementsModule() {}

func (m *simTransportModule) Start() error {
	return nil
}

func (m *simTransportModule) Stop() {
	close(m.stopChan)
}

func (m *simTransportModule) Send(dest t.NodeID, msg *messagepb.Message) error {
	m.sendMessage(msg, dest)
	return nil
}

func (m *simTransportModule) CloseOldConnections(_ map[t.NodeID]t.NodeAddress) {
}

func (m *simTransportModule) Connect(_ map[t.NodeID]t.NodeAddress) {
	go m.handleOutChan(m.SimTransport.Simulation.Spawn())
}

// WaitFor returns immediately, since the simulated transport does not need to wait for anything.
func (m *simTransportModule) WaitFor(_ int) {
}

func (m *simTransportModule) ApplyEvents(ctx context.Context, eventList *events.EventList) error {
	_, err := modules.ApplyEventsSequentially(eventList, func(e *eventpb.Event) (*events.EventList, error) {
		return events.EmptyList(), m.applyEvent(ctx, e)
	})
	return err
}

func (m *simTransportModule) applyEvent(ctx context.Context, e *eventpb.Event) error {
	switch e := e.Type.(type) {
	case *eventpb.Event_Init:
		// do nothing
	case *eventpb.Event_SendMessage:
		targets := t.NodeIDSlice(e.SendMessage.Destinations)
		m.multicastMessage(ctx, e.SendMessage.Msg, targets)
	default:
		return fmt.Errorf("unexpected type of Net event: %T", e)
	}

	return nil
}

func (m *simTransportModule) multicastMessage(_ context.Context, msg *messagepb.Message, targets []t.NodeID) {
	for _, target := range targets {
		m.sendMessage(msg, target)
	}
}

func (m *simTransportModule) sendMessage(msg *messagepb.Message, target t.NodeID) {
	proc := m.SimTransport.Simulation.Spawn()

	done := make(chan struct{})
	go func() {
		select {
		case <-m.stopChan:
		case <-done:
			return
		}
		proc.Kill()
	}()

	d := m.SimTransport.delayFn(m.id, target)
	go func() {
		defer close(done)

		if !proc.Delay(d) {
			return
		}

		destModule, ok := m.SimTransport.nodes[target]
		if !ok {
			panic(fmt.Sprintf("Destination node does not exist: %v", target))
		}
		proc.Send(destModule.simChan, message{m.id, target, msg})

		proc.Exit()
	}()
}

func (m *simTransportModule) EventsOut() <-chan *events.EventList {
	return m.outChan
}

func (m *simTransportModule) handleOutChan(proc *testsim.Process) {
	go func() {
		<-m.stopChan
		proc.Kill()
	}()

	for {
		v, ok := proc.Recv(m.simChan)
		if !ok {
			return
		}
		msg := v.(message)

		destModule := t.ModuleID(msg.message.DestModule)
		eventList := events.ListOf(events.MessageReceived(destModule, msg.from, msg.message))

		select {
		case eventsOut := <-m.outChan:
			eventsOut.PushBackList(eventList)
			m.outChan <- eventsOut
		default:
			m.outChan <- eventList
		}
		m.SimNode.SendEvents(proc, eventList)
	}
}

type message struct {
	from    t.NodeID
	to      t.NodeID
	message *messagepb.Message
}
