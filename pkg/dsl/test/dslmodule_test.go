package test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/cryptopb"
	cryptopbdsl "github.com/filecoin-project/mir/pkg/pb/cryptopb/dsl"
	cryptopbevents "github.com/filecoin-project/mir/pkg/pb/cryptopb/events"
	cryptopbtypes "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	messagepbtypes "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	testerpbevents "github.com/filecoin-project/mir/pkg/pb/testerpb/events"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	"github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/mathutil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
	// TODO: Try removing the dependency on the crypto module. (Does that mean completely removing some tests?)
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
	m := dsl.NewModule(mc.Self)

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
			return es.Errorf("bye")
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

	dsl.UponStateUpdates(m, func() error {
		if len(testingStrings) >= 3 {
			dsl.EmitEvent(m, stdevents.NewTestString(
				"reports",
				fmt.Sprintf("Collected at least 3 NewTest strings: %v", testingStrings),
			))
		}
		return nil
	})

	UponTestingUint(m, func(u uint64) error {
		uintsSum += u
		return nil
	})

	dsl.UponStateUpdates(m, func() error {
		for uintsSum >= lastReportedUint+100 {
			lastReportedUint += 100
			dsl.EmitEvent(m, stdevents.NewTestUint64(
				"reports",
				lastReportedUint,
			))
		}
		return nil
	})

	dsl.UponStateUpdates(m, func() error {
		if uintsSum > 1000 {
			return errors.New("too much")
		}
		return nil
	})

	// TODO: Write tests for UponStateUpdate and especially
	//   for the different behavior of UponStateUpdate and UponstateUpdates.

	return m
}

func TestDslModule_ApplyEvents(t *testing.T) {
	mc := defaultSimpleModuleConfig()

	tests := map[string]struct {
		eventsIn  *stdtypes.EventList
		eventsOut *stdtypes.EventList
		err       error
	}{
		"empty": {
			eventsIn:  stdtypes.EmptyList(),
			eventsOut: stdtypes.EmptyList(),
			err:       nil,
		},
		"hello world": {
			eventsIn:  stdtypes.ListOf(stdevents.NewTestString(mc.Self, "hello")),
			eventsOut: stdtypes.ListOf(stdevents.NewTestString(mc.Replies, "world"), stdevents.NewTestUint64(mc.Replies, 42)),
			err:       nil,
		},
		"test error": {
			eventsIn:  stdtypes.ListOf(stdevents.NewTestString(mc.Self, "good")),
			eventsOut: stdtypes.EmptyList(),
			err:       errors.New("bye"),
		},
		"test simple condition": {
			eventsIn: stdtypes.ListOf(
				stdevents.NewTestString(mc.Self, "foo"), stdevents.NewTestString(mc.Self, "bar"),
				stdevents.NewTestString(mc.Self, "baz"), stdevents.NewTestString(mc.Self, "quz")),
			eventsOut: stdtypes.ListOf(
				stdevents.NewTestString(mc.Reports, "Collected at least 3 NewTest strings: [foo bar baz quz]")),
		},
		"test multiple handlers for one event and a loop condition": {
			eventsIn: stdtypes.ListOf(
				stdevents.NewTestUint64(mc.Self, 0), stdevents.NewTestUint64(mc.Self, 17), stdevents.NewTestUint64(mc.Self, 105),
				stdevents.NewTestUint64(mc.Self, 182), stdevents.NewTestUint64(mc.Self, 42), stdevents.NewTestUint64(mc.Self, 222),
				stdevents.NewTestUint64(mc.Self, 14)),
			// if the number is below 100, the module will reply with a string representation of the number.
			// the module will also add up all received values and will emit reports 100, 200, and so on if these
			// thresholds are passed at the end of the batch. In this example, the total sum is 582.
			eventsOut: stdtypes.ListOf(
				stdevents.NewTestString(mc.Replies, "0"), stdevents.NewTestString(mc.Replies, "17"),
				stdevents.NewTestString(mc.Replies, "42"), stdevents.NewTestString(mc.Replies, "14"),
				stdevents.NewTestUint64(mc.Reports, 100), stdevents.NewTestUint64(mc.Reports, 200),
				stdevents.NewTestUint64(mc.Reports, 300), stdevents.NewTestUint64(mc.Reports, 400),
				stdevents.NewTestUint64(mc.Reports, 500)),
		},
		"test unknown message event type": {
			eventsIn: stdtypes.ListOf(transportpbevents.SendMessage(
				mc.Self,
				&messagepbtypes.Message{},
				[]types.NodeID{}).Pb(),
			),
			eventsOut: stdtypes.EmptyList(),
			err:       nil,
		},
		"test unknown event type": {
			eventsIn:  stdtypes.ListOf(testerpbevents.Tester(mc.Self).Pb()),
			eventsOut: stdtypes.EmptyList(),
			err:       errors.New("unknown event type '*eventpb.Event_Tester'"),
		},
		"test failed condition": {
			eventsIn:  stdtypes.ListOf(stdevents.NewTestUint64(mc.Self, 2000)),
			eventsOut: stdtypes.EmptyList(),
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

func newContextTestingModule(mc *contextTestingModuleModuleConfig) dsl.Module {
	m := dsl.NewModule(mc.Self)

	UponTestingString(m, func(s string) error {
		cryptopbdsl.SignRequest(
			m,
			mc.Crypto,
			&cryptopbtypes.SignedData{Data: [][]byte{[]byte(s)}},
			&testingStringContext{s},
		)
		return nil
	})

	cryptopbdsl.UponSignResult(m, func(signature []byte, context *testingStringContext) error {
		EmitTestingString(m, mc.Signed, fmt.Sprintf("%s: %s", context.s, string(signature)))
		return nil
	})

	UponTestingUint(m, func(u uint64) error {
		if u < 10 {
			msg := &cryptopbtypes.SignedData{Data: [][]byte{[]byte("uint"), []byte(strconv.FormatUint(u, 10))}}

			var signatures [][]byte
			var nodeIDs []types.NodeID
			for i := uint64(0); i < u; i++ {
				signatures = append(signatures, []byte(strconv.FormatUint(i, 10)))
				nodeIDs = append(nodeIDs, types.NodeID(strconv.FormatUint(i, 10)))
			}

			// NB: avoid using primitive types as the context in the actual implementation, prefer named structs,
			//     remember that the context type is used to match requests with responses.
			cryptopbdsl.VerifySigs(m, mc.Crypto, sliceutil.Repeat(msg, int(u)), signatures, nodeIDs, &u)
		}
		return nil
	})

	cryptopbdsl.UponSigsVerified(m, func(nodeIDs []types.NodeID, errs []error, allOK bool, context *uint64) error {
		if allOK {
			for _, nodeID := range nodeIDs {
				EmitTestingString(m, mc.Verified, fmt.Sprintf("%v: %v verified", *context, nodeID))
			}
		}
		return nil
	})

	cryptopbdsl.UponSigsVerified(m, func(nodeIDs []types.NodeID, errs []error, allOK bool, context *uint64) error {
		if allOK {
			EmitTestingUint(m, mc.Verified, *context)
		}
		return nil
	})

	return m
}

func TestDslModule_ContextRecoveryAndCleanup(t *testing.T) {
	testCases := map[string]func(mc *contextTestingModuleModuleConfig, m dsl.Module){
		"empty": func(mc *contextTestingModuleModuleConfig, m dsl.Module) {},

		"request response": func(mc *contextTestingModuleModuleConfig, m dsl.Module) {
			eventsOut, err := m.ApplyEvents(stdtypes.ListOf(stdevents.NewTestString(mc.Self, "hello")))
			assert.Nil(t, err)
			assert.Equal(t, 1, eventsOut.Len())

			iter := eventsOut.Iterator()
			signOrigin := iter.Next().(*eventpb.Event).Type.(*eventpb.Event_Crypto).Crypto.Type.(*cryptopb.Event_SignRequest).SignRequest.Origin

			eventsOut, err = m.ApplyEvents(stdtypes.ListOf(cryptopbevents.SignResult(
				mc.Self,
				[]byte("world"),
				cryptopbtypes.SignOriginFromPb(signOrigin),
			).Pb()))
			assert.Nil(t, err)
			assert.Equal(t, []stdtypes.Event{stdevents.NewTestString(mc.Signed, "hello: world")}, eventsOut.Slice())
		},

		"response without request": func(mc *contextTestingModuleModuleConfig, m dsl.Module) {
			assert.Panics(t, func() {
				// Context with id 42 doesn't exist. The module should panic.
				_, _ = m.ApplyEvents(stdtypes.ListOf(
					cryptopbevents.SignResult(mc.Self, []byte{}, DslSignOrigin(mc.Self, dsl.ContextID(42))).Pb()))
			})
		},

		"check context is disposed": func(mc *contextTestingModuleModuleConfig, m dsl.Module) {
			eventsOut, err := m.ApplyEvents(stdtypes.ListOf(stdevents.NewTestString(mc.Self, "hello")))
			assert.Nil(t, err)
			assert.Equal(t, 1, eventsOut.Len())

			iter := eventsOut.Iterator()
			signOrigin := cryptopbtypes.SignOriginFromPb(iter.Next().(*eventpb.Event).Type.(*eventpb.Event_Crypto).Crypto.Type.(*cryptopb.Event_SignRequest).SignRequest.Origin)

			eventsOut, err = m.ApplyEvents(stdtypes.ListOf(
				cryptopbevents.SignResult(mc.Self, []byte("world"), signOrigin).Pb(),
			))
			assert.Nil(t, err)
			assert.Equal(t, []stdtypes.Event{stdevents.NewTestString(mc.Signed, "hello: world")}, eventsOut.Slice())

			assert.Panics(t, func() {
				// This reply is sent for the second time.
				// The context should already be disposed of and the module should panic.
				_, _ = m.ApplyEvents(stdtypes.ListOf(
					cryptopbevents.SignResult(mc.Self, []byte("world"), signOrigin).Pb()),
				)
			})
		},

		"check multiple handlers for response": func(mc *contextTestingModuleModuleConfig, m dsl.Module) {
			eventsOut, err := m.ApplyEvents(stdtypes.ListOf(stdevents.NewTestUint64(mc.Self, 8)))
			assert.Nil(t, err)
			assert.Equal(t, 1, eventsOut.Len())

			iter := eventsOut.Iterator()
			sigVerEvent := iter.Next().(*eventpb.Event).Type.(*eventpb.Event_Crypto).Crypto.Type.(*cryptopb.Event_VerifySigs).VerifySigs
			sigVerNodes := sigVerEvent.NodeIds
			assert.Equal(t, 8, len(sigVerNodes))
			sigVerOrigin := cryptopbtypes.SigVerOriginFromPb(sigVerEvent.Origin)

			// send some unrelated events to make sure the context is preserved and does not get overwritten
			_, err = m.ApplyEvents(stdtypes.ListOf(stdevents.NewTestString(mc.Self, "hello")))
			assert.Nil(t, err)
			_, err = m.ApplyEvents(stdtypes.ListOf(stdevents.NewTestUint64(mc.Self, 3)))
			assert.Nil(t, err)
			_, err = m.ApplyEvents(stdtypes.ListOf(stdevents.NewTestUint64(mc.Self, 16), stdevents.NewTestString(mc.Self, "foo")))
			assert.Nil(t, err)

			// construct a response for the signature verification request.
			var nilErr error
			sigsVerifiedEvent := cryptopbevents.SigsVerified(
				/*destModule*/ mc.Self,
				/*origin*/ sigVerOrigin,
				/*nodeIDs*/ types.NodeIDSlice(sigVerNodes),
				/*errors*/ sliceutil.Repeat(nilErr, 8),
				/*allOk*/ true,
			).Pb()

			eventsOut, err = m.ApplyEvents(stdtypes.ListOf(sigsVerifiedEvent))
			assert.Nil(t, err)

			var expectedResponse []stdtypes.Event
			for i := 0; i < 8; i++ {
				expectedResponse = append(expectedResponse, stdevents.NewTestString(mc.Verified, fmt.Sprintf("8: %v verified", i)))
			}
			expectedResponse = append(expectedResponse, stdevents.NewTestUint64(mc.Verified, 8))

			assert.Equal(t, expectedResponse, eventsOut.Slice())

			assert.Panics(t, func() {
				// This reply is sent for the second time.
				// The context should already be disposed of and the module should panic.
				_, _ = m.ApplyEvents(stdtypes.ListOf(sigsVerifiedEvent))
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

func DslSignOrigin(module types.ModuleID, contextID dsl.ContextID) *cryptopbtypes.SignOrigin {
	return &cryptopbtypes.SignOrigin{
		Module: module,
		Type: &cryptopbtypes.SignOrigin_Dsl{
			Dsl: dsl.Origin(contextID),
		},
	}
}

// dsl wrappers (similar to the ones in pkg/dsl/events.go)

func EmitTestingString(m dsl.Module, dest types.ModuleID, s string) {
	dsl.EmitEvent(m, stdevents.NewTestString(dest, s))
}

func EmitTestingUint(m dsl.Module, dest types.ModuleID, u uint64) {
	dsl.EmitEvent(m, stdevents.NewTestUint64(dest, u))
}

func UponTestingString(m dsl.Module, handler func(s string) error) {
	dsl.UponEvent(m, func(ev *stdevents.TestString) error {
		return handler(ev.Value)
	})
}

func UponTestingUint(m dsl.Module, handler func(u uint64) error) {
	dsl.UponEvent(m, func(ev *stdevents.TestUint64) error {
		return handler(ev.Value)
	})
}
