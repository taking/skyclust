package adapters

import (
	"context"
	"time"

	"skyclust/internal/domain"
	"skyclust/internal/infrastructure/messaging"
)

// EventAdapter adapts messaging.Publisher to domain.EventService
type EventAdapter struct {
	publisher *messaging.Publisher
}

// NewEventAdapter creates a new event adapter
func NewEventAdapter(publisher *messaging.Publisher) domain.EventService {
	return &EventAdapter{publisher: publisher}
}

// Publish publishes an event
func (a *EventAdapter) Publish(ctx context.Context, eventType string, data interface{}) error {
	// Convert to domain event
	event := domain.DomainEvent{
		Type:      eventType,
		Data:      a.convertDataToMap(data),
		Timestamp: time.Now(),
	}

	// Publish using messaging publisher's PublishToNATS method
	// Convert DomainEvent to messaging.Event format
	messagingEvent := messaging.Event{
		Type:      event.Type,
		Data:      event.Data,
		Timestamp: event.Timestamp.Unix(),
	}
	return a.publisher.PublishToNATS(ctx, eventType, messagingEvent)
}

// Subscribe subscribes to events
func (a *EventAdapter) Subscribe(ctx context.Context, eventType string, handler func(data interface{}) error) error {
	// Note: messaging.Publisher doesn't have Subscribe method
	// This would need to be implemented at the infrastructure level
	// For now, return nil (no-op)
	return nil
}

// convertDataToMap converts data to map[string]interface{}
func (a *EventAdapter) convertDataToMap(data interface{}) map[string]interface{} {
	if dataMap, ok := data.(map[string]interface{}); ok {
		return dataMap
	}

	// Try to convert to map
	result := make(map[string]interface{})
	if data != nil {
		result["data"] = data
	}
	return result
}
