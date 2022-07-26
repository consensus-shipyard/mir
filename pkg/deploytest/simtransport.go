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
)

type MessageDelayFn func(from, to t.NodeID) time.Duration

type SimTransport struct {
	*Simulation
	delayFn MessageDelayFn
	nodes   map[t.NodeID]*simTransportModule
}

func NewSimTransport(s *Simulation, nodeIDs []t.NodeID, delayFn MessageDelayFn) *SimTransport {
	t := &SimTransport{
		Simulation: s,
		delayFn:    delayFn,
		nodes:      make(map[t.NodeID]*simTransportModule, len(nodeIDs)),
	}

	for _, id := range nodeIDs {
		t.nodes[id] = newModule(t, id, s.Node(id))
	}

	return t
}

func (t *SimTransport) Link(source t.NodeID) net.Transport {
	return t.nodes[source]
}

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
	m.sendMessage(context.Background(), msg, dest)
	return nil
}

func (m *simTransportModule) Connect(ctx context.Context) {
	go m.handleOutChan(ctx, m.SimTransport.Simulation.Spawn())

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

func (m *simTransportModule) multicastMessage(ctx context.Context, msg *messagepb.Message, targets []t.NodeID) {
	for _, target := range targets {
		m.sendMessage(ctx, msg, target)
	}
}

func (m *simTransportModule) sendMessage(ctx context.Context, msg *messagepb.Message, target t.NodeID) {
	proc := m.SimTransport.Simulation.Spawn()

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
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

		destModule := m.SimTransport.nodes[target]
		proc.Send(destModule.simChan, message{m.id, target, msg})

		proc.Exit()
	}()
}

func (m *simTransportModule) EventsOut() <-chan *events.EventList {
	return m.outChan
}

func (m *simTransportModule) handleOutChan(ctx context.Context, proc *testsim.Process) {
	go func() {
		select {
		case <-m.stopChan:
		case <-ctx.Done():
		}
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
