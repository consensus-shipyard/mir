package contextstore

// ContextStore can be used to store arbitrary data under an automatically deterministically generated unique id.
type ContextStore[T any] interface {
	// Store stores the given data in the ContextStore and returns a unique id.
	// The data can be later recovered or disposed of using this id.
	Store(t T) ItemID

	// Recover returns the data stored under the provided id.
	// Note that the data will continue to exist in the ContextStore.
	// In order to dispose of the data, call s.Dispose(id) or s.RecoverAndDispose(id).
	Recover(id ItemID) T

	// Dispose removes the data from the ContextStore.
	Dispose(id ItemID)

	// RecoverAndDispose returns the data stored under the provided id and removes it from the ContextStore.
	RecoverAndDispose(id ItemID) T
}

// ItemID is used to uniquely identify entries of the ContextStore.
type ItemID uint64

// Pb returns the protobuf representation of ItemID.
func (i ItemID) Pb() uint64 {
	return uint64(i)
}
