package services

import (
	"skyclust/internal/domain"
)

// NotificationService defines the interface for notification operations
type NotificationService interface {
	// CreateNotification creates a new notification
	CreateNotification(notification *domain.Notification) error

	// GetNotification retrieves a notification by ID
	GetNotification(id string) (*domain.Notification, error)

	// GetNotifications retrieves notifications for a user
	GetNotifications(userID string, limit, offset int) ([]*domain.Notification, error)

	// UpdateNotification updates a notification
	UpdateNotification(notification *domain.Notification) error

	// DeleteNotification deletes a notification
	DeleteNotification(id string) error

	// MarkAsRead marks a notification as read
	MarkAsRead(id string) error

	// MarkAllAsRead marks all notifications as read for a user
	MarkAllAsRead(userID string) error

	// GetUnreadCount returns the count of unread notifications for a user
	GetUnreadCount(userID string) (int, error)

	// SendNotification sends a notification to a user
	SendNotification(userID string, title, message string, notificationType string) error

	// GetNotificationSettings retrieves notification settings for a user
	GetNotificationSettings(userID string) (interface{}, error)

	// UpdateNotificationSettings updates notification settings for a user
	UpdateNotificationSettings(userID string, settings interface{}) error
}
