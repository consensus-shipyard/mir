package dsl

import (
	"reflect"

	es "github.com/go-errors/errors"

	cs "github.com/filecoin-project/mir/pkg/contextstore"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type dslModuleImpl struct {
	moduleID                 t.ModuleID
	defaultEventHandler      func(ev events.Event) error
	eventHandlers            map[reflect.Type][]func(ev events.Event) error
	pbEventHandlers          map[reflect.Type][]func(ev events.Event) error
	stateUpdateHandlers      []func() error
	stateUpdateBatchHandlers []func() error
	outputEvents             *events.EventList
	// contextStore is used to store and recover context on asynchronous operations such as signature verification.
	contextStore cs.ContextStore[any]
	// eventCleanupContextIDs is used to dispose of the used up entries in contextStore.
	eventCleanupContextIDs map[ContextID]struct{}
}

// Handle is used to manage internal state of the dsl module.
type Handle struct {
	impl *dslModuleImpl
}

// ContextID is used to address the internal context store of the dsl module.
type ContextID = cs.ItemID

// Module allows creating passive modules in a very natural declarative way.
type Module interface {
	modules.PassiveModule

	// DslHandle is used to manage internal state of the dsl module.
	DslHandle() Handle

	// ModuleID returns the identifier of the module.
	// TODO: consider moving this method to modules.Module.
	ModuleID() t.ModuleID
}

// NewModule creates a new dsl module with a given id.
func NewModule(moduleID t.ModuleID) Module {
	return &dslModuleImpl{
		moduleID:               moduleID,
		defaultEventHandler:    failExceptForInitAndTransport,
		eventHandlers:          make(map[reflect.Type][]func(ev events.Event) error),
		pbEventHandlers:        make(map[reflect.Type][]func(ev events.Event) error),
		outputEvents:           &events.EventList{},
		contextStore:           cs.NewSequentialContextStore[any](),
		eventCleanupContextIDs: make(map[ContextID]struct{}),
	}
}

// DslHandle is used to manage internal state of the dsl module.
func (m *dslModuleImpl) DslHandle() Handle {
	return Handle{m}
}

// ModuleID returns the identifier of the module.
func (m *dslModuleImpl) ModuleID() t.ModuleID {
	return m.moduleID
}

// UponEvent registers an event handler for module m.
// This event handler will be called every time an event of type EvWrapper is received.
// NB: This function works with the (legacy) protoc-generated types and is likely to be
// removed in the future, with UponMirEvent taking its place.
func UponEvent[T events.Event](m Module, handler func(ev T) error) {
	reflectType := reflectutil.TypeOf[T]()

	m.DslHandle().impl.eventHandlers[reflectType] = append(m.DslHandle().impl.eventHandlers[reflectType],
		func(ev events.Event) error {
			return handler(ev.(T))
		})
}

// UponMirEvent registers an event handler for module m.
// This event handler will be called every time an event of type EvWrapper is received.
// NB: this function works with the Mir-generated types.
// For all other types, use the general UponEvent.
// We treat Mir-generated events specially, since all are grouped under a single common *eventpb.Event type
// and we need to make a distinction of different sub-types.
func UponMirEvent[EvWrapper eventpbtypes.Event_TypeWrapper[Ev], Ev any](m Module, handler func(ev *Ev) error) {
	var zeroW EvWrapper
	evWrapperType := zeroW.MirReflect().PbType()

	m.DslHandle().impl.pbEventHandlers[evWrapperType] = append(m.DslHandle().impl.pbEventHandlers[evWrapperType],
		func(ev events.Event) error {
			return handler(eventpbtypes.EventFromPb(ev.(*eventpb.Event)).Type.(EvWrapper).Unwrap())
		})
}

func UponOtherEvent(m Module, handler func(ev events.Event) error) {
	m.DslHandle().impl.defaultEventHandler = handler
}

// UponStateUpdate registers a special type of handler that is invoked after the processing of every event.
// The handler is intended to represent a conditional action: it is supposed to check some predicate on the state
// and perform actions if the predicate is satisfied. This is, however, not enforced in any way, and the handler can
// contain arbitrary code. Use with care, especially if the execution overhead of the handler is not negligible.
// Execution of these handlers might easily have a severe impact on performance.
func UponStateUpdate(m Module, handler func() error) {
	impl := m.DslHandle().impl
	impl.stateUpdateHandlers = append(impl.stateUpdateHandlers, handler)
}

// UponStateUpdates registers a special type of handler that is invoked each time after processing a batch of events.
// It is a less resource-intensive alternative to UponStateUpdate that is useful if intermediate states of the module
// are not relevant (as not all intermediate states of the module are observed by handlers
// registered using UponStateUpdates).
//
// ATTENTION: The handler always called after applying a *batch of* events, not after individual event application.
// This can lead to unintuitive behavior. For example, a counter module maintaining an internal counter variable
// that is incremented on every event could register a state update handler checking for counter == 10.
// Due to (unpredictable) event application batching, the counter can go from 0 to 20
// without having triggered the condition handler, if events 10 and 11 are applied in the same batch.
func UponStateUpdates(m Module, handler func() error) {
	impl := m.DslHandle().impl
	impl.stateUpdateBatchHandlers = append(impl.stateUpdateBatchHandlers, handler)
}

// StoreContext stores the given data and returns an automatically deterministically generated unique id.
// The data can be later recovered or disposed of using this id.
func (h Handle) StoreContext(context any) ContextID {
	return h.impl.contextStore.Store(context)
}

// CleanupContext schedules a disposal of context with the given id after the current batch of events is processed.
// NB: the context cannot be disposed of immediately because there may be more event handlers for this event that may
// need this context.
func (h Handle) CleanupContext(id ContextID) {
	h.impl.eventCleanupContextIDs[id] = struct{}{}
}

// RecoverAndRetainContext recovers the context with the given id and retains it in the internal context store so that
// it can be recovered again later. Only use this function when expecting to receive multiple events with the same
// context. In case of a typical request-response semantic, use RecoverAndCleanupContext.
func (h Handle) RecoverAndRetainContext(id cs.ItemID) any {
	return h.impl.contextStore.Recover(id)
}

// RecoverAndCleanupContext recovers the context with te given id and schedules a disposal of this context after the
// current batch of events is processed.
// NB: the context cannot be disposed of immediately because there may be more event handlers for this event that may
// need this context.
func (h Handle) RecoverAndCleanupContext(id ContextID) any {
	res := h.RecoverAndRetainContext(id)
	h.CleanupContext(id)
	return res
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (m *dslModuleImpl) ImplementsModule() {}

// EmitEvent adds the event to the queue of output events
// NB: This function works with the (legacy) protoc-generated types and is likely to be
// removed in the future, with EmitMirEvent taking its place.
func EmitEvent(m Module, ev events.Event) {
	m.DslHandle().impl.outputEvents.PushBack(ev)
}

// EmitMirEvent adds the event to the queue of output events
// NB: this function works with the Mir-generated types.
// For the (legacy) protoc-generated types, EmitEvent can be used.
func EmitMirEvent(m Module, ev *eventpbtypes.Event) {
	m.DslHandle().impl.outputEvents.PushBack(ev.Pb())
}

// ApplyEvents applies a list of input events to the module, making it advance its state
// and returns a (potentially empty) list of output events that the application of the input events results in.
func (m *dslModuleImpl) ApplyEvents(evs *events.EventList) (*events.EventList, error) {
	// Run event handlers.
	iter := evs.Iterator()
	for ev := iter.Next(); ev != nil; ev = iter.Next() {

		var handlers []func(event events.Event) error
		var ok bool
		switch e := ev.(type) {
		case *eventpb.Event:
			// TODO: This special treatment of PB events is ugly.
			// It stems from the DSL module's support for generated code.
			// Adapt the generated (and the DSL module) code so this special treatment is not necessary any more.
			handlers, ok = m.pbEventHandlers[reflect.TypeOf(e.Type)]
		default:
			handlers, ok = m.eventHandlers[reflect.TypeOf(ev)]
		}

		// If no specific handler was defined for this event type, execute the default handler.
		if !ok {
			err := m.defaultEventHandler(ev)
			if err != nil {
				return nil, err
			}
		}

		// Execute all handlers registered for the event type.
		for _, h := range handlers {
			err := h(ev)
			if err != nil {
				return nil, err
			}
		}

		// Execute state update handlers.
		for _, h := range m.stateUpdateHandlers {
			err := h()
			if err != nil {
				return nil, err
			}
		}
	}

	// Run batched state update handlers.
	for _, condition := range m.stateUpdateBatchHandlers {
		err := condition()

		if err != nil {
			return nil, err
		}
	}

	// Cleanup used up context store entries
	if len(m.eventCleanupContextIDs) > 0 {
		for id := range m.eventCleanupContextIDs {
			m.contextStore.Dispose(id)
		}
		m.eventCleanupContextIDs = make(map[ContextID]struct{})
	}

	outputEvents := m.outputEvents
	m.outputEvents = &events.EventList{}
	return outputEvents, nil
}

// The failExceptForInit returns an error for every received event type except for Init.
// For convenience, if this is used as the default event handler,
// it is not considered an error to not have handlers for the Init event.
func failExceptForInit(ev events.Event) error { //nolint
	pbev, ok := ev.(*eventpb.Event)
	if !ok {
		return es.Errorf("unknown event type '%T'", ev)
	}

	if reflect.TypeOf(pbev.Type) == reflectutil.TypeOf[*eventpb.Event_Init]() {
		return nil
	}
	return es.Errorf("unknown event type '%T'", pbev.Type)
}

// The failExceptForInitAndTransport is just like failExceptForInit except
// that if the module does not tolerate a transport event, it gracefully
// ignores it. This prevents external nodes from crashing the node by sending
// transport events to modules that do not tolerate it.
func failExceptForInitAndTransport(ev events.Event) error {
	pbev, ok := ev.(*eventpb.Event)
	if !ok {
		return es.Errorf("unknown event type '%T'", ev)
	}

	if reflect.TypeOf(pbev.Type) == reflectutil.TypeOf[*eventpb.Event_Transport]() {
		return nil
	}

	return failExceptForInit(ev)
}
