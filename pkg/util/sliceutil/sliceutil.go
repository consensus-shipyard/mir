package sliceutil

// Repeat returns a slice with value repeated n times.
func Repeat[T any](value T, n int) []T {
	arr := make([]T, n)
	for i := 0; i < n; i++ {
		arr[i] = value
	}
	return arr
}

func Generate[T any](n int, f func(i int) T) []T {
	arr := make([]T, n)
	for i := 0; i < n; i++ {
		arr[i] = f(i)
	}
	return arr
}

func Transform[T, R any](ts []T, f func(i int, t T) R) []R {
	rs := make([]R, len(ts))
	for i, t := range ts {
		rs[i] = f(i, t)
	}
	return rs
}

func Filter[T any](ts []T, f func(i int, t T) bool) []T {
	var res []T
	for i, t := range ts {
		if f(i, t) {
			res = append(res, t)
		}
	}
	return res
}

// Any returns an arbitrary element of the slice.
// If the slice is not empty, the second return value is true.
// Otherwise, Any returns the zero value of the slice's element type and false.
func Any[T any](ts []T) (T, bool) {
	if len(ts) == 0 {
		var res T
		return res, false
	}

	return ts[0], true
}
