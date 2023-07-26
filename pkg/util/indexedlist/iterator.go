package indexedlist

type Iterator[K comparable, V any] struct {
	list       *IndexedList[K, V]
	lastOutput *element[K, V]
	stopped    bool
}

func (i *Iterator[K, V]) Next() (K, V, bool) {
	i.list.lock.Lock()
	defer i.list.lock.Unlock()

	var e *element[K, V]
	var zeroKey K
	var zeroVal V

	// Check if iteration has stopped.
	if i.stopped {
		return zeroKey, zeroVal, false
	}

	// Select the next element.
	if i.lastOutput == nil {
		e = i.list.first
	} else {
		e = i.lastOutput.next
	}

	// If there is no next element, return zero value.
	if e == nil {
		return zeroKey, zeroVal, false
	}

	i.lastOutput = e
	return e.key, e.val, true
}

func (i *Iterator[K, V]) NextWhile(predicate func(K, V) bool) ([]K, []V, bool) {
	i.list.lock.Lock()
	defer i.list.lock.Unlock()

	keys := make([]K, 0)
	vals := make([]V, 0)

	// Check if iteration has stopped.
	if i.stopped {
		return keys, vals, false
	}

	// Select the next element.
	var e *element[K, V]
	if i.lastOutput == nil {
		e = i.list.first
	} else {
		e = i.lastOutput.next
	}

	for ; e != nil && predicate(e.key, e.val); e = e.next {
		// If there is a next element satisfying the predicate, add it to the output.
		keys = append(keys, e.key)
		vals = append(vals, e.val)
		i.lastOutput = e
	}

	return keys, vals, true
}
