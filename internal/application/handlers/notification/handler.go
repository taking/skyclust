package notification

import (
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/readability"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler: 알림 관리 작업을 처리하는 핸들러
type Handler struct {
	*handlers.BaseHandler
	notificationService domain.NotificationService
	readabilityHelper   *readability.ReadabilityHelper
}

// NewHandler: 새로운 알림 핸들러를 생성합니다
func NewHandler(notificationService domain.NotificationService) *Handler {
	return &Handler{
		BaseHandler:         handlers.NewBaseHandler("notification"),
		notificationService: notificationService,
		readabilityHelper:   readability.NewReadabilityHelper(),
	}
}

// GetNotifications: 알림 목록을 조회합니다 (데코레이터 패턴 사용)
func (h *Handler) GetNotifications(c *gin.Context) {
	handler := h.Compose(
		h.getNotificationsHandler(),
		h.StandardCRUDDecorators("get_notifications")...,
	)

	handler(c)
}

// getNotificationsHandler: 알림 조회의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) getNotificationsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_notifications")
			return
		}
		filters := h.parseNotificationFilters(c)

		h.logNotificationsRequest(c, userID, filters)

		notifications, total, err := h.notificationService.GetNotifications(
			c.Request.Context(),
			userID.String(),
			filters.Limit,
			filters.Offset,
			filters.UnreadOnly,
			filters.Category,
			filters.Priority,
		)
		if err != nil {
			h.HandleError(c, err, "get_notifications")
			return
		}

		responses := h.convertToNotificationResponses(notifications)

		// Calculate page from offset for pagination metadata
		page := (filters.Offset / filters.Limit) + 1
		if filters.Offset == 0 {
			page = 1
		}
		paginationMeta := h.CalculatePaginationMeta(int64(total), page, filters.Limit)

		h.OK(c, gin.H{
			"notifications": responses,
			"pagination":    paginationMeta,
		}, "Notifications retrieved successfully")
	}
}

// GetNotification: 특정 알림을 조회합니다 (데코레이터 패턴 사용)
func (h *Handler) GetNotification(c *gin.Context) {
	handler := h.Compose(
		h.getNotificationHandler(),
		h.StandardCRUDDecorators("get_notification")...,
	)

	handler(c)
}

// getNotificationHandler: 알림 조회의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) getNotificationHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_notifications")
			return
		}
		notificationID, err := h.ExtractPathParam(c, "id")
		if err != nil {
			h.HandleError(c, err, "get_notification")
			return
		}

		h.logNotificationRequest(c, userID, notificationID)

		notification, err := h.notificationService.GetNotification(
			c.Request.Context(),
			userID.String(),
			notificationID.String(),
		)
		if err != nil {
			h.HandleError(c, err, "get_notification")
			return
		}

		h.OK(c, notification, "Notification retrieved successfully")
	}
}

// UpdateNotification: 알림을 업데이트합니다 (RESTful: PATCH /notifications/:id)
// 요청 본문을 통해 읽음 상태 업데이트 지원: {"read": true/false}
func (h *Handler) UpdateNotification(c *gin.Context) {
	handler := h.Compose(
		h.updateNotificationHandler(),
		h.StandardCRUDDecorators("update_notification")...,
	)

	handler(c)
}

// updateNotificationHandler: 알림 업데이트의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) updateNotificationHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_notifications")
			return
		}
		notificationID, err := h.ExtractPathParam(c, "id")
		if err != nil {
			h.HandleError(c, err, "get_notification")
			return
		}

		// Parse request body
		var req UpdateNotificationRequest
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "update_notification")
			return
		}

		// If no read status provided, default to read=true for backward compatibility
		readStatus := true
		if req.Read != nil {
			readStatus = *req.Read
		}

		// Get existing notification
		notification, err := h.notificationService.GetNotification(
			c.Request.Context(),
			userID.String(),
			notificationID.String(),
		)
		if err != nil {
			h.HandleError(c, err, "update_notification")
			return
		}

		// Update read status
		notification.IsRead = readStatus
		if readStatus {
			now := time.Now()
			notification.ReadAt = &now
		} else {
			notification.ReadAt = nil
		}

		// Update notification
		err = h.notificationService.UpdateNotification(
			c.Request.Context(),
			notification,
		)
		if err != nil {
			h.HandleError(c, err, "update_notification")
			return
		}

		h.logNotificationUpdateSuccess(c, userID, notificationID, readStatus)
		h.OK(c, gin.H{
			"id":   notificationID.String(),
			"read": readStatus,
		}, "Notification updated successfully")
	}
}

// UpdateNotifications: 알림을 일괄 업데이트합니다 (RESTful: PATCH /notifications)
// 요청 본문을 통해 일괄 읽음 상태 업데이트 지원: {"read": true/false, "notification_ids": [...]}
func (h *Handler) UpdateNotifications(c *gin.Context) {
	handler := h.Compose(
		h.updateNotificationsHandler(),
		h.StandardCRUDDecorators("update_notifications")...,
	)

	handler(c)
}

// updateNotificationsHandler: 여러 알림 업데이트의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) updateNotificationsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_notifications")
			return
		}

		// Parse request body
		var req UpdateNotificationsRequest
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "update_notifications")
			return
		}

		// If no read status provided, default to read=true for backward compatibility
		readStatus := true
		if req.Read != nil {
			readStatus = *req.Read
		}

		updatedCount := int64(0)

		// If specific notification IDs provided, update only those
		if len(req.NotificationIDs) > 0 {
			now := time.Now()
			for _, notificationID := range req.NotificationIDs {
				notification, err := h.notificationService.GetNotification(
					c.Request.Context(),
					userID.String(),
					notificationID,
				)
				if err != nil {
					// Skip invalid IDs, continue with others
					h.LogWarn(c, "Failed to get notification for bulk update",
						zap.String("notification_id", notificationID),
						zap.Error(err))
					continue
				}

				notification.IsRead = readStatus
				if readStatus {
					notification.ReadAt = &now
				} else {
					notification.ReadAt = nil
				}

				err = h.notificationService.UpdateNotification(
					c.Request.Context(),
					notification,
				)
				if err != nil {
					h.LogWarn(c, "Failed to update notification in bulk update",
						zap.String("notification_id", notificationID),
						zap.Error(err))
					continue
				}
				updatedCount++
			}
		} else {
			// Update all notifications for the user
			if readStatus {
				// Use service method for marking all as read (more efficient)
				err := h.notificationService.MarkAllAsRead(
					c.Request.Context(),
					userID.String(),
				)
				if err != nil {
					h.HandleError(c, err, "update_notifications")
					return
				}
				// Get stats to determine updated count
				stats, err := h.notificationService.GetNotificationStats(
					c.Request.Context(),
					userID.String(),
				)
				if err == nil {
					updatedCount = int64(stats.UnreadNotifications)
				}
			} else {
				// Mark all as unread - need to update each notification
				notifications, total, err := h.notificationService.GetNotifications(
					c.Request.Context(),
					userID.String(),
					1000, // Get a large batch
					0,
					false,
					"",
					"",
				)
				if err != nil {
					h.HandleError(c, err, "update_notifications")
					return
				}

				// Update all notifications to unread
				for _, notification := range notifications {
					notification.IsRead = false
					notification.ReadAt = nil
					err = h.notificationService.UpdateNotification(
						c.Request.Context(),
						notification,
					)
					if err == nil {
						updatedCount++
					}
				}
				updatedCount = int64(total) // Use total from query
			}
		}

		h.logNotificationsUpdateSuccess(c, userID, updatedCount, readStatus)
		h.OK(c, gin.H{
			"updated_count": updatedCount,
			"read":          readStatus,
		}, "Notifications updated successfully")
	}
}

// DeleteNotification: 알림을 삭제합니다 (데코레이터 패턴 사용)
func (h *Handler) DeleteNotification(c *gin.Context) {
	handler := h.Compose(
		h.deleteNotificationHandler(),
		h.StandardCRUDDecorators("delete_notification")...,
	)

	handler(c)
}

// deleteNotificationHandler: 알림 삭제의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) deleteNotificationHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_notifications")
			return
		}
		notificationID, err := h.ExtractPathParam(c, "id")
		if err != nil {
			h.HandleError(c, err, "get_notification")
			return
		}

		h.logNotificationDeletionAttempt(c, userID, notificationID)

		err = h.notificationService.DeleteNotification(
			c.Request.Context(),
			userID.String(),
			notificationID.String(),
		)
		if err != nil {
			h.HandleError(c, err, "delete_notification")
			return
		}

		h.logNotificationDeletionSuccess(c, userID, notificationID)
		h.OK(c, gin.H{
			"id": notificationID.String(),
		}, "Notification deleted successfully")
	}
}

// DeleteNotifications: 여러 알림을 삭제합니다 (데코레이터 패턴 사용)
func (h *Handler) DeleteNotifications(c *gin.Context) {
	handler := h.Compose(
		h.deleteNotificationsHandler(),
		h.StandardCRUDDecorators("delete_notifications")...,
	)

	handler(c)
}

// deleteNotificationsHandler: 여러 알림 삭제의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) deleteNotificationsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_notifications")
			return
		}

		h.logBulkDeletionAttempt(c, userID)

		var req struct {
			NotificationIDs []string `json:"notification_ids" binding:"required"`
		}
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "delete_notifications")
			return
		}

		err = h.notificationService.DeleteNotifications(
			c.Request.Context(),
			userID.String(),
			req.NotificationIDs,
		)
		if err != nil {
			h.HandleError(c, err, "delete_notifications")
			return
		}

		h.logBulkDeletionSuccess(c, userID)
		h.OK(c, gin.H{
			"message":          "Notifications deleted successfully",
			"deleted_count":    len(req.NotificationIDs),
			"notification_ids": req.NotificationIDs,
		}, "Notifications deleted successfully")
	}
}

// GetNotificationPreferences: 알림 설정을 조회합니다 (데코레이터 패턴 사용)
func (h *Handler) GetNotificationPreferences(c *gin.Context) {
	handler := h.Compose(
		h.getNotificationPreferencesHandler(),
		h.StandardCRUDDecorators("get_notification_preferences")...,
	)

	handler(c)
}

// getNotificationPreferencesHandler: 알림 설정 조회의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) getNotificationPreferencesHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_notifications")
			return
		}

		h.logPreferencesRequest(c, userID)

		preferences, err := h.notificationService.GetNotificationPreferences(
			c.Request.Context(),
			userID.String(),
		)
		if err != nil {
			h.HandleError(c, err, "get_notification_preferences")
			return
		}

		h.OK(c, preferences, "Notification preferences retrieved successfully")
	}
}

// UpdateNotificationPreferences: 알림 설정을 업데이트합니다 (데코레이터 패턴 사용)
func (h *Handler) UpdateNotificationPreferences(c *gin.Context) {
	var req domain.UpdateNotificationPreferencesRequest

	handler := h.Compose(
		h.updateNotificationPreferencesHandler(req),
		h.StandardCRUDDecorators("update_notification_preferences")...,
	)

	handler(c)
}

// updateNotificationPreferencesHandler: 알림 설정 업데이트의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) updateNotificationPreferencesHandler(req domain.UpdateNotificationPreferencesRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		var req domain.UpdateNotificationPreferencesRequest
		if err := h.ExtractValidatedRequest(c, &req); err != nil {
			h.HandleError(c, err, "update_notification_preferences")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_notifications")
			return
		}

		h.logPreferencesUpdateAttempt(c, userID, req)

		// Get existing preferences or create new ones
		preferences, err := h.notificationService.GetNotificationPreferences(
			c.Request.Context(),
			userID.String(),
		)
		if err != nil {
			// If not found, create default preferences
			preferences = &domain.NotificationPreferences{
				UserID:                userID.String(),
				EmailEnabled:          true,
				PushEnabled:           true,
				BrowserEnabled:        true,
				InAppEnabled:          true,
				SystemNotifications:   true,
				VMNotifications:       true,
				CostNotifications:     true,
				SecurityNotifications: true,
				LowPriorityEnabled:    true,
				MediumPriorityEnabled: true,
				HighPriorityEnabled:   true,
				UrgentPriorityEnabled: true,
				Timezone:              "UTC",
			}
		}

		// Update preferences with request data
		if req.EmailEnabled != nil {
			preferences.EmailEnabled = *req.EmailEnabled
		}
		if req.PushEnabled != nil {
			preferences.PushEnabled = *req.PushEnabled
		}
		if req.BrowserEnabled != nil {
			preferences.BrowserEnabled = *req.BrowserEnabled
		}
		if req.InAppEnabled != nil {
			preferences.InAppEnabled = *req.InAppEnabled
		}
		if req.SystemNotifications != nil {
			preferences.SystemNotifications = *req.SystemNotifications
		}
		if req.VMNotifications != nil {
			preferences.VMNotifications = *req.VMNotifications
		}
		if req.CostNotifications != nil {
			preferences.CostNotifications = *req.CostNotifications
		}
		if req.SecurityNotifications != nil {
			preferences.SecurityNotifications = *req.SecurityNotifications
		}
		if req.LowPriorityEnabled != nil {
			preferences.LowPriorityEnabled = *req.LowPriorityEnabled
		}
		if req.MediumPriorityEnabled != nil {
			preferences.MediumPriorityEnabled = *req.MediumPriorityEnabled
		}
		if req.HighPriorityEnabled != nil {
			preferences.HighPriorityEnabled = *req.HighPriorityEnabled
		}
		if req.UrgentPriorityEnabled != nil {
			preferences.UrgentPriorityEnabled = *req.UrgentPriorityEnabled
		}
		if req.QuietHoursStart != "" {
			preferences.QuietHoursStart = req.QuietHoursStart
		}
		if req.QuietHoursEnd != "" {
			preferences.QuietHoursEnd = req.QuietHoursEnd
		}
		if req.Timezone != "" {
			preferences.Timezone = req.Timezone
		}

		err = h.notificationService.UpdateNotificationPreferences(
			c.Request.Context(),
			userID.String(),
			preferences,
		)
		if err != nil {
			h.HandleError(c, err, "update_notification_preferences")
			return
		}

		h.logPreferencesUpdateSuccess(c, userID, req)
		h.OK(c, preferences, "Notification preferences updated successfully")
	}
}

// GetNotificationStats: 알림 통계를 조회합니다 (데코레이터 패턴 사용)
func (h *Handler) GetNotificationStats(c *gin.Context) {
	handler := h.Compose(
		h.getNotificationStatsHandler(),
		h.StandardCRUDDecorators("get_notification_stats")...,
	)

	handler(c)
}

// getNotificationStatsHandler: 알림 통계 조회의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) getNotificationStatsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_notifications")
			return
		}

		h.logStatsRequest(c, userID)

		stats, err := h.notificationService.GetNotificationStats(
			c.Request.Context(),
			userID.String(),
		)
		if err != nil {
			h.HandleError(c, err, "get_notification_stats")
			return
		}

		h.OK(c, stats, "Notification statistics retrieved successfully")
	}
}

// SendTestNotification: 테스트 알림을 전송합니다 (데코레이터 패턴 사용)
func (h *Handler) SendTestNotification(c *gin.Context) {
	handler := h.Compose(
		h.sendTestNotificationHandler(),
		h.StandardCRUDDecorators("send_test_notification")...,
	)

	handler(c)
}

// sendTestNotificationHandler: 테스트 알림 전송의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) sendTestNotificationHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_notifications")
			return
		}

		h.logTestNotificationAttempt(c, userID)

		// Create test notification
		testNotification := &domain.Notification{
			ID:        uuid.New().String(),
			UserID:    userID.String(),
			Type:      "info",
			Title:     "Test Notification",
			Message:   "This is a test notification",
			Category:  "system",
			Priority:  "low",
			IsRead:    false,
			CreatedAt: time.Now(),
		}

		err = h.notificationService.SendNotification(
			c.Request.Context(),
			userID.String(),
			testNotification,
		)
		if err != nil {
			h.HandleError(c, err, "send_test_notification")
			return
		}

		h.logTestNotificationSuccess(c, userID, gin.H{
			"id":         testNotification.ID,
			"title":      testNotification.Title,
			"message":    testNotification.Message,
			"type":       testNotification.Type,
			"created_at": testNotification.CreatedAt,
		})
		h.OK(c, gin.H{
			"message":      "Test notification sent",
			"notification": testNotification,
		}, "Test notification sent successfully")
	}
}

// 헬퍼 메서드들

// NotificationFilters: 알림 필터링을 위한 구조체
type NotificationFilters struct {
	Limit      int
	Offset     int
	UnreadOnly bool
	Category   string
	Priority   string
}

// parseNotificationFilters: 쿼리 파라미터로부터 알림 필터를 파싱합니다
func (h *Handler) parseNotificationFilters(c *gin.Context) NotificationFilters {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	unreadOnlyStr := c.DefaultQuery("unread_only", "false")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	unreadOnly := unreadOnlyStr == "true"
	category := c.Query("category")
	priority := c.Query("priority")

	return NotificationFilters{
		Limit:      limit,
		Offset:     offset,
		UnreadOnly: unreadOnly,
		Category:   category,
		Priority:   priority,
	}
}

// convertToNotificationResponses: 도메인 알림을 응답 DTO로 변환합니다
func (h *Handler) convertToNotificationResponses(notifications []*domain.Notification) []*NotificationResponse {
	responses := make([]*NotificationResponse, len(notifications))
	for i, notification := range notifications {
		responses[i] = &NotificationResponse{
			ID:        notification.ID,
			UserID:    notification.UserID,
			Title:     notification.Title,
			Message:   notification.Message,
			Type:      notification.Type,
			Status:    notification.Priority,
			Data:      map[string]interface{}{"data": notification.Data},
			CreatedAt: notification.CreatedAt,
			ReadAt:    notification.ReadAt,
		}
	}
	return responses
}

// 로깅 헬퍼 메서드들

// logNotificationsRequest: 알림 목록 조회 요청 로그를 기록합니다
func (h *Handler) logNotificationsRequest(c *gin.Context, userID uuid.UUID, filters NotificationFilters) {
	h.LogBusinessEvent(c, "notifications_requested", userID.String(), "", map[string]interface{}{
		"limit":       filters.Limit,
		"offset":      filters.Offset,
		"unread_only": filters.UnreadOnly,
	})
}

// logNotificationRequest: 알림 조회 요청 로그를 기록합니다
func (h *Handler) logNotificationRequest(c *gin.Context, userID uuid.UUID, notificationID uuid.UUID) {
	h.LogBusinessEvent(c, "notification_requested", userID.String(), notificationID.String(), map[string]interface{}{
		"notification_id": notificationID.String(),
	})
}

// logNotificationUpdateSuccess: 알림 업데이트 성공 로그를 기록합니다
func (h *Handler) logNotificationUpdateSuccess(c *gin.Context, userID uuid.UUID, notificationID uuid.UUID, readStatus bool) {
	h.LogBusinessEvent(c, "notification_updated", userID.String(), notificationID.String(), map[string]interface{}{
		"notification_id": notificationID.String(),
		"read":            readStatus,
	})
}

// logNotificationsUpdateSuccess: 여러 알림 업데이트 성공 로그를 기록합니다
func (h *Handler) logNotificationsUpdateSuccess(c *gin.Context, userID uuid.UUID, updatedCount int64, readStatus bool) {
	h.LogBusinessEvent(c, "notifications_updated", userID.String(), "", map[string]interface{}{
		"updated_count": updatedCount,
		"read":          readStatus,
	})
}

// logNotificationDeletionAttempt: 알림 삭제 시도 로그를 기록합니다
func (h *Handler) logNotificationDeletionAttempt(c *gin.Context, userID uuid.UUID, notificationID uuid.UUID) {
	h.LogBusinessEvent(c, "notification_deletion_attempted", userID.String(), notificationID.String(), map[string]interface{}{
		"notification_id": notificationID.String(),
	})
}

// logNotificationDeletionSuccess: 알림 삭제 성공 로그를 기록합니다
func (h *Handler) logNotificationDeletionSuccess(c *gin.Context, userID uuid.UUID, notificationID uuid.UUID) {
	h.LogBusinessEvent(c, "notification_deleted", userID.String(), notificationID.String(), map[string]interface{}{
		"notification_id": notificationID.String(),
	})
}

// logBulkDeletionAttempt: 일괄 삭제 시도 로그를 기록합니다
func (h *Handler) logBulkDeletionAttempt(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "bulk_deletion_attempted", userID.String(), "", map[string]interface{}{
		"operation": "bulk_delete_notifications",
	})
}

// logBulkDeletionSuccess: 일괄 삭제 성공 로그를 기록합니다
func (h *Handler) logBulkDeletionSuccess(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "notifications_bulk_deleted", userID.String(), "", map[string]interface{}{
		"operation": "bulk_delete_notifications",
	})
}

// logPreferencesRequest: 알림 설정 조회 요청 로그를 기록합니다
func (h *Handler) logPreferencesRequest(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "preferences_requested", userID.String(), "", map[string]interface{}{
		"operation": "get_notification_preferences",
	})
}

// logPreferencesUpdateAttempt: 알림 설정 업데이트 시도 로그를 기록합니다
func (h *Handler) logPreferencesUpdateAttempt(c *gin.Context, userID uuid.UUID, req domain.UpdateNotificationPreferencesRequest) {
	h.LogBusinessEvent(c, "preferences_update_attempted", userID.String(), "", map[string]interface{}{
		"operation":   "update_notification_preferences",
		"preferences": req,
	})
}

// logPreferencesUpdateSuccess: 알림 설정 업데이트 성공 로그를 기록합니다
func (h *Handler) logPreferencesUpdateSuccess(c *gin.Context, userID uuid.UUID, req domain.UpdateNotificationPreferencesRequest) {
	h.LogBusinessEvent(c, "preferences_updated", userID.String(), "", map[string]interface{}{
		"operation":   "update_notification_preferences",
		"preferences": req,
	})
}

// logStatsRequest: 알림 통계 조회 요청 로그를 기록합니다
func (h *Handler) logStatsRequest(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "stats_requested", userID.String(), "", map[string]interface{}{
		"operation": "get_notification_stats",
	})
}

// logTestNotificationAttempt: 테스트 알림 전송 시도 로그를 기록합니다
func (h *Handler) logTestNotificationAttempt(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "test_notification_attempted", userID.String(), "", map[string]interface{}{
		"operation": "send_test_notification",
	})
}

// logTestNotificationSuccess: 테스트 알림 전송 성공 로그를 기록합니다
func (h *Handler) logTestNotificationSuccess(c *gin.Context, userID uuid.UUID, notification gin.H) {
	h.LogBusinessEvent(c, "test_notification_sent", userID.String(), "", map[string]interface{}{
		"operation":    "send_test_notification",
		"notification": notification,
	})
}
