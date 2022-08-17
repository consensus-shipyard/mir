package types

import "strings"

const Separator = "/"

// TODO: Mention in the documentation that the Separator is a special sequence
//       that must not be contained in a module ID.
//       Technically, the only constraint is that no two modules whose IDs share the same prefix
//       up to the first separator can coexist in a Node.
//       However, we might still want to reserve the separator for future, more elaborate,
//       native support for structured modules.

// ModuleID represents an identifier of a module.
// The intention is for it to correspond to a path in the module hierarchy.
// However, technically, the Mir Node only cares for the ID's prefix up to the first separator and ignores the rest.
// The rest of the ID can be used for any module-specific purposes.
type ModuleID string

// Pb converts a ModuleID to a type used in a Protobuf message.
func (mid ModuleID) Pb() string {
	return string(mid)
}

// Top returns the ID of the top-level module of the path, stripped of the IDs of the submodules.
func (mid ModuleID) Top() ModuleID {
	top, _, _ := strings.Cut(string(mid), Separator)
	return ModuleID(top)
}

// Sub returns the identifier of a submodule within the top-level module, stripped of the top-level module identifier.
func (mid ModuleID) Sub() ModuleID {
	_, sub, _ := strings.Cut(string(mid), Separator)
	return ModuleID(sub)
}

// Then combines the module ID with a relative path to its submodule in a single module ID.
func (mid ModuleID) Then(submodule ModuleID) ModuleID {
	return ModuleID(string(mid) + Separator + string(submodule))
}
