package domain

import (
	"context"
)

// OutboxRepository defines the interface for outbox event persistence
// Outbox 이벤트 영속성을 위한 인터페이스 정의
type OutboxRepository interface {
	// Create stores a new outbox event
	// 새로운 outbox 이벤트를 저장
	Create(ctx context.Context, event *OutboxEvent) error

	// GetPendingEvents retrieves pending events up to the specified limit
	// 지정된 제한까지 대기 중인 이벤트를 조회
	GetPendingEvents(ctx context.Context, limit int) ([]*OutboxEvent, error)

	// UpdateStatus updates the status of an outbox event
	// Outbox 이벤트의 상태를 업데이트
	UpdateStatus(ctx context.Context, id string, status OutboxEventStatus, errorMsg *string) error

	// MarkAsProcessing marks events as processing
	// 이벤트를 처리 중으로 표시
	MarkAsProcessing(ctx context.Context, ids []string) error

	// MarkAsPublished marks an event as published
	// 이벤트를 발행됨으로 표시
	MarkAsPublished(ctx context.Context, id string) error

	// IncrementRetryCount increments the retry count for an event
	// 이벤트의 재시도 횟수를 증가
	IncrementRetryCount(ctx context.Context, id string) error

	// GetFailedEvents retrieves failed events for manual review
	// 수동 검토를 위한 실패한 이벤트를 조회
	GetFailedEvents(ctx context.Context, limit int) ([]*OutboxEvent, error)

	// DeleteOldEvents deletes events older than the specified duration
	// 지정된 기간보다 오래된 이벤트를 삭제
	DeleteOldEvents(ctx context.Context, olderThanDays int) error
}
