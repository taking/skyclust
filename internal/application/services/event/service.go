package event

import (
	"context"
	"skyclust/internal/domain"
	"skyclust/internal/infrastructure/messaging"
)

// eventService implements the event business logic
type eventService struct {
	eventBus messaging.Bus
}

// NewService creates a new event service
func NewService(eventBus messaging.Bus) domain.EventService {
	return &eventService{
		eventBus: eventBus,
	}
}

// Publish publishes an event
func (s *eventService) Publish(ctx context.Context, eventType string, data interface{}) error {
	event := messaging.Event{
		Type: eventType,
		Data: map[string]interface{}{
			"data": data,
		},
	}
	return s.eventBus.Publish(ctx, event)
}

// Subscribe subscribes to an event type
func (s *eventService) Subscribe(ctx context.Context, eventType string, handler func(data interface{}) error) error {
	eventHandler := &eventHandler{handler: handler}
	return s.eventBus.Subscribe(eventType, eventHandler)
}

// eventHandler implements the EventHandler interface
type eventHandler struct {
	handler func(data interface{}) error
}

// Handle handles an event
func (h *eventHandler) Handle(ctx context.Context, event messaging.Event) error {
	return h.handler(event.Data)
}
