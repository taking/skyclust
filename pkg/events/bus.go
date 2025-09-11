package events

import (
	"context"
	"sync"
)

// Event represents an event in the system
type Event struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	WorkspaceID string                 `json:"workspace_id"`
	UserID      string                 `json:"user_id"`
	Provider    string                 `json:"provider,omitempty"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   int64                  `json:"timestamp"`
}

// Bus defines the event bus interface
type Bus interface {
	// Event publishing
	Publish(ctx context.Context, event *Event) error
	PublishToWorkspace(ctx context.Context, workspaceID string, event *Event) error
	PublishToUser(ctx context.Context, userID string, event *Event) error

	// Event subscription
	Subscribe(eventType string, handler EventHandler) error
	Unsubscribe(eventType string, handler EventHandler) error

	// Event handling
	HandleEvent(ctx context.Context, event *Event) error
}

// EventHandler defines the event handler interface
type EventHandler interface {
	Handle(ctx context.Context, event *Event) error
}

// NewBus creates a new event bus
func NewBus() Bus {
	return &bus{
		handlers: make(map[string][]EventHandler),
	}
}

// NewNATSBus creates a new NATS-powered event bus
func NewNATSBus(natsService interface{}) Bus {
	return &natsBus{
		nats:     natsService,
		handlers: make(map[string][]EventHandler),
	}
}

type bus struct {
	handlers map[string][]EventHandler
	mutex    sync.RWMutex
}

// Publish publishes an event to all subscribers
func (b *bus) Publish(ctx context.Context, event *Event) error {
	b.mutex.RLock()
	handlers := b.handlers[event.Type]
	b.mutex.RUnlock()

	for _, handler := range handlers {
		go func(h EventHandler) {
			if err := h.Handle(ctx, event); err != nil {
				// Log error in production
			}
		}(handler)
	}

	return nil
}

// PublishToWorkspace publishes an event to a specific workspace
func (b *bus) PublishToWorkspace(ctx context.Context, workspaceID string, event *Event) error {
	event.WorkspaceID = workspaceID
	return b.Publish(ctx, event)
}

// PublishToUser publishes an event to a specific user
func (b *bus) PublishToUser(ctx context.Context, userID string, event *Event) error {
	event.UserID = userID
	return b.Publish(ctx, event)
}

// Subscribe subscribes to an event type
func (b *bus) Subscribe(eventType string, handler EventHandler) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
	return nil
}

// Unsubscribe unsubscribes from an event type
func (b *bus) Unsubscribe(eventType string, handler EventHandler) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	handlers := b.handlers[eventType]
	for i, h := range handlers {
		if h == handler {
			b.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	return nil
}

// HandleEvent handles an event
func (b *bus) HandleEvent(ctx context.Context, event *Event) error {
	return b.Publish(ctx, event)
}

// natsBus is a NATS-powered event bus
type natsBus struct {
	nats     interface{}
	handlers map[string][]EventHandler
	mutex    sync.RWMutex
}

// Publish publishes an event to all subscribers
func (nb *natsBus) Publish(ctx context.Context, event *Event) error {
	// Publish to NATS
	if natsService, ok := nb.nats.(interface {
		PublishToWorkspace(workspaceID string, event interface{}) error
		PublishToUser(userID string, event interface{}) error
	}); ok {
		if event.WorkspaceID != "" {
			if err := natsService.PublishToWorkspace(event.WorkspaceID, event); err != nil {
				return err
			}
		}

		if event.UserID != "" {
			if err := natsService.PublishToUser(event.UserID, event); err != nil {
				return err
			}
		}
	}

	// Also call local handlers
	nb.mutex.RLock()
	handlers := nb.handlers[event.Type]
	nb.mutex.RUnlock()

	for _, handler := range handlers {
		go func(h EventHandler) {
			if err := h.Handle(ctx, event); err != nil {
				// Log error in production
			}
		}(handler)
	}

	return nil
}

// PublishToWorkspace publishes an event to a specific workspace
func (nb *natsBus) PublishToWorkspace(ctx context.Context, workspaceID string, event *Event) error {
	event.WorkspaceID = workspaceID
	return nb.Publish(ctx, event)
}

// PublishToUser publishes an event to a specific user
func (nb *natsBus) PublishToUser(ctx context.Context, userID string, event *Event) error {
	event.UserID = userID
	return nb.Publish(ctx, event)
}

// Subscribe subscribes to an event type
func (nb *natsBus) Subscribe(eventType string, handler EventHandler) error {
	nb.mutex.Lock()
	defer nb.mutex.Unlock()

	nb.handlers[eventType] = append(nb.handlers[eventType], handler)
	return nil
}

// Unsubscribe unsubscribes from an event type
func (nb *natsBus) Unsubscribe(eventType string, handler EventHandler) error {
	nb.mutex.Lock()
	defer nb.mutex.Unlock()

	handlers := nb.handlers[eventType]
	for i, h := range handlers {
		if h == handler {
			nb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	return nil
}

// HandleEvent handles an event
func (nb *natsBus) HandleEvent(ctx context.Context, event *Event) error {
	return nb.Publish(ctx, event)
}
