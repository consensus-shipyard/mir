package sliceutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

type args[T comparable] struct {
	Item1 []T
	Item2 []T
}
type testCase[T comparable] struct {
	name string
	args args[T]
	want bool
}

func TestContainsAll1(t *testing.T) {
	tcString := []testCase[string]{
		{
			name: "n1",
			args: args[string]{
				Item1: []string{"a", "b"},
				Item2: []string{"a", "b"},
			},
			want: true,
		},
		{
			name: "n2",
			args: args[string]{
				Item1: []string{"a", "b"},
				Item2: []string{"b"},
			},
			want: true,
		},
		{
			name: "n3",
			args: args[string]{
				Item1: []string{"a", "b"},
				Item2: []string{"b", "d"},
			},
			want: false,
		},
		{
			name: "n4",
			args: args[string]{
				Item1: []string{"a", "b"},
				Item2: []string{"c", "d"},
			},
			want: false,
		},
	}

	for _, tc := range tcString {
		t.Run(tc.name, func(t *testing.T) {
			actual := sliceutil.ContainsAll(tc.args.Item1, tc.args.Item2)
			assert.Equal(t, tc.want, actual)
		})
	}

	tcInt := []testCase[int]{
		{
			name: "testInt-1",
			args: args[int]{
				Item1: []int{1, 2},
				Item2: []int{2},
			},
			want: true,
		},
		{
			name: "testInt-2",
			args: args[int]{
				Item1: []int{1, 2},
				Item2: []int{2, 1},
			},
			want: true,
		},
		{
			name: "testInt-3",
			args: args[int]{
				Item1: []int{1, 2},
				Item2: []int{3, 4},
			},
			want: false,
		},
		{
			name: "testInt-4",
			args: args[int]{
				Item1: []int{1, 2},
				Item2: []int{1, 3},
			},
			want: false,
		},
	}
	for _, tc := range tcInt {
		t.Run(tc.name, func(t *testing.T) {
			actual := sliceutil.ContainsAll(tc.args.Item1, tc.args.Item2)
			assert.Equal(t, tc.want, actual)
		})
	}
}
