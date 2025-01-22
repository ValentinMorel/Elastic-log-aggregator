package server

import (
	"context"
	"log"

	proto "log-aggregator/pb/logmsg"
)

var broadcast = make(chan *proto.LogMessage, 100) // Buffered channel to broadcast logs

type LogService struct {
	proto.UnimplementedLogServiceServer
	storage *Storage
	metrics *Metrics
}

func NewLogService() *LogService {
	return &LogService{
		storage: NewStorage(),
		metrics: NewMetrics(),
	}
}

func (s *LogService) SendLogs(stream proto.LogService_SendLogsServer) error {
	for {
		logMsg, err := stream.Recv()
		if err != nil {
			return stream.SendAndClose(&proto.LogResponse{Status: "Logs received successfully"})
		}

		// Increment active sources
		s.metrics.IncrementSource(logMsg.Source)

		// Save the log to storage
		s.storage.SaveLog(logMsg)

		// Broadcast to WebSocket clients
		broadcast <- logMsg

		log.Printf("Received log from %s: %s", logMsg.Source, logMsg.Message)
	}
}

func (s *LogService) QueryLogs(query *proto.LogQuery, stream proto.LogService_QueryLogsServer) error {
	logs := s.storage.QueryLogs(query)
	for _, logMsg := range logs {
		if err := stream.Send(logMsg); err != nil {
			return err
		}
	}
	return nil
}

func (s *LogService) Metrics(ctx context.Context, empty *proto.Empty) (*proto.MetricsResponse, error) {
	return &proto.MetricsResponse{
		ActiveSources:   int32(s.metrics.GetActiveSources()),
		AlertsTriggered: int32(s.metrics.GetAlertsTriggered()),
	}, nil
}
