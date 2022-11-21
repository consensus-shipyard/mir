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

## Nodes, Modules, and Events

Mir is a framework for implementing distributed protocols (also referred to as distributed algorithms)
meant to run on a distributed system.
The basic unit of a distributed system is a *node*.
Each node locally executes (its portion of) a protocol,
sending and receiving *messages* to and from other nodes over a communication network.

Mir models a node of such a distributed system and presents the consumer (the programmer using Mir)
with a [`Node`](/node.go) abstraction.
The main task of a `Node` is to process events.

A node contains one or multiple [`Module`](/pkg/modules/modules.go)s that implement the logic for event processing.
Each module independently consumes, processes, and outputs events.
This approach bears resemblance to the [actor model](https://en.wikipedia.org/wiki/Actor_model),
where events exchanged between modules correspond to messages exchanged between actors.

The Node implements an event loop where all events created by modules are stored in a buffer and
distributed to their corresponding target modules for processing.
For example, when the networking module receives a protocol message over the network,
it generates a `MessageReceived` event (containing the received message)
that the node implementation routes to the protocol module, which processes the message,
potentially outputting `SendMessage` events that the Node implementation routes back to the networking module.

The architecture described above enables a powerful debugging approach.
All Events in the event loop can, in debug mode, be recorded, inspected, or even modified and replayed to the Node
using a debugging interface.

In practice, when instantiating a Node, the consumer of Mir provides implementations of these modules to Mir.
For example, instantiating a node of a state machine replication system might look as follows:

```go
    // Example Node instantiation
    node, err := mir.NewNode(
		/* some more technical arguments ... */
		&modules.Modules{
			// ... 
			"app":      NewChatApp(),
			"protocol": TOBProtocol,
			"net":      grpcNetworking,
			"crypto":   ecdsaCrypto,
		},
		eventInterceptor,
		writeAheadLog,
	)
```

![Example Mir Node](/docs/images/high-level-architecture.png)

Here the consumer provides modules for networking
(implements sending and receiving messages over the network),
the protocol logic (using some total-order broadcast protocol),
the application (implementing the logic of the replicated app),
and a cryptographic module (able to produce and verify digital signatures using ECDSA).
the `eventInterceptor` implements recording of the events passed between the modules
for analysis and debugging purposes.
The `writeAheadLog` is a special module that enables the node to recover from crashes.
For more details, see the [Documentation](/docs).

The programmer working with Mir is free to provide own implementations of these modules,
but Mir already comes bundled with several module implementations out of the box.

### Relation to the Mir-BFT protocol

[Mir-BFT](https://arxiv.org/abs/1906.05552) is a scalable atomic broadcast protocol.
The Mir framework initially started as an implementation of that protocol - thus the related name -
but has since been made completely independent of Mir-BFT.
Even the implementation of the Mir-BFT protocol itself has been abandoned and replaced by its successor,
[ISS](/pkg/iss), which is intended to be the first protocol implemented within Mir.
However, since Mir is designed to be modular and versatile,
ISS is just one (the first) of the protocols implemented in Mir.

## Current Status

This library is in development.
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
Make sure (by configuring your shell interpreter) that the directory with Go binaries (usually `~/go/bin` by default) is included in the `PATH` environment variable.
On a default Ubuntu Linux system, for example, this can be achieved by running
```shell
echo 'PATH=$PATH:~/go/bin' >> ~/.profile
```
and restarting the terminal.

Once the dependencies are ready, the generated sources can be updated by executing `go generate ./...` in the root directory.

## Documentation

For a description of the design and inner workings of the library, see [Mir Library Overview](/docs).
We also keep a log of [Architecture Decision Records (ADRs)](/docs/architecture-decision-records).

For a small demo application, see [/samples/chat-demo](/samples/chat-demo)

For an automated deployment of Mir on a set of remote machines, see the [remote deployment instructions](/deployment).

### Getting started

To get started using (and contributing to) Mir, in addition to this README, we recommend the following:
1. Watch the first [introductory video](https://www.youtube.com/watch?v=UZKpqrJtMxc)
2. Read the [Mir Library Overview](/docs)
3. Watch the second [introductory video](https://www.youtube.com/watch?v=Vmt3ZeKIJIM). (Very low-level coding, this is **not** how Mir coding works in real life - it is for developers to understand how Mir internally works. Realistic coding videos will follow soon.)
4. Check out the [chat-demo](/samples/chat-demo) sample application to learn how to use Mir for state machine replication.
5. To see an example of using a DSL module (allowing to write pseudocode-like code for the protocol logic),
   look at the [implementation of Byzantine Consistent Broadcast (BCB)](/pkg/bcb)
   being used in the corresponding [sample application](/samples/bcb-demo).
   Original pseudocode can also be found in
   [these lecture notes](https://dcl.epfl.ch/site/_media/education/sdc_byzconsensus.pdf)
   (Algorithm 4 (Echo broadcast [Rei94])).
6. A more complex example of DSL code is the implementation of the [SMR availability layer](/pkg/availability)
   (concretely the `multisigcollector`).

To learn about the first complex system being built on Mir,
have a look at [Trantor](https://hackmd.io/P59lk4hnSBKN5ki5OblSFg?view), a complex SMR syste being implemented using Mir.

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
