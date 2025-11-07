package postgres

import (
	"context"
	"fmt"
	"time"

	"skyclust/internal/domain"
	"skyclust/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// outboxRepository implements the OutboxRepository interface
// OutboxRepository 인터페이스 구현
type outboxRepository struct {
	db *gorm.DB
}

// NewOutboxRepository creates a new outbox repository
// 새로운 outbox repository 생성
func NewOutboxRepository(db *gorm.DB) domain.OutboxRepository {
	return &outboxRepository{db: db}
}

// Create stores a new outbox event
// 새로운 outbox 이벤트를 저장
func (r *outboxRepository) Create(ctx context.Context, event *domain.OutboxEvent) error {
	// 트랜잭션 지원: context에 트랜잭션이 있으면 사용
	db := GetTransaction(ctx, r.db)

	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	if event.Status == "" {
		event.Status = domain.OutboxStatusPending
	}

	if err := db.WithContext(ctx).Create(event).Error; err != nil {
		logger.Errorf("Failed to create outbox event: %v", err)
		return fmt.Errorf("failed to create outbox event: %w", err)
	}

	logger.Debug(fmt.Sprintf("Created outbox event: id=%s, topic=%s, event_type=%s",
		event.ID.String(), event.Topic, event.EventType))

	return nil
}

// GetPendingEvents retrieves pending events up to the specified limit
// 지정된 제한까지 대기 중인 이벤트를 조회
func (r *outboxRepository) GetPendingEvents(ctx context.Context, limit int) ([]*domain.OutboxEvent, error) {
	var events []*domain.OutboxEvent

	query := r.db.WithContext(ctx).
		Where("status = ?", domain.OutboxStatusPending).
		Order("created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&events).Error; err != nil {
		logger.Errorf("Failed to get pending outbox events: %v", err)
		return nil, fmt.Errorf("failed to get pending outbox events: %w", err)
	}

	return events, nil
}

// UpdateStatus updates the status of an outbox event
// Outbox 이벤트의 상태를 업데이트
func (r *outboxRepository) UpdateStatus(ctx context.Context, id string, status domain.OutboxEventStatus, errorMsg *string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if errorMsg != nil {
		updates["last_error"] = *errorMsg
	}

	if status == domain.OutboxStatusPublished {
		now := time.Now()
		updates["published_at"] = &now
	}

	if err := r.db.WithContext(ctx).
		Model(&domain.OutboxEvent{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		logger.Errorf("Failed to update outbox event status: %v", err)
		return fmt.Errorf("failed to update outbox event status: %w", err)
	}

	return nil
}

// MarkAsProcessing marks events as processing
// 이벤트를 처리 중으로 표시
func (r *outboxRepository) MarkAsProcessing(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).
		Model(&domain.OutboxEvent{}).
		Where("id IN ?", ids).
		Update("status", domain.OutboxStatusProcessing).Error; err != nil {
		logger.Errorf("Failed to mark outbox events as processing: %v", err)
		return fmt.Errorf("failed to mark outbox events as processing: %w", err)
	}

	return nil
}

// MarkAsPublished marks an event as published
// 이벤트를 발행됨으로 표시
func (r *outboxRepository) MarkAsPublished(ctx context.Context, id string) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&domain.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       domain.OutboxStatusPublished,
			"published_at": &now,
		}).Error; err != nil {
		logger.Errorf("Failed to mark outbox event as published: %v", err)
		return fmt.Errorf("failed to mark outbox event as published: %w", err)
	}

	return nil
}

// IncrementRetryCount increments the retry count for an event
// 이벤트의 재시도 횟수를 증가
func (r *outboxRepository) IncrementRetryCount(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).
		Model(&domain.OutboxEvent{}).
		Where("id = ?", id).
		UpdateColumn("retry_count", gorm.Expr("retry_count + 1")).Error; err != nil {
		logger.Errorf("Failed to increment retry count: %v", err)
		return fmt.Errorf("failed to increment retry count: %w", err)
	}

	return nil
}

// GetFailedEvents retrieves failed events for manual review
// 수동 검토를 위한 실패한 이벤트를 조회
func (r *outboxRepository) GetFailedEvents(ctx context.Context, limit int) ([]*domain.OutboxEvent, error) {
	var events []*domain.OutboxEvent

	query := r.db.WithContext(ctx).
		Where("status = ?", domain.OutboxStatusFailed).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&events).Error; err != nil {
		logger.Errorf("Failed to get failed outbox events: %v", err)
		return nil, fmt.Errorf("failed to get failed outbox events: %w", err)
	}

	return events, nil
}

// DeleteOldEvents deletes events older than the specified duration
// 지정된 기간보다 오래된 이벤트를 삭제
func (r *outboxRepository) DeleteOldEvents(ctx context.Context, olderThanDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)

	result := r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", domain.OutboxStatusPublished, cutoffDate).
		Delete(&domain.OutboxEvent{})

	if result.Error != nil {
		logger.Errorf("Failed to delete old outbox events: %v", result.Error)
		return fmt.Errorf("failed to delete old outbox events: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		logger.Info(fmt.Sprintf("Deleted old outbox events: count=%d, cutoff_date=%s",
			result.RowsAffected, cutoffDate))
	}

	return nil
}
