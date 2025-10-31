package service

import (
	"context"
	"fmt"
	"skyclust/internal/domain"
	"time"

	"go.uber.org/zap"
)

type NotificationService struct {
	logger              *zap.Logger
	notificationRepo    domain.NotificationRepository
	preferencesRepo     domain.NotificationPreferencesRepository
	auditLogRepo        domain.AuditLogRepository
	userRepo            domain.UserRepository
	workspaceRepo       domain.WorkspaceRepository
	eventService        domain.EventService
}

func NewNotificationService(
	logger *zap.Logger,
	notificationRepo domain.NotificationRepository,
	preferencesRepo domain.NotificationPreferencesRepository,
	auditLogRepo domain.AuditLogRepository,
	userRepo domain.UserRepository,
	workspaceRepo domain.WorkspaceRepository,
	eventService domain.EventService,
) domain.NotificationService {
	return &NotificationService{
		logger:           logger,
		notificationRepo: notificationRepo,
		preferencesRepo:  preferencesRepo,
		auditLogRepo:     auditLogRepo,
		userRepo:         userRepo,
		workspaceRepo:    workspaceRepo,
		eventService:     eventService,
	}
}

// NotificationType represents different types of notifications
type NotificationType string

const (
	NotificationTypeInfo    NotificationType = "info"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeError   NotificationType = "error"
	NotificationTypeSuccess NotificationType = "success"
	NotificationTypeBudget  NotificationType = "budget"
	NotificationTypeVM      NotificationType = "vm"
	NotificationTypeSystem  NotificationType = "system"
)

// NotificationPriority represents notification priority levels
type NotificationPriority string

const (
	PriorityLow      NotificationPriority = "low"
	PriorityMedium   NotificationPriority = "medium"
	PriorityHigh     NotificationPriority = "high"
	PriorityCritical NotificationPriority = "critical"
)

// NotificationChannel represents delivery channels
type NotificationChannel string

const (
	ChannelInApp   NotificationChannel = "in_app"
	ChannelEmail   NotificationChannel = "email"
	ChannelBrowser NotificationChannel = "browser"
	ChannelSMS     NotificationChannel = "sms"
	ChannelWebhook NotificationChannel = "webhook"
)

// Notification represents a notification message
type Notification struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	WorkspaceID string                 `json:"workspace_id,omitempty"`
	Type        NotificationType       `json:"type"`
	Priority    NotificationPriority   `json:"priority"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Channels    []NotificationChannel  `json:"channels"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Read        bool                   `json:"read"`
	CreatedAt   time.Time              `json:"created_at"`
	ReadAt      *time.Time             `json:"read_at,omitempty"`
}

// NotificationTemplate represents a notification template
type NotificationTemplate struct {
	ID        string                `json:"id"`
	Name      string                `json:"name"`
	Type      NotificationType      `json:"type"`
	Priority  NotificationPriority  `json:"priority"`
	Title     string                `json:"title"`
	Message   string                `json:"message"`
	Channels  []NotificationChannel `json:"channels"`
	Variables []string              `json:"variables"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	UserID      string                                     `json:"user_id"`
	Email       bool                                       `json:"email"`
	Browser     bool                                       `json:"browser"`
	SMS         bool                                       `json:"sms"`
	InApp       bool                                       `json:"in_app"`
	Webhook     bool                                       `json:"webhook"`
	Preferences map[NotificationType]bool                  `json:"preferences"`
	Channels    map[NotificationType][]NotificationChannel `json:"channels"`
}

// SendToWorkspace sends a notification to all users in a workspace
func (s *NotificationService) SendToWorkspace(ctx context.Context, workspaceID string, notification *Notification) error {
	// TODO: Implement GetWorkspaceMembers in workspace repository
	// For now, we'll just log the notification
	s.logger.Info("Workspace notification sent",
		zap.String("workspace_id", workspaceID),
		zap.String("notification_id", notification.ID))

	return nil
}

// GetUserNotifications retrieves notifications for a user
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID string, limit, offset int) ([]*Notification, error) {
	// This would typically query a notifications table
	// For now, we'll return mock data
	return s.getMockNotifications(userID, limit, offset), nil
}

// CreateTemplate creates a notification template
func (s *NotificationService) CreateTemplate(ctx context.Context, template *NotificationTemplate) error {
	// This would typically save to a templates table
	// For now, we'll just log the action
	s.logger.Info("Notification template created", zap.String("template_id", template.ID))
	return nil
}

// SendTemplateNotification sends a notification using a template
func (s *NotificationService) SendTemplateNotification(ctx context.Context, templateID string, userID string, variables map[string]interface{}) error {
	// This would typically load the template and substitute variables
	// For now, we'll create a basic notification
	notification := &Notification{
		ID:        fmt.Sprintf("template-%s-%d", templateID, time.Now().Unix()),
		UserID:    userID,
		Type:      NotificationTypeInfo,
		Priority:  PriorityMedium,
		Title:     "Template Notification",
		Message:   "This is a template-based notification",
		Channels:  []NotificationChannel{ChannelInApp},
		Data:      variables,
		CreatedAt: time.Now(),
	}

	// Convert to domain.Notification
	domainNotification := &domain.Notification{
		ID:        notification.ID,
		UserID:    notification.UserID,
		Type:      string(notification.Type),
		Title:     notification.Title,
		Message:   notification.Message,
		Category:  "system", // Default category
		Priority:  string(notification.Priority),
		IsRead:    notification.Read,
		Data:      fmt.Sprintf("%v", notification.Data),
		CreatedAt: notification.CreatedAt,
	}

	return s.SendNotification(ctx, notification.UserID, domainNotification)
}

// Helper methods

func (s *NotificationService) getMockNotifications(userID string, limit, offset int) []*Notification {
	notifications := []*Notification{
		{
			ID:        "notif-1",
			UserID:    userID,
			Type:      NotificationTypeInfo,
			Priority:  PriorityMedium,
			Title:     "Welcome to SkyClust",
			Message:   "Your account has been successfully created.",
			Channels:  []NotificationChannel{ChannelInApp},
			Read:      false,
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:        "notif-2",
			UserID:    userID,
			Type:      NotificationTypeBudget,
			Priority:  PriorityHigh,
			Title:     "Budget Alert",
			Message:   "Your workspace budget is at 85% of the limit.",
			Channels:  []NotificationChannel{ChannelInApp, ChannelEmail},
			Read:      false,
			CreatedAt: time.Now().Add(-30 * time.Minute),
		},
		{
			ID:        "notif-3",
			UserID:    userID,
			Type:      NotificationTypeVM,
			Priority:  PriorityMedium,
			Title:     "VM Status Update",
			Message:   "VM 'web-server-01' has started successfully.",
			Channels:  []NotificationChannel{ChannelInApp},
			Read:      true,
			CreatedAt: time.Now().Add(-2 * time.Hour),
			ReadAt:    &[]time.Time{time.Now().Add(-1 * time.Hour)}[0],
		},
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start >= len(notifications) {
		return []*Notification{}
	}
	if end > len(notifications) {
		end = len(notifications)
	}

	return notifications[start:end]
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(ctx context.Context, notification *domain.Notification) error {
	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to create notification: %v", err), 500)
	}

	// Publish event if event service is available
	if s.eventService != nil {
		_ = s.eventService.Publish(ctx, "notification.created", notification)
	}

	return nil
}

// GetNotification gets a notification by ID
func (s *NotificationService) GetNotification(ctx context.Context, userID, notificationID string) (*domain.Notification, error) {
	notification, err := s.notificationRepo.GetByID(ctx, userID, notificationID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("notification not found: %s", notificationID), 404)
	}
	return notification, nil
}

// GetNotifications gets notifications for a user
func (s *NotificationService) GetNotifications(ctx context.Context, userID string, limit, offset int, unreadOnly bool, category, priority string) ([]*domain.Notification, int, error) {
	notifications, total, err := s.notificationRepo.GetByUserID(ctx, userID, limit, offset, unreadOnly, category, priority)
	if err != nil {
		return nil, 0, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get notifications: %v", err), 500)
	}
	return notifications, total, nil
}

// UpdateNotification updates a notification
func (s *NotificationService) UpdateNotification(ctx context.Context, notification *domain.Notification) error {
	if err := s.notificationRepo.Update(ctx, notification); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to update notification: %v", err), 500)
	}
	return nil
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(ctx context.Context, userID, notificationID string) error {
	if err := s.notificationRepo.Delete(ctx, userID, notificationID); err != nil {
		return domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("notification not found: %s", notificationID), 404)
	}
	return nil
}

// DeleteNotifications deletes multiple notifications
func (s *NotificationService) DeleteNotifications(ctx context.Context, userID string, notificationIDs []string) error {
	if err := s.notificationRepo.DeleteMultiple(ctx, userID, notificationIDs); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to delete notifications: %v", err), 500)
	}
	return nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, userID, notificationID string) error {
	if err := s.notificationRepo.MarkAsRead(ctx, userID, notificationID); err != nil {
		return domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("notification not found: %s", notificationID), 404)
	}
	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	if err := s.notificationRepo.MarkAllAsRead(ctx, userID); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to mark all notifications as read: %v", err), 500)
	}
	return nil
}

// GetNotificationPreferences gets notification preferences for a user
func (s *NotificationService) GetNotificationPreferences(ctx context.Context, userID string) (*domain.NotificationPreferences, error) {
	preferences, err := s.preferencesRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get notification preferences: %v", err), 500)
	}
	return preferences, nil
}

// UpdateNotificationPreferences updates notification preferences for a user
func (s *NotificationService) UpdateNotificationPreferences(ctx context.Context, userID string, preferences *domain.NotificationPreferences) error {
	// Ensure userID matches
	preferences.UserID = userID
	
	// Use Upsert to create or update
	if err := s.preferencesRepo.Upsert(ctx, preferences); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to update notification preferences: %v", err), 500)
	}
	return nil
}

// GetNotificationStats gets notification statistics for a user
func (s *NotificationService) GetNotificationStats(ctx context.Context, userID string) (*domain.NotificationStats, error) {
	stats, err := s.notificationRepo.GetStats(ctx, userID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get notification stats: %v", err), 500)
	}
	return stats, nil
}

// SendNotification sends a notification to a user
func (s *NotificationService) SendNotification(ctx context.Context, userID string, notification *domain.Notification) error {
	// Ensure userID matches
	notification.UserID = userID
	
	// Create notification in database
	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to send notification: %v", err), 500)
	}
	
	// Publish event if event service is available
	if s.eventService != nil {
		_ = s.eventService.Publish(ctx, "notification.created", notification)
	}
	
	return nil
}

// SendBulkNotification sends a notification to multiple users
func (s *NotificationService) SendBulkNotification(ctx context.Context, userIDs []string, notification *domain.Notification) error {
	// Create notification for each user
	for _, userID := range userIDs {
		bulkNotification := *notification // Copy
		bulkNotification.UserID = userID
		// Ensure unique ID per user (append userID to notification ID)
		bulkNotification.ID = fmt.Sprintf("%s-%s", notification.ID, userID)

		if err := s.notificationRepo.Create(ctx, &bulkNotification); err != nil {
			s.logger.Warn("Failed to send notification to user",
				zap.String("user_id", userID),
				zap.Error(err))
			// Continue with other users instead of failing completely
			continue
		}

		// Publish event if event service is available
		if s.eventService != nil {
			_ = s.eventService.Publish(ctx, "notification.created", &bulkNotification)
		}
	}

	s.logger.Info("Bulk notification sent",
		zap.Strings("user_ids", userIDs),
		zap.String("title", notification.Title),
		zap.Int("count", len(userIDs)))

	return nil
}

// CleanupOldNotifications removes old notifications
func (s *NotificationService) CleanupOldNotifications(ctx context.Context, olderThan time.Duration) error {
	// Call repository cleanup method
	if err := s.notificationRepo.CleanupOld(ctx, olderThan); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to cleanup old notifications: %v", err), 500)
	}

	s.logger.Info("Cleanup old notifications completed",
		zap.Duration("older_than", olderThan))

	return nil
}
