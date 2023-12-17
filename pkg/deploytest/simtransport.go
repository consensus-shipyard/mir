// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package deploytest

import (
	"context"
	"fmt"
	"time"

	"github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/stdtypes"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	messagepbtypes "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	"github.com/filecoin-project/mir/pkg/pb/transportpb"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/testsim"
	"github.com/filecoin-project/mir/pkg/util/libp2p"
)

type MessageDelayFn func(from, to stdtypes.NodeID) time.Duration

type SimTransport struct {
	*Simulation
	delayFn       MessageDelayFn
	nodes         map[stdtypes.NodeID]*simTransportModule
	nodeIDsWeight map[stdtypes.NodeID]types.VoteWeight
}

func NewSimTransport(s *Simulation, nodeIDsWeight map[stdtypes.NodeID]types.VoteWeight, delayFn MessageDelayFn) *SimTransport {
	st := &SimTransport{
		Simulation:    s,
		delayFn:       delayFn,
		nodes:         make(map[stdtypes.NodeID]*simTransportModule, len(nodeIDsWeight)),
		nodeIDsWeight: nodeIDsWeight,
	}

	for id := range nodeIDsWeight {
		st.nodes[id] = newModule(st, id, s.Node(id))
	}

	return st
}

func (st *SimTransport) Link(source stdtypes.NodeID) (net.Transport, error) {
	return st.nodes[source], nil
}

func (st *SimTransport) Membership() *trantorpbtypes.Membership {
	membership := &trantorpbtypes.Membership{make(map[stdtypes.NodeID]*trantorpbtypes.NodeIdentity)} // nolint:govet

	// Dummy addresses. Never actually used.
	for nID := range st.nodes {
		membership.Nodes[nID] = &trantorpbtypes.NodeIdentity{ // nolint:govet
			nID,
			libp2p.NewDummyHostAddr(0, 0).String(),
			nil,
			st.nodeIDsWeight[nID],
		}
	}

	return membership
}

func (st *SimTransport) Close() {}

type simTransportModule struct {
	*SimTransport
	*SimNode
	id       stdtypes.NodeID
	outChan  chan *stdtypes.EventList
	simChan  *testsim.Chan
	stopChan chan struct{}
}

func newModule(t *SimTransport, id stdtypes.NodeID, node *SimNode) *simTransportModule {
	return &simTransportModule{
		SimTransport: t,
		SimNode:      node,
		id:           id,
		outChan:      make(chan *stdtypes.EventList, 1),
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

func (m *simTransportModule) Send(dest stdtypes.NodeID, msg *messagepb.Message) error {
	m.sendMessage(msg, dest)
	return nil
}

func (m *simTransportModule) CloseOldConnections(_ *trantorpbtypes.Membership) {
}

func (m *simTransportModule) Connect(_ *trantorpbtypes.Membership) {
	go m.handleOutChan(m.SimTransport.Simulation.Spawn())
}

// WaitFor returns immediately, since the simulated transport does not need to wait for anything.
func (m *simTransportModule) WaitFor(_ int) error {
	return nil
}

func (m *simTransportModule) ApplyEvents(ctx context.Context, eventList *stdtypes.EventList) error {
	_, err := modules.ApplyEventsSequentially(eventList, func(e stdtypes.Event) (*stdtypes.EventList, error) {
		return stdtypes.EmptyList(), m.applyEvent(ctx, e)
	})
	return err
}

func (m *simTransportModule) applyEvent(ctx context.Context, event stdtypes.Event) error {

	// We only support proto events.
	pbevent, ok := event.(*eventpb.Event)
	if !ok {
		return es.Errorf("The simulation timer module only supports proto events, received %T", event)
	}

	switch e := pbevent.Type.(type) {
	case *eventpb.Event_Init:
		// do nothing
	case *eventpb.Event_Transport:
		switch e := e.Transport.Type.(type) {
		case *transportpb.Event_SendMessage:
			targets := stdtypes.NodeIDSlice(e.SendMessage.Destinations)
			m.multicastMessage(ctx, e.SendMessage.Msg, targets)
		default:
			return es.Errorf("unexpected transport event type: %T", e)
		}
	default:
		return es.Errorf("unexpected type of Net event: %T", e)
	}

	return nil
}

func (m *simTransportModule) multicastMessage(_ context.Context, msg *messagepb.Message, targets []stdtypes.NodeID) {
	for _, target := range targets {
		m.sendMessage(msg, target)
	}
}

func (m *simTransportModule) sendMessage(msg *messagepb.Message, target stdtypes.NodeID) {
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

func (m *simTransportModule) EventsOut() <-chan *stdtypes.EventList {
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

		destModule := stdtypes.ModuleID(msg.message.DestModule)
		eventList := stdtypes.ListOf(transportpbevents.MessageReceived(
			destModule,
			msg.from,
			messagepbtypes.MessageFromPb(msg.message),
		).Pb())

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
	from    stdtypes.NodeID
	to      stdtypes.NodeID
	message *messagepb.Message
}
