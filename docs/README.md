# Mir Library Overview

Mir is a framework for implementing, debugging, and analyzing distributed protocols.
This document describes the architecture and inner workings of the Mir library implementing the framework.
We also keep a [log of design choices](/docs/architecture-decision-records) using Architecture Decision Records (ADRs).
In a nutshell, Mir presents the abstractino of an *Node* containing several *Modules*
that interact with each other by producing and consuming *Events*.
A group of modules configured to work together is called a *System*.

## Node

The top-level component of the library is the [Node](/node.go).
The user of the library instantiates a `Node` object which serves as the interface to the Mir library.
Here we focus on the inner workings of a node.
A node contains named [modules](#modules) (specified at node instantiation)
that interact through events (borrowing concepts from the actor model).
The node implements an event loop that orchestrates the processing of those events.


## Events

`Event`s are serializable objects (implemented using Protocol Buffers) that internally serve as messages
(not to be confused with [protocol messages](/protos/messagepb/messagepb.proto)) passed between the node's modules.
Events are created either by the node itself upon input from outside
(the library user calling the `Node.InjectEvent()` method) or by the node's modules.
Each event has a `DestModule` field containing the identifier of the module that should process it.
The node's modules, independently of each other, take events as input, process them,
and produce new events as output.
All created events are stored in the [`EventBuffer`](/eventbuffer.go) within the node
(the `EventBuffer` contains a separate FIFO queue for each module present in the node),
from where they are dispatched (by the `Node.process()` method) to the appropriate module for processing.
This processing creates new events which are in turn added to the `EventBuffer`, and so on.

For debugging purposes, all events can be recorded using an event Interceptor ([see below](#interceptor))
and inspected or replayed in debug mode using the [mircat](/cmd/mircat/) utility.

### Follow-up events

In certain cases, it is necessary that certain events are processed only after other events have been processed.
For example, before sending a protocol message and delivering a request batch to the application,
the protocol state might need to be persisted in a write-ahead log (WAL)
to prevent unwanted behavior in case of a crash and recovery.
In such a case, the protocol module creates three event: one for persisting the protocol state
(to be processed by the WAL module), one for sending a protocol message (to be processed by the Net module),
and one for delivering the protocol batch to the application (to be processed by the App module).
Since the WAL, the Net, and the App modules work independently and in parallel,
there is no guarantee about which of those modules will be the first to process events.

An `Event` thus contains a `Next` field, pointing to a slice of follow-up events
that can only be processed after the event itself has been processed by the appropriate module.
In the example above, the events for the Net and App module would be included (as a 2-element list)
in the `Next` field of the "parent" WAL event.

When a module (the WAL module in this case) processes an event, it strips off all follow-up events and
only processes the parent event.
It then adds all follow-up events to its output, as if they resulted from processing the parent event.
In the example case, the events for the Net and App modules will be added to the event buffer at the same time,
without any guarantee about the order of their processing.
Follow-up events can be arbitrarily nested to express more complex dependency trees.

Note that a follow-up event is not guaranteed to be processed immediately after the parent event.
Mir does not even guarantee that a follow-up event will be processed before other events emitted by the same module.

### Event representation

Mir uses Protocol Buffers (Protobuf) to represent events.
Protobuf definitions are located in the [/protos](/protos) directory,
with the `Event` object ("message" in protobuf terminology)
defined in [/protos/eventpb/eventpb.proto](/protos/eventpb/eventpb.proto).

The various types of events are defined in the `Event` object using the `oneof` construct.
Whenever new event types are introduced, the `Event` object definition must be updated accordingly.
If new `.proto` files are added, corresponding lines must be added to [/protos/generate.go](/protos/generate.go).

Whenever any Protobuf definitions are updated or added, the corresponding Go code must be regenerated
(by running `make generate` from Mir the root directory).

> Note: This approach is not ideal, as adding new event types requires modifying the `Event` object
> defined in the Mir library codebase.
> Moreover, for historical reasons, there is no clean approach to defining event types in general.
> This is likely to change in the future.

## Modules

A module is a component of the node that consumes, processes, and produces events.
It has to implement the [`Module`](/pkg/modules/modules.go) interface and, depending on its type,
the [`PassiveModule`](/pkg/modules/passivemodule.go) or [`AcingModule`](/pkg/modules/activemodule.go) interface.
The two types of modules are described in more detail later in this section.

At node instantiation, the user (i.e. the programmer using Mir)
specifies a set of named modules that will make up the node in a [`Modules`](/pkg/modules/modules.go) map, e.g.:
```go
    // Example Node instantiation 
	node, err := mir.NewNode(
		&modules.Modules{
			// ...
			"app":      NewChatApp(),
			"protocol": TOBProtocol,
			"net":      grpcNetworking,
			"crypto":   ecdsaCrypto,
		}, 
		/* some more technical arguments ... */
	)
```
Here the `NewChatApp()`, `TOBProtocol`, `grpcNetworking`, and `ecdsaCrypto` are instances of modules
associated with the IDs "app", "protocol", "net", and "crypto", respectively.
The user is free to provide own implementations of these modules,
but Mir already comes bundled with several module implementations out of the box.

### Module IDs

Each module is identified by a unique [`ModuleID`](/pkg/types/moduleid.go)
(currently represented as a string) within its node.
Note however, that it is common that two different instances of a module exist under the same ID on two different nodes.
A module ID is composed of a top-level identifier (itself a valid `ModuleID`) and an optional suffix.
When routing an event to a module, the node implementation only considers the top-level identifier
in the event's `DestModule` field and ignores the suffix.
This allows for implementation of submodules (modules within modules),
where the top-level module can decide to further dispatch an event based on the suffix of the ID.
See [`ModuleID`](/pkg/types/moduleid.go) for details on constructing module IDs and accessing their different parts.

### Module configuration

While the configuration of a modules is entirely module-specific, the convention is to use two separate data structures:
- `ModuleConfig` only specifies identifiers of other modules 
  that the module interacts with or otherwise needs to be aware of.
  E.g., a module implementing the logic of a protocol that needs to interact with a crypto module
  for the computation of signatures might include a `Crypto` field it its `ModuleConfig`
  that needs to be set to the ID of a module to which events requesting signature computation can be sent.
  Each module implementation defines its own `ModuleConfig` struct in its package, containing the appropriate fields.
- `ModuleParams` defines all parameters specific to the module's implemented logic
  such as various buffer capacities or timeouts for protocol-specific actions.

> Note: Even some (legacy) modules within the Mir library itself might not follow this convention.

### Passive and active modules

#### Passive module

The `PassiveModule` is the simpler type of module.
It defines logic for transforming input events into output events.
It passively waits to be invoked by the node (by the node calling the `PassiveModule.ApplyEvents` method)
with a list of events to process.
When invoked, it processes the given events, possibly updating its internal state,
and returns a (potentially empty) list of new events.

Passive modules are useful for stateless logic, e.g., computation of hashes and signatures.
They are also useful for holding state and performing transformations on it in response to external events,
such as the protocol logic.

#### Active module

An `ActiveModule` also consumes, processes, and produces events.
However, unlike the `PassiveModule` which only produces output events synchronously as a reaction to input events,
an `ActiveModule` can also produce output events on its own, without being triggered by an input event.
Instead of returning events from an invocation of the `ActiveModule.ApplyEvents` method,
the `ActiveModule` exposes a channel to which new events can be written any time.

Active modules are useful for receiving input from the outside world such as network messages, user input, or timeouts.

#### Example: usage of active and passive modules

1. The implementation of a networking module (active) receives a message over the network.
2. The networking module (spontaneously, from the point of view of the node)
   writes a `MessageReceived` event to its output channel.
3. The node implementation reads the event from the channel
   and dispatches it to the corresponding protocol module (passive).
4. The protocol module (as a reaction to the `MessageReceived` event)
   emits a `TimeoutRequest` event to the timer module (active).
5. The timer module implementation sets up an internal timeout using the operating system's timer mechanism.
6. After the timer expires, the timer module (spontaneously, from the point of view of the node)
   writes a `Timeout` event to its output channel.
7. The node implementation reads the event from the channel and dispatches it to the protocol module
   which can react to it and possibly (synchronously) emits other events.

### Special modules

There are two special modules that are treated differently than other modules by the node:
the write-ahead log (WAL) and the interceptor.

#### Write-ahead log (WAL)

The WAL module implements a persistent write-ahead log for the case of crashes and restarts.
It has the interface of a regular passive module, but has a distinguished role in the node.
At node instantiation, the WAL is not passed together with the other modules, but as a separate argument to `NewNode()`.

The WAL can persist events to stable storage and be used during crash-recovery to restore the state of the node's modules
by emitting all the stored events at node startup.
More precisely, whenever any module needs to append a persistent entry to the write-ahead log,
it outputs an event destined to the WAL module.
The WAL module simply persists (a serialized form of) this event.
On node startup, if a non-empty write-ahead log is passed to the node at instantiation,
the node loads all events stored in the WAL and adds them to the event buffer.
before it processes events from any other module.
It is then up to the individual modules to re-initialize their state
based on the events received this way.

#### Interceptor

The interceptor is technically not a module, but a special component of the node
specified separately at node instantiation.

The interceptor intercepts events in the same order as they are being passed to modules by the node implementation.
The default interceptor implementation provided by Mir logs all events to disk, producing what we call the event log.
This allows to inspect and replay the event log later on using the [`mircat` utility](/cmd/mircat).

The Interceptor module is not essential and would probably be disabled in a production environment,
but it is priceless for debugging. The user can use a custom interceptor by implementing the
[`Interceptor`](/pkg/eventlog/interceptor.go) interface and passing the corresponding object to `mir.NewNode()`.
E.g., a custom interceptor might only log certain events of interest
or perform a different action on the intercepted events.

#### Difference between the WAL and the (default) interceptor

Note that both the Write-Ahead Log (WAL) and the Interceptor produce logs of Events in stable storage.
Their purposes, however, are very different and largely orthogonal.

**_The WAL_** produces the **_write-ahead log_**
and is crucial for protocol correctness during recovery after restarting a Node.
It is explicitly used by other modules for persisting **_only certain events_** that are crucial for recovery.
The implementation of these modules (e.g., the protocol logic) decides what to store there
and the same logic must be capable to reinitialize itself when those stored events are played back to it on restart.

**_The Interceptor_** produces the **_event log_**.
This is a list of **_all events_** that occurred and is meant only for debugging (not for recovery).
The _Node_'s modules have no influence on what is intercepted.
The event log is intended to be processed by the [`mircat` utility](/cmd/mircat)
to gain more insight into what exactly is happening inside the _Node_.

### Module operation

Each module operates independently of the other modules and only interacts with them through events.
The logic of each module is evaluated concurrently by a separate goroutine.
The implementation of the module can, in principle, execute any code, including spawning additional goroutines.

It is good practice, however, that processing one event by the module implementation
is atomic with respect to the processing of other events.
If in addition, event processing is deterministic for all modules,
the interceptor combined with the [`mircat` utility](/cmd/mircat) become a very powerful debugging tool.

## Systems

A system is a collection of modules that are logically related and configured to work together.
For example, the [Trantor system](/pkg/systems/trantor)
(an implementation of the [Trantor ordering protocol](https://hackmd.io/P59lk4hnSBKN5ki5OblSFg?view))
that comes bundled with Mir can be instantiated as a single abstraction in a user-friendly way.
The `Trantor.Modules()` method then returns a set of named configured modules
that can directly be passed to `mir.NewNode()`.

> Note: At the time of writing, Mir only comes with a single system - the Trantor system.
> Even this system is likely to evolve in the future.
