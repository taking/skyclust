package messaging

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

// NATSService provides NATS connection management
type NATSService struct {
	url  string
	conn *nats.Conn
}

// NewNATSService creates a new NATS service
func NewNATSService(url string) *NATSService {
	return &NATSService{
		url: url,
	}
}

// Connect connects to NATS
func (s *NATSService) Connect() error {
	conn, err := nats.Connect(s.url)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	s.conn = conn
	return nil
}

// Close closes the NATS connection
func (s *NATSService) Close() error {
	if s.conn != nil {
		s.conn.Close()
	}
	return nil
}

// GetConnection returns the NATS connection
func (s *NATSService) GetConnection() *nats.Conn {
	return s.conn
}
