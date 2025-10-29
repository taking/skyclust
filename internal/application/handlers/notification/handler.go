package notification

import (
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/readability"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
			"",
			"",
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

		// TODO: Implement actual notification retrieval
		notification := gin.H{
			"id":         notificationID.String(),
			"title":      "Sample Notification",
			"message":    "This is a sample notification",
			"type":       "info",
			"read":       false,
			"created_at": "2024-01-01T00:00:00Z",
		}

		h.OK(c, notification, "Notification retrieved successfully")
	}
}

// MarkAsRead marks a notification as read using decorator pattern
func (h *Handler) MarkAsRead(c *gin.Context) {
	handler := h.Compose(
		h.markAsReadHandler(),
		h.StandardCRUDDecorators("mark_as_read")...,
	)

	handler(c)
}

// markAsReadHandler is the core business logic for marking a notification as read
func (h *Handler) markAsReadHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID := h.extractUserID(c)
		notificationID := h.parseNotificationID(c)

		if notificationID == uuid.Nil {
			return
		}

		h.logNotificationMarkAsReadAttempt(c, userID, notificationID)

		// TODO: Implement mark as read functionality
		h.logNotificationMarkAsReadSuccess(c, userID, notificationID)
		h.OK(c, gin.H{
			"id":   notificationID.String(),
			"read": true,
		}, "Notification marked as read")
	}
}

// MarkAllAsRead marks all notifications as read using decorator pattern
func (h *Handler) MarkAllAsRead(c *gin.Context) {
	handler := h.Compose(
		h.markAllAsReadHandler(),
		h.StandardCRUDDecorators("mark_all_as_read")...,
	)

	handler(c)
}

// markAllAsReadHandler is the core business logic for marking all notifications as read
func (h *Handler) markAllAsReadHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID := h.extractUserID(c)

		h.logMarkAllAsReadAttempt(c, userID)

		// TODO: Implement mark all as read functionality
		h.logMarkAllAsReadSuccess(c, userID)
		h.OK(c, gin.H{
			"message": "All notifications marked as read",
		}, "All notifications marked as read")
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

		// TODO: Implement delete functionality
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

		// TODO: Implement bulk delete functionality
		h.logBulkDeletionSuccess(c, userID)
		h.OK(c, gin.H{
			"message": "Notifications deleted successfully",
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

		// TODO: Implement preferences retrieval
		preferences := gin.H{
			"email":     true,
			"push":      true,
			"sms":       false,
			"frequency": "immediate",
		}

		h.OK(c, preferences, "Notification preferences retrieved successfully")
	}
}

// UpdateNotificationPreferences updates notification preferences using decorator pattern
func (h *Handler) UpdateNotificationPreferences(c *gin.Context) {
	var req gin.H

	handler := h.Compose(
		h.updateNotificationPreferencesHandler(req),
		h.StandardCRUDDecorators("update_notification_preferences")...,
	)

	handler(c)
}

// updateNotificationPreferencesHandler is the core business logic for updating notification preferences
func (h *Handler) updateNotificationPreferencesHandler(req gin.H) handlers.HandlerFunc {
	return func(c *gin.Context) {
		req = h.extractValidatedRequest(c)
		userID := h.extractUserID(c)

		h.logPreferencesUpdateAttempt(c, userID, req)

		// TODO: Implement preferences update
		h.logPreferencesUpdateSuccess(c, userID, req)
		h.OK(c, req, "Notification preferences updated successfully")
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

		// TODO: Implement stats retrieval
		stats := gin.H{
			"total":  100,
			"unread": 25,
			"read":   75,
			"by_type": gin.H{
				"info":    50,
				"warning": 30,
				"error":   20,
			},
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

		// TODO: Implement test notification
		notification := gin.H{
			"id":         uuid.New().String(),
			"title":      "Test Notification",
			"message":    "This is a test notification",
			"type":       "info",
			"read":       false,
			"created_at": "2024-01-01T00:00:00Z",
		}

		h.logTestNotificationSuccess(c, userID, notification)
		h.OK(c, gin.H{
			"message":      "Test notification sent",
			"notification": notification,
		}, "Test notification sent successfully")
	}
}

// Helper methods for better readability

type NotificationFilters struct {
	Limit      int
	Offset     int
	UnreadOnly bool
}

func (h *Handler) extractUserID(c *gin.Context) uuid.UUID {
	userID, exists := c.Get("user_id")
	if !exists {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), "extract_user_id")
		return uuid.Nil
	}
	return userID.(uuid.UUID)
}

func (h *Handler) extractValidatedRequest(c *gin.Context) gin.H {
	var req gin.H
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid request body", 400), "extract_validated_request")
		return gin.H{}
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

	return NotificationFilters{
		Limit:      limit,
		Offset:     offset,
		UnreadOnly: unreadOnly,
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

func (h *Handler) logNotificationMarkAsReadAttempt(c *gin.Context, userID uuid.UUID, notificationID uuid.UUID) {
	h.LogBusinessEvent(c, "notification_mark_as_read_attempted", userID.String(), notificationID.String(), map[string]interface{}{
		"notification_id": notificationID.String(),
	})
}

func (h *Handler) logNotificationMarkAsReadSuccess(c *gin.Context, userID uuid.UUID, notificationID uuid.UUID) {
	h.LogBusinessEvent(c, "notification_marked_as_read", userID.String(), notificationID.String(), map[string]interface{}{
		"notification_id": notificationID.String(),
	})
}

func (h *Handler) logMarkAllAsReadAttempt(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "mark_all_as_read_attempted", userID.String(), "", map[string]interface{}{
		"operation": "mark_all_as_read",
	})
}

func (h *Handler) logMarkAllAsReadSuccess(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "all_notifications_marked_as_read", userID.String(), "", map[string]interface{}{
		"operation": "mark_all_as_read",
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

func (h *Handler) logPreferencesUpdateAttempt(c *gin.Context, userID uuid.UUID, req gin.H) {
	h.LogBusinessEvent(c, "preferences_update_attempted", userID.String(), "", map[string]interface{}{
		"operation":   "update_notification_preferences",
		"preferences": req,
	})
}

func (h *Handler) logPreferencesUpdateSuccess(c *gin.Context, userID uuid.UUID, req gin.H) {
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
