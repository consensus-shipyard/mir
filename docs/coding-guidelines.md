# Guidelines and conventions contributed code

[DRAFT]

TODO: Fill in 1-2 introductory sentences and more content.

## Logging

Each module that produces any log output should have a logger parameter.
Important events during the module's operation should produce a log output.
In the context of distributed protocols, the log levels have the following semantics:

### FATAL
This should never happen and is the cause of immediate system halt.
No further execution is meaningful when the error occurs.
This log entry should provide the immediate reason for a crash.

### ERROR
This should never happen.
An error is logged if and only if the functionality of the system might be compromised
and the system might not provide the advertised guarantees.
This should only be the case if there is a bug in the implementation
or if some assumptions of the implemented protocol (e.g. on the upper bound of faulty nodes) have been violated.

### WARNING
A warning indicates that something potentially suspicious, but technically legitimate is happening.
When a warning is output, the distributed system as a whole is still fully functional,
even tough parts of it might not be any more.
A warning is the sign of something being sub-optimal, but accounted for.
For example, a warning should be triggered by the reception of a malformed message (another node might be faulty),
a buffer running out of space and dropping messages (network asynchrony or sub-optimal configuration),
a node (local or remote) severely lagging behind (slow node), etc.
If all the nodes are correct and operating smoothly (no significant slow-down)
and the network does not drop or significantly delay messages, no warnings should be logged.

### INFO
An event that indicates the system making progress in its operation.
Info-level log events should be used sparingly, such that the stream of info-level output gives good insight
of what the node is currently doing, but is still observable in real time by a human.
For example, a block being delivered, the configuration having changed, or a checkpoint being created,
all are typical info-level events.

### DEBUG
Debug events provide insight into the internal workings of the system.
The debug events can be produced at a higher rate than a human can perceive
and are intended for recording and studying later.
Interpreting a debug-level log entry can require deep knowledge of the details of the implementation.

### TRACE
Very fine-grained log messages, providing a detailed insight about what code is being executed.
Trace-level logging is not expected to occur all throughout the code, but should only be inserted when needed.
