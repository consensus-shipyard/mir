// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

// Package testsim implements a deterministic execution runtime driven
// by simulated logical time.
//
// The runtime controls execution by means of processes. The process
// represent a context of execution within the simulation runtime.
// Each process can be either active or inactive. Active processes are
// allowed to execute while inactive processes are blocked waiting for
// certain conditions to hold before they can continue execution.
//
// Processes can synchronize and exchange values through channels
// provided by the runtime. The channel represents an object which
// processes can send to or receive values from. An attempt to perform
// the send or receive operation on a channel deactivates the process
// and blocks waiting until another process performs the complementary
// operation on the channel. Once that happens, both processes become
// active and can continue execution, whereas the receiving process
// obtains the value supplied by the sending process. In case multiple
// processes get blocked trying to send/receive on the same channel,
// they complete the operations in FIFO order.
//
// Processes can also get blocked by waiting for a specified duration
// of simulated time to expire. Such process gets activated back once
// the simulation reaches that point of simulated time.
//
// The simulation proceeds in steps. Each step represents an instance
// of simulated time and lasts as long as there is an active process
// to execute. Execution within the same step is considered happening
// instantaneously in terms of simulated time. Once there is no more
// active process, the simulation can proceed to the next step that
// corresponds to the next action scheduled in the simulated time.
package testsim

import (
	"container/heap"
	"math/rand"
	"sync"
	"time"
)

// RandDuration returns a uniformly distributed pseudo-random duration
// in the range of [min, max].
func RandDuration(r *rand.Rand, min, max time.Duration) time.Duration {
	return min + time.Duration(r.Int63n(int64(max-min+1)))
}

// Runtime provides a mechanism to run a simulation with logical time.
type Runtime struct {
	*rand.Rand

	activeProcLock sync.Mutex
	nrActiveProcs  uint64
	resumeCond     sync.Cond

	now       int64
	queue     actionQueue
	queueLock sync.Mutex

	stopChan chan struct{}
}

// NewRuntime creates a new simulation runtime given a source of
// pseudo-random numbers.
func NewRuntime(rnd *rand.Rand) *Runtime {
	r := &Runtime{
		Rand:     rnd,
		stopChan: make(chan struct{}),
	}
	r.resumeCond = *sync.NewCond(&r.activeProcLock)

	return r
}

// Spawn creates a new active process.
func (r *Runtime) Spawn() *Process {
	return newProcess(r)
}

// Run executes the simulation until there is no more active process
// or scheduled action.
func (r *Runtime) Run() {
	r.waitQuiescence()
	r.queueLock.Lock()
	for r.queue.Len() > 0 {
		r.doStepLocked()
	}
	r.queueLock.Unlock()
}

// RunFor works the same as Run, but stops execution after expiration
// of the duration d of simulated time.
func (r *Runtime) RunFor(d time.Duration) {
	r.waitQuiescence()
	t := r.timeAfter(d)
	r.queueLock.Lock()
	for r.queue.Len() > 0 && r.queue.top().deadline <= t {
		r.doStepLocked()
	}
	r.queueLock.Unlock()
}

// Step executes the next scheduled action in the simulation and waits
// until there is no more active process. It returns true if there was
// an action to execute.
func (r *Runtime) Step() (ok bool) {
	r.waitQuiescence()
	r.queueLock.Lock()
	if r.queue.Len() > 0 {
		r.doStepLocked()
		ok = true
	}
	r.queueLock.Unlock()
	return
}

// Stop kills all processes and terminates the simulation.
func (r *Runtime) Stop() {
	close(r.stopChan)
}

// timeAfter calculates the simulated time after the duration d.
func (r *Runtime) timeAfter(d time.Duration) int64 {
	t := r.now + int64(d)
	if t < 0 {
		panic("Time overflow")
	}
	return t
}

// scheduleAfter arranges the function f to be called after expiration
// of the duration d of simulated time.
func (r *Runtime) scheduleAfter(d time.Duration, f func()) *action {
	return r.schedule(r.timeAfter(d), f)
}

// action represents a scheduled call to a function.
type action struct {
	index    int
	deadline int64
	fn       func()
}

// schedule arranges the function f to be called at simulated time t.
func (r *Runtime) schedule(t int64, f func()) *action {
	e := &action{
		deadline: t,
		fn:       f,
	}
	r.queueLock.Lock()
	heap.Push(&r.queue, e)
	r.queueLock.Unlock()
	return e
}

// unschedule cancels the supplied scheduled action.
func (r *Runtime) unschedule(e *action) {
	r.queueLock.Lock()
	if e.index >= 0 {
		heap.Remove(&r.queue, e.index)
		e.index = -1
	}
	r.queueLock.Unlock()
}

// doStepLocked executes the next scheduled action and then waits
// until there is no more active process.
func (r *Runtime) doStepLocked() {
	e := heap.Pop(&r.queue).(*action)
	r.now = e.deadline
	r.queueLock.Unlock()
	e.fn()
	r.waitQuiescence()
	r.queueLock.Lock()
}

// waitQuiescence waits until there is no more active process.
func (r *Runtime) waitQuiescence() {
	r.activeProcLock.Lock()
	for r.nrActiveProcs > 0 {
		r.resumeCond.Wait()
	}
	r.activeProcLock.Unlock()
}

// addActiveProcess notifies the runtime about process activation.
func (r *Runtime) addActiveProcess() {
	r.activeProcLock.Lock()
	r.nrActiveProcs++
	r.activeProcLock.Unlock()
}

// rmActiveProcess notifies the runtime about process deactivation.
func (r *Runtime) rmActiveProcess() {
	r.activeProcLock.Lock()
	r.nrActiveProcs--
	if r.nrActiveProcs == 0 {
		r.resumeCond.Broadcast()
	}
	r.activeProcLock.Unlock()
}
