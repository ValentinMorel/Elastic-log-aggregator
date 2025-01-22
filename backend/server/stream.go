package server

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var upgrader = websocket.Upgrader{}

func HandleWebsocketConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer ws.Close()

	clients[ws] = true
	defer delete(clients, ws)

	for {
		logMsg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(logMsg)
			if err != nil {
				log.Printf("Error writing to WebSocket: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
