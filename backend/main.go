package main

import (
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Accepting all requests
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

	// Start the gRPC Server
	go func() {
		listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer()
		reflection.Register(grpcServer)

		server := server.NewLogService()
		proto.RegisterLogServiceServer(grpcServer, server)

		log.Println("gRPC server running on port 50051")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// WebSocket Server for Alerts
	http.HandleFunc("/alerts", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}
		defer ws.Close()

		for alert := range alerts.GetBroadcastChannel() {
			if err := ws.WriteJSON(alert); err != nil {
				log.Printf("WebSocket write error: %v", err)
				break
			}
		}
	})

	log.Println("WebSocket server running on port 8080")
	if err := http.ListenAndServe(":8080", http.HandlerFunc(server.HandleWebsocketConnections)); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
	<-quit
}
