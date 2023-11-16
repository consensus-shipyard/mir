package wsserver

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

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
				log.Println(err)
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
		log.Printf("Upgrade failed: %v", err)
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
	log.Println("Starting servers...")
	wss.setupHttpRoutes()
	go wss.sendHandler()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", wss.port), nil))
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
