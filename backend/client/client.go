package main

import (
	"context"
	"log"
	proto "log-aggregator/pb/logmsg"
	"time"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	client := proto.NewLogServiceClient(conn)
	logMessage := []*proto.LogMessage{
		{Source: "source1", LogLevel: "INFO", Message: "log message 1"},
		{Source: "source2", LogLevel: "INFO", Message: "log message 2"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.SendLogs(ctx)
	if err != nil {
		log.Fatalf("Failed to send logs: %v", err)
	}

	for _, logMsg := range logMessage {
		if err := stream.Send(logMsg); err != nil {
			log.Fatalf("Failed to send log message: %v", err)
		}
	}
	response, err := stream.CloseAndRecv()
	if err != nil {
		log.Printf("Empty response: %v", err)
		return
	}

	log.Printf("Response: %s", response.Status)

}
