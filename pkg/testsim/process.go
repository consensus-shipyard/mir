// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package testsim

import (
	"sync"
	"time"
)

// Process represents a context of execution in the simulation.
type Process struct {
	*Runtime
	lock     sync.Mutex
	active   bool
	killed   bool
	killChan chan struct{}
}

// newProcess creates a new active process within the runtime.
func newProcess(r *Runtime) *Process {
	p := &Process{
		Runtime:  r,
		killChan: make(chan struct{}),
	}
	p.activateLocked()

	return p
}

// Fork creates a new active process.
func (p *Process) Fork() *Process { return p.Runtime.Spawn() }

// Yield suspends and resumes the execution in a separate simulation
// step. It returns false in case the process was killed while
// waiting.
func (p *Process) Yield() (ok bool) { return p.Delay(0) }

// Delay suspends the execution and resumes it after d amount of
// simulated time. It returns false in case the process was killed
// while waiting.
func (p *Process) Delay(d time.Duration) (ok bool) {
	done := make(chan struct{})
	e := p.Runtime.scheduleAfter(d, func() {
		p.activate()
		close(done)
	})
	p.deactivate()

	select {
	case <-done:
		return true
	case <-p.killChan:
		p.Runtime.unschedule(e)
		return false
	case <-p.Runtime.stopChan:
		return false
	}
}

// Send attempts to send the value v to the channel c. It either
// resumes the process waiting in Recv operation on the channel, or
// blocks until another process attempts to receive from the channel.
// It returns false in case the process was killed while waiting.
func (p *Process) Send(c *Chan, v any) (ok bool) {
	return c.send(p, v)
}

// Recv attempts to receive a value from the channel. It either
// resumes the process waiting in Send operation on the channel, or
// blocks until another process attempts to send to the channel. It
// returns the received value and true once the operation is complete.
// It returns nil and false if the process was killed while waiting.
func (p *Process) Recv(c *Chan) (v any, ok bool) {
	return c.recv(p)
}

// Exit terminates the process normally.
func (p *Process) Exit() {
	p.deactivate()
}

// Kill immediately resumes and terminates the process.
func (p *Process) Kill() {
	p.lock.Lock()
	p.deactivateLocked()
	p.killed = true
	p.lock.Unlock()

	close(p.killChan)
}

func (p *Process) activate() bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.activateLocked()
}
func (p *Process) deactivate() bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.deactivateLocked()
}

func (p *Process) activateLocked() bool {
	if !p.active && !p.killed {
		p.active = true
		p.Runtime.addActiveProcess()
		return true
	}
	return false
}

func (p *Process) deactivateLocked() bool {
	if p.active {
		p.active = false
		p.Runtime.rmActiveProcess()
		return true
	}
	return false
}
