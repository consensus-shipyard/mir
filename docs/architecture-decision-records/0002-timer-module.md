# Managing Time through a dedicated abstraction

* Status: accepted
* Deciders: @matejpavlovic, @sergefdrv, @dnkolegov
* Date: 2022-05-02 (last update)

## Context and Problem Statement

Time is currently represented by *Tick* events periodically injected in the modules.
Each module counts the ticks to obtain a notion of time and acts accordingly.
This makes it inconvenient for the programmer to use time (e.g. for implementing timeouts),
having to implement logic to explicitly deal with Tick events (e.g. counting them).
It would be better to have access to higher-level time abstractions like timeouts when implementing protocols.

## Decision Drivers

* Simplicity of reasoning about the implemented protocol
* Traceability, reproducibility (deterministic execution), and ease of debugging of the protocol
* Programming convenience, code readability
* Simplicity of use of the Mir library for the consumer

## Considered Options

### Separate *Timer* module:

Replace the *Tick* abstraction by a higher-level *Timer* module with 2 main functionalities:
- **Delay**: Emits a given event after a given amount of time.
- **Repeat**: Repeatedly emits a given event with a given frequency until stopped.

If the protocol (or any other module) wishes to invoke one of the Timer functions,
it emits a corresponding `Delay` or `Repeat` event that the Node routes to the *Timer* module.
The event includes, apart from the delay / period, the event to be later emitted by the *Timer* module.
Similarly to WAL entries, each invocation of the Timer would be labeled with a "retention index" that could,
the same way as the WAL, be garbage-collected,
canceling all future event emissions associated with a certain (or lower) retention index.

Note that this approach does not introduce any non-determinism in the protocol execution,
as the time is still implemented as delayed injection of events.
The same sequence of events still deterministically induces exactly the same behavior of the prtocol implementation.
In this sense, the the separate timer module is equivalent to inserting *Ticks* in the protocol module.
The only difference is that the protocol does not have to perform the counting itself -
it happens in a separate *Timer* module

#### Example: view change timeout in a PBFT instance in ISS

At initialization of a PBFT instance, the PBFT protocol must set a "view change timer"
after the expiration of which it enters view change, unless it delivers a batch with sequence number 0 before the expiration.

1.  Thus, at initilization of the PBFT protocol, the Protocol module creates 2 Events:
  - `vct := ViewChangeTimeout{Sn: 0}`
  - `d := Delay{DelayedEvent: vct, Duration: 5000}`  
    Note that the `Delay` event contains the `ViewChangeTimeout` event for sequence number 0.
2.  The Protocol module emits the `Delay` event and the Node implementation saves it in the buffer of the event loop of the Node.
3.  Eventually, the Node implementation reads the `Delay` event from the buffer and routes it to the *Timer* module
    (by calling `applyEvent(d)` on the *Timer* module).
4.  The *Timer* module locally saves the `ViewChangeTimeout` event contained in the `Delay` event
    and sets up an OS-level physical timer for a time that corresponds to 5000 time units (e.g. milliseconds).
5.  When the operating system timer expires, the *Timer* module emits the associated `ViewChangeTimeout` event
    that the Node implementation saves in the buffer of the event loop.
7.  Eventually, the Node implementation picks up the `ViewChangeTimeout` from the buffer and submits it to the Protocol module
    (by calling `applyEvent(vct)` on the Protocol module).
8.  On application of the `ViewChangeTimeout` event, the protocol checks whether batch 0 has already been committed.
  - If yes, it ignores the event. (Note that this makes explicit cancellation of the timeout on committing batch 0 unnecessary.)
  - If not, it enters view change.

*Note*: In practice, there might be more meta-information attached to the events from this example, which are omitted for clarity.

### Protocol-local *Timer* abstraction on top of *Ticks*

Keep the *Ticks* but implement a *Timer* object in a separate package that would expose higher-level functionalities
similar to (or same as) the ones described above.
A *Timer* can be instantiated by a protocol implementation (only one is generally needed)
The protocol implementation would only need to feed *Ticks* to the abstraction (and nothing else but the abstraction).


## Decision Outcome

Chosen option: **Separate *Timer* Module**, because it either improves on, or at least does not compromise
any of the above decision drivers, removes the necessity of *Ticks* altogether
and more naturally addresses the common use case of using a timer as a follow-up event.

### Positive Consequences

See "Good" points below.

### Negative Consequences <!-- optional -->

See "Bad" points below.

## Pros and Cons of the Options

### Separate *Timer* module

* Good, because it improves on programming convenience and code readability
  and does not sacrifice traceability, reproducibility, or simplicity of reasoning about implemented algorithms.
* Good, because it improves simplicity of use for the library consumer,
  who does not need to care about explicitly passing *Ticks* to the Node instance.
* Good, because the implementation of the algorithm does not need to deal with *Ticks* at all.
* Good, because it naturally supports the common use case of using a timer as a follow-up event,
  without having to implement additional protocol logic.
  For example, if the protocol needs to periodically start sending a message
  only after another event is persisted in the WAL,
  it can achieve this by simply emitting one event with a follow-up event attached to it.
* Bad, because it increases the complexity of the core Node framework by introducing another module.

### Protocol-local *Timer* abstraction on top of *Ticks*

* Good, because it improves on programming convenience (although a bit less than Option 1) and code readability
  and does not sacrifice traceability, reproducibility, or simplicity of reasoning about implemented algorithms.
* Good, because it does not necessitate any new modules.
* Good, because if used in sub-modules (e.g. an ISS SB instance), garbage-collection and canceling timers comes "for free"
  together with the garbage-collection of the sub-module.
* Bad, because setting of timers that depend on the execution of another event is not naturally supported
  and needs to be explicitly implemented every time in the protocol code.
* Bad, because it requires augmenting the WAL by a feedback mechanism
  to explicitly notify the protocol about having persisted the required events.
  Otherwise, setting timers is hard (if not impossible) to make dependent on the execution of WAL events.
  (But admittedly, it is likely that the WAL feedback mechanism will be necessary in the future anyway,
  regardless of timers.)
* Bad, because of making it harder to use natural untis of time in the protocol configuration, since tick counting depends on the tick interval determined somewhere else.
