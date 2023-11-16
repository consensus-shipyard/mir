package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventList_Constructors(t *testing.T) {
	testCases := map[string]struct {
		list     *EventList
		expected []Event
	}{
		"EmptyList":    {EmptyList(), nil},
		"empty ListOf": {ListOf(), nil},
		"one item": {
			list:     ListOf(NewTestString("testmodule", "hello")),
			expected: []Event{NewTestString("testmodule", "hello")},
		},
		"three items": {
			list: ListOf(NewTestString("testmodule", "hello"), NewTestString("testmodule", "world"),
				NewTestUint("testmodule", 42)),
			expected: []Event{NewTestString("testmodule", "hello"), NewTestString("testmodule", "world"),
				NewTestUint("testmodule", 42)},
		},
	}

	for testName, tc := range testCases {
		tc := tc
		t.Run(testName, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.list.Slice())
		})
	}
}
