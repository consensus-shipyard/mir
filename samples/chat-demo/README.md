# Chat Demo Application

This is a small application for demonstrating the usage of Mir.
The comments in the code explain in detail how the library is instantiated and used.

This is a very minimal application that is not challenging the library in any way,
since at the time of writing this application, most of the library was composed of stubs.
As the library matures, the application can be extended to support state transfers, node restarts, etc.

## Running the application

Running a chat node requires 2 arguments to be passed on the command line:
- the membership file containing the IDs and network addresses of all the nodes in the system and
- the node's own ID.

Sample membership files for different system sizes running all no the local machine are provided as `membership-*`.
For example, 4 nodes can to be started on the local machine,
e.g. in 4 different terminal windows, as follows (from the root repository directory):

```bash
go run ./samples/chat-demo -i samples/chat-demo/membership-4 0
go run ./samples/chat-demo -i samples/chat-demo/membership-4 1
go run ./samples/chat-demo -i samples/chat-demo/membership-4 2
go run ./samples/chat-demo -i samples/chat-demo/membership-4 3
```

If you have `tmux` installed you can run the nodes using the `run.sh` script:
```bash
./samples/chat-demo/run.sh
```
and then stop them using this command:
```bash
tmux kill-session -t demo
```

When all 4 nodes are started, it may take a moment until they connect to each other.
Once all four nodes print a prompt the user to type their messages, the demo app can be used.
The processes read from standard input line by line
and replicate, in the same order, all messages across all processes.
The process stops after reading EOF.

## Remote deployment

For an automated deployment of the chat demo application on a set of remote machines,
see the [remote deployment instructions](/deployment).
