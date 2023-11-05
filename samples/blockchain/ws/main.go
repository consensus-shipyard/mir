package main

import "github.com/filecoin-project/mir/samples/blockchain/ws/ws"

func main() {
	ws, send, recv := ws.NewWsServer(8080)
	go ws.StartServers()

	for {
		msg := <-recv
		println("Received message...")
		println(msg.Payload)
		send <- msg
	}

}
