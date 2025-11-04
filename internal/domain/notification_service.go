package domain

import (
	"context"
	"time"
)

// NotificationService 알림 서비스 인터페이스
type NotificationService interface {
	// 알림 CRUD
	CreateNotification(ctx context.Context, notification *Notification) error
	GetNotification(ctx context.Context, userID, notificationID string) (*Notification, error)
	GetNotifications(ctx context.Context, userID string, limit, offset int, unreadOnly bool, category, priority string) ([]*Notification, int, error)
	UpdateNotification(ctx context.Context, notification *Notification) error
	DeleteNotification(ctx context.Context, userID, notificationID string) error
	DeleteNotifications(ctx context.Context, userID string, notificationIDs []string) error

	// 읽음 처리
	MarkAsRead(ctx context.Context, userID, notificationID string) error
	MarkAllAsRead(ctx context.Context, userID string) error

	// 알림 설정
	GetNotificationPreferences(ctx context.Context, userID string) (*NotificationPreferences, error)
	UpdateNotificationPreferences(ctx context.Context, userID string, preferences *NotificationPreferences) error

	// 통계
	GetNotificationStats(ctx context.Context, userID string) (*NotificationStats, error)

	// 알림 전송
	SendNotification(ctx context.Context, userID string, notification *Notification) error
	SendBulkNotification(ctx context.Context, userIDs []string, notification *Notification) error

	// 알림 정리
	CleanupOldNotifications(ctx context.Context, olderThan time.Duration) error
}

