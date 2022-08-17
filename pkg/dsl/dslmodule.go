package dsl

import (
	"fmt"
	"reflect"

	cs "github.com/filecoin-project/mir/pkg/contextstore"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	eventpbtypes "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type dslModuleImpl struct {
	moduleID            t.ModuleID
	defaultEventHandler func(ev *eventpb.Event) error
	eventHandlers       map[reflect.Type][]func(ev *eventpb.Event) error
	conditionHandlers   []func() error
	outputEvents        *events.EventList
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
		defaultEventHandler:    failExceptForInit,
		eventHandlers:          make(map[reflect.Type][]func(ev *eventpb.Event) error),
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
func UponEvent[EvWrapper eventpb.Event_TypeWrapper[Ev], Ev any](m Module, handler func(ev *Ev) error) {
	evWrapperType := reflectutil.TypeOf[EvWrapper]()

	m.DslHandle().impl.eventHandlers[evWrapperType] = append(m.DslHandle().impl.eventHandlers[evWrapperType],
		func(ev *eventpb.Event) error {
			return handler(ev.Type.(EvWrapper).Unwrap())
		})
}

// UponMirEvent registers an event handler for module m.
// This event handler will be called every time an event of type EvWrapper is received.
// NB: this function works with the Mir-generated types.
// For the (legacy) protoc-generated types, UponEvent can be used.
func UponMirEvent[EvWrapper eventpbtypes.Event_TypeWrapper[Ev], Ev any](m Module, handler func(ev *Ev) error) {
	var zeroW EvWrapper
	evWrapperType := zeroW.MirReflect().PbType()

	m.DslHandle().impl.eventHandlers[evWrapperType] = append(m.DslHandle().impl.eventHandlers[evWrapperType],
		func(ev *eventpb.Event) error {
			return handler(eventpbtypes.EventFromPb(ev).Type.(EvWrapper).Unwrap())
		})
}

func UponOtherEvent(m Module, handler func(ev *eventpb.Event) error) {
	m.DslHandle().impl.defaultEventHandler = handler
}

// UponCondition registers a special type of handler that will be invoked each time after processing a batch of events.
// The handler is assumed to represent a conditional action: it is supposed to check some predicate on the state
// and perform actions if the predicate is satisfied.
func UponCondition(m Module, handler func() error) {
	impl := m.DslHandle().impl
	impl.conditionHandlers = append(impl.conditionHandlers, handler)
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
func EmitEvent(m Module, ev *eventpb.Event) {
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
		handlers, ok := m.eventHandlers[reflect.TypeOf(ev.Type)]

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
	}

	// Run condition handlers.
	for _, condition := range m.conditionHandlers {
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
func failExceptForInit(ev *eventpb.Event) error {
	if reflect.TypeOf(ev.Type) == reflectutil.TypeOf[*eventpb.Event_Init]() {
		return nil
	}
	return fmt.Errorf("unknown event type '%T'", ev.Type)
}
