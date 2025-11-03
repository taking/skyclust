package domain

import (
	"context"
	"time"
)

// NotificationRepository 알림 저장소 인터페이스
type NotificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
	GetByID(ctx context.Context, userID, notificationID string) (*Notification, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int, unreadOnly bool, category, priority string) ([]*Notification, int, error)
	Update(ctx context.Context, notification *Notification) error
	Delete(ctx context.Context, userID, notificationID string) error
	DeleteMultiple(ctx context.Context, userID string, notificationIDs []string) error
	MarkAsRead(ctx context.Context, userID, notificationID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	GetStats(ctx context.Context, userID string) (*NotificationStats, error)
	CleanupOld(ctx context.Context, olderThan time.Duration) error
}

// NotificationPreferencesRepository 알림 설정 저장소 인터페이스
type NotificationPreferencesRepository interface {
	GetByUserID(ctx context.Context, userID string) (*NotificationPreferences, error)
	Create(ctx context.Context, preferences *NotificationPreferences) error
	Update(ctx context.Context, preferences *NotificationPreferences) error
	Upsert(ctx context.Context, preferences *NotificationPreferences) error
}

