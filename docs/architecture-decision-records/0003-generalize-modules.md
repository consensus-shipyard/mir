# Generalize Modules

* Status: proposed
* Deciders: @matejpavlovic, @sergefdrv, @dnkolegov
* Date: 2022-05-28 (last update)

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

* **Active and passive modules.**
  Generalize the concept of modules, such that:
  - Modules have a unified interface for processing events,
    where each module internally inspects the incoming events and decides on the appropriate action.
    The `Protocol` and some other modules already do have such a unified interface - use it for other modules as well.
  - Instead of dedicated slots for each type of module, use a variable number of generic modules of only 2 types:
    - `PassiveModule`:
      A module that simply transforms input events to output events whenever invoked by Mir.
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
      Like the `PassiveModule`, the `ActiveModule` consumes input events. However,
      unlike the `PassiveModule`, the `ActiveModule` does not synchronously return receive input
      and return output in a function call.
      Instead, the `ActiveModule` asynchronously reads input events from a channel
      and writes produced events to an output channel.
      In an `ActiveModule`, the produced events might or might not depend on the input events.
      This makes the `ActiveModule` suitable for injecting external events
      (e.g., reception of a message over the network) in the Node.
      To prevent the module from flooding the Node with too many external events,
      the node might ask it to temporarily back off and function essentially as a `PassiveModule`.
      ```go
      type ActiveModule interface {

        // Run performs the module's processing.
        // It usually reads events from eventsIn, processes them, and writes new events to eventsOut.
        // Note, however, that the implementation may produce output events even without receiving any input.
        // Before Run is called, ActiveModule guarantees not to read or write,
        // respectively, from eventsIn and to eventsOut.
        // Run must return after ctx is canceled.
        Run(ctx context.Context, eventsIn <-chan *events.EventList, eventsOut chan<- *events.EventList) error

        // BackOff asks the module to back off, i.e., to not produce new events
        // except in immediate reaction to input events, until ctx is canceled.
        // After BackOff is called and before ctx is canceled,
        // the module will only produce a bounded number of output events and only in reaction to processing input events.
        // In other words, if BackOff has been called, ctx has not been canceled, and the input channel is not written to,
        // the module will eventually stop producing output events.
        //
        // This function is necessary for guaranteeing that the internal Node event buffers do not grow indefinitely.
        // This could potentially happen when external events are injected into the system through active modules
        // faster than the system can process them.
        // In such a case, the Node asks all active modules to back off
        // until enough events in the internal buffers have been processed.
        BackOff(ctx context.Context)
        
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
