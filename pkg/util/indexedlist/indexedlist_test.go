package indexedlist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexedList_Append(t *testing.T) {
	il := New[string, int]()
	keys, vals := il.Append([]string{"a", "b", "c"}, []int{0, 1, 2})
	assert.Equal(t, 3, il.Len())
	assert.Equal(t, []string{"a", "b", "c"}, keys)
	assert.Equal(t, []int{0, 1, 2}, vals)

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

	keys, vals = il.Append([]string{"b", "d"}, []int{1, 3})
	assert.Equal(t, 4, il.Len())
	assert.Equal(t, []string{"d"}, keys)
	assert.Equal(t, []int{3}, vals)
}

func TestIndexedList_LookUp(t *testing.T) {
	il := New[string, int]()
	il.Append([]string{"a", "b", "c"}, []int{0, 1, 2})
	foundKeys, foundVals, missingKeys := il.LookUp([]string{"c", "a", "d"})
	assert.Equal(t, []string{"c", "a"}, foundKeys)
	assert.Equal(t, []int{2, 0}, foundVals)
	assert.Equal(t, []string{"d"}, missingKeys)
}

func TestIndexedList_Remove(t *testing.T) {
	il := New[string, int]()
	il.Append([]string{"a", "b", "c"}, []int{0, 1, 2})
	assert.Equal(t, 3, il.Len())

	keys, vals := il.Remove([]string{"b"})
	assert.Equal(t, "b", keys[0])
	assert.Equal(t, 1, vals[0])
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

func TestIndexedList_RemoveSelected(t *testing.T) {
	il := New[string, int]()
	il.Append([]string{"a", "b", "c"}, []int{0, 1, 2})
	assert.Equal(t, 3, il.Len())

	keys, vals := il.RemoveSelected(func(key string, val int) bool {
		return key == "b" || key == "c"
	})
	assert.Equal(t, []string{"b", "c"}, keys)
	assert.Equal(t, []int{1, 2}, vals)
	assert.Equal(t, 1, il.Len())

	iter := il.Iterator(0)
	key, val, ok := iter.Next()
	assert.Equal(t, "a", key)
	assert.Equal(t, 0, val)
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

	keys, vals := il.Remove([]string{"a"})
	assert.Equal(t, "a", keys[0])
	assert.Equal(t, 0, vals[0])
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

	keys, vals := il.Remove([]string{"b"})
	assert.Equal(t, "b", keys[0])
	assert.Equal(t, 1, vals[0])

	key, val, ok = iter.Next()
	assert.Equal(t, "c", key)
	assert.Equal(t, 2, val)
	assert.True(t, ok)
	key, val, ok = iter.Next()
	assert.Equal(t, "", key)
	assert.Equal(t, 0, val)
	assert.False(t, ok)
}

func TestIterator_NextWhile(t *testing.T) {
	il := New[string, int]()
	il.Append([]string{"a", "b", "c"}, []int{0, 1, 2})
	assert.Equal(t, 3, il.Len())

	// Condition allowing all elements to be traversed.
	iter := il.Iterator(0)
	sum := 0
	keys, vals, ok := iter.NextWhile(func(key string, val int) bool {
		if sum+val <= 10 {
			sum += val
			return true
		}
		return false
	})
	assert.True(t, ok)
	assert.Equal(t, 3, len(keys))
	assert.Equal(t, 3, len(vals))
	assert.Equal(t, "a", keys[0])
	assert.Equal(t, 0, vals[0])
	assert.Equal(t, "b", keys[1])
	assert.Equal(t, 1, vals[1])
	assert.Equal(t, "c", keys[2])
	assert.Equal(t, 2, vals[2])
	_, _, ok = iter.Next()
	assert.False(t, ok)

	// Condition allowing only part of the elements to be traversed.
	iter = il.Iterator(0)
	sum = 0
	keys, vals, ok = iter.NextWhile(func(key string, val int) bool {
		if sum+val <= 1 {
			sum += val
			return true
		}
		return false
	})
	assert.True(t, ok)
	assert.Equal(t, 2, len(keys))
	assert.Equal(t, 2, len(vals))
	assert.Equal(t, "a", keys[0])
	assert.Equal(t, 0, vals[0])
	assert.Equal(t, "b", keys[1])
	assert.Equal(t, 1, vals[1])
	key, val, ok := iter.Next()
	assert.Equal(t, "c", key)
	assert.Equal(t, 2, val)
	assert.True(t, ok)
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
