// This file is in a separate package to avoid an import cycle between the stdtypes and stdevents packages.
package testing

import (
	"testing"

	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
	"github.com/stretchr/testify/assert"
)

func TestEventList_Constructors(t *testing.T) {
	testCases := map[string]struct {
		list     *stdtypes.EventList
		expected []stdtypes.Event
	}{
		"EmptyList":    {stdtypes.EmptyList(), nil},
		"empty ListOf": {stdtypes.ListOf(), nil},
		"one item": {
			list:     stdtypes.ListOf(stdevents.NewTestString("testmodule", "hello")),
			expected: []stdtypes.Event{stdevents.NewTestString("testmodule", "hello")},
		},
		"three items": {
			list: stdtypes.ListOf(stdevents.NewTestString("testmodule", "hello"), stdevents.NewTestString("testmodule", "world"),
				stdevents.NewTestUint64("testmodule", 42)),
			expected: []stdtypes.Event{stdevents.NewTestString("testmodule", "hello"), stdevents.NewTestString("testmodule", "world"),
				stdevents.NewTestUint64("testmodule", 42)},
		},
	}

	for testName, tc := range testCases {
		tc := tc
		t.Run(testName, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.list.Slice())
		})
	}
}
