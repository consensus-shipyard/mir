package contextstore

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContextStore_SimpleTest(t *testing.T) {
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
}
