package indexedlist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexedList_Append(t *testing.T) {
	il := New[string, int]()
	il.Append([]string{"a", "b", "c"}, []int{0, 1, 2})
	assert.Equal(t, 3, il.Len())

	iter := il.Iterator(0)
	key, val, ok := iter.Next()
	assert.Equal(t, "a", key)
	assert.Equal(t, 0, val)
	assert.True(t, ok)
	key, val, ok = iter.Next()
	assert.Equal(t, "b", key)
	assert.Equal(t, 1, val)
	assert.True(t, ok)
	key, val, ok = iter.Next()
	assert.Equal(t, "c", key)
	assert.Equal(t, 2, val)
	assert.True(t, ok)
	key, val, ok = iter.Next()
	assert.Equal(t, "", key)
	assert.Equal(t, 0, val)
	assert.False(t, ok)
	key, val, ok = iter.Next()
	assert.Equal(t, "", key)
	assert.Equal(t, 0, val)
	assert.False(t, ok)
}

func TestIndexedList_Remove(t *testing.T) {
	il := New[string, int]()
	il.Append([]string{"a", "b", "c"}, []int{0, 1, 2})
	assert.Equal(t, 3, il.Len())

	il.Remove([]string{"b"})
	assert.Equal(t, 2, il.Len())

	iter := il.Iterator(0)
	key, val, ok := iter.Next()
	assert.Equal(t, "a", key)
	assert.Equal(t, 0, val)
	assert.True(t, ok)
	key, val, ok = iter.Next()
	assert.Equal(t, "c", key)
	assert.Equal(t, 2, val)
	assert.True(t, ok)
	key, val, ok = iter.Next()
	assert.Equal(t, "", key)
	assert.Equal(t, 0, val)
	assert.False(t, ok)
}

func TestIndexedList_EmptyIterator(t *testing.T) {
	il := New[string, int]()
	il.Append([]string{"a"}, []int{0})
	assert.Equal(t, 1, il.Len())

	iter := il.Iterator(0)

	il.Remove([]string{"a"})
	assert.Equal(t, 0, il.Len())

	key, val, ok := iter.Next()
	assert.Equal(t, "", key)
	assert.Equal(t, 0, val)
	assert.False(t, ok)
}

func TestIndexedList_IteratorModify(t *testing.T) {
	il := New[string, int]()
	il.Append([]string{"a", "b", "c"}, []int{0, 1, 2})
	assert.Equal(t, 3, il.Len())

	iter := il.Iterator(0)
	key, val, ok := iter.Next()
	assert.Equal(t, "a", key)
	assert.Equal(t, 0, val)
	assert.True(t, ok)

	il.Remove([]string{"b"})

	key, val, ok = iter.Next()
	assert.Equal(t, "c", key)
	assert.Equal(t, 2, val)
	assert.True(t, ok)
	key, val, ok = iter.Next()
	assert.Equal(t, "", key)
	assert.Equal(t, 0, val)
	assert.False(t, ok)
}

func TestIndexedList_GarbageCollect(t *testing.T) {
	il := New[string, int]()
	il.Append([]string{"a", "b", "c"}, []int{0, 1, 2})
	assert.Equal(t, 3, il.Len())

	it0 := il.Iterator(0)
	it1 := il.Iterator(1)

	key, val, ok := it0.Next()
	assert.Equal(t, "a", key)
	assert.Equal(t, 0, val)
	assert.True(t, ok)
	key, val, ok = it1.Next()
	assert.Equal(t, "a", key)
	assert.Equal(t, 0, val)
	assert.True(t, ok)

	assert.Equal(t, 2, len(il.iterators))
	il.GarbageCollect(1)
	assert.Equal(t, 1, len(il.iterators))

	key, val, ok = it0.Next()
	assert.Equal(t, "", key)
	assert.Equal(t, 0, val)
	assert.False(t, ok)
	key, val, ok = it1.Next()
	assert.Equal(t, "b", key)
	assert.Equal(t, 1, val)
	assert.True(t, ok)

	it2 := il.Iterator(0)
	assert.True(t, it2.stopped)
}
