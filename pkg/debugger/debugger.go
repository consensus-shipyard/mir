package debugger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"

	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	t "github.com/filecoin-project/mir/pkg/types"
)

const (
	ReadBufferSize  = 1024
	WriteBufferSize = 1024
	BasePort        = 8080
)

type WSWriter struct {
	// ... websocket server variables ...
	conn        *websocket.Conn
	upgrader    websocket.Upgrader
	eventSignal chan map[string]interface{}
}

// InterceptorInit initializes the interceptor according to input boolean
// The interceptor is set to nil if the user decides to not use the debugger
func InterceptorInit(
	debugger bool,
	ownID t.NodeID,
) (*eventlog.Recorder, error) {
	var interceptor *eventlog.Recorder
	var err error
	if debugger {
		interceptor, err = eventlog.NewRecorder(
			ownID,
			fmt.Sprintf("./node%s", ownID),
			logging.ConsoleInfoLogger,
			eventlog.EventWriterOpt(writerFactory),
			eventlog.SyncWriteOpt(),
		)
		if err != nil {
			panic(err)
		}
		fmt.Println("Interceptor created successfully")
	} else {
		interceptor = nil
		err = nil
	}
	return interceptor, err
}

// Flush does nothing at the moment
func (wsw *WSWriter) Flush() error {
	if wsw.conn == nil {
		return nil
	}
	return nil
}

// Close closes the connection
func (wsw *WSWriter) Close() error {
	if wsw.conn == nil {
		return nil
	}
	return wsw.conn.Close()
}

// Write sends every event to the frontend and then wits for a message detailing how to proceed with that event
// The returned EventList contains the accepted events
func (wsw *WSWriter) Write(list *events.EventList, timestamp int64) (*events.EventList, error) {
	for wsw.conn == nil {
		fmt.Println("No connection.")
		time.Sleep(time.Millisecond * 100)
	}
	if list.Len() == 0 {
		fmt.Println("No events to print.")
		return list, nil
	}

	acceptedEvents := events.EmptyList()
	iter := list.Iterator()

	for event := iter.Next(); event != nil; event = iter.Next() {
		// Create a new JSON object with a timestamp field
		timestamp := time.Now()
		logData := map[string]interface{}{
			"event":     event,
			"timestamp": timestamp,
		}

		// Marshal the JSON data
		message, err := json.Marshal(logData)
		if err != nil {
			panic(err)
		}

		// Send the JSON message over WebSocket
		if err := wsw.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return list, fmt.Errorf("error sending message over WebSocket: %w", err)
		}

		eventAction := <-wsw.eventSignal
		actionType, ok := eventAction["type"].(string)
		value, ok := eventAction["value"].(string)
		if ok {
			panic(ok)
		}
		acceptedEvents, _ = EventAction(actionType, value, acceptedEvents, event)
	}
	return acceptedEvents, nil
}

// EventAction decides, based on the input what exactly is done next with the current event
func EventAction(
	actionType string,
	actionValue string,
	acceptedEvents *events.EventList,
	currentEvent *eventpb.Event,
) (*events.EventList, error) {
	if actionType == "accept" {
		acceptedEvents.PushBack(currentEvent)
	} else if actionType == "decline" {
		// do nothing
	} else if actionType == "delay" {
		//TODO: delay the event based on the value specified in actionValue
	} else if actionType == "modify" {
		//TODO
	}
	return acceptedEvents, nil
}

func (wsw *WSWriter) HandleClientSignal(signal map[string]interface{}) {
	wsw.eventSignal <- signal
}

// newWSWriter creates a new WSWriter that establishes a websocket connection
func newWSWriter(port string) *WSWriter {
	fmt.Println("Starting newWSWriter")

	// Create a new WSWriter object
	wsWriter := &WSWriter{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  ReadBufferSize,
			WriteBufferSize: WriteBufferSize,
		},
		eventSignal: make(chan map[string]interface{}),
	}

	// Create an Async go routine that waits for the connection
	go func() {
		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			wsWriter.upgrader.CheckOrigin = func(r *http.Request) bool { return true } // Allow opening the connection by HTML file
			conn, err := wsWriter.upgrader.Upgrade(w, r, nil)
			if err != nil {
				panic(err)
			}

			fmt.Println("WebSocket connection established")

			// Update the attribute of the WSWriter object with the established connection
			wsWriter.conn = conn
			// go routine for incoming messages
			go func() {
				defer func(conn *websocket.Conn) {
					err := conn.Close()
					if err != nil {

					}
				}(conn) // Ensure the connection is closed when the function exits

				for {
					_, message, err := conn.ReadMessage()
					if err != nil {
						break
					}

					var signal map[string]interface{}
					err = json.Unmarshal(message, &signal)
					if err != nil {
						continue
					}
					wsWriter.HandleClientSignal(signal)
				}
			}()
		})

		server := &http.Server{
			Addr:         port,
			Handler:      nil,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  15 * time.Second,
		}

		err := server.ListenAndServe()
		if err != nil {
			panic(err)
		}
		select {}
	}()
	return wsWriter
}

// writerFactory creates and returns a WebSocket-based event writer
// It determines the port for the WebSocket server based on the given nodeID and a base port value
func writerFactory(dest string, nodeID t.NodeID, logger logging.Logger) (eventlog.EventWriter, error) {
	ownPort, err := strconv.Atoi(string(nodeID))
	if err != nil {
		panic(err)
	}
	ownPort += BasePort
	return newWSWriter(fmt.Sprintf(":%d", ownPort)), nil
}
