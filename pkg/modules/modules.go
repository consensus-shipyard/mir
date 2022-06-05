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

// Modules represents a collection of named modules used by the Node.
type Modules map[t.ModuleID]Module
