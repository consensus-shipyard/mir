package indexedlist

import (
	"fmt"
	"sync"

	tt "github.com/filecoin-project/mir/pkg/trantor/types"
)

type element[K comparable, V any] struct {
	prev *element[K, V]
	next *element[K, V]
	key  K
	val  V
}

// IndexedList is a doubly-linked list with a map index.
// Appending, prepending, and removal of arbitrary elements is in constant time and thread-safe.
type IndexedList[K comparable, V any] struct {
	// Any modification (or read of potentially concurrently modified values) requires acquiring this lock.
	lock sync.Mutex

	// index of elements in this list. Needed for constant-time random access (and removal).
	index map[K]*element[K, V]

	// first points to the start of a doubly linked list of elements in the list.
	// Necessary for constant-time prepending and appending, while still being able to traverse in consistent order.
	// If set to nil, no elements are in the list (and last also must be nil).
	first *element[K, V]

	// End of the doubly linked list of request.
	// Pointer to the end necessary for using the list as a FIFO queue.
	// If set to nil, no requests are in the bucket (and FirstRequest also must be nil).
	last *element[K, V]

	// Iterators over this list, each associated with a retention index.
	// When GarbageCollect is called with a retention index retIdx,
	// all iterators associated with an index lower than retIdx stop outputting any element
	// and their associated resources are freed.
	iterators map[tt.RetentionIndex][]*Iterator[K, V]

	// Minimal retention index of the iterators associated with this list.
	// Any call to Iterator with a lower retention index than retIdx returns a stopped iterator.
	retIdx tt.RetentionIndex
}

// New returns a new empty IndexedList with K as the index type and V as the value type.
func New[K comparable, V any]() *IndexedList[K, V] {
	return &IndexedList[K, V]{
		index:     make(map[K]*element[K, V]),
		iterators: make(map[tt.RetentionIndex][]*Iterator[K, V]),
	}
}

// Len returns the number of items stored in the list.
func (il *IndexedList[K, V]) Len() int {
	return len(il.index)
}

// Append adds multiple key-value pairs to the end of the list, in the given order.
// If the two arguments have different lengths, Append panics.
func (il *IndexedList[K, V]) Append(keys []K, vals []V) {
	il.lock.Lock()
	defer il.lock.Unlock()

	if len(keys) != len(vals) {
		panic(fmt.Sprintf("key and value slices have different lengths: %d, %d", len(keys), len(vals)))
	}

	for i, k := range keys {
		// TODO: This can be optimized by creating a chain of elements directly and then appending it as a whole.
		il.appendElement(&element[K, V]{
			key: k,
			val: vals[i],
		})
	}
}

// appendElement appends a given element to the end of the list if not already present (anywhere in the list).
// ATTENTION: The list must be locked when calling this function!
func (il *IndexedList[K, V]) appendElement(e *element[K, V]) {

	// Ignore element if one with the same key already exists.
	if _, ok := il.index[e.key]; ok {
		return
	}

	// Add element to the index.
	il.index[e.key] = e

	// Append element.
	if il.first == nil {
		il.first = e
	} else {
		il.last.next = e
		e.prev = il.last
	}
	il.last = e

}

// prependElement prepends a given element to the start of the list if not already present (anywhere in the list).
// ATTENTION: The list must be locked when calling this function!
func (il *IndexedList[K, V]) prependElement(e *element[K, V]) {

	// Ignore element if one with the same key already exists.
	if _, ok := il.index[e.key]; ok {
		return
	}

	// Add element to the index.
	il.index[e.key] = e

	// Prepend element
	if il.first == nil {
		il.last = e
	} else {
		il.first.prev = e
		e.next = il.first
	}
	il.first = e

}

// Remove removes list entries with the given keys, if they are in the list.
func (il *IndexedList[K, V]) Remove(keys []K) {
	for _, key := range keys {
		if e, ok := il.index[key]; ok {
			il.removeElement(e)
		}
	}
}

// removeElement removes an element from the list, including the index.
// ATTENTION: The list must be locked when calling this function!
// ATTENTION: The element must be in the list. If it is not, removeElement will corrupt the state
// of both this list and the list the element is actually in (if any).
func (il *IndexedList[K, V]) removeElement(e *element[K, V]) {

	// Remove element from the index.
	delete(il.index, e.key)

	// Adapt iterators pointing to the element being removed.
	// TODO: This makes removal quite inefficient.
	//  One can optimize this by referring to the relevant iterators from the element itself
	//  or, as a compromise, just flag elements as iterated over and call an unmodified updateIterators only for those.
	il.updateIterators(e)

	if e.next != nil { // Element is not last in the list.
		e.next.prev = e.prev
	} else { // Element is last in the list
		il.last = e.prev
	}
	if e.prev != nil { // Element is not first in the list
		e.prev.next = e.next
	} else { // Element is first in the list
		il.first = e.next
	}

}

// updateIterators adapts the state of all iterators currently traversing the list to the removal of the given element.
func (il *IndexedList[K, V]) updateIterators(e *element[K, V]) {
	for _, iterators := range il.iterators {
		for _, iterator := range iterators {
			if iterator.lastOutput == e {
				iterator.lastOutput = e.prev
			}
		}
	}
}

// Iterator creates a new iterator for the list, initialized to the list's start.
// The list can be modified concurrently with being iterated over.
func (il *IndexedList[K, V]) Iterator(retIdx tt.RetentionIndex) *Iterator[K, V] {
	il.lock.Lock()
	defer il.lock.Unlock()

	// If the given retention index already has been garbage-collected,
	// return a stopped iterator.
	if retIdx < il.retIdx {
		return &Iterator[K, V]{
			list:       il,
			lastOutput: nil,
			stopped:    true,
		}
	}

	iterator := &Iterator[K, V]{
		list:       il,
		lastOutput: nil,
		stopped:    false,
	}

	il.iterators[retIdx] = append(il.iterators[retIdx], iterator)

	return iterator
}

func (il *IndexedList[K, V]) GarbageCollect(retIdx tt.RetentionIndex) {
	il.lock.Lock()
	defer il.lock.Unlock()

	for ; il.retIdx < retIdx; il.retIdx++ {
		for _, iterator := range il.iterators[il.retIdx] {
			iterator.stopped = true
		}
		delete(il.iterators, il.retIdx)
	}
}
