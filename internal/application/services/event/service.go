package event

import (
	"context"
	"skyclust/internal/domain"
	"skyclust/internal/infrastructure/messaging"
)

// eventService: 이벤트 비즈니스 로직 구현체
type eventService struct {
	eventBus messaging.Bus
}

// NewService: 새로운 이벤트 서비스를 생성합니다
func NewService(eventBus messaging.Bus) domain.EventService {
	return &eventService{
		eventBus: eventBus,
	}
}

// Publish: 이벤트를 발행합니다
func (s *eventService) Publish(ctx context.Context, eventType string, data interface{}) error {
	event := messaging.Event{
		Type: eventType,
		Data: map[string]interface{}{
			"data": data,
		},
	}
	return s.eventBus.Publish(ctx, event)
}

// Subscribe: 이벤트 타입에 구독합니다
func (s *eventService) Subscribe(ctx context.Context, eventType string, handler func(data interface{}) error) error {
	eventHandler := &eventHandler{handler: handler}
	return s.eventBus.Subscribe(eventType, eventHandler)
}

// eventHandler: EventHandler 인터페이스 구현체
type eventHandler struct {
	handler func(data interface{}) error
}

// Handle: 이벤트를 처리합니다
func (h *eventHandler) Handle(ctx context.Context, event messaging.Event) error {
	return h.handler(event.Data)
}
