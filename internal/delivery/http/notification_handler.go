/**
 * Notification Handler
 * 알림 관련 HTTP 핸들러
 */

package http

import (
	"skyclust/internal/domain"
	"net/http"
	"strconv"
	"time"

	"skyclust/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	notificationService domain.NotificationService
	logger              *zap.Logger
}

// NewNotificationHandler 알림 핸들러 생성
func NewNotificationHandler(notificationService domain.NotificationService, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		logger:              logger,
	}
}

// GetNotifications 알림 목록 조회
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		ErrorResponse(c, http.StatusUnauthorized, "User ID is required", "MISSING_USER_ID")
		return
	}

	// 쿼리 파라미터 파싱
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	unreadOnly := c.Query("unread_only") == "true"
	category := c.Query("category")
	priority := c.Query("priority")

	// 알림 목록 조회
	notifications, total, err := h.notificationService.GetNotifications(
		c.Request.Context(),
		userID,
		limit,
		offset,
		unreadOnly,
		category,
		priority,
	)
	if err != nil {
		h.logger.Error("Failed to get notifications", zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to get notifications", "ERROR")
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"notifications": notifications,
		"total":         total,
		"limit":         limit,
		"offset":        offset,
	}, "Success")
}

// GetNotification 알림 상세 조회
func (h *NotificationHandler) GetNotification(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		ErrorResponse(c, http.StatusUnauthorized, "User ID is required", "MISSING_USER_ID")
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		ErrorResponse(c, http.StatusBadRequest, "Notification ID is required", "ERROR")
		return
	}

	// 알림 조회
	notification, err := h.notificationService.GetNotification(
		c.Request.Context(),
		userID,
		notificationID,
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusNotFound, "Notification not found", "ERROR")
			return
		}
		logger.Errorf("Failed to get notification: %v", err)
		ErrorResponse(c, http.StatusInternalServerError, "Failed to get notification", "ERROR")
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{"notification": notification}, "Success")
}

// MarkAsRead 알림 읽음 처리
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		ErrorResponse(c, http.StatusUnauthorized, "User ID is required", "MISSING_USER_ID")
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		ErrorResponse(c, http.StatusBadRequest, "Notification ID is required", "ERROR")
		return
	}

	// 알림 읽음 처리
	err := h.notificationService.MarkAsRead(
		c.Request.Context(),
		userID,
		notificationID,
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusNotFound, "Notification not found", "ERROR")
			return
		}
		h.logger.Error("Failed to mark notification as read", zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to mark notification as read", "ERROR")
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{"message": "Notification marked as read"}, "Success")
}

// MarkAllAsRead 모든 알림 읽음 처리
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		ErrorResponse(c, http.StatusUnauthorized, "User ID is required", "MISSING_USER_ID")
		return
	}

	// 모든 알림 읽음 처리
	err := h.notificationService.MarkAllAsRead(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to mark all notifications as read", zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to mark all notifications as read", "ERROR")
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{"message": "All notifications marked as read"}, "Success")
}

// DeleteNotification 알림 삭제
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		ErrorResponse(c, http.StatusUnauthorized, "User ID is required", "MISSING_USER_ID")
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		ErrorResponse(c, http.StatusBadRequest, "Notification ID is required", "ERROR")
		return
	}

	// 알림 삭제
	err := h.notificationService.DeleteNotification(
		c.Request.Context(),
		userID,
		notificationID,
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(c, http.StatusNotFound, "Notification not found", "ERROR")
			return
		}
		h.logger.Error("Failed to delete notification", zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to delete notification", "ERROR")
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{"message": "Notification deleted"}, "Success")
}

// DeleteNotifications 여러 알림 삭제
func (h *NotificationHandler) DeleteNotifications(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		ErrorResponse(c, http.StatusUnauthorized, "User ID is required", "MISSING_USER_ID")
		return
	}

	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid request body", "ERROR")
		return
	}

	// 여러 알림 삭제
	err := h.notificationService.DeleteNotifications(
		c.Request.Context(),
		userID,
		req.IDs,
	)
	if err != nil {
		logger.Errorf("Failed to delete notifications: %v", err)
		ErrorResponse(c, http.StatusInternalServerError, "Failed to delete notifications", "ERROR")
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{"message": "Notifications deleted"}, "Success")
}

// GetNotificationPreferences 알림 설정 조회
func (h *NotificationHandler) GetNotificationPreferences(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		ErrorResponse(c, http.StatusUnauthorized, "User ID is required", "MISSING_USER_ID")
		return
	}

	// 알림 설정 조회
	preferences, err := h.notificationService.GetNotificationPreferences(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		h.logger.Error("Failed to get notification preferences", zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to get notification preferences", "ERROR")
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{"preferences": preferences}, "Success")
}

// UpdateNotificationPreferences 알림 설정 업데이트
func (h *NotificationHandler) UpdateNotificationPreferences(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		ErrorResponse(c, http.StatusUnauthorized, "User ID is required", "MISSING_USER_ID")
		return
	}

	var req domain.NotificationPreferences
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid request body", "ERROR")
		return
	}

	// 알림 설정 업데이트
	err := h.notificationService.UpdateNotificationPreferences(
		c.Request.Context(),
		userID,
		&req,
	)
	if err != nil {
		h.logger.Error("Failed to update notification preferences", zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to update notification preferences", "ERROR")
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{"message": "Notification preferences updated"}, "Success")
}

// GetNotificationStats 알림 통계 조회
func (h *NotificationHandler) GetNotificationStats(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		ErrorResponse(c, http.StatusUnauthorized, "User ID is required", "MISSING_USER_ID")
		return
	}

	// 알림 통계 조회
	stats, err := h.notificationService.GetNotificationStats(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		h.logger.Error("Failed to get notification stats", zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to get notification stats", "ERROR")
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{"stats": stats}, "Success")
}

// SendTestNotification 테스트 알림 전송
func (h *NotificationHandler) SendTestNotification(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		ErrorResponse(c, http.StatusUnauthorized, "User ID is required", "MISSING_USER_ID")
		return
	}

	var req struct {
		Type     string `json:"type" binding:"required"`
		Title    string `json:"title" binding:"required"`
		Message  string `json:"message" binding:"required"`
		Category string `json:"category"`
		Priority string `json:"priority"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid request body", "ERROR")
		return
	}

	// 테스트 알림 생성
	notification := &domain.Notification{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      req.Type,
		Title:     req.Title,
		Message:   req.Message,
		Category:  req.Category,
		Priority:  req.Priority,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	// 알림 저장 및 전송
	err := h.notificationService.CreateNotification(
		c.Request.Context(),
		notification,
	)
	if err != nil {
		h.logger.Error("Failed to send test notification", zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to send test notification", "ERROR")
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{"message": "Test notification sent", "notification": notification}, "Success")
}
