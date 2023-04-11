// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package deploytest

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/testsim"
	t "github.com/filecoin-project/mir/pkg/types"
)

// Simulation represents a test deployment in the simulation runtime.
type Simulation struct {
	*testsim.Runtime
	nodes map[t.NodeID]*SimNode
}

// EventDelayFn defines a function to provide event processing delay.
type EventDelayFn func(e *eventpb.Event) time.Duration

func NewSimulation(rnd *rand.Rand, nodeIDs []t.NodeID, delayFn EventDelayFn) *Simulation {
	s := &Simulation{
		Runtime: testsim.NewRuntime(rnd),
		nodes:   make(map[t.NodeID]*SimNode, len(nodeIDs)),
	}

	for _, id := range nodeIDs {
		s.nodes[id] = newNode(s, id, delayFn)
	}

	return s
}

func (s *Simulation) Node(id t.NodeID) *SimNode {
	return s.nodes[id]
}

// SimNode represents a Mir node deployed in the simulation runtime.
type SimNode struct {
	*Simulation
	id          t.NodeID
	delayFn     EventDelayFn
	moduleChans map[t.ModuleID]*testsim.Chan
}

func newNode(s *Simulation, id t.NodeID, delayFn EventDelayFn) *SimNode {
	n := &SimNode{
		Simulation:  s,
		id:          id,
		delayFn:     delayFn,
		moduleChans: make(map[t.ModuleID]*testsim.Chan),
	}
	return n
}

// SendEvents notifies simulation about the list of emitted events on
// behalf of the given process.
func (n *SimNode) SendEvents(proc *testsim.Process, eventList *events.EventList) {
	moduleIDs := make([]t.ModuleID, 0)
	eventsMap := make(map[t.ModuleID]*events.EventList)

	it := eventList.Iterator()
	for e := it.Next(); e != nil; e = it.Next() {
		m := t.ModuleID(e.DestModule).Top()
		if eventsMap[m] == nil {
			eventsMap[m] = events.EmptyList()
			moduleIDs = append(moduleIDs, m)
		}
		eventsMap[m].PushBack(e)
	}

	n.Simulation.Rand.Shuffle(len(moduleIDs), func(i, j int) {
		moduleIDs[i], moduleIDs[j] = moduleIDs[j], moduleIDs[i]
	})

	for _, m := range moduleIDs {
		ch, ok := n.moduleChans[m.Top()]
		if !ok {
			panic(fmt.Sprintf("destination module does not exist: %v", m.Top()))
		}
		proc.Send(ch, eventsMap[m.Top()])
		proc.Yield() // wait until the receiver blocks
	}
}

func (n *SimNode) recvEvents(proc *testsim.Process, simChan *testsim.Chan) (eventList *events.EventList, ok bool) {
	v, ok := proc.Recv(simChan)
	if !ok {
		return nil, false
	}
	return v.(*events.EventList), true
}

// WrapModules wraps the modules to be used in simulation. Mir nodes
// in the simulation deployment should be given the wrapped modules.
func (n *SimNode) WrapModules(mods modules.Modules) modules.Modules {
	wrapped := make(modules.Modules, len(mods))
	for k, v := range mods {
		wrapped[k] = n.WrapModule(k, v)
	}
	return wrapped
}

// WrapModule wraps the module to be used in simulation.
func (n *SimNode) WrapModule(id t.ModuleID, m modules.Module) modules.Module {
	moduleChan := testsim.NewChan()
	n.moduleChans[id.Top()] = moduleChan

	switch m := m.(type) {
	case modules.ActiveModule:
		return n.wrapActive(m, moduleChan)
	case modules.PassiveModule:
		return n.wrapPassive(m, moduleChan)
	default:
		panic(fmt.Sprintf("Unexpected module type: %v %T", m, m))
	}
}

// Start initiates simulation with the events from the write-ahead log
// on behalf of the given process. To be called concurrently with
// mir.Node.Run().
func (n *SimNode) Start(proc *testsim.Process) {
	initEvents := events.EmptyList()
	for m := range n.moduleChans {
		initEvents.PushBack(events.Init(m))
	}
	n.SendEvents(proc, initEvents)
}

type applyEventsFn func(ctx context.Context, eventList *events.EventList) (*events.EventList, error)

type eventsIn struct {
	ctx       context.Context
	eventList *events.EventList
}

type eventsOut struct {
	eventList *events.EventList
	err       error
}

type simModule struct {
	*SimNode
	inChan  chan eventsIn
	outChan chan eventsOut
	simChan *testsim.Chan
	wg      sync.WaitGroup
}

func newSimModule(n *SimNode, m modules.Module, simChan *testsim.Chan) *simModule {
	var applyFn applyEventsFn
	switch m := m.(type) {
	case modules.PassiveModule:
		applyFn = func(_ context.Context, eventList *events.EventList) (*events.EventList, error) {
			return m.ApplyEvents(eventList)
		}
	case modules.ActiveModule:
		applyFn = func(ctx context.Context, eventList *events.EventList) (*events.EventList, error) {
			return events.EmptyList(), m.ApplyEvents(ctx, eventList)
		}
	default:
		panic(fmt.Sprintf("Unexpected module type: %v %T", m, m))
	}

	sm := &simModule{
		SimNode: n,
		inChan:  make(chan eventsIn, 1),
		outChan: make(chan eventsOut, 1),
		simChan: simChan,
	}

	go sm.run(n.Spawn(), applyFn)

	return sm
}

func (m *simModule) run(proc *testsim.Process, applyFn applyEventsFn) {
	defer m.wg.Done()

	origEvents := events.EmptyList()
	for {
		if origEvents.Len() == 0 {
			newOrigEvents, ok := m.SimNode.recvEvents(proc, m.simChan)
			if !ok {
				return
			}
			origEvents.PushBackList(newOrigEvents)
		}

		in := <-m.inChan

		for origEvents.Len() < in.eventList.Len() {
			newOrigEvents, ok := m.SimNode.recvEvents(proc, m.simChan)
			if !ok {
				return
			}
			origEvents.PushBackList(newOrigEvents)
		}

		it := in.eventList.Iterator()
		for e := it.Next(); e != nil; e = it.Next() {
			if !proc.Delay(m.SimNode.delayFn(e)) {
				return
			}
		}

		var out eventsOut
		out.eventList, out.err = applyFn(in.ctx, in.eventList)

		if out.err == nil {
			it := origEvents.Iterator()

			// First, collect from the original event list
			// follow-ups for each event in the event list
			// passed by the Mir node to ApplyEvents
			followUps := events.EmptyList()
			for i := 0; i < in.eventList.Len(); i++ {
				followUps.PushBackSlice(it.Next().Next)
			}

			// Then keep only the rest of the events in
			// the original event list
			origEvents = events.EmptyList()
			for e := it.Next(); e != nil; e = it.Next() {
				origEvents.PushBack(e)
			}

			// After that, append the event list returned
			// by ApplyEvents to the follow-up event list
			followUps.PushBackList(out.eventList)

			// Send the events in a new concurrent process
			// because some of them may target this module
			go func(proc *testsim.Process) {
				m.outChan <- out
				m.SimNode.SendEvents(proc, followUps)
				proc.Exit()
			}(proc.Fork())

			proc.Yield() // wait until any other process blocks
		} else {
			panic(out.err)
		}
	}
}

func (m *simModule) applyEvents(ctx context.Context, eventList *events.EventList) (eventsOut *events.EventList, err error) {
	m.inChan <- eventsIn{ctx, eventList}
	out := <-m.outChan
	return out.eventList, out.err
}

type passiveSimModule struct {
	modules.PassiveModule
	*simModule
}

func (n *SimNode) wrapPassive(m modules.PassiveModule, simChan *testsim.Chan) modules.PassiveModule {
	return &passiveSimModule{m, newSimModule(n, m, simChan)}
}

func (m *passiveSimModule) ApplyEvents(eventList *events.EventList) (*events.EventList, error) {
	return m.applyEvents(context.Background(), eventList)
}

type activeSimModule struct {
	modules.ActiveModule
	*simModule
}

func (n *SimNode) wrapActive(m modules.ActiveModule, simChan *testsim.Chan) modules.ActiveModule {
	return &activeSimModule{m, newSimModule(n, m, simChan)}
}

func (m *activeSimModule) ApplyEvents(ctx context.Context, eventList *events.EventList) error {
	_, err := m.applyEvents(ctx, eventList)
	return err
}
