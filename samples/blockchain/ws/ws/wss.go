package ws

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WsMessage struct {
	MessageType int
	Payload     []byte
}

type connection struct {
	id       string
	server   *wsServer
	ws       *websocket.Conn
	sendChan chan WsMessage
	recvChan chan WsMessage
}

type wsServer struct {
	port            int
	connections     map[string]connection
	sendChan        chan WsMessage
	recvChan        chan WsMessage
	connectionsLock sync.Mutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (wss *wsServer) handleHome(w http.ResponseWriter, r *http.Request) {
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
				panic(err)
			}
		}
	}()

	// receiving msgs
	for {
		messageType, p, err := wsc.ws.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		wsc.recvChan <- WsMessage{messageType, p}
	}
}

func (wss *wsServer) handleWs(w http.ResponseWriter, r *http.Request) {
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
	conn := connection{id, wss, ws, make(chan WsMessage), wss.recvChan}
	wss.connections[id] = conn
	wss.connectionsLock.Unlock()
	go conn.wsConnection()
}

func (wss *wsServer) setupHttpRoutes() {
	http.HandleFunc("/", wss.handleHome)
	http.HandleFunc("/ws", wss.handleWs)
}

func (wss *wsServer) sendHandler() {
	for {
		message := <-wss.sendChan
		for _, conn := range wss.connections {
			conn.sendChan <- message
		}
	}
}

func (wss *wsServer) StartServers() {
	log.Println("Starting servers...")
	wss.setupHttpRoutes()
	go wss.sendHandler()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", wss.port), nil))
}

func NewWsServer(port int) (*wsServer, chan WsMessage, chan WsMessage) {
	wss := &wsServer{
		port:        port,
		connections: make(map[string]connection),
		sendChan:    make(chan WsMessage),
		recvChan:    make(chan WsMessage),
	}

	return wss, wss.sendChan, wss.recvChan
}
