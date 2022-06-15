/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package modules provides interfaces of modules that serve as building blocks of a Node.
// Implementations of those interfaces are not contained by this package
// and are expected to be provided by other packages.
package modules

import (
	"github.com/filecoin-project/mir/pkg/pb/statuspb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// Module generalizes the ActiveModule and PassiveModule types.
type Module interface {

	// ImplementsModule only serves the purpose of indicating that this is a Module and must not be called.
	ImplementsModule()

	// Status returns the current state of the module.
	// Mostly for debugging purposes.
	// TODO: This functionality is not yet implemented and all the Status() implementations are stubs. Fix that.
	Status() (s *statuspb.ProtocolStatus, err error)
}

// The Modules structs groups the modules a Node consists of.
type Modules map[t.ModuleID]Module
