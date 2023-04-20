package factorymodule

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/factorypb"
	factorypbevents "github.com/filecoin-project/mir/pkg/pb/factorypb/events"
	factorypbtypes "github.com/filecoin-project/mir/pkg/pb/factorypb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	tp "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/testlogger"
)

const (
	echoFactoryID = tp.ModuleID("echoFactory")
)

type echoModule struct {
	t      *testing.T
	id     tp.ModuleID
	prefix string
}

func (em *echoModule) ImplementsModule() {}

func (em *echoModule) ApplyEvents(evts *events.EventList) (*events.EventList, error) {
	return modules.ApplyEventsSequentially(evts, em.applyEvent)
}

func (em *echoModule) applyEvent(event *eventpb.Event) (*events.EventList, error) {

	// Convenience variable
	destModuleID := tp.ModuleID(event.DestModule)

	assert.Equal(em.t, em.id, destModuleID)
	switch e := event.Type.(type) {
	case *eventpb.Event_Init:
		return events.ListOf(events.TestingString(destModuleID.Top(), string(em.id)+" Init")), nil
	case *eventpb.Event_TestingString:
		return events.ListOf(events.TestingString(destModuleID.Top(), em.prefix+e.TestingString.GetValue())), nil
	default:
		return nil, fmt.Errorf("unknown echo module event type: %T", e)
	}
}

func newEchoFactory(t *testing.T, logger logging.Logger) *FactoryModule {
	return New(
		echoFactoryID,
		DefaultParams(func(id tp.ModuleID, params *factorypb.GeneratorParams) (modules.PassiveModule, error) {
			return &echoModule{
				t:      t,
				id:     id,
				prefix: params.Type.(*factorypb.GeneratorParams_EchoTestModule).EchoTestModule.Prefix,
			}, nil
		}),
		logger)
}

func TestFactoryModule(t *testing.T) {
	defer goleak.VerifyNone(t)

	var echoFactory modules.PassiveModule
	logger := testlogger.New()

	testCases := map[string]func(t *testing.T){

		"00 Instantiate": func(t *testing.T) {
			echoFactory = newEchoFactory(t, logger)
			evOut, err := echoFactory.ApplyEvents(events.ListOf(factorypbevents.NewModule(
				echoFactoryID,
				echoFactoryID.Then("inst0"),
				0,
				EchoModuleParams("Inst 0: "),
			).Pb()))
			assert.NoError(t, err)
			assert.Equal(t, 1, evOut.Len())
			assert.Equal(t, echoFactoryID.Pb(), evOut.Slice()[0].DestModule)
			assert.Equal(t,
				string(echoFactoryID.Then("inst0"))+" Init",
				evOut.Slice()[0].Type.(*eventpb.Event_TestingString).TestingString.GetValue(),
			)
		},

		"01 Invoke": func(t *testing.T) {
			evOut, err := echoFactory.ApplyEvents(events.ListOf(events.TestingString(
				echoFactoryID.Then("inst0"),
				"Hi!"),
			))
			assert.NoError(t, err)
			assert.Equal(t, 1, evOut.Len())
			assert.Equal(t,
				"Inst 0: Hi!",
				evOut.Slice()[0].Type.(*eventpb.Event_TestingString).TestingString.GetValue(),
			)
		},

		"02 Instantiate many": func(t *testing.T) {
			for i := 1; i <= 5; i++ {
				evOut, err := echoFactory.ApplyEvents(events.ListOf(factorypbevents.NewModule(
					echoFactoryID,
					echoFactoryID.Then(tp.ModuleID(fmt.Sprintf("inst%d", i))),
					tt.RetentionIndex(i),
					EchoModuleParams(fmt.Sprintf("Inst %d: ", i)),
				).Pb()))
				assert.NoError(t, err)
				assert.Equal(t, 1, evOut.Len())
				assert.Equal(t, echoFactoryID.Pb(), evOut.Slice()[0].DestModule)
				assert.Equal(t,
					string(echoFactoryID.Then(tp.ModuleID(fmt.Sprintf("inst%d", i))))+" Init",
					evOut.Slice()[0].Type.(*eventpb.Event_TestingString).TestingString.GetValue(),
				)
			}
		},

		"03 Invoke many": func(t *testing.T) {
			evList := events.EmptyList()
			for i := 5; i >= 0; i-- {
				evList.PushBack(events.TestingString(
					echoFactoryID.Then(tp.ModuleID(fmt.Sprintf("inst%d", i))),
					"Hi!"),
				)
			}
			evOut, err := echoFactory.ApplyEvents(evList)
			assert.NoError(t, err)
			assert.Equal(t, 6, evOut.Len())

			sortedOutput := evOut.Slice()

			sort.Slice(sortedOutput, func(i, j int) bool {
				return sortedOutput[i].Type.(*eventpb.Event_TestingString).TestingString.GetValue() <
					sortedOutput[j].Type.(*eventpb.Event_TestingString).TestingString.GetValue()
			})

			for i := 0; i <= 5; i++ {
				assert.Equal(t, echoFactoryID.Pb(), sortedOutput[0].DestModule)
				assert.Equal(t,
					fmt.Sprintf("Inst %d: Hi!", i),
					sortedOutput[i].Type.(*eventpb.Event_TestingString).TestingString.GetValue(),
				)
			}
		},

		"04 Wrong event type": func(t *testing.T) {
			wrongEvent := events.TestingUint(echoFactoryID, 42)
			evOut, err := echoFactory.ApplyEvents(events.ListOf(wrongEvent))
			if assert.Error(t, err) {
				assert.Equal(t, fmt.Errorf("unsupported event type for factory module: %T", wrongEvent.Type), err)
			}
			assert.Nil(t, evOut)
		},

		"05 Wrong submodule ID": func(t *testing.T) {
			wrongEvent := events.TestingUint(echoFactoryID.Then("non-existent-module"), 42)
			evOut, err := echoFactory.ApplyEvents(events.ListOf(wrongEvent))
			assert.NoError(t, err)
			assert.Equal(t, 0, evOut.Len())
			logger.CheckFirstEntry(t, logging.LevelDebug, "Ignoring submodule event. Destination module not found.",
				"moduleID", echoFactoryID.Then("non-existent-module"),
				"eventType", fmt.Sprintf("%T", wrongEvent.Type),
				"eventValue", fmt.Sprintf("%v", wrongEvent.Type))
			logger.CheckEmpty(t)
		},

		"06 Garbage-collect some": func(t *testing.T) {
			evOut, err := echoFactory.ApplyEvents(events.ListOf(factorypbevents.GarbageCollect(
				echoFactoryID,
				3,
			).Pb()))
			assert.NoError(t, err)
			assert.Equal(t, 0, evOut.Len())
		},

		"07 Invoke garbage-collected": func(t *testing.T) {
			evList := events.EmptyList()
			for i := 5; i >= 0; i-- {
				evList.PushBack(events.TestingString(
					echoFactoryID.Then(tp.ModuleID(fmt.Sprintf("inst%d", i))),
					"Hi!"),
				)
			}
			evSlice := evList.Slice()
			evOut, err := echoFactory.ApplyEvents(evList)
			assert.NoError(t, err)
			assert.Equal(t, 3, evOut.Len())

			sortedOutput := evOut.Slice()

			sort.Slice(sortedOutput, func(i, j int) bool {
				return sortedOutput[i].Type.(*eventpb.Event_TestingString).TestingString.GetValue() <
					sortedOutput[j].Type.(*eventpb.Event_TestingString).TestingString.GetValue()
			})

			for i := 0; i < 3; i++ {
				logger.CheckAnyEntry(t, logging.LevelDebug, "Ignoring submodule event. Destination module not found.",
					"moduleID", echoFactoryID.Then(tp.ModuleID(fmt.Sprintf("inst%d", i))),
					"eventType", fmt.Sprintf("%T", evSlice[i].Type),
					"eventValue", fmt.Sprintf("%v", evSlice[i].Type),
				)
			}

			for i := 3; i <= 5; i++ {
				assert.Equal(t, echoFactoryID.Pb(), sortedOutput[i-3].DestModule)
				assert.Equal(t,
					fmt.Sprintf("Inst %d: Hi!", i),
					sortedOutput[i-3].Type.(*eventpb.Event_TestingString).TestingString.GetValue(),
				)
			}

			logger.CheckEmpty(t)
		},
	}

	maputil.IterateSorted(testCases, func(testName string, testFunc func(t *testing.T)) bool {
		t.Run(testName, testFunc)
		return true
	})
}

func EchoModuleParams(prefix string) *factorypbtypes.GeneratorParams {
	return &factorypbtypes.GeneratorParams{
		Type: &factorypbtypes.GeneratorParams_EchoTestModule{
			EchoTestModule: &factorypbtypes.EchoModuleParams{
				Prefix: prefix,
			},
		}}
}
