package utils

import "slices"

// Reverse reverses the order of the elements in a slice and returns the result
func Reverse[S ~[]T, T any](slice S) S {
	slices.Reverse(slice)
	return slice
}
