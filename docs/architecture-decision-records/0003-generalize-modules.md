# Generalize Modules

* Status: accepted
* Deciders: @matejpavlovic, @sergefdrv, @dnkolegov
* Date: 2022-05-31 (last update)

## Context and Problem Statement

The current architecture defines several kinds of modules, each with a module-specific interface.
There is also exactly one of each kind of module
(i.e., one `Protocol` module instance, one `App`, one `ClientTracker`, etc.).
The event handling code then takes each module and wraps it almost the same simple interface
(consuming input events and producing output events) using a module-specific wrapper.
This wrapper almost always ends up being just a simple de-multiplexer that,
based on the incoming event type, calls the appropriate module-specific method of the module.

This approach:
- Pollutes the Node's event handling logic by the above-mentioned wrappers (which can be considered boilerplate code).
  Whenever a new type of module is being defined, the core Mir code needs to be modified
  to add the corresponding wrapper.
- Unnecessarily constrains the user of Mir to use exactly one of each type of module, 
  even if a protocol might naturally be split up in several modules working in parallel.
- The selection of the types of modules supported is tailored to a SMR application, but not necessarily to other uses.

## Considered Options

* **Active and passive module types.**
  Generalize the concept of modules, such that:
  - Modules have a unified interface for processing events,
    where each module internally inspects the incoming events and decides on the appropriate action.
    The `Protocol` and some other modules already do have such a unified interface - use it for other modules as well.
  - Instead of dedicated slots for each type of module, use a variable number of generic modules of only 2 types:
    - `PassiveModule`:
      A module that simply defines the logic for transforming input events to output events whenever invoked by Mir.
      Never injects external events in the system.
      A passive module has the following interface (corresponding to some existing modules, like `Protocol`):
      ```go
      type PassiveModule interface {

        // ApplyEvent applies a single input event to the module, making it advance its state
        // and return a (potentially empty) list of output events that the application of the input event results in.
        ApplyEvent(event *eventpb.Event) *events.EventList
        
        // Status returns the current state of the module.
        // Mostly for debugging purposes.
        Status() (s Status, err error) // The Status type is a placeholder not intended to be defined by this ADR.
      }
      ```
    - `ActiveModule`:
      The `ActiveModule` asynchronously reads input events from a channel
      and writes produced events to an output channel.
      
      In an `ActiveModule`, the produced events might or might not depend on the input events.
      This makes the `ActiveModule` suitable for injecting external events
      (e.g., reception of a message over the network) in the Node.
      To prevent the module from flooding the Node with too many external events,
      the node might, at any time, stop processing the `ActiveModule`s output events
      for an arbitrarily long time period.
      This is expected to happen when internal node buffers become full of unprocessed events
      and the node needs to wait until they free up.
      ```go
      type ActiveModule interface {

        // Run performs the module's processing.
        // It usually reads events from eventsIn, processes them, and writes new events to eventsOut.
        // Note, however, that the implementation may produce output events even without receiving any input.
        // Before Run is called, ActiveModule guarantees not to read or write,
        // respectively, from eventsIn and to eventsOut.
        //
        // In general, Run is expected to block until explicitly asked to stop
        // (through a module-specific event or by canceling ctx)
        // and thus should most likely be run in its own goroutine.
        // Run must return after ctx is canceled.
        //
        // Note that the node does not guarantee to always read events from eventsOut.
        // The node might decide at any moment to stop reading from eventsOut for an arbitrary amount of time
        // (e.g. if the Node's internal event buffers become full and the Node needs to wait until they free up).
        //
        // The implementation of Run is required to always read from eventsIn.
        // I.e., when an event is written to eventsIn, Run must eventually read it, regardless of the module's state
        // or any external factors
        // (e.g., the processing of events must not depend on some other goroutine reading from eventsOut).
        Run(ctx context.Context, eventsIn <-chan *events.EventList, eventsOut chan<- *events.EventList) error
		
        // Status returns the current state of the module.
        // If Run has been called and has not returned when calling Status,
        // there is no guarantee about which input events are taken into account
        // when creating the snapshot of the returned state.
        // Mostly for debugging purposes.
        Status() (s *statuspb.ProtocolStatus, err error)
      }
      ```
  Note that this approach still requires the user of Mir to define the module-specific events in Mir,
  as well as the way they are dispatched.
  Generalizing event dispatching is also desirable, but can be treated as a separate issue in a different ADR.
## Decision Outcome

Chosen option: Active and passive modules, because it addresses the issue well,
can easily be implemented, and is the only proposed one.
Moreover, almost as a side effect, it explicitly addresses
the concern of growing event buffers in the Node's event loop. 

### Positive Consequences <!-- optional -->

* Makes Mir's code more uniform, simpler, and cleaner.
* Increases the expressiveness of mir by giving the user significantly more flexibility in how to compose the modules.
* Does not impact the internal logic of the existing modules
  (only the way they are attached to the node needs to be modified)

### Negative Consequences <!-- optional -->

* Requires some implementation overhead.
