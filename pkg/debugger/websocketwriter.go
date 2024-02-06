package debugger

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
)

const (
	ReadBufferSize  = 1024
	WriteBufferSize = 1024
)

type WSWriter struct {
	// ... websocket server variables ...
	conn        *websocket.Conn
	upgrader    websocket.Upgrader
	eventSignal chan map[string]string
	WSMessage   struct {
		Type  string `json:"Type"`
		Value string `json:"Value"`
	}
	logger logging.Logger
}

// newWSWriter creates a new WSWriter that establishes a websocket connection
func newWSWriter(port string, logger logging.Logger) *WSWriter {

	// Create a new WSWriter object
	wsWriter := &WSWriter{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  ReadBufferSize,
			WriteBufferSize: WriteBufferSize,
		},
		eventSignal: make(chan map[string]string),
		logger:      logger,
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

			var signal map[string]string
			err = json.Unmarshal(message, &signal)
			if err != nil {
				panic(err)
			}

			// Check if the signal is a 'close' command
			if signal["Type"] == "close" && signal["Value"] == "" {
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

// Write sends every event to the frontend and then waits for a message detailing how to proceed with that event
// The returned EventList contains the accepted events
func (wsw *WSWriter) Write(list *events.EventList, _ int64) (*events.EventList, error) {
	for wsw.conn == nil {
		wsw.logger.Log(logging.LevelInfo, "No connection")
		time.Sleep(time.Millisecond * 100) // TODO: Why do we sleep here? Do we need it?
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

		action := <-wsw.eventSignal
		acceptedEvents, _ = eventAction(action["Type"], action["Value"], acceptedEvents, event)
	}
	return acceptedEvents, nil
}

func (wsw *WSWriter) HandleClientSignal(signal map[string]string) {
	wsw.eventSignal <- signal
}

// EventAction decides, based on the input what exactly is done next with the current event
func eventAction(
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
