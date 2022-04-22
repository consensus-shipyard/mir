# ADR 1: Types of Node IDs and Client IDs

## Context

Each node and each client in Mir is uniquely identified by an ID.
The types of the IDs are currently defined in [/pkg/types/types.go](/pkg/types/types.go)
as unsigned 64-bit integers.
Using this type for IDs is simple and can easily and efficiently be used as map keys.
Moreover, in test deployments, numeric IDs can be used to derive default values for node information
like, for example, the port number the node is listening on.

Such numeric IDs, however, create a hurdle when using Mir in systems where nodes' and clients' identifiers
are already fixed (e.g. if Mir is only used to implement one component of a bigger system
like [Eudico](https://github.com/filecoin-project/eudico))
and cannot easily be converted to and from this type (e.g. public keys).
Another translation layer between system-native IDs and IDs used inside Mir would be necessary.
Using strings instead enables a more direct use of the system's native IDs,
as almost anything can be converted to a string.
When it comes to the (very frequent) use of IDs as map keys, strings can still be used for this purpose.

## Decision

We will use strings as the underlying type for node and client IDs.

## Status

Proposed (Apr 20, 2022)
Accepted (Apr 21, 2022)

## Consequences

In the protocol logic implementation, nothing should change, since the abstract NodeID and ClientID data types are used.

The biggest change will be adapting the protocol buffer definitions,
since it does not encapsulate the node or client IDs in a separate message type.

The configuration of dummy test deployments will need to be adapted wherever it relies
on inferring information (like port number) about a node only from the node's ID.

Node and client IDs native to systems where Mir is used (e.g. Eudico) will be easily representable.

## Related issues

[#5: Make Node ID and Client ID string types](https://github.com/filecoin-project/mir/issues/5)