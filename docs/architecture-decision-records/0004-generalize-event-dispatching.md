# Generalize Event Dispatching

* Status: accepted
* Deciders: @matejpavlovic, @sergefdrv, @dnkolegov
* Date: 2022-06-01 (last update)

This is a natural adaptation to changes proposed in [ADR-0003 - Generalize modules](0003-generalize-modules.md)

## Context and Problem Statement

The more flexible approach to modules introduced in [ADR-0003 - Generalize modules](0003-generalize-modules.md)
cannot be used to its full potential while each module is defined statically.
E.g., Mir is still constrained to use exactly one `App` module, one `Net` module, one `Protocol` module, etc.,
even if those have a more general interface of either `ActiveModule` or `PassiveModule`.

Event types still need to be routed to the appropriate modules by the dispatching code based on event type,
requiring the dispatching code to know about each event type and which module it belongs to.
Thus, when creating new event types, the use of Mir still needs to modify the dispatching code,
statically "hard-coding" this information in Mir.

## Considered Options

* **Module naming.**
  First, associate each module with a string name. Instead of the current static approach of instantiating modules
  ```go
  Modules{
    App:      newAppModule()
    Net:      newNetModule()
    Protocol: newProtocolModule()
  }
  ```
  instantiate the modules as follows
  ```go

  type Module interface {
    ImplementsModule()
  }
  
  type PassiveModule interface {
    Module
    // ...
  }
  
  type ActiveModule interface {
    Module
    // ...
  }
  
  type Modules map[string]Module

  var modules = map[string]Module{
    "app":      newAppModule()
    "protocol": newProtocolModule()
    "net":      newNetModule()
  }
  ```
  The exact interface of `ActiveModule` and `PassiveModule` (replaced by `// ...`) in the snippet above is described in
  [ADR-0003 - Generalize modules](0003-generalize-modules.md).

  Second, add a `Destination` field to events, extending the `Event` Protobuf definition as follows
  ```protobuf
  // Event represents a state event to be injected into the state machine
  message Event {
    oneof type {
      // ...
    }
  
    // Identifier of the module this Event should be routed to.
    string Destination = 200;

    // A list of follow-up events to process after this event has been processed.
    // This field is used if events need to be processed in a particular order.
    // For example, a message sending event must only be processed
    // after the corresponding entry has been persisted in the write-ahead log (WAL).
    // In this case, the WAL append event would be this event
    // and the next field would contain the message sending event.
    repeated Event next = 100;
  }
  
  ```
  This field will hold the name of the module (key in the above map)
  that informs the dispatching code where to route the event.

## Decision Outcome

Chosen option: Module naming, because it addresses the issue well,
can be implemented rather easily, and is the only proposed one.

### Positive Consequences <!-- optional -->

* Truly enables the flexibility in designing general modules and composing them at will.
* Removes the requirement of hard-coding the dispatching of every single event type in the Mir core code.
* Opens the door to further improvements, in particular fully dynamic module management through the existing modules.
  For example, the system can easily be extended in the future to give a module the means 
  for dynamically creating and removing other modules.

### Negative Consequences <!-- optional -->

* Involves some refactoring of existing code, especially adding explicit destinations to already existing events.
