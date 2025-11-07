package messaging

import (
	"context"
	"fmt"
	"time"

	"skyclust/internal/domain"

	"go.uber.org/zap"
)

// OutboxWorker processes events from the outbox table and publishes them to NATS
// Outbox 테이블의 이벤트를 처리하여 NATS로 발행하는 워커
type OutboxWorker struct {
	outboxRepo domain.OutboxRepository
	publisher  *Publisher
	logger     *zap.Logger
	config     OutboxWorkerConfig
	stopCh     chan struct{}
}

// OutboxWorkerConfig contains configuration for the outbox worker
// Outbox 워커 설정
type OutboxWorkerConfig struct {
	// BatchSize is the number of events to process in each batch
	// 각 배치에서 처리할 이벤트 수
	BatchSize int

	// PollInterval is the interval between polling for new events
	// 새 이벤트를 폴링하는 간격
	PollInterval time.Duration

	// MaxRetries is the maximum number of retries before marking an event as failed
	// 이벤트를 실패로 표시하기 전 최대 재시도 횟수
	MaxRetries int

	// RetryDelay is the delay between retries
	// 재시도 간 지연 시간
	RetryDelay time.Duration
}

// DefaultOutboxWorkerConfig returns default configuration for the outbox worker
// Outbox 워커의 기본 설정 반환
func DefaultOutboxWorkerConfig() OutboxWorkerConfig {
	return OutboxWorkerConfig{
		BatchSize:    10,
		PollInterval: 5 * time.Second,
		MaxRetries:   3,
		RetryDelay:   1 * time.Second,
	}
}

// NewOutboxWorker creates a new outbox worker
// 새로운 outbox 워커 생성
func NewOutboxWorker(outboxRepo domain.OutboxRepository, publisher *Publisher, logger *zap.Logger, config OutboxWorkerConfig) *OutboxWorker {
	if config.BatchSize == 0 {
		config.BatchSize = 10
	}
	if config.PollInterval == 0 {
		config.PollInterval = 5 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}

	return &OutboxWorker{
		outboxRepo: outboxRepo,
		publisher:  publisher,
		logger:     logger,
		config:     config,
		stopCh:     make(chan struct{}),
	}
}

// Start starts the outbox worker
// Outbox 워커 시작
func (w *OutboxWorker) Start(ctx context.Context) error {
	w.logger.Info("Starting outbox worker",
		zap.Int("batch_size", w.config.BatchSize),
		zap.Duration("poll_interval", w.config.PollInterval))

	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Outbox worker stopped (context cancelled)")
			return ctx.Err()
		case <-w.stopCh:
			w.logger.Info("Outbox worker stopped")
			return nil
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				w.logger.Warn("Error processing outbox batch",
					zap.Error(err))
				// Continue processing even if one batch fails
				// 한 배치가 실패해도 계속 처리
			}
		}
	}
}

// Stop stops the outbox worker
// Outbox 워커 중지
func (w *OutboxWorker) Stop() {
	close(w.stopCh)
}

// processBatch processes a batch of pending events
// 대기 중인 이벤트 배치 처리
func (w *OutboxWorker) processBatch(ctx context.Context) error {
	// Get pending events
	// 대기 중인 이벤트 조회
	events, err := w.outboxRepo.GetPendingEvents(ctx, w.config.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to get pending events: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	w.logger.Debug("Processing outbox batch",
		zap.Int("count", len(events)))

	// Mark events as processing
	// 이벤트를 처리 중으로 표시
	eventIDs := make([]string, len(events))
	for i, event := range events {
		eventIDs[i] = event.ID.String()
	}

	if err := w.outboxRepo.MarkAsProcessing(ctx, eventIDs); err != nil {
		return fmt.Errorf("failed to mark events as processing: %w", err)
	}

	// Process each event
	// 각 이벤트 처리
	for _, event := range events {
		if err := w.processEvent(ctx, event); err != nil {
			w.logger.Warn("Failed to process outbox event",
				zap.String("event_id", event.ID.String()),
				zap.String("topic", event.Topic),
				zap.Error(err))

			// Increment retry count
			// 재시도 횟수 증가
			if err := w.outboxRepo.IncrementRetryCount(ctx, event.ID.String()); err != nil {
				w.logger.Error("Failed to increment retry count",
					zap.String("event_id", event.ID.String()),
					zap.Error(err))
			}

			// Check if max retries exceeded
			// 최대 재시도 횟수 초과 확인
			if event.RetryCount+1 >= w.config.MaxRetries {
				errorMsg := err.Error()
				if updateErr := w.outboxRepo.UpdateStatus(ctx, event.ID.String(), domain.OutboxStatusFailed, &errorMsg); updateErr != nil {
					w.logger.Error("Failed to mark event as failed",
						zap.String("event_id", event.ID.String()),
						zap.Error(updateErr))
				}
			} else {
				// Reset to pending for retry
				// 재시도를 위해 pending으로 재설정
				if updateErr := w.outboxRepo.UpdateStatus(ctx, event.ID.String(), domain.OutboxStatusPending, nil); updateErr != nil {
					w.logger.Error("Failed to reset event status",
						zap.String("event_id", event.ID.String()),
						zap.Error(updateErr))
				}
			}

			// Wait before retrying
			// 재시도 전 대기
			time.Sleep(w.config.RetryDelay)
		} else {
			// Mark as published
			// 발행됨으로 표시
			if err := w.outboxRepo.MarkAsPublished(ctx, event.ID.String()); err != nil {
				w.logger.Error("Failed to mark event as published",
					zap.String("event_id", event.ID.String()),
					zap.Error(err))
			}
		}
	}

	return nil
}

// processEvent processes a single outbox event
// 단일 outbox 이벤트 처리
func (w *OutboxWorker) processEvent(ctx context.Context, event *domain.OutboxEvent) error {
	// Convert JSONBMap to map[string]interface{}
	// JSONBMap을 map[string]interface{}로 변환
	data := make(map[string]interface{})
	for k, v := range event.Data {
		data[k] = v
	}

	// Create Event from OutboxEvent
	// OutboxEvent에서 Event 생성
	messagingEvent := Event{
		Type:      event.Topic,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	// Set workspace ID if available
	// workspace ID가 있으면 설정
	if event.WorkspaceID != nil {
		messagingEvent.WorkspaceID = event.WorkspaceID.String()
	}

	// Publish to NATS
	// NATS로 발행
	if err := w.publisher.bus.Publish(ctx, messagingEvent); err != nil {
		return fmt.Errorf("failed to publish event to NATS: %w", err)
	}

	w.logger.Debug("Published outbox event to NATS",
		zap.String("event_id", event.ID.String()),
		zap.String("topic", event.Topic))

	return nil
}
