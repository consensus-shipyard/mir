## Websocket Interface

### Connection

* Each Mir module, when running in 'debugger' mode, starts a websocket server at the port address `xxxx` given at runtime using the flag `-port=xxxx` (e.g., `go run ./pinpong -d -port=8080 0` runs the websocket server of the node on port 8080).
* Before the node starts, it waits for a websocket client to connect. (The frontend will act as the client.)
* The frontend connects to the specified port using `ws://localhost:${this.port}/ws`.
* Upon connection, the server module sends an "init" event.
* Upon receiving the "init" event, the frontend responds with the JSON `{"Type": "start"}`.
* The connection closes when the server stops or the client disconnects using a message of `{"Type": "close"}`.
  .

### Event Format (JSON)

```json
{
  "event":     "protojson.Marshal(event)",
  "timestamp": "timestamp"
}

```
_(To note, there is a double JSON parsing for the event)_


### Frontend Response Format (JSON)

```json
{
  "Type": "",
  "Value": ""
}

```

### Supported Response Types, Expected Values, and Actions

| Type    | Value                                                             | Frontend Action                                          | Backend Action                                                                |
|---------|-------------------------------------------------------------------|----------------------------------------------------------|-------------------------------------------------------------------------------|
| start   | null                                                              | Initialize the connection                                | Send the first event                                                           |
| accept  | null                                                              | Add accepted event to the log and wait for the next event | Update the list of events with the accepted one and send the next event        |
| decline | null                                                              | Drop the declined event and wait for the next event      | Drop the declined event and send the next event                               |
| replace | Modified event (in the format mentioned [above](#event-format-json)) | Add modified event to the log and wait for the next event | Update the list of events with the modified event received and send next event |
| close   | null                                                              | Close the websocket connection                           | Stop sending data to the client                                               |
