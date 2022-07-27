// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package testsim

import (
	"sync"
)

// Chan represents a channel for synchronization and communication
// between processes in the simulated runtime.
type Chan struct {
	lock sync.Mutex

	recvProcs []*Process
	recvChans []chan any

	sendProcs []*Process
	sendChans []chan any
}

// NewChan creates a new channel in the simulation runtime.
func NewChan() *Chan {
	return &Chan{}
}

// send performs the Send operation by the process on the channel.
func (c *Chan) send(p *Process, v any) (ok bool) {
	var ch chan any

	c.lock.Lock()
	if len(c.recvProcs) > 0 {
		ch = c.recvChans[0]
		c.recvProcs[0].activate()
		c.recvProcs[0], c.recvProcs = nil, c.recvProcs[1:]
		c.recvChans[0], c.recvChans = nil, c.recvChans[1:]
	} else {
		ch = make(chan any)
		c.sendChans = append(c.sendChans, ch)
		c.sendProcs = append(c.sendProcs, p)
		p.deactivate()
	}
	c.lock.Unlock()

	select {
	case ch <- v:
		return true
	case <-p.killChan:
		return false
	case <-p.Runtime.stopChan:
		return false
	}
}

// recv performs the Recv operation by the process on the channel.
func (c *Chan) recv(p *Process) (v any, ok bool) {
	var ch chan any

	c.lock.Lock()
	if len(c.sendProcs) > 0 {
		ch = c.sendChans[0]
		c.sendProcs[0].activate()
		c.sendProcs[0], c.sendProcs = nil, c.sendProcs[1:]
		c.sendChans[0], c.sendChans = nil, c.sendChans[1:]
	} else {
		ch = make(chan any)
		c.recvChans = append(c.recvChans, ch)
		c.recvProcs = append(c.recvProcs, p)
		p.deactivate()
	}
	c.lock.Unlock()

	select {
	case v = <-ch:
		return v, true
	case <-p.killChan:
		return nil, false
	case <-p.Runtime.stopChan:
		return nil, false
	}
}
