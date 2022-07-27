package dsl

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/mathutil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
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
		if uintsSum > 1000 {
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
			eventsIn:  events.ListOf(events.TestingUint(mc.Self, 2000)),
			eventsOut: events.EmptyList(),
			err:       errors.New("too much"),
		},
	}

	for testName, tc := range tests {
		eventsIn := tc.eventsIn
		tc := tc
		t.Run(testName, func(t *testing.T) {
			m := newSimpleTestingModule(mc)
			eventsOutList, err := m.ApplyEvents(eventsIn)

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

type contextTestingModuleModuleConfig struct {
	Self     types.ModuleID
	Crypto   types.ModuleID
	Hasher   types.ModuleID
	Timer    types.ModuleID
	Signed   types.ModuleID
	Hashed   types.ModuleID
	Verified types.ModuleID
}

func defaultContextTestingModuleConfig() *contextTestingModuleModuleConfig {
	return &contextTestingModuleModuleConfig{
		Self:     "testing",
		Crypto:   "crypto",
		Hasher:   "hasher",
		Timer:    "timer",
		Signed:   "signed",
		Hashed:   "hashed",
		Verified: "verified",
	}
}

type testingStringContext struct {
	s string
}

func newContextTestingModule(mc *contextTestingModuleModuleConfig) Module {
	m := NewModule(mc.Self)

	UponTestingString(m, func(s string) error {
		SignRequest(m, mc.Crypto, [][]byte{[]byte(s)}, &testingStringContext{s})
		HashOneMessage(m, mc.Hasher, [][]byte{[]byte(s)}, &testingStringContext{s})
		return nil
	})

	UponSignResult(m, func(signature []byte, context *testingStringContext) error {
		EmitTestingString(m, mc.Signed, fmt.Sprintf("%s: %s", context.s, string(signature)))
		return nil
	})

	UponHashResult(m, func(hashes [][]byte, context *testingStringContext) error {
		if len(hashes) != 1 {
			return fmt.Errorf("unexpected number of hashes: %v", hashes)
		}
		EmitTestingString(m, mc.Hashed, fmt.Sprintf("%s: %s", context.s, string(hashes[0])))
		return nil
	})

	UponTestingUint(m, func(u uint64) error {
		if u < 10 {
			msg := [][]byte{[]byte("uint"), []byte(strconv.FormatUint(u, 10))}

			var signatures [][]byte
			var nodeIDs []types.NodeID
			for i := uint64(0); i < u; i++ {
				signatures = append(signatures, []byte(strconv.FormatUint(i, 10)))
				nodeIDs = append(nodeIDs, types.NodeID(strconv.FormatUint(i, 10)))
			}

			// NB: avoid using primitive types as the context in the actual implementation, prefer named structs,
			//     remember that the context type is used to match requests with responses.
			VerifyNodeSigs(m, mc.Crypto, sliceutil.Repeat(msg, int(u)), signatures, nodeIDs, &u)
		}
		return nil
	})

	UponOneNodeSigVerified(m, func(nodeID types.NodeID, err error, context *uint64) error {
		if err == nil {
			EmitTestingString(m, mc.Verified, fmt.Sprintf("%v: %v verified", *context, nodeID))
		}
		return nil
	})

	UponNodeSigsVerified(m, func(nodeIDs []types.NodeID, errs []error, allOK bool, context *uint64) error {
		if allOK {
			EmitTestingUint(m, mc.Verified, *context)
		}
		return nil
	})

	return m
}

func TestDslModule_ContextRecoveryAndCleanup(t *testing.T) {
	testCases := map[string]func(mc *contextTestingModuleModuleConfig, m Module){
		"empty": func(mc *contextTestingModuleModuleConfig, m Module) {},

		"request response": func(mc *contextTestingModuleModuleConfig, m Module) {
			eventsOut, err := m.ApplyEvents(events.ListOf(events.TestingString(mc.Self, "hello")))
			assert.Nil(t, err)
			assert.Equal(t, 2, eventsOut.Len())

			iter := eventsOut.Iterator()
			signOrigin := iter.Next().Type.(*eventpb.Event_SignRequest).SignRequest.Origin
			hashOrigin := iter.Next().Type.(*eventpb.Event_HashRequest).HashRequest.Origin

			eventsOut, err = m.ApplyEvents(events.ListOf(events.SignResult(mc.Self, []byte("world"), signOrigin)))
			assert.Nil(t, err)
			assert.Equal(t, []*eventpb.Event{events.TestingString(mc.Signed, "hello: world")}, eventsOut.Slice())

			eventsOut, err = m.ApplyEvents(events.ListOf(events.HashResult(mc.Self, [][]byte{[]byte("world")}, hashOrigin)))
			assert.Nil(t, err)
			assert.Equal(t, []*eventpb.Event{events.TestingString(mc.Hashed, "hello: world")}, eventsOut.Slice())
		},

		"response without request": func(mc *contextTestingModuleModuleConfig, m Module) {
			assert.Panics(t, func() {
				// Context with id 42 doesn't exist. The module should panic.
				_, _ = m.ApplyEvents(events.ListOf(
					events.SignResult(mc.Self, []byte{}, DslSignOrigin(mc.Self, ContextID(42)))))
			})
		},

		"check context is disposed": func(mc *contextTestingModuleModuleConfig, m Module) {
			eventsOut, err := m.ApplyEvents(events.ListOf(events.TestingString(mc.Self, "hello")))
			assert.Nil(t, err)
			assert.Equal(t, 2, eventsOut.Len())

			iter := eventsOut.Iterator()
			signOrigin := iter.Next().Type.(*eventpb.Event_SignRequest).SignRequest.Origin
			_ = iter.Next().Type.(*eventpb.Event_HashRequest).HashRequest.Origin

			eventsOut, err = m.ApplyEvents(events.ListOf(events.SignResult(mc.Self, []byte("world"), signOrigin)))
			assert.Nil(t, err)
			assert.Equal(t, []*eventpb.Event{events.TestingString(mc.Signed, "hello: world")}, eventsOut.Slice())

			assert.Panics(t, func() {
				// This reply is sent for the second time.
				// The context should already be disposed of and the module should panic.
				_, _ = m.ApplyEvents(events.ListOf(events.SignResult(mc.Self, []byte("world"), signOrigin)))
			})
		},

		"check multiple handlers for response": func(mc *contextTestingModuleModuleConfig, m Module) {
			eventsOut, err := m.ApplyEvents(events.ListOf(events.TestingUint(mc.Self, 8)))
			assert.Nil(t, err)
			assert.Equal(t, 1, eventsOut.Len())

			iter := eventsOut.Iterator()
			sigVerEvent := iter.Next().Type.(*eventpb.Event_VerifyNodeSigs).VerifyNodeSigs
			sigVerNodes := sigVerEvent.NodeIds
			assert.Equal(t, 8, len(sigVerNodes))
			sigVerOrigin := sigVerEvent.Origin

			// send some undelated events to make sure the context is preserved and does not get overwritten
			_, err = m.ApplyEvents(events.ListOf(events.TestingString(mc.Self, "hello")))
			assert.Nil(t, err)
			_, err = m.ApplyEvents(events.ListOf(events.TestingUint(mc.Self, 3)))
			assert.Nil(t, err)
			_, err = m.ApplyEvents(events.ListOf(events.TestingUint(mc.Self, 16), events.TestingString(mc.Self, "foo")))
			assert.Nil(t, err)

			// construct a response for the signature verification request.
			sigsVerifiedEvent := events.NodeSigsVerified(
				/*destModule*/ mc.Self,
				/*valid*/ sliceutil.Repeat(true, 8),
				/*errors*/ sliceutil.Repeat("", 8),
				/*nodeIDs*/ types.NodeIDSlice(sigVerNodes),
				/*origin*/ sigVerOrigin,
				/*allOk*/ true,
			)

			eventsOut, err = m.ApplyEvents(events.ListOf(sigsVerifiedEvent))
			assert.Nil(t, err)

			var expectedResponse []*eventpb.Event
			for i := 0; i < 8; i++ {
				expectedResponse = append(expectedResponse, events.TestingString(mc.Verified, fmt.Sprintf("8: %v verified", i)))
			}
			expectedResponse = append(expectedResponse, events.TestingUint(mc.Verified, 8))

			assert.Equal(t, expectedResponse, eventsOut.Slice())

			assert.Panics(t, func() {
				// This reply is sent for the second time.
				// The context should already be disposed of and the module should panic.
				_, _ = m.ApplyEvents(events.ListOf(sigsVerifiedEvent))
			})
		},
	}

	for testName, tc := range testCases {
		tc := tc
		t.Run(testName, func(t *testing.T) {
			mc := defaultContextTestingModuleConfig()
			m := newContextTestingModule(mc)
			tc(mc, m)
		})
	}

}

// event wrappers (similar to the ones in pkg/events/events.go)

func DslSignOrigin(module types.ModuleID, contextID ContextID) *eventpb.SignOrigin {
	return &eventpb.SignOrigin{
		Module: module.Pb(),
		Type: &eventpb.SignOrigin_Dsl{
			Dsl: Origin(contextID),
		},
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
	UponEvent[*eventpb.Event_TestingString](m, func(ev *wrapperspb.StringValue) error {
		return handler(ev.Value)
	})
}

func UponTestingUint(m Module, handler func(u uint64) error) {
	UponEvent[*eventpb.Event_TestingUint](m, func(ev *wrapperspb.UInt64Value) error {
		return handler(ev.Value)
	})
}
