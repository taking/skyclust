package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSBus implements event bus using NATS
type NATSBus struct {
	conn *nats.Conn
}

// NewNATSBus creates a new NATS event bus
func NewNATSBus(conn *nats.Conn) *NATSBus {
	return &NATSBus{
		conn: conn,
	}
}

// Publish publishes an event
func (b *NATSBus) Publish(ctx context.Context, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	subject := fmt.Sprintf("events.%s", event.Type)
	return b.conn.Publish(subject, data)
}

// Subscribe subscribes to an event type
func (b *NATSBus) Subscribe(eventType string, handler EventHandler) error {
	subject := fmt.Sprintf("events.%s", eventType)

	_, err := b.conn.Subscribe(subject, func(msg *nats.Msg) {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			// Log error but don't fail
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		//nolint:staticcheck // SA9003: intentionally ignore handler errors to prevent messaging failure
		if err := handler.Handle(ctx, event); err != nil {
			// Log error but don't fail
			// TODO: Add proper logging here
		}
	})

	return err
}

// PublishToWorkspace publishes an event to a specific workspace
func (b *NATSBus) PublishToWorkspace(ctx context.Context, workspaceID string, event *Event) error {
	event.WorkspaceID = workspaceID
	return b.Publish(ctx, *event)
}

// PublishToUser publishes an event to a specific user
func (b *NATSBus) PublishToUser(ctx context.Context, userID string, event *Event) error {
	event.UserID = userID
	return b.Publish(ctx, *event)
}

// Health checks the health of the event bus
func (b *NATSBus) Health(ctx context.Context) error {
	if !b.conn.IsConnected() {
		return fmt.Errorf("NATS connection is not healthy")
	}
	return nil
}
