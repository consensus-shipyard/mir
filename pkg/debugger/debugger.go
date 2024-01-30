package debugger

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
)

type WSWriter struct {
	// ... websocket server variables ...
	conn        *websocket.Conn
	upgrader    websocket.Upgrader
	eventSignal chan map[string]interface{}
	WSMessage   struct {
		Type  string `json:"Type"`
		Value string `json:"Value"`
	}
}

// InterceptorInit initializes the interceptor according to input boolean
// The interceptor is set to nil if the user decides to not use the debugger
func InterceptorInit(
	ownID t.NodeID,
	port string,
) (*eventlog.Recorder, error) {
	// writerFactory creates and returns a WebSocket-based event writer
	writerFactory := func(_ string, ownID t.NodeID, _ logging.Logger) (eventlog.EventWriter, error) {
		return newWSWriter(fmt.Sprintf(":%s", port)), nil
	}

	var interceptor *eventlog.Recorder
	var err error
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
	return interceptor, err
}

// Flush does nothing at the moment
func (wsw *WSWriter) Flush() error {
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
func (wsw *WSWriter) Write(list *events.EventList, _ int64) (*events.EventList, error) {
	for wsw.conn == nil {
		fmt.Println("No connection.")
		time.Sleep(time.Millisecond * 100)
	}
	if list.Len() == 0 {
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
		actionType, _ := eventAction["Type"].(string)
		value, _ := eventAction["Value"].(string)
		acceptedEvents, _ = EventAction(actionType, value, acceptedEvents, event)
	}
	return acceptedEvents, nil
}

// EventAction decides, based on the input what exactly is done next with the current event
func EventAction(
	actionType string,
	_ string,
	acceptedEvents *events.EventList,
	currentEvent *eventpb.Event,
) (*events.EventList, error) {
	if actionType == "accept" {
		acceptedEvents.PushBack(currentEvent)
	}
	return acceptedEvents, nil
}

func (wsw *WSWriter) HandleClientSignal(signal map[string]interface{}) {
	wsw.eventSignal <- signal
}

// newWSWriter creates a new WSWriter that establishes a websocket connection
func newWSWriter(port string) *WSWriter {

	// Create a new WSWriter object
	wsWriter := &WSWriter{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  ReadBufferSize,
			WriteBufferSize: WriteBufferSize,
		},
		eventSignal: make(chan map[string]interface{}),
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsWriter.upgrader.CheckOrigin = func(r *http.Request) bool { return true } // Allow opening the connection by HTML file
		conn, err := wsWriter.upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}

		wsWriter.conn = conn
		defer func() {
			err := wsWriter.Close()
			if err != nil {
				panic(err)
			}
		}() // Ensure the connection is closed when the function exits

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil || messageType != websocket.TextMessage {
				break
			}

			var signal map[string]interface{}
			err = json.Unmarshal(message, &signal)
			if err != nil {
				panic(err)
			}

			signalType, typeOk := signal["Type"].(string)
			signalValue, valueOk := signal["Value"].(string)
			if !typeOk || !valueOk {
				panic(fmt.Sprintf("Invalid signal format: Type or Value key missing or not a string in %+v", signal))
			}

			// Check if the signal is a 'close' command
			if signalType == "close" && signalValue == "" {
				break
			}

			wsWriter.HandleClientSignal(signal)
		}
	})

	// Create an Async go routine that waits for the connection
	go func() {
		server := &http.Server{
			Addr:         port,
			Handler:      nil,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  15 * time.Second,
		}

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	return wsWriter
}
