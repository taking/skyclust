package messaging

import (
	"context"
	"sync"
)

// Bus defines the interface for event publishing
type Bus interface {
	Publish(ctx context.Context, event Event) error
	PublishToWorkspace(ctx context.Context, workspaceID string, event *Event) error
	PublishToUser(ctx context.Context, userID string, event *Event) error
	Subscribe(eventType string, handler EventHandler) error
	Health(ctx context.Context) error
}

// EventHandler defines the interface for event handlers
type EventHandler interface {
	Handle(ctx context.Context, event Event) error
}

// Event represents a domain event
type Event struct {
	Type        string                 `json:"type"`
	WorkspaceID string                 `json:"workspace_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   int64                  `json:"timestamp"`
}

// LocalBus implements a local event bus
type LocalBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

// NewLocalBus creates a new local event bus
func NewLocalBus() *LocalBus {
	return &LocalBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Publish publishes an event
func (b *LocalBus) Publish(ctx context.Context, event Event) error {
	b.mu.RLock()
	handlers := b.handlers[event.Type]
	b.mu.RUnlock()

	for _, handler := range handlers {
		go func(h EventHandler) {
			//nolint:staticcheck // SA9003: intentionally ignore handler errors to prevent event publishing failure
			if err := h.Handle(ctx, event); err != nil {
				// Log error but don't fail the publish
				// TODO: Add proper logging here
			}
		}(handler)
	}

	return nil
}

// Subscribe subscribes to an event type
func (b *LocalBus) Subscribe(eventType string, handler EventHandler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
	return nil
}

// PublishToWorkspace publishes an event to a specific workspace
func (b *LocalBus) PublishToWorkspace(ctx context.Context, workspaceID string, event *Event) error {
	event.WorkspaceID = workspaceID
	return b.Publish(ctx, *event)
}

// PublishToUser publishes an event to a specific user
func (b *LocalBus) PublishToUser(ctx context.Context, userID string, event *Event) error {
	event.UserID = userID
	return b.Publish(ctx, *event)
}

// Health checks the health of the event bus
func (b *LocalBus) Health(ctx context.Context) error {
	return nil
}
