[![Go Reference](https://pkg.go.dev/badge/github.com/filecoin-project/mir.svg)](https://pkg.go.dev/github.com/filecoin-project/mir)
[![Mir Test](https://github.com/filecoin-project/mir/actions/workflows/test.yml/badge.svg)](https://github.com/filecoin-project/mir/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/filecoin-project/mir)](https://goreportcard.com/report/github.com/filecoin-project/mir)

# Mir - The Distributed Protocol Implementation Framework

Mir is a framework for implementing, debugging, and analyzing distributed protocols.
It has the form of a library that provides abstractions representing different components of a distributed system
and an engine orchestrating their interaction.

Mir aims to be general enough to enable implementing distributed protocols in a way that is agnostic to
network transport, storage, and cryptographic primitive implementation.
All these (and other) usual components of a distributed protocol implementation
are encapsulated in abstractions with well-defined interfaces.
While Mir provides some implementations of those abstractions to be used directly "out of the box",
the consumer is free to provide their own implementations.

The first intended use of Mir is as a scalable and efficient
[consensus layer in Filecoin subnets](https://github.com/protocol/ConsensusLab/issues/9)
and, potentially, as a Byzantine fault-tolerant ordering service in Hyperledger Fabric.
However, Mir hopes to be a building block of a next generation of distributed systems, being used by many applications.

## Structure and Usage

Mir is a framework for implementing distributed protocols (also referred to as distributed algorithms)
meant to run on a distributed system.
The basic unit of a distributed system is a *node*.
Each node locally executes (its portion of) a protocol,
sending and receiving *messages* to and from other nodes over a communication network.

Mir models a node of such a distributed system and presents the consumer (the programmer using Mir)
with a [*Node*](/node.go) abstraction.
The Node receives *requests*, processes them (while coordinating with other Nodes using the distributed protocol),
and eventually delivers them to a (consumer-defined) *application*.
Fundamentally, Mir's purpose can be summed up simply as receiving requests from the consumer
and delivering them to an application as prescribed by some protocol.
The [ISS](/pkg/iss) protocol, for example, being a total order broadcast protocol,
guarantees that all requests received by the nodes will be delivered to the application in the same order.

Note that the application need not necessarily be an end-user application -
any program using Mir is an application from Mir's point of view.
While the application logic is (except for the included sample demos) always expected to
be provided by the Mir consumer, this need not be the case for the protocol.
While the consumer is free to provide a custom protocol component,
Mir will provide out-of-the-box implementations of different distributed protocols for the consumer to select
(the first of them being ISS).

We now describe how the above (providing an application, selecting a protocol, etc.) works in practice.
This is where Mir's modular design comes into play.
The Node abstraction mentioned above is implemented as a Go struct that contains multiple *modules*.
In short, when instantiating a Node, the consumer of Mir provides implementations of these modules to Mir.
For example, instantiating a node might look as follows:

```go
    // Example Node instantiation adapted from samples/chat-demo/main.go
    node, err := mir.NewNode(/*some more arguments*/ &modules.Modules{
        Net:      grpcNetworking,
        // ...
        Protocol: issProtocol,
        App:      NewChatApp(reqStore),
        Crypto:   ecdsaCrypto,
    })
```

Here the consumer provides modules for networking
(implements sending and receiving messages over the network, in this case using gRPC),
the protocol logic (using the ISS protocol), the application (implementing the logic of a chat app), and a cryptographic
module (able to produce and verify digital signatures using the ECDSA algorithm).
There are more modules a Node is using, some of them always have to be provided,
some can be left out and Mir will fall back to default built-in implementations.
Some modules, even though they always need to be explicitly provided at Node instantiation,
are part of Mir and can themselves be instantiated easily using Mir library functions.

Inside the Node, the modules interact using *Events*.
Each module independently consumes, processes, and outputs Events.
This approach bears resemblance to the [actor model](https://en.wikipedia.org/wiki/Actor_model),
where Events exchanged between modules correspond to messages exchanged between actors.

The Node implements an event loop, where all Events created by modules are stored in a buffer and, based on their types,
distributed to their corresponding modules for processing.
For example, when the networking module receives a protocol message over the network,
it generates a `MessageReceived` Event (containing the received message)
that the Node implementation routes to the protocol module, which processes the message,
potentially outputting `SendMessage` Events that the Node implementation routes back to the networking module.

The architecture described above enables a powerful debugging approach.
All Events in the event loop can, in debug mode, be recorded, inspected, or even replayed to the Node
using a debugging interface.

The high-level architecture of a Node is depicted in the figure below.
For more details, see the [module interfaces](/pkg/modules)
and a more detailed description of each module in the [Documentation](/docs).
![High-level architecture of a Mir Node](/docs/images/high-level-architecture.png)

### Relation to the Mir-BFT protocol

[Mir-BFT](https://arxiv.org/abs/1906.05552) is a scalable atomic broadcast protocol.
The Mir framework initially started as an implementation of that protocol - thus the related name -
but has since been made completely independent of Mir-BFT.
Even the implementation of the Mir-BFT protocol itself has been abandoned and replaced by its successor,
[ISS](/pkg/iss), which is intended to be the first protocol implemented within Mir.
However, since Mir is designed to be modular and versatile,
ISS is just one (the first) of the protocols implemented in Mir.

## Current Status

This library is in development and not usable yet.
This document describes what the library *should become* rather than what it *currently is*.
This document itself is more than likely to still change.
You are more than welcome to contribute to accelerating the development of the Mir library
as an open-source contributor.
Have a look at the [Contributions section](#contributing) if you want to help out!

## Compiling and running tests

Assuming [Go](https://go.dev/) version 1.18 or higher is installed, the tests can be run by executing `go test ./...` in the root folder of the repository.
The dependencies should be downloaded and installed automatically. Some of the dependencies may require [gcc](https://gcc.gnu.org/) installed.
On Ubuntu Linux, it can be done by invoking `sudo apt install gcc`.

If the sources have been updated, it is possible that some of the generated source files need to be updated as well.
More specifically, the Mir library relies on [Protocol Buffers](https://developers.google.com/protocol-buffers) and [gomock](https://github.com/golang/mock).
The `protoc` compiler and the corresponding Go plugin need to be installed as well as `mockgen`.
On Ubuntu Linux, those can be installed using

```shell
sudo snap install --classic protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/golang/mock/mockgen@v1.6.0
```

Once installed, the generated sources can be updated by executing `go generate ./...` in the root directory.

## Documentation

For a description of the design and inner workings of the library, see [Mir Library Architecture](/docs).
We also keep a log of [Architecture Decision Records (ADRs)](/docs/architecture-decision-records).

For a small demo application, see [/samples/chat-demo](/samples/chat-demo)

## Contributing

**Contributions are more than welcome!**

If you want to contribute, have a look at the open [issues](https://github.com/filecoin-project/mir/issues).
If you have any questions (specific or general),
do not hesitate to drop an email to the active [maintainer(s)](/MAINTAINERS.md)
or write a message to the developers in the
public Slack channel [#mir-dev](https://filecoinproject.slack.com/archives/C03C77HN3AS) of the Filecoin project.

## Active maintainer(s)

- [Matej Pavlovic](https://github.com/matejpavlovic) (matopavlovic@gmail.com)

## License

The Mir library source code is made available under the Apache License, version 2.0 (Apache-2.0), located in the
[LICENSE file](LICENSE).

## Acknowledgments

This project is a continuation of the development started under the name
[`mirbft`](https://github.com/hyperledger-labs/mirbft)
as a [Hyperledger Lab](https://labs.hyperledger.org/labs/mir-bft.html).
