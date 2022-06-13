/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package modules provides interfaces of modules that serve as building blocks of a Node.
// Implementations of those interfaces are not contained by this package
// and are expected to be provided by other packages.
package modules

import (
	t "github.com/filecoin-project/mir/pkg/types"
)

// Module generalizes the ActiveModule and PassiveModule types.
type Module interface {
	ImplementsModule()
}

// The Modules structs groups the modules a Node consists of.
type Modules struct {
	Net           Net              // Sends messages produced by Mir through the network.
	WAL           WAL              // Implements a persistent write-ahead log for the case of crashes and restarts.
	ClientTracker ClientTracker    // Keeps the state related to clients and validates submitted requests.
	RequestStore  RequestStore     // Provides persistent storage for request data.
	Protocol      Protocol         // Implements the logic of the distributed protocol.
	Interceptor   EventInterceptor // Intercepts and logs all internal _Events_ for debugging purposes.
	Timer         Timer            // Tracks real time (e.g. for timeouts) and injects events accordingly.

	GenericModules map[t.ModuleID]Module
}
