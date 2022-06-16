package maputil

// GetKeys returns a slice containing all keys of map m in arbitrary order.
func GetKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
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
//   keys, values := GetKeysAndValues(m)
// is equivalent to:
//   keys := GetKeys(m)
//   values := GetValuesOf(m, keys)
func GetKeysAndValues[K comparable, V any](m map[K]V) ([]K, []V) {
	keys := make([]K, 0, len(m))
	values := make([]V, 0, len(m))
	for k, v := range m {
		keys = append(keys, k)
		values = append(values, v)
	}
	return keys, values
}
