package sliceutil

// Repeat returns a slice with value repeated n times.
func Repeat[T any](value T, n int) []T {
	arr := make([]T, n)
	for i := 0; i < n; i++ {
		arr[i] = value
	}
	return arr
}
