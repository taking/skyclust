package domain

import (
	"time"

	"github.com/google/uuid"
)

// OutboxEvent represents an event stored in the outbox table
// Outbox 테이블에 저장된 이벤트를 나타냄
type OutboxEvent struct {
	ID          uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Topic       string            `json:"topic" gorm:"not null;size:255;index"`
	EventType   string            `json:"event_type" gorm:"not null;size:100;index"`
	Data        JSONBMap          `json:"data" gorm:"type:jsonb;not null"`
	WorkspaceID *uuid.UUID        `json:"workspace_id" gorm:"type:uuid;index"`
	Status      OutboxEventStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	RetryCount  int               `json:"retry_count" gorm:"default:0"`
	LastError   *string           `json:"last_error" gorm:"type:text"`
	CreatedAt   time.Time         `json:"created_at" gorm:"not null;index"`
	PublishedAt *time.Time        `json:"published_at" gorm:"index"`
}

// OutboxEventStatus represents the status of an outbox event
// Outbox 이벤트의 상태를 나타냄
type OutboxEventStatus string

const (
	// OutboxStatusPending indicates the event is waiting to be published
	// 이벤트가 발행 대기 중임을 나타냄
	OutboxStatusPending OutboxEventStatus = "pending"

	// OutboxStatusProcessing indicates the event is being processed
	// 이벤트가 처리 중임을 나타냄
	OutboxStatusProcessing OutboxEventStatus = "processing"

	// OutboxStatusPublished indicates the event has been successfully published
	// 이벤트가 성공적으로 발행되었음을 나타냄
	OutboxStatusPublished OutboxEventStatus = "published"

	// OutboxStatusFailed indicates the event failed to publish after retries
	// 이벤트가 재시도 후에도 발행에 실패했음을 나타냄
	OutboxStatusFailed OutboxEventStatus = "failed"
)

// TableName specifies the table name for OutboxEvent
// OutboxEvent의 테이블 이름을 지정
func (OutboxEvent) TableName() string {
	return "outbox_events"
}
