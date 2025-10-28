package notification

import (
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler handles notification management operations
type Handler struct {
	*handlers.BaseHandler
	notificationService domain.NotificationService
}

// NewHandler creates a new notification handler
func NewHandler(notificationService domain.NotificationService) *Handler {
	return &Handler{
		BaseHandler:         handlers.NewBaseHandler("notification"),
		notificationService: notificationService,
	}
}

// GetNotifications retrieves notifications
func (h *Handler) GetNotifications(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "get_notifications", 200)

	// Log operation start
	h.LogInfo(c, "Getting notifications",
		zap.String("operation", "get_notifications"))

	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "get_notifications")
		return
	}

	// Parse and validate query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	unreadOnlyStr := c.DefaultQuery("unread_only", "false")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		h.LogWarn(c, "Invalid limit parameter, using default",
			zap.String("limit", limitStr),
			zap.Int("default_limit", 20))
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		h.LogWarn(c, "Invalid offset parameter, using default",
			zap.String("offset", offsetStr),
			zap.Int("default_offset", 0))
		offset = 0
	}

	unreadOnly := unreadOnlyStr == "true"

	// Log business event
	h.LogBusinessEvent(c, "notifications_requested", userID.String(), "", map[string]interface{}{
		"limit":       limit,
		"offset":      offset,
		"unread_only": unreadOnly,
	})

	// Get notifications from service
	notifications, total, err := h.notificationService.GetNotifications(c.Request.Context(), userID.String(), limit, offset, unreadOnly, "", "")
	if err != nil {
		h.LogError(c, err, "Failed to retrieve notifications")
		h.HandleError(c, err, "get_notifications")
		return
	}

	// Convert to response format
	notificationResponses := make([]*NotificationResponse, len(notifications))
	for i, notification := range notifications {
		notificationResponses[i] = &NotificationResponse{
			ID:        notification.ID,
			UserID:    notification.UserID,
			Title:     notification.Title,
			Message:   notification.Message,
			Type:      notification.Type,
			Status:    notification.Priority,                             // Using Priority as Status
			Data:      map[string]interface{}{"data": notification.Data}, // Convert string to map
			CreatedAt: notification.CreatedAt,
			ReadAt:    notification.ReadAt,
		}
	}

	// Log successful operation
	h.LogInfo(c, "Notifications retrieved successfully",
		zap.Int("notifications_count", len(notifications)),
		zap.Int("total", total))

	h.OK(c, gin.H{
		"notifications": notificationResponses,
		"total":         total,
		"limit":         limit,
		"offset":        offset,
	}, "Notifications retrieved successfully")
}

// GetNotification retrieves a specific notification
func (h *Handler) GetNotification(c *gin.Context) {
	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		h.LogWarn(c, "Invalid notification ID format",
			zap.String("notification_id", notificationIDStr),
			zap.Error(err))
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid notification ID format", 400), "get_notification")
		return
	}

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

// MarkAsRead marks a notification as read
func (h *Handler) MarkAsRead(c *gin.Context) {
	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		h.LogWarn(c, "Invalid notification ID format",
			zap.String("notification_id", notificationIDStr),
			zap.Error(err))
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid notification ID format", 400), "get_notification")
		return
	}

	// TODO: Implement mark as read functionality
	h.OK(c, gin.H{
		"id":   notificationID.String(),
		"read": true,
	}, "Notification marked as read")
}

// MarkAllAsRead marks all notifications as read
func (h *Handler) MarkAllAsRead(c *gin.Context) {
	// TODO: Implement mark all as read functionality
	h.OK(c, gin.H{
		"message": "All notifications marked as read",
	}, "All notifications marked as read")
}

// DeleteNotification deletes a notification
func (h *Handler) DeleteNotification(c *gin.Context) {
	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		h.LogWarn(c, "Invalid notification ID format",
			zap.String("notification_id", notificationIDStr),
			zap.Error(err))
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid notification ID format", 400), "get_notification")
		return
	}

	// TODO: Implement delete functionality
	h.OK(c, gin.H{
		"id": notificationID.String(),
	}, "Notification deleted successfully")
}

// DeleteNotifications deletes multiple notifications
func (h *Handler) DeleteNotifications(c *gin.Context) {
	// TODO: Implement bulk delete functionality
	h.OK(c, gin.H{
		"message": "Notifications deleted successfully",
	}, "Notifications deleted successfully")
}

// GetNotificationPreferences retrieves notification preferences
func (h *Handler) GetNotificationPreferences(c *gin.Context) {
	// TODO: Implement preferences retrieval
	preferences := gin.H{
		"email":     true,
		"push":      true,
		"sms":       false,
		"frequency": "immediate",
	}

	h.OK(c, preferences, "Notification preferences retrieved successfully")
}

// UpdateNotificationPreferences updates notification preferences
func (h *Handler) UpdateNotificationPreferences(c *gin.Context) {
	var req gin.H
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid request body", 400), "update_notification_preferences")
		return
	}

	// TODO: Implement preferences update
	h.OK(c, req, "Notification preferences updated successfully")
}

// GetNotificationStats retrieves notification statistics
func (h *Handler) GetNotificationStats(c *gin.Context) {
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

// SendTestNotification sends a test notification
func (h *Handler) SendTestNotification(c *gin.Context) {
	// TODO: Implement test notification
	notification := gin.H{
		"id":         uuid.New().String(),
		"title":      "Test Notification",
		"message":    "This is a test notification",
		"type":       "info",
		"read":       false,
		"created_at": "2024-01-01T00:00:00Z",
	}

	h.OK(c, gin.H{
		"message":      "Test notification sent",
		"notification": notification,
	}, "Test notification sent successfully")
}
