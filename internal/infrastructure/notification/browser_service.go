/**
 * Browser Notification Service
 * 브라우저 알림 전송 서비스
 */

package notification

import (
	"skyclust/internal/domain"
	"context"
	"encoding/json"
	"time"

	"skyclust/pkg/logger"
)

type BrowserService struct {
	// 브라우저 알림은 클라이언트 측에서 처리되므로
	// 서버에서는 알림 데이터만 준비
}

// NewBrowserService 브라우저 서비스 생성
func NewBrowserService() *BrowserService {
	return &BrowserService{}
}

// SendNotification 브라우저 알림 데이터 준비
func (s *BrowserService) SendNotification(ctx context.Context, userID string, notification *domain.Notification) error {
	// 브라우저 알림은 클라이언트 측에서 처리되므로
	// 여기서는 로깅만 수행
	logger.Infof("Browser notification prepared for user %s: %s", userID, notification.Title)
	return nil
}

// SendBulkNotification 여러 사용자에게 브라우저 알림 데이터 준비
func (s *BrowserService) SendBulkNotification(ctx context.Context, userIDs []string, notification *domain.Notification) error {
	// 브라우저 알림은 클라이언트 측에서 처리되므로
	// 여기서는 로깅만 수행
	logger.Infof("Browser notification prepared for %d users: %s", len(userIDs), notification.Title)
	return nil
}

// BrowserNotificationData 브라우저 알림 데이터 구조체
type BrowserNotificationData struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Category  string                 `json:"category"`
	Priority  string                 `json:"priority"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
}

// PrepareBrowserNotification 브라우저 알림 데이터 준비
func (s *BrowserService) PrepareBrowserNotification(notification *domain.Notification) (*BrowserNotificationData, error) {
	// 데이터 파싱
	var data map[string]interface{}
	if notification.Data != "" {
		if err := json.Unmarshal([]byte(notification.Data), &data); err != nil {
			logger.Warnf("Failed to parse notification data: %v", err)
			data = make(map[string]interface{})
		}
	} else {
		data = make(map[string]interface{})
	}

	return &BrowserNotificationData{
		ID:        notification.ID,
		UserID:    notification.UserID,
		Type:      notification.Type,
		Title:     notification.Title,
		Message:   notification.Message,
		Category:  notification.Category,
		Priority:  notification.Priority,
		Data:      data,
		CreatedAt: notification.CreatedAt,
	}, nil
}

// GetNotificationOptions 브라우저 알림 옵션 생성
func (s *BrowserService) GetNotificationOptions(notification *domain.Notification) map[string]interface{} {
	options := map[string]interface{}{
		"body":  notification.Message,
		"icon":  "/icons/notification-icon.png",
		"badge": "/icons/badge-icon.png",
		"tag":   notification.ID,
		"data": map[string]interface{}{
			"notification_id": notification.ID,
			"type":            notification.Type,
			"category":        notification.Category,
			"priority":        notification.Priority,
		},
	}

	// 우선순위에 따른 설정
	switch notification.Priority {
	case "urgent":
		options["requireInteraction"] = true
		options["silent"] = false
	case "high":
		options["requireInteraction"] = false
		options["silent"] = false
	case "medium":
		options["requireInteraction"] = false
		options["silent"] = false
	case "low":
		options["requireInteraction"] = false
		options["silent"] = true
	}

	// 타입에 따른 아이콘 설정
	switch notification.Type {
	case "success":
		options["icon"] = "/icons/success-icon.png"
	case "warning":
		options["icon"] = "/icons/warning-icon.png"
	case "error":
		options["icon"] = "/icons/error-icon.png"
	case "info":
		options["icon"] = "/icons/info-icon.png"
	}

	// 카테고리별 추가 설정
	switch notification.Category {
	case "vm":
		options["tag"] = "vm-" + notification.ID
	case "cost":
		options["tag"] = "cost-" + notification.ID
	case "security":
		options["tag"] = "security-" + notification.ID
		options["requireInteraction"] = true
	case "system":
		options["tag"] = "system-" + notification.ID
	}

	return options
}
