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

// Handler handles notification management operations using improved patterns
type Handler struct {
	*handlers.BaseHandler
	notificationService domain.NotificationService
	readabilityHelper   *readability.ReadabilityHelper
}

// NewHandler creates a new notification handler
func NewHandler(notificationService domain.NotificationService) *Handler {
	return &Handler{
		BaseHandler:         handlers.NewBaseHandler("notification"),
		notificationService: notificationService,
		readabilityHelper:   readability.NewReadabilityHelper(),
	}
}

// GetNotifications retrieves notifications using decorator pattern
func (h *Handler) GetNotifications(c *gin.Context) {
	handler := h.Compose(
		h.getNotificationsHandler(),
		h.StandardCRUDDecorators("get_notifications")...,
	)

	handler(c)
}

// getNotificationsHandler is the core business logic for getting notifications
func (h *Handler) getNotificationsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID := h.extractUserID(c)
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
		pagination := h.calculatePagination(total, filters.Limit, filters.Offset)

		h.OK(c, gin.H{
			"notifications": responses,
			"pagination":    pagination,
		}, "Notifications retrieved successfully")
	}
}

// GetNotification retrieves a specific notification using decorator pattern
func (h *Handler) GetNotification(c *gin.Context) {
	handler := h.Compose(
		h.getNotificationHandler(),
		h.StandardCRUDDecorators("get_notification")...,
	)

	handler(c)
}

// getNotificationHandler is the core business logic for getting a notification
func (h *Handler) getNotificationHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID := h.extractUserID(c)
		notificationID := h.parseNotificationID(c)

		if notificationID == uuid.Nil {
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

// UpdateNotification updates a notification (RESTful: PATCH /notifications/:id)
// Supports updating read status via request body: {"read": true/false}
func (h *Handler) UpdateNotification(c *gin.Context) {
	handler := h.Compose(
		h.updateNotificationHandler(),
		h.StandardCRUDDecorators("update_notification")...,
	)

	handler(c)
}

// updateNotificationHandler is the core business logic for updating a notification
func (h *Handler) updateNotificationHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID := h.extractUserID(c)
		notificationID := h.parseNotificationID(c)

		if notificationID == uuid.Nil {
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

// UpdateNotifications updates notifications in bulk (RESTful: PATCH /notifications)
// Supports bulk read status update via request body: {"read": true/false, "notification_ids": [...]}
func (h *Handler) UpdateNotifications(c *gin.Context) {
	handler := h.Compose(
		h.updateNotificationsHandler(),
		h.StandardCRUDDecorators("update_notifications")...,
	)

	handler(c)
}

// updateNotificationsHandler is the core business logic for updating multiple notifications
func (h *Handler) updateNotificationsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID := h.extractUserID(c)

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

// DeleteNotification deletes a notification using decorator pattern
func (h *Handler) DeleteNotification(c *gin.Context) {
	handler := h.Compose(
		h.deleteNotificationHandler(),
		h.StandardCRUDDecorators("delete_notification")...,
	)

	handler(c)
}

// deleteNotificationHandler is the core business logic for deleting a notification
func (h *Handler) deleteNotificationHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID := h.extractUserID(c)
		notificationID := h.parseNotificationID(c)

		if notificationID == uuid.Nil {
			return
		}

	h.logNotificationDeletionAttempt(c, userID, notificationID)

	err := h.notificationService.DeleteNotification(
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

// DeleteNotifications deletes multiple notifications using decorator pattern
func (h *Handler) DeleteNotifications(c *gin.Context) {
	handler := h.Compose(
		h.deleteNotificationsHandler(),
		h.StandardCRUDDecorators("delete_notifications")...,
	)

	handler(c)
}

// deleteNotificationsHandler is the core business logic for deleting multiple notifications
func (h *Handler) deleteNotificationsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID := h.extractUserID(c)

	h.logBulkDeletionAttempt(c, userID)

	var req struct {
		NotificationIDs []string `json:"notification_ids" binding:"required"`
	}
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "delete_notifications")
		return
	}

	err := h.notificationService.DeleteNotifications(
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
		"message":            "Notifications deleted successfully",
		"deleted_count":      len(req.NotificationIDs),
		"notification_ids":   req.NotificationIDs,
	}, "Notifications deleted successfully")
	}
}

// GetNotificationPreferences retrieves notification preferences using decorator pattern
func (h *Handler) GetNotificationPreferences(c *gin.Context) {
	handler := h.Compose(
		h.getNotificationPreferencesHandler(),
		h.StandardCRUDDecorators("get_notification_preferences")...,
	)

	handler(c)
}

// getNotificationPreferencesHandler is the core business logic for getting notification preferences
func (h *Handler) getNotificationPreferencesHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID := h.extractUserID(c)

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

// UpdateNotificationPreferences updates notification preferences using decorator pattern
func (h *Handler) UpdateNotificationPreferences(c *gin.Context) {
	var req domain.UpdateNotificationPreferencesRequest

	handler := h.Compose(
		h.updateNotificationPreferencesHandler(req),
		h.StandardCRUDDecorators("update_notification_preferences")...,
	)

	handler(c)
}

// updateNotificationPreferencesHandler is the core business logic for updating notification preferences
func (h *Handler) updateNotificationPreferencesHandler(req domain.UpdateNotificationPreferencesRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		req = h.extractValidatedUpdatePreferencesRequest(c)
		userID := h.extractUserID(c)

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

// GetNotificationStats retrieves notification statistics using decorator pattern
func (h *Handler) GetNotificationStats(c *gin.Context) {
	handler := h.Compose(
		h.getNotificationStatsHandler(),
		h.StandardCRUDDecorators("get_notification_stats")...,
	)

	handler(c)
}

// getNotificationStatsHandler is the core business logic for getting notification statistics
func (h *Handler) getNotificationStatsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID := h.extractUserID(c)

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

// SendTestNotification sends a test notification using decorator pattern
func (h *Handler) SendTestNotification(c *gin.Context) {
	handler := h.Compose(
		h.sendTestNotificationHandler(),
		h.StandardCRUDDecorators("send_test_notification")...,
	)

	handler(c)
}

// sendTestNotificationHandler is the core business logic for sending a test notification
func (h *Handler) sendTestNotificationHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID := h.extractUserID(c)

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

	err := h.notificationService.SendNotification(
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

// Helper methods for better readability

type NotificationFilters struct {
	Limit      int
	Offset     int
	UnreadOnly bool
	Category   string
	Priority   string
}

func (h *Handler) extractUserID(c *gin.Context) uuid.UUID {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), "extract_user_id")
		return uuid.Nil
	}
	
	// Convert to uuid.UUID (handle both string and uuid.UUID types)
	switch v := userIDValue.(type) {
	case uuid.UUID:
		return v
	case string:
		parsedUserID, err := uuid.Parse(v)
		if err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "Invalid user ID format", 401), "extract_user_id")
			return uuid.Nil
		}
		return parsedUserID
	default:
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "Invalid user ID type", 401), "extract_user_id")
		return uuid.Nil
	}
}

func (h *Handler) extractValidatedUpdatePreferencesRequest(c *gin.Context) domain.UpdateNotificationPreferencesRequest {
	var req domain.UpdateNotificationPreferencesRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_notification_preferences")
		return domain.UpdateNotificationPreferencesRequest{}
	}
	return req
}

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

func (h *Handler) calculatePagination(total int, limit, offset int) gin.H {
	totalPages := (total + limit - 1) / limit
	currentPage := (offset / limit) + 1

	return gin.H{
		"total":        total,
		"limit":        limit,
		"offset":       offset,
		"current_page": currentPage,
		"total_pages":  totalPages,
	}
}

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

func (h *Handler) parseNotificationID(c *gin.Context) uuid.UUID {
	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid notification ID format", 400), "parse_notification_id")
		return uuid.Nil
	}
	return notificationID
}

// Logging helper methods

func (h *Handler) logNotificationsRequest(c *gin.Context, userID uuid.UUID, filters NotificationFilters) {
	h.LogBusinessEvent(c, "notifications_requested", userID.String(), "", map[string]interface{}{
		"limit":       filters.Limit,
		"offset":      filters.Offset,
		"unread_only": filters.UnreadOnly,
	})
}

func (h *Handler) logNotificationRequest(c *gin.Context, userID uuid.UUID, notificationID uuid.UUID) {
	h.LogBusinessEvent(c, "notification_requested", userID.String(), notificationID.String(), map[string]interface{}{
		"notification_id": notificationID.String(),
	})
}

func (h *Handler) logNotificationUpdateSuccess(c *gin.Context, userID uuid.UUID, notificationID uuid.UUID, readStatus bool) {
	h.LogBusinessEvent(c, "notification_updated", userID.String(), notificationID.String(), map[string]interface{}{
		"notification_id": notificationID.String(),
		"read":           readStatus,
	})
}

func (h *Handler) logNotificationsUpdateSuccess(c *gin.Context, userID uuid.UUID, updatedCount int64, readStatus bool) {
	h.LogBusinessEvent(c, "notifications_updated", userID.String(), "", map[string]interface{}{
		"updated_count": updatedCount,
		"read":           readStatus,
	})
}

func (h *Handler) logNotificationDeletionAttempt(c *gin.Context, userID uuid.UUID, notificationID uuid.UUID) {
	h.LogBusinessEvent(c, "notification_deletion_attempted", userID.String(), notificationID.String(), map[string]interface{}{
		"notification_id": notificationID.String(),
	})
}

func (h *Handler) logNotificationDeletionSuccess(c *gin.Context, userID uuid.UUID, notificationID uuid.UUID) {
	h.LogBusinessEvent(c, "notification_deleted", userID.String(), notificationID.String(), map[string]interface{}{
		"notification_id": notificationID.String(),
	})
}

func (h *Handler) logBulkDeletionAttempt(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "bulk_deletion_attempted", userID.String(), "", map[string]interface{}{
		"operation": "bulk_delete_notifications",
	})
}

func (h *Handler) logBulkDeletionSuccess(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "notifications_bulk_deleted", userID.String(), "", map[string]interface{}{
		"operation": "bulk_delete_notifications",
	})
}

func (h *Handler) logPreferencesRequest(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "preferences_requested", userID.String(), "", map[string]interface{}{
		"operation": "get_notification_preferences",
	})
}

func (h *Handler) logPreferencesUpdateAttempt(c *gin.Context, userID uuid.UUID, req domain.UpdateNotificationPreferencesRequest) {
	h.LogBusinessEvent(c, "preferences_update_attempted", userID.String(), "", map[string]interface{}{
		"operation":   "update_notification_preferences",
		"preferences": req,
	})
}

func (h *Handler) logPreferencesUpdateSuccess(c *gin.Context, userID uuid.UUID, req domain.UpdateNotificationPreferencesRequest) {
	h.LogBusinessEvent(c, "preferences_updated", userID.String(), "", map[string]interface{}{
		"operation":   "update_notification_preferences",
		"preferences": req,
	})
}

func (h *Handler) logStatsRequest(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "stats_requested", userID.String(), "", map[string]interface{}{
		"operation": "get_notification_stats",
	})
}

func (h *Handler) logTestNotificationAttempt(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "test_notification_attempted", userID.String(), "", map[string]interface{}{
		"operation": "send_test_notification",
	})
}

func (h *Handler) logTestNotificationSuccess(c *gin.Context, userID uuid.UUID, notification gin.H) {
	h.LogBusinessEvent(c, "test_notification_sent", userID.String(), "", map[string]interface{}{
		"operation":    "send_test_notification",
		"notification": notification,
	})
}
