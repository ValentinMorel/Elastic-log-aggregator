package server

import (
	"io"
	proto "log-aggregator/pb/logmsg"
	"testing"

	"google.golang.org/grpc"
)

func TestSendLogs(t *testing.T) {
	tests := []struct {
		name    string
		logs    []*proto.LogMessage
		wantErr bool
	}{
		{
			name: "Single log message",
			logs: []*proto.LogMessage{
				{Source: "source1", Message: "log message 1"},
			},
			wantErr: false,
		},
		{
			name: "Multiple log messages",
			logs: []*proto.LogMessage{
				{Source: "source1", Message: "log message 1"},
				{Source: "source2", Message: "log message 2"},
			},
			wantErr: false,
		},
		{
			name:    "No log messages",
			logs:    []*proto.LogMessage{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewLogService()
			stream := &mockLogService_SendLogsServer{
				recv: tt.logs,
			}

			err := s.SendLogs(stream)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendLogs() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(stream.sent) != 1 {
				t.Errorf("Expected 1 response, got %d", len(stream.sent))
			}

			if stream.sent[0].Status != "Logs received successfully" {
				t.Errorf("Expected status 'Logs received successfully', got %s", stream.sent[0].Status)
			}
		})
	}
}

type mockLogService_SendLogsServer struct {
	grpc.ServerStream
	recv []*proto.LogMessage
	sent []*proto.LogResponse
}

func (m *mockLogService_SendLogsServer) Recv() (*proto.LogMessage, error) {
	if len(m.recv) == 0 {
		return nil, io.EOF
	}
	msg := m.recv[0]
	m.recv = m.recv[1:]
	return msg, nil
}

func (m *mockLogService_SendLogsServer) SendAndClose(resp *proto.LogResponse) error {
	m.sent = append(m.sent, resp)
	return nil
}
