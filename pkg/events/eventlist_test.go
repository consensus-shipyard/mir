package events

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/mir/pkg/pb/eventpb"
)

func TestEventList_Constructors(t *testing.T) {
	testCases := map[string]struct {
		list     *EventList
		expected []*eventpb.Event
	}{
		"EmptyList":    {EmptyList(), nil},
		"empty ListOf": {ListOf(), nil},
		"one item": {
			list:     ListOf(TestingString("testmodule", "hello")),
			expected: []*eventpb.Event{TestingString("testmodule", "hello")},
		},
		"three items": {
			list: ListOf(TestingString("testmodule", "hello"), TestingString("testmodule", "world"),
				TestingUint("testmodule", 42)),
			expected: []*eventpb.Event{TestingString("testmodule", "hello"), TestingString("testmodule", "world"),
				TestingUint("testmodule", 42)},
		},
	}

	for testName, tc := range testCases {
		tc := tc
		t.Run(testName, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.list.Slice())
		})
	}
}
