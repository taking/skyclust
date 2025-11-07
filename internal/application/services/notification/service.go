package notification

import (
	"context"
	"fmt"
	"skyclust/internal/domain"
	"time"

	"go.uber.org/zap"
)

// Service: 알림 서비스 구현체
type Service struct {
	logger           *zap.Logger
	notificationRepo domain.NotificationRepository
	preferencesRepo  domain.NotificationPreferencesRepository
	auditLogRepo     domain.AuditLogRepository
	userRepo         domain.UserRepository
	workspaceRepo    domain.WorkspaceRepository
	eventService     domain.EventService
}

// NewService: 새로운 알림 서비스를 생성합니다
func NewService(
	logger *zap.Logger,
	notificationRepo domain.NotificationRepository,
	preferencesRepo domain.NotificationPreferencesRepository,
	auditLogRepo domain.AuditLogRepository,
	userRepo domain.UserRepository,
	workspaceRepo domain.WorkspaceRepository,
	eventService domain.EventService,
) domain.NotificationService {
	return &Service{
		logger:           logger,
		notificationRepo: notificationRepo,
		preferencesRepo:  preferencesRepo,
		auditLogRepo:     auditLogRepo,
		userRepo:         userRepo,
		workspaceRepo:    workspaceRepo,
		eventService:     eventService,
	}
}

// SendToWorkspace: 워크스페이스의 모든 사용자에게 알림을 전송합니다
func (s *Service) SendToWorkspace(ctx context.Context, workspaceID string, notification *Notification) error {
	// TODO: Implement GetWorkspaceMembers in workspace repository
	// For now, we'll just log the notification
	s.logger.Info("Workspace notification sent",
		zap.String("workspace_id", workspaceID),
		zap.String("notification_id", notification.ID))

	return nil
}

// GetUserNotifications: 사용자의 알림을 조회합니다
func (s *Service) GetUserNotifications(ctx context.Context, userID string, limit, offset int) ([]*Notification, error) {
	// This would typically query a notifications table
	// For now, we'll return mock data
	return s.getMockNotifications(userID, limit, offset), nil
}

// CreateTemplate: 알림 템플릿을 생성합니다
func (s *Service) CreateTemplate(ctx context.Context, template *NotificationTemplate) error {
	// This would typically save to a templates table
	// For now, we'll just log the action
	s.logger.Info("Notification template created", zap.String("template_id", template.ID))
	return nil
}

// SendTemplateNotification: 템플릿을 사용하여 알림을 전송합니다
func (s *Service) SendTemplateNotification(ctx context.Context, templateID string, userID string, variables map[string]interface{}) error {
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

func (s *Service) getMockNotifications(userID string, limit, offset int) []*Notification {
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

// CreateNotification: 새로운 알림을 생성합니다
func (s *Service) CreateNotification(ctx context.Context, notification *domain.Notification) error {
	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to create notification: %v", err), 500)
	}

	// Publish event if event service is available
	if s.eventService != nil {
		_ = s.eventService.Publish(ctx, "notification.created", notification)
	}

	return nil
}

// GetNotification: ID로 알림을 조회합니다
func (s *Service) GetNotification(ctx context.Context, userID, notificationID string) (*domain.Notification, error) {
	notification, err := s.notificationRepo.GetByID(ctx, userID, notificationID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("notification not found: %s", notificationID), 404)
	}
	return notification, nil
}

// GetNotifications: 사용자의 알림 목록을 조회합니다
func (s *Service) GetNotifications(ctx context.Context, userID string, limit, offset int, unreadOnly bool, category, priority string) ([]*domain.Notification, int, error) {
	notifications, total, err := s.notificationRepo.GetByUserID(ctx, userID, limit, offset, unreadOnly, category, priority)
	if err != nil {
		return nil, 0, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get notifications: %v", err), 500)
	}
	return notifications, total, nil
}

// UpdateNotification: 알림을 업데이트합니다
func (s *Service) UpdateNotification(ctx context.Context, notification *domain.Notification) error {
	if err := s.notificationRepo.Update(ctx, notification); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to update notification: %v", err), 500)
	}
	return nil
}

// DeleteNotification: 알림을 삭제합니다
func (s *Service) DeleteNotification(ctx context.Context, userID, notificationID string) error {
	if err := s.notificationRepo.Delete(ctx, userID, notificationID); err != nil {
		return domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("notification not found: %s", notificationID), 404)
	}
	return nil
}

// DeleteNotifications: 여러 알림을 삭제합니다
func (s *Service) DeleteNotifications(ctx context.Context, userID string, notificationIDs []string) error {
	if err := s.notificationRepo.DeleteMultiple(ctx, userID, notificationIDs); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to delete notifications: %v", err), 500)
	}
	return nil
}

// MarkAsRead: 알림을 읽음으로 표시합니다
func (s *Service) MarkAsRead(ctx context.Context, userID, notificationID string) error {
	if err := s.notificationRepo.MarkAsRead(ctx, userID, notificationID); err != nil {
		return domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("notification not found: %s", notificationID), 404)
	}
	return nil
}

// MarkAllAsRead: 사용자의 모든 알림을 읽음으로 표시합니다
func (s *Service) MarkAllAsRead(ctx context.Context, userID string) error {
	if err := s.notificationRepo.MarkAllAsRead(ctx, userID); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to mark all notifications as read: %v", err), 500)
	}
	return nil
}

// GetNotificationPreferences: 사용자의 알림 설정을 조회합니다
func (s *Service) GetNotificationPreferences(ctx context.Context, userID string) (*domain.NotificationPreferences, error) {
	preferences, err := s.preferencesRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get notification preferences: %v", err), 500)
	}
	return preferences, nil
}

// UpdateNotificationPreferences: 사용자의 알림 설정을 업데이트합니다
func (s *Service) UpdateNotificationPreferences(ctx context.Context, userID string, preferences *domain.NotificationPreferences) error {
	// Ensure userID matches
	preferences.UserID = userID

	// Use Upsert to create or update
	if err := s.preferencesRepo.Upsert(ctx, preferences); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to update notification preferences: %v", err), 500)
	}
	return nil
}

// GetNotificationStats: 사용자의 알림 통계를 조회합니다
func (s *Service) GetNotificationStats(ctx context.Context, userID string) (*domain.NotificationStats, error) {
	stats, err := s.notificationRepo.GetStats(ctx, userID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get notification stats: %v", err), 500)
	}
	return stats, nil
}

// SendNotification: 사용자에게 알림을 전송합니다
func (s *Service) SendNotification(ctx context.Context, userID string, notification *domain.Notification) error {
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

// SendBulkNotification: 여러 사용자에게 알림을 전송합니다
func (s *Service) SendBulkNotification(ctx context.Context, userIDs []string, notification *domain.Notification) error {
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

// CleanupOldNotifications: 오래된 알림을 제거합니다
func (s *Service) CleanupOldNotifications(ctx context.Context, olderThan time.Duration) error {
	// Call repository cleanup method
	if err := s.notificationRepo.CleanupOld(ctx, olderThan); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to cleanup old notifications: %v", err), 500)
	}

	s.logger.Info("Cleanup old notifications completed",
		zap.Duration("older_than", olderThan))

	return nil
}
