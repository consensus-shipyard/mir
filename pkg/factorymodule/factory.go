package factorymodule

import (
	"fmt"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/factorymodulepb"
	t "github.com/filecoin-project/mir/pkg/types"
)

// TODO: Add support for active modules as well.

type moduleGenerator func(id t.ModuleID, params *factorymodulepb.GeneratorParams) (modules.PassiveModule, error)

type FactoryModule struct {
	ownID     t.ModuleID
	generator moduleGenerator

	submodules      map[t.ModuleID]modules.PassiveModule
	moduleRetention map[t.RetentionIndex][]t.ModuleID
	retIdx          t.RetentionIndex

	logger logging.Logger
}

func NewFactoryModule(id t.ModuleID, generator moduleGenerator, logger logging.Logger) *FactoryModule {
	return &FactoryModule{
		ownID:     id,
		generator: generator,

		submodules:      make(map[t.ModuleID]modules.PassiveModule),
		moduleRetention: make(map[t.RetentionIndex][]t.ModuleID),
		retIdx:          0,

		logger: logger,
	}
}

func (fm *FactoryModule) ImplementsModule() {
}

func (fm *FactoryModule) ApplyEvents(evts *events.EventList) (*events.EventList, error) {
	// TODO: Perform event processing in parallel (applyEvent will need to be made thread-safe).
	//       The idea is to have one internal thread per submodule, distribute the events to them through channels,
	//       and wait until all are processed.
	return modules.ApplyEventsSequentially(evts, fm.applyEvent)
}

func (fm *FactoryModule) applyEvent(event *eventpb.Event) (*events.EventList, error) {

	if t.ModuleID(event.DestModule) == fm.ownID {
		switch e := event.Type.(type) {
		case *eventpb.Event_Init:
			return events.EmptyList(), nil // Nothing to do at initialization.
		case *eventpb.Event_Factory:
			switch e := e.Factory.Type.(type) {
			case *factorymodulepb.Factory_NewModule:
				return fm.applyNewModule(e.NewModule)
			case *factorymodulepb.Factory_GarbageCollect:
				return fm.applyGarbageCollect(e.GarbageCollect)
			default:
				return nil, fmt.Errorf("unsupported factory event subtype: %T", e)
			}
		default:
			return nil, fmt.Errorf("unsupported event type for factory module: %T", e)
		}
	}
	return fm.forwardEvent(event)
}

func (fm *FactoryModule) applyNewModule(newModule *factorymodulepb.NewModule) (*events.EventList, error) {

	// Convenience variables
	id := t.ModuleID(newModule.ModuleId)
	retIdx := t.RetentionIndex(newModule.RetentionIndex)

	// The new module's ID must have the factory's ID as a prefix.
	if id.Top() != fm.ownID {
		return nil, fmt.Errorf("submodule (%v) must have the factory's ID (%v) as a prefix", id, fm.ownID)
	}

	// Skip creation of submodules that should have been already garbage-collected.
	if retIdx < fm.retIdx {
		fm.logger.Log(logging.LevelWarn, "Ignoring new module instantiation with low retention index.",
			"moduleID", id, "currentRetIdx", fm.retIdx, "moduleRetIdx", retIdx)
		return events.EmptyList(), nil
	}

	// Create new instance of the submodule.
	if submodule, err := fm.generator(id, newModule.Params); err == nil {
		fm.submodules[id] = submodule
	} else {
		return nil, err
	}

	// Assign the newly created submodule to its retention index.
	fm.moduleRetention[retIdx] = append(fm.moduleRetention[retIdx], id)

	// Initialize new submodule and return the resulting events.
	return fm.submodules[id].ApplyEvents(events.ListOf(events.Init(id)))
}

func (fm *FactoryModule) applyGarbageCollect(gc *factorymodulepb.GarbageCollect) (*events.EventList, error) {
	// While the new retention index is larger than the current one
	for t.RetentionIndex(gc.RetentionIndex) > fm.retIdx {

		// Delete all modules associated with the current retention index.
		for _, mID := range fm.moduleRetention[fm.retIdx] {
			delete(fm.submodules, mID)
		}

		// Increase current retention index.
		delete(fm.moduleRetention, fm.retIdx)
		fm.retIdx++
	}

	return events.EmptyList(), nil
}

func (fm *FactoryModule) forwardEvent(event *eventpb.Event) (*events.EventList, error) {

	// Convenience variable.
	mID := t.ModuleID(event.DestModule)

	var submodule modules.PassiveModule
	var ok bool
	if submodule, ok = fm.submodules[mID]; !ok {
		fm.logger.Log(logging.LevelWarn, "Ignoring submodule event. Destination module not found.", "moduleID", mID)
		return events.EmptyList(), nil
	}

	// TODO: This might be inefficient. Try to not forward events one by one.
	//       Especially once parallel processing is supported.
	return submodule.ApplyEvents(events.ListOf(event))
}
