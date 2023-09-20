# Trantor: Modular State Machine Replication

Trantor is a modular state machine replication (SMR) system.
It decomposes the concept of SMR into multiple smaller components and draws carefully designed,
simple yet powerful abstractions around them.
Trantor aims at being practical, not neglecting important technical aspects
such as checkpointing, state transfer, garbage collection, reconfiguration, or weighted voting,
making them an integral part of the design.
Trantor’s modularity allows it to be flexible, maintainable, adaptable, and future-proof.
Components such as the total-order broadcast protocol can easily be swapped in and out, potentially even at runtime.

Even though the focus of Trantor is not on performance,
a preliminary performance evaluation of our Byzantine fault-tolerant implementation
shows an attractive throughput of over 30k tx/s with a latency of under 1.3 seconds (and 0.5 seconds at 5’000 tx/s)
at a moderate scale of 32 replicas dispersed over 5 different continents,
despite a naive implementation of many of Trantor’s components.

To learn more about Trantor, have a look at the detailed
[Trantor design document](https://github.com/consensus-shipyard/trantor-doc/blob/main/main.pdf).
