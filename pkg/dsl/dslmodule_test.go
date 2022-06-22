package dsl

import (
	"errors"
	"fmt"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/mathutil"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"strconv"
	"testing"
)

type simpleModuleConfig struct {
	Self    types.ModuleID
	Replies types.ModuleID
	Reports types.ModuleID
}

func defaultSimpleModuleConfig() *simpleModuleConfig {
	return &simpleModuleConfig{
		Self:    "testing",
		Replies: "replies",
		Reports: "reports",
	}
}

func newSimpleTestingModule(mc *simpleModuleConfig) modules.PassiveModule {
	m := NewModule(mc.Self)

	// state
	var testingStrings []string
	var uintsSum uint64
	var lastReportedUint uint64

	UponTestingString(m, func(s string) error {
		if s == "hello" {
			EmitTestingString(m, mc.Replies, "world")
			EmitTestingUint(m, mc.Replies, 42)
		}
		return nil
	})

	UponTestingString(m, func(s string) error {
		if s == "good" {
			// By design, this event will be lost due to the error.
			EmitTestingString(m, mc.Replies, "lost")
			return fmt.Errorf("bye")
		}
		return nil
	})

	UponTestingUint(m, func(u uint64) error {
		if u < 100 {
			EmitTestingString(m, mc.Replies, strconv.FormatUint(u, 10))
		}
		return nil
	})

	UponTestingString(m, func(s string) error {
		testingStrings = append(testingStrings, s)
		return nil
	})

	UponCondition(m, func() error {
		if len(testingStrings) >= 3 {
			EmitEvent(m, &eventpb.Event{
				DestModule: "reports",
				Type: &eventpb.Event_TestingString{
					TestingString: wrapperspb.String(fmt.Sprintf("Collected at least 3 testing strings: %v",
						testingStrings)),
				},
			})
		}
		return nil
	})

	UponTestingUint(m, func(u uint64) error {
		uintsSum += u
		return nil
	})

	UponCondition(m, func() error {
		for uintsSum >= lastReportedUint+100 {
			lastReportedUint += 100
			EmitEvent(m, &eventpb.Event{
				DestModule: "reports",
				Type: &eventpb.Event_TestingUint{
					TestingUint: wrapperspb.UInt64(lastReportedUint),
				},
			})
		}
		return nil
	})

	UponCondition(m, func() error {
		if uintsSum > 100_000 {
			return errors.New("too much")
		}
		return nil
	})

	return m
}

func TestDslModule_ApplyEvents(t *testing.T) {
	mc := defaultSimpleModuleConfig()

	tests := map[string]struct {
		eventsIn  *events.EventList
		eventsOut *events.EventList
		err       error
	}{
		"empty": {
			eventsIn:  events.EmptyList(),
			eventsOut: events.EmptyList(),
			err:       nil,
		},
		"hello world": {
			eventsIn:  events.ListOf(events.TestingString(mc.Self, "hello")),
			eventsOut: events.ListOf(events.TestingString(mc.Replies, "world"), events.TestingUint(mc.Replies, 42)),
			err:       nil,
		},
		"test error": {
			eventsIn:  events.ListOf(events.TestingString(mc.Self, "good")),
			eventsOut: events.EmptyList(),
			err:       errors.New("bye"),
		},
		"test simple condition": {
			eventsIn: events.ListOf(
				events.TestingString(mc.Self, "foo"), events.TestingString(mc.Self, "bar"),
				events.TestingString(mc.Self, "baz"), events.TestingString(mc.Self, "quz")),
			eventsOut: events.ListOf(
				events.TestingString(mc.Reports, "Collected at least 3 testing strings: [foo bar baz quz]")),
		},
		"test multiple handlers for one event and a loop condition": {
			eventsIn: events.ListOf(
				events.TestingUint(mc.Self, 0), events.TestingUint(mc.Self, 17), events.TestingUint(mc.Self, 105),
				events.TestingUint(mc.Self, 182), events.TestingUint(mc.Self, 42), events.TestingUint(mc.Self, 222),
				events.TestingUint(mc.Self, 14)),
			// if the number is below 100, the module will reply with a string representation of the number.
			// the module will also add up all received values and will emit reports 100, 200, and so on if these
			// thresholds are passed at the end of the batch. In this example, the total sum is 582.
			eventsOut: events.ListOf(
				events.TestingString(mc.Replies, "0"), events.TestingString(mc.Replies, "17"),
				events.TestingString(mc.Replies, "42"), events.TestingString(mc.Replies, "14"),
				events.TestingUint(mc.Reports, 100), events.TestingUint(mc.Reports, 200),
				events.TestingUint(mc.Reports, 300), events.TestingUint(mc.Reports, 400),
				events.TestingUint(mc.Reports, 500)),
		},
		"test unknown event type": {
			eventsIn:  events.ListOf(events.SendMessage(mc.Self, &messagepb.Message{}, []types.NodeID{})),
			eventsOut: events.EmptyList(),
			err:       errors.New("unknown event type '*eventpb.Event_SendMessage'"),
		},
		"test failed condition": {
			eventsIn:  events.ListOf(events.TestingUint(mc.Self, 999_999_999)),
			eventsOut: events.EmptyList(),
			err:       errors.New("too much"),
		},
	}

	for testName, tc := range tests {
		t.Run(testName, func(t *testing.T) {
			m := newSimpleTestingModule(mc)
			eventsOutList, err := m.ApplyEvents(tc.eventsIn)

			if tc.err != nil {
				assert.Equal(t, tc.err.Error(), err.Error())
				assert.Nil(t, eventsOutList)
				return
			}
			assert.Nil(t, err)

			expectedEventsOut := tc.eventsOut.Slice()
			eventsOut := eventsOutList.Slice()

			assert.Equal(t, len(expectedEventsOut), len(eventsOut))

			i := 0
			for i < mathutil.Min(len(expectedEventsOut), len(eventsOut)) {
				assert.EqualValues(t, expectedEventsOut[i], eventsOut[i])
				i++
			}

			for i < len(expectedEventsOut) {
				t.Errorf("expected event %v", expectedEventsOut[i])
				i++
			}

			for i < len(eventsOut) {
				t.Errorf("unexpected event %v", eventsOut[i])
				i++
			}
		})
	}
}

// dsl wrappers (similar to the ones in pkg/dsl/events.go)

func EmitTestingString(m Module, dest types.ModuleID, s string) {
	EmitEvent(m, events.TestingString(dest, s))
}

func EmitTestingUint(m Module, dest types.ModuleID, u uint64) {
	EmitEvent(m, events.TestingUint(dest, u))
}

func UponTestingString(m Module, handler func(s string) error) {
	RegisterEventHandler(m, func(ev *eventpb.Event_TestingString) error {
		return handler(ev.TestingString.Value)
	})
}

func UponTestingUint(m Module, handler func(u uint64) error) {
	RegisterEventHandler(m, func(ev *eventpb.Event_TestingUint) error {
		return handler(ev.TestingUint.Value)
	})
}
