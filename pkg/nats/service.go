package nats

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// Service defines the NATS service interface
type Service interface {
	// Connection management
	Connect(url string) error
	Close() error
	IsConnected() bool

	// Publishing
	Publish(subject string, data interface{}) error
	PublishToWorkspace(workspaceID string, event interface{}) error
	PublishToUser(userID string, event interface{}) error

	// Subscribing
	Subscribe(subject string, handler MessageHandler) error
	SubscribeToWorkspace(workspaceID string, handler MessageHandler) error
	SubscribeToUser(userID string, handler MessageHandler) error
	Unsubscribe(subject string) error

	// Request-Reply
	Request(subject string, data interface{}, timeout time.Duration) (*nats.Msg, error)
	Reply(subject string, handler RequestHandler) error
}

// MessageHandler defines the message handler interface
type MessageHandler func(msg *nats.Msg) error

// RequestHandler defines the request handler interface
type RequestHandler func(msg *nats.Msg) (interface{}, error)

// NewService creates a new NATS service
func NewService() Service {
	return &service{
		conn: nil,
	}
}

type service struct {
	conn *nats.Conn
}

// Connect connects to NATS server
func (s *service) Connect(url string) error {
	opts := []nats.Option{
		nats.Name("CMP Server"),
		nats.Timeout(10 * time.Second),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(-1),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("NATS disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconnected to %v", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Printf("NATS connection closed")
		}),
	}

	conn, err := nats.Connect(url, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	s.conn = conn
	log.Printf("Connected to NATS at %s", url)
	return nil
}

// Close closes the NATS connection
func (s *service) Close() error {
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	return nil
}

// IsConnected checks if NATS is connected
func (s *service) IsConnected() bool {
	return s.conn != nil && s.conn.IsConnected()
}

// Publish publishes a message to a subject
func (s *service) Publish(subject string, data interface{}) error {
	if !s.IsConnected() {
		return fmt.Errorf("NATS not connected")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	return s.conn.Publish(subject, jsonData)
}

// PublishToWorkspace publishes an event to a workspace
func (s *service) PublishToWorkspace(workspaceID string, event interface{}) error {
	subject := fmt.Sprintf("workspace.%s.events", workspaceID)
	return s.Publish(subject, event)
}

// PublishToUser publishes an event to a user
func (s *service) PublishToUser(userID string, event interface{}) error {
	subject := fmt.Sprintf("user.%s.events", userID)
	return s.Publish(subject, event)
}

// Subscribe subscribes to a subject
func (s *service) Subscribe(subject string, handler MessageHandler) error {
	if !s.IsConnected() {
		return fmt.Errorf("NATS not connected")
	}

	_, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		if err := handler(msg); err != nil {
			log.Printf("Error handling message: %v", err)
		}
	})

	return err
}

// SubscribeToWorkspace subscribes to workspace events
func (s *service) SubscribeToWorkspace(workspaceID string, handler MessageHandler) error {
	subject := fmt.Sprintf("workspace.%s.events", workspaceID)
	return s.Subscribe(subject, handler)
}

// SubscribeToUser subscribes to user events
func (s *service) SubscribeToUser(userID string, handler MessageHandler) error {
	subject := fmt.Sprintf("user.%s.events", userID)
	return s.Subscribe(subject, handler)
}

// Unsubscribe unsubscribes from a subject
func (s *service) Unsubscribe(subject string) error {
	if !s.IsConnected() {
		return fmt.Errorf("NATS not connected")
	}

	// For now, we'll just return success
	// In a real implementation, you would track subscriptions
	return nil
}

// Request sends a request and waits for a reply
func (s *service) Request(subject string, data interface{}, timeout time.Duration) (*nats.Msg, error) {
	if !s.IsConnected() {
		return nil, fmt.Errorf("NATS not connected")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	return s.conn.Request(subject, jsonData, timeout)
}

// Reply sets up a reply handler for a subject
func (s *service) Reply(subject string, handler RequestHandler) error {
	if !s.IsConnected() {
		return fmt.Errorf("NATS not connected")
	}

	_, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		response, err := handler(msg)
		if err != nil {
			log.Printf("Error handling request: %v", err)
			return
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling response: %v", err)
			return
		}

		msg.Respond(jsonData)
	})

	return err
}
