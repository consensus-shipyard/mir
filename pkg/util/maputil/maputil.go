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

// GetValues returns a slice containing all values of map m in arbitrary order.
func GetValues[K comparable, V any](m map[K]V) []V {
	vals := make([]V, 0, len(m))
	for _, v := range m {
		vals = append(vals, v)
	}
	return vals
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

func IterateSortedCustom[K comparable, V any](m map[K]V, f func(K, V) bool, less func(K, K) bool) {
	sl := make([]struct {
		k K
		v V
	}, 0)
	for k, v := range m {
		sl = append(sl, struct {
			k K
			v V
		}{k: k, v: v})
	}

	sort.Slice(sl, func(i, j int) bool {
		return less(sl[i].k, sl[j].k)
	})

	for _, item := range sl {
		if !f(item.k, item.v) {
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

// AnyKey returns an arbitrary key of the map m.
// If m is not empty, the second return value is true.
// If the map is empty, AnyKey returns the zero value of the m's key type and false.
func AnyKey[K comparable, V any](m map[K]V) (K, bool) {

	for key := range m {
		return key, true
	}

	var zeroVal K
	return zeroVal, false
}

// AnyVal returns an arbitrary value of the map m.
// If m is not empty, the second return value is true.
// If the map is empty, AnyVal returns the zero value of the m's value type and false.
func AnyVal[K comparable, V any](m map[K]V) (V, bool) {

	for _, val := range m {
		return val, true
	}

	var zeroVal V
	return zeroVal, false
}

func Transform[Ki comparable, Vi any, Ko comparable, Vo any](mi map[Ki]Vi, f func(Ki, Vi) (Ko, Vo)) map[Ko]Vo {
	mo := make(map[Ko]Vo, len(mi))
	for ki, vi := range mi {
		ko, vo := f(ki, vi)
		mo[ko] = vo
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

func RemoveAll[K comparable, V any](m map[K]V, f func(key K, value V) bool) map[K]V {
	removed := make(map[K]V)
	for k, v := range m {
		if f(k, v) {
			removed[k] = v
			delete(m, k)
		}
	}
	return removed
}
