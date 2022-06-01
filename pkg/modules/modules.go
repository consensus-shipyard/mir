/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package modules provides interfaces of modules that serve as building blocks of a Node.
// Implementations of those interfaces are not contained by this package
// and are expected to be provided by other packages.
package modules

import (
	"fmt"
	"github.com/filecoin-project/mir/pkg/clients"
	"github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/reqstore"
	"github.com/filecoin-project/mir/pkg/timer"
)

// The Modules structs groups the modules a Node consists of.
type Modules struct {
	Hasher        PassiveModule // Computes hashes of requests and other data.
	Crypto        PassiveModule // Performs cryptographic operations (except for computing hashes)
	App           PassiveModule // Implements user application logic. The user is expected to provide this module.
	WAL           PassiveModule // Implements a persistent write-ahead log for the case of crashes and restarts.
	ClientTracker PassiveModule // Keeps the state related to clients and validates submitted requests.
	RequestStore  PassiveModule // Provides persistent storage for request data.
	Protocol      PassiveModule // Implements the logic of the distributed protocol.

	Net   ActiveModule // Sends messages produced by Mir through the network.
	Timer ActiveModule // Tracks real time (e.g. for timeouts) and injects events accordingly.

	PassiveModules map[string]PassiveModule
	ActiveModules  map[string]ActiveModule

	Interceptor eventlog.Interceptor // Intercepts and logs all internal _Events_ for debugging purposes.
}

// Defaults takes a Modules object (as a value, not a pointer to it) and returns a pointer to a new Modules object
// with default modules inserted in fields where no module has been specified.
func Defaults(m Modules) (*Modules, error) {
	if m.Net == nil {
		// TODO: Change this when a Net implementation exists.
		panic("no default Net implementation")
	}

	if m.Hasher == nil {
		m.Hasher = &crypto.SHA256Hasher{}
	}

	if m.App == nil {
		return nil, fmt.Errorf("no default App implementation")
	}

	if m.ClientTracker == nil {
		// TODO: Change this to the real default client tracker once implemented.
		//       Also, make the "iss" default protocol module more explicit.
		m.ClientTracker = clients.SigningTracker("iss", nil)
	}

	if m.RequestStore == nil {
		m.RequestStore = reqstore.NewVolatileRequestStore()
	}

	if m.Protocol == nil {
		// TODO: Use default protocol once implemented.
		return nil, fmt.Errorf("no default protocol implementation")
	}

	if m.Crypto == nil {
		// TODO: Use default crypto once implemented and tested.
		return nil, fmt.Errorf("no default crypto implementation")
	}

	if m.Timer == nil {
		m.Timer = &timer.Timer{}
	}

	// The Interceptor can stay nil, in which case Events will simply not be intercepted.

	// The WAL can stay nil, in which case no write-ahead log will be written
	// and the node will not be able to restart.

	return &m, nil
}
