package contextstore

// ContextStore can be used to store arbitrary data under an automatically deterministically generated unique id.
type ContextStore[T any] interface {
	Store(t T) ItemID
	Recover(id ItemID) T
	Dispose(id ItemID)
	RecoverAndDispose(id ItemID) T
}

// ItemID is used to uniquely identify entries of the ContextStore.
type ItemID uint64

// Pb returns the protobuf representation of ItemID.
func (i ItemID) Pb() uint64 {
	return uint64(i)
}
