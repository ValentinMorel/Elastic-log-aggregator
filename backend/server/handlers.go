package server

import (
	"context"
	"log"
	"time"

	proto "log-aggregator/pb/logmsg"
)

var Broadcast = make(chan *proto.LogMessage) // Buffered channel to broadcast logs

type LogService struct {
	proto.UnimplementedLogServiceServer
	Storage *Storage
	Metrics *MetricsData
}

func NewLogService() *LogService {
	return &LogService{
		Storage: NewStorage(),
		Metrics: NewMetrics(),
	}
}

func (s *LogService) SendLogs(stream proto.LogService_SendLogsServer) error {
	for {
		logMsg, err := stream.Recv()
		if err != nil {
			return stream.SendAndClose(&proto.LogResponse{Status: "Logs received successfully"})
		}

		logEntry := &proto.LogMessage{
			Timestamp: time.Now().UTC().Format(time.RFC3339), // Add timestamp
			Source:    logMsg.Source,
			LogLevel:  logMsg.LogLevel,
			Message:   logMsg.Message,
		}

		// Increment active sources
		s.Metrics.IncrementSource(logMsg.Source)

		log.Println("Active sources:", s.Metrics.GetActiveSources())

		// Save the log to storage
		s.Storage.SaveLog(logEntry)

		// Broadcast to WebSocket clients
		log.Println("Broadcasting log message")
		Broadcast <- logMsg

		log.Printf("Received log from %s: %s", logMsg.Source, logMsg.Message)
		return nil
	}
}

func (s *LogService) QueryLogs(query *proto.LogQuery, stream proto.LogService_QueryLogsServer) error {
	logs := s.Storage.QueryLogs(query)
	for _, logMsg := range logs {
		if err := stream.Send(logMsg); err != nil {
			return err
		}
	}
	return nil
}

func (s *LogService) GetMetrics(ctx context.Context, empty *proto.Empty) (*proto.MetricsResponse, error) {
	return &proto.MetricsResponse{
		ActiveSources:   int32(s.Metrics.GetActiveSources()),
		AlertsTriggered: int32(s.Metrics.GetAlertsTriggered()),
	}, nil
}
