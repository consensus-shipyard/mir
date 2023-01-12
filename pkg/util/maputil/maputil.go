package maputil

import (
	"fmt"
	"sort"

	"golang.org/x/exp/constraints"
)

// GetKeys returns a slice containing all keys of map m in arbitrary order.
func GetKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func GetSortedKeys[K constraints.Ordered, V any](m map[K]V) []K {
	keys := GetKeys(m)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

// GetValuesOf returns a slice containing the values from map m corresponding to the provided keys, in the same order
// as the keys.
func GetValuesOf[K comparable, V any](m map[K]V, keys []K) []V {
	values := make([]V, 0, len(keys))
	for _, k := range keys {
		values = append(values, m[k])
	}
	return values
}

// GetKeysAndValues returns a slice containing all keys of map m (in arbitrary order) and a slice containing the
// corresponding values, in the corresponding order.
//
// statement
//
//	keys, values := GetKeysAndValues(m)
//
// is equivalent to:
//
//	keys := GetKeys(m)
//	values := GetValuesOf(m, keys)
func GetKeysAndValues[K comparable, V any](m map[K]V) ([]K, []V) {
	keys := make([]K, 0, len(m))
	values := make([]V, 0, len(m))
	for k, v := range m {
		keys = append(keys, k)
		values = append(values, v)
	}
	return keys, values
}

func IterateSorted[K constraints.Ordered, V any](m map[K]V, f func(key K, value V) (cont bool)) {

	for _, k := range GetSortedKeys(m) {
		if !f(k, m[k]) {
			break
		}
	}
}

func Copy[K comparable, V any](m map[K]V) map[K]V {
	newMap := make(map[K]V, len(m))
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}

// Any returns an arbitrary element of the map m.
// If m is not empty, the second return value is true.
// If the map is empty, Any returns the zero value of the m's value type and false.
func Any[K comparable, V any](m map[K]V) (V, bool) {

	for _, val := range m {
		return val, true
	}

	var zeroVal V
	return zeroVal, false
}

func Transform[Ki comparable, Vi any, Ko comparable, Vo any](mi map[Ki]Vi, kt func(Ki) Ko, vt func(Vi) Vo) map[Ko]Vo {
	mo := make(map[Ko]Vo, len(mi))
	for ki, vi := range mi {
		mo[kt(ki)] = vt(vi)
	}
	return mo
}

// FromSlices constructs and returns a map from two separate slices of keys and corresponding values.
// FromSlices panics if the number of keys differs from the number of values.
func FromSlices[K comparable, V any](keys []K, vals []V) map[K]V {
	if len(keys) != len(vals) {
		panic(fmt.Sprintf("number of keys (%d) and number of values (%d) must match", len(keys), len(vals)))
	}

	m := make(map[K]V, len(keys))
	for i, key := range keys {
		m[key] = vals[i]
	}

	return m
}
