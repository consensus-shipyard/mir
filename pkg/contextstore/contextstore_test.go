package contextstore

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSequentialContextStoreImpl_RecoverAndDispose(t *testing.T) {
	cs := NewSequentialContextStore[string]()
	helloID := cs.Store("Hello")
	worldID := cs.Store("World")

	assert.Equal(t, "World", cs.Recover(worldID))
	assert.Equal(t, "Hello", cs.Recover(helloID))

	cs.Dispose(worldID)
	assert.Panics(t, func() {
		cs.Recover(worldID)
	})

	assert.Equal(t, "Hello", cs.RecoverAndDispose(helloID))
	assert.Panics(t, func() {
		cs.RecoverAndDispose(helloID)
	})

	assert.NotPanics(t, func() {
		cs.Dispose(worldID)
	})

	assert.NotPanics(t, func() {
		cs.Dispose(helloID)
	})
}
