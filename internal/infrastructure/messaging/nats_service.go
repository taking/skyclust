package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"skyclust/pkg/logger"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL     string
	Cluster string
	Subject string
}

// NATSService implements messaging using NATS
type NATSService struct {
	conn   *nats.Conn
	config NATSConfig
}

// NewNATSService creates a new NATS service
func NewNATSService(config NATSConfig) (*NATSService, error) {
	opts := []nats.Option{
		nats.Name("CMP-Server"),
		nats.Timeout(10 * time.Second),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(5),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("NATS reconnected")
		}),
		nats.DisconnectHandler(func(nc *nats.Conn) {
			logger.Warn("NATS disconnected")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			logger.Warn("NATS connection closed")
		}),
	}

	conn, err := nats.Connect(config.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	logger.Info("Successfully connected to NATS")
	return &NATSService{
		conn:   conn,
		config: config,
	}, nil
}

// Publish publishes an event (implements Bus interface)
func (n *NATSService) Publish(ctx context.Context, event Event) error {
	subject := fmt.Sprintf("cmp.events.%s", event.Type)
	message, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return n.conn.Publish(subject, message)
}

// PublishToWorkspace publishes an event to a specific workspace
func (n *NATSService) PublishToWorkspace(ctx context.Context, workspaceID string, event *Event) error {
	event.WorkspaceID = workspaceID
	return n.Publish(ctx, *event)
}

// PublishToUser publishes an event to a specific user
func (n *NATSService) PublishToUser(ctx context.Context, userID string, event *Event) error {
	event.UserID = userID
	return n.Publish(ctx, *event)
}

// Subscribe subscribes to an event type (implements Bus interface)
func (n *NATSService) Subscribe(eventType string, handler EventHandler) error {
	subject := fmt.Sprintf("cmp.events.%s", eventType)
	_, err := n.conn.Subscribe(subject, func(msg *nats.Msg) {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			logger.Errorf("Failed to unmarshal event: %v", err)
			return
		}

		if err := handler.Handle(context.Background(), event); err != nil {
			logger.Errorf("Error processing event: %v", err)
		}
	})
	return err
}

// PublishMessage publishes a message to a subject (legacy method)
func (n *NATSService) PublishMessage(ctx context.Context, subject string, data interface{}) error {
	message, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return n.conn.Publish(subject, message)
}

// SubscribeToSubject subscribes to a subject with a handler
func (n *NATSService) SubscribeToSubject(ctx context.Context, subject string, handler func([]byte) error) error {
	_, err := n.conn.Subscribe(subject, func(msg *nats.Msg) {
		if err := handler(msg.Data); err != nil {
			logger.Errorf("Error processing message: %v", err)
		}
	})
	return err
}

// SubscribeWithQueue subscribes to a subject with queue group
func (n *NATSService) SubscribeWithQueue(ctx context.Context, subject, queue string, handler func([]byte) error) error {
	_, err := n.conn.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		if err := handler(msg.Data); err != nil {
			logger.Errorf("Error processing message: %v", err)
		}
	})
	return err
}

// Request sends a request and waits for a response
func (n *NATSService) Request(ctx context.Context, subject string, data interface{}, timeout time.Duration) ([]byte, error) {
	message, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	msg, err := n.conn.Request(subject, message, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return msg.Data, nil
}

// Close closes the NATS connection
func (n *NATSService) Close() {
	n.conn.Close()
}

// Health checks NATS health
func (n *NATSService) Health(ctx context.Context) error {
	if !n.conn.IsConnected() {
		return fmt.Errorf("NATS not connected")
	}
	return nil
}

// GetStats returns NATS connection statistics
func (n *NATSService) GetStats() map[string]interface{} {
	stats := n.conn.Statistics
	return map[string]interface{}{
		"in_msgs":    stats.InMsgs,
		"out_msgs":   stats.OutMsgs,
		"in_bytes":   stats.InBytes,
		"out_bytes":  stats.OutBytes,
		"reconnects": stats.Reconnects,
		"connected":  n.conn.IsConnected(),
	}
}
