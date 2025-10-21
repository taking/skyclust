package notification

import (
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/responses"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			responses.DomainError(c, domainErr)
		} else {
			responses.InternalServerError(c, "Failed to get user ID from token")
		}
		return
	}

	// Parse query parameters
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

	// Get notifications from service
	notifications, total, err := h.notificationService.GetNotifications(c.Request.Context(), userID.String(), limit, offset, unreadOnly, "", "")
	if err != nil {
		responses.InternalServerError(c, "Failed to retrieve notifications")
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

	responses.OK(c, gin.H{
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
		responses.BadRequest(c, "Invalid notification ID format")
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

	responses.OK(c, notification, "Notification retrieved successfully")
}

// MarkAsRead marks a notification as read
func (h *Handler) MarkAsRead(c *gin.Context) {
	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		responses.BadRequest(c, "Invalid notification ID format")
		return
	}

	// TODO: Implement mark as read functionality
	responses.OK(c, gin.H{
		"id":   notificationID.String(),
		"read": true,
	}, "Notification marked as read")
}

// MarkAllAsRead marks all notifications as read
func (h *Handler) MarkAllAsRead(c *gin.Context) {
	// TODO: Implement mark all as read functionality
	responses.OK(c, gin.H{
		"message": "All notifications marked as read",
	}, "All notifications marked as read")
}

// DeleteNotification deletes a notification
func (h *Handler) DeleteNotification(c *gin.Context) {
	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		responses.BadRequest(c, "Invalid notification ID format")
		return
	}

	// TODO: Implement delete functionality
	responses.OK(c, gin.H{
		"id": notificationID.String(),
	}, "Notification deleted successfully")
}

// DeleteNotifications deletes multiple notifications
func (h *Handler) DeleteNotifications(c *gin.Context) {
	// TODO: Implement bulk delete functionality
	responses.OK(c, gin.H{
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

	responses.OK(c, preferences, "Notification preferences retrieved successfully")
}

// UpdateNotificationPreferences updates notification preferences
func (h *Handler) UpdateNotificationPreferences(c *gin.Context) {
	var req gin.H
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, "Invalid request body")
		return
	}

	// TODO: Implement preferences update
	responses.OK(c, req, "Notification preferences updated successfully")
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

	responses.OK(c, stats, "Notification statistics retrieved successfully")
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

	responses.OK(c, gin.H{
		"message":      "Test notification sent",
		"notification": notification,
	}, "Test notification sent successfully")
}
