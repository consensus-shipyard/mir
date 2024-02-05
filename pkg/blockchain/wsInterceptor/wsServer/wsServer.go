package wsServer

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

/**
 * Websocket server
 * =================
 *
 * A simple websocket server which accepts connections and sends messages to all connected clients.
 * The messages are sent to the server through a channel.
 *
 * Note: This is a very shotty implementation. For example, it crashes if a client disconnects.
 * It is only intended for debugging purposes.
 */

type WsMessage struct {
	MessageType int
	Payload     []byte
}

type connection struct {
	id       string
	server   *WsServer
	ws       *websocket.Conn
	sendChan chan WsMessage
}

type WsServer struct {
	SendChan        chan WsMessage
	port            int
	connections     map[string]connection
	connectionsLock sync.Mutex
	logger          logging.Logger
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (wss *WsServer) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("Hello world!"))
}

func (wsc *connection) wsConnection() {
	log.Println("wsConnection")
	//  close connection
	defer wsc.ws.Close()
	// close up channels and remove connection
	defer func() {
		wsc.server.connectionsLock.Lock()
		delete(wsc.server.connections, wsc.id)
		close(wsc.sendChan)
		wsc.server.connectionsLock.Unlock()
	}()

	// sending msgs
	go func() {
		for {
			message := <-wsc.sendChan
			err := wsc.ws.WriteMessage(message.MessageType, message.Payload)
			if err != nil {
				wsc.server.logger.Log(logging.LevelError, "Error sending WS message: ", err)
			}
		}
	}()

	// "receiving" msgs
	for {
		wsc.ws.ReadMessage()
		wsc.server.logger.Log(logging.LevelWarn, "Received WS message - ignored")
	}
}

func (wss *WsServer) handleWs(w http.ResponseWriter, r *http.Request) {
	// cors *
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		wss.logger.Log(logging.LevelError, "Upgrade failed: ", err)
		return
	}
	wss.connectionsLock.Lock()
	id := ""
	for {
		id = uuid.New().String()
		if _, ok := wss.connections[id]; !ok {
			break
		}
	}
	conn := connection{id, wss, ws, make(chan WsMessage)}
	wss.connections[id] = conn
	wss.connectionsLock.Unlock()
	go conn.wsConnection()
}

func (wss *WsServer) setupHttpRoutes() {
	http.HandleFunc("/", wss.handleHome)
	http.HandleFunc("/ws", wss.handleWs)
}

func (wss *WsServer) sendHandler() {
	for {
		message := <-wss.SendChan
		for _, conn := range wss.connections {
			conn.sendChan <- message
		}
	}
}

func (wss *WsServer) StartServers() {
	wss.logger.Log(logging.LevelInfo, "Starting servers...")
	wss.setupHttpRoutes()
	go wss.sendHandler()
	if err := http.ListenAndServe(fmt.Sprintf(":%d", wss.port), nil); err != nil {
		wss.logger.Log(logging.LevelError, "ListenAndServe: ", err)
	}
}

func NewWsServer(port int, logger logging.Logger) *WsServer {
	wss := &WsServer{
		port:        port,
		connections: make(map[string]connection),
		SendChan:    make(chan WsMessage),
		logger:      logger,
	}

	return wss
}
