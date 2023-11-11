package events

import t "github.com/filecoin-project/mir/pkg/types"

type Event interface {

	// Src returns the module that emitted the event.
	// While this information is not always necessary for the system operation,
	// it is useful for analyzing event traces and debugging.
	Src() t.ModuleID

	// Dest returns the destination module of the event.
	Dest() t.ModuleID

	// ToBytes returns a serialized representation of the event
	// as a slice of bytes from which the event can be reconstructed.
	// Note that ToBytes does not necessarily guarantee the output to be deterministic.
	// Even multiple subsequent calls to ToBytes on the same event object might return different byte slices.
	// If an error occurs during serialization, ToBytes returns a nil byte slice and a non-nil error.
	ToBytes() ([]byte, error)

	// ToString returns a human-readable representation of the event.
	// While not used by the runtime itself, it can be used by associated tools.
	// Conventionally, this representation is JSON.
	ToString() string
}

// Note: The method names ToBytes() and ToString() were chosen instead of the more elegant Bytes() and String(),
// because the String() method is a special one that would override Go's default string representation.
// That would make it impossible for the String() (now renamed to ToString()) to access Go's default string
// representation, as it would enter into recursion.
