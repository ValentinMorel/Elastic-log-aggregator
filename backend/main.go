package main

import (
	"encoding/json"
	"log"
	proto "log-aggregator/pb/logmsg"
	"log-aggregator/server"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/olivere/elastic/v7"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var clients = make(map[*websocket.Conn]bool)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Check the request's Origin header
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:3000" {
			return true
		}
		return false
	},
}

func main() {
	quit := make(chan bool, 1)
	log.Println("Starting Distributed Logging System...")

	// Initialize Elasticsearch client
	elasticClient, err := elastic.NewClient(elastic.SetURL("http://localhost:9200"))
	if err != nil {
		log.Fatalf("Failed to connect to Elasticsearch: %v", err)
	}

	// Initialize Alerts
	alerts := server.NewAlerts(elasticClient)
	alerts.AddRule("ERROR", 10, 1*time.Minute, "High error rate detected")
	alerts.StartMonitoring()

	logService := server.NewLogService()

	// Start the gRPC Server
	go func(service *server.LogService) {
		listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer()
		reflection.Register(grpcServer)

		proto.RegisterLogServiceServer(grpcServer, service)

		log.Println("gRPC server running on port 50051")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}(logService)

	http.HandleFunc("/ws", handleWS)

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		// Optional: add CORS header if needed
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Example single snapshot
		data := struct {
			ActiveSources   int `json:"activeSources"`
			AlertsTriggered int `json:"alertsTriggered"`
		}{
			ActiveSources:   logService.Metrics.GetActiveSources(), // e.g. logService.Metrics().GetActiveSources()
			AlertsTriggered: alerts.GetAlertsTriggered(),           // from your alerts struct
		}

		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding JSON for /metrics HTTP: %v", err)
		}

	})

	log.Println("WebSocket server running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
	<-quit
}

// handleWS upgrades the connection to WebSocket, and listens for broadcast messages
func handleWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer ws.Close()

	clients[ws] = true

	log.Println("New client connected to /ws")

	for {
		// When something is broadcast on the channel, send it to all clients
		msg := <-server.Broadcast
		for client := range clients {
			if err := client.WriteJSON(msg); err != nil {
				log.Printf("WebSocket write error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
