/**
 * Push Notification Service
 * 푸시 알림 전송 서비스
 */

package notification

import (
	"bytes"
	"skyclust/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"skyclust/pkg/logger"
)

type PushService struct {
	fcmServerKey string
	fcmURL       string
}

// FCMRequest FCM 요청 구조체
type FCMRequest struct {
	To           string                 `json:"to"`
	Notification *FCMNotification       `json:"notification"`
	Data         map[string]interface{} `json:"data"`
	Priority     string                 `json:"priority"`
}

// FCMNotification FCM 알림 구조체
type FCMNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon"`
	Sound string `json:"sound"`
}

// FCMResponse FCM 응답 구조체
type FCMResponse struct {
	Success int `json:"success"`
	Failure int `json:"failure"`
	Results []struct {
		MessageID string `json:"message_id"`
		Error     string `json:"error"`
	} `json:"results"`
}

// NewPushService 푸시 서비스 생성
func NewPushService(fcmServerKey string) *PushService {
	return &PushService{
		fcmServerKey: fcmServerKey,
		fcmURL:       "https://fcm.googleapis.com/fcm/send",
	}
}

// SendNotification 푸시 알림 전송
func (s *PushService) SendNotification(ctx context.Context, deviceToken string, notification *domain.Notification) error {
	// FCM 요청 생성
	req := FCMRequest{
		To: deviceToken,
		Notification: &FCMNotification{
			Title: notification.Title,
			Body:  notification.Message,
			Icon:  "notification_icon",
			Sound: "default",
		},
		Data: map[string]interface{}{
			"notification_id": notification.ID,
			"type":            notification.Type,
			"category":        notification.Category,
			"priority":        notification.Priority,
			"created_at":      notification.CreatedAt.Unix(),
		},
		Priority: s.getPriority(notification.Priority),
	}

	// JSON 인코딩
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal FCM request: %w", err)
	}

	// HTTP 요청 생성
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.fcmURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// 헤더 설정
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "key="+s.fcmServerKey)

	// HTTP 클라이언트 생성
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 요청 전송
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send FCM request: %w", err)
	}
	defer resp.Body.Close()

	// 응답 처리
	var fcmResp FCMResponse
	if err := json.NewDecoder(resp.Body).Decode(&fcmResp); err != nil {
		return fmt.Errorf("failed to decode FCM response: %w", err)
	}

	// 결과 확인
	if fcmResp.Failure > 0 {
		logger.Errorf("FCM notification failed: %+v", fcmResp)
		return fmt.Errorf("FCM notification failed: %d failures", fcmResp.Failure)
	}

	logger.Infof("Push notification sent successfully to device %s", deviceToken)
	return nil
}

// SendBulkNotification 여러 기기에 푸시 알림 전송
func (s *PushService) SendBulkNotification(ctx context.Context, deviceTokens []string, notification *domain.Notification) error {
	if len(deviceTokens) == 0 {
		return nil
	}

	// FCM 요청 생성 (다중 수신자)
	req := FCMRequest{
		To: "", // 다중 수신자용으로 비워둠
		Notification: &FCMNotification{
			Title: notification.Title,
			Body:  notification.Message,
			Icon:  "notification_icon",
			Sound: "default",
		},
		Data: map[string]interface{}{
			"notification_id": notification.ID,
			"type":            notification.Type,
			"category":        notification.Category,
			"priority":        notification.Priority,
			"created_at":      notification.CreatedAt.Unix(),
		},
		Priority: s.getPriority(notification.Priority),
	}

	// JSON 인코딩
	_, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal FCM request: %w", err)
	}

	// 각 기기에 개별 전송 (FCM은 다중 수신자 지원이 제한적)
	var lastErr error
	successCount := 0

	for _, deviceToken := range deviceTokens {
		// 개별 요청 생성
		individualReq := FCMRequest{
			To:           deviceToken,
			Notification: req.Notification,
			Data:         req.Data,
			Priority:     req.Priority,
		}

		individualJsonData, err := json.Marshal(individualReq)
		if err != nil {
			logger.Errorf("Failed to marshal individual FCM request: %v", err)
			continue
		}

		// HTTP 요청 생성
		httpReq, err := http.NewRequestWithContext(ctx, "POST", s.fcmURL, bytes.NewBuffer(individualJsonData))
		if err != nil {
			logger.Errorf("Failed to create individual HTTP request: %v", err)
			continue
		}

		// 헤더 설정
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "key="+s.fcmServerKey)

		// HTTP 클라이언트 생성
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		// 요청 전송
		resp, err := client.Do(httpReq)
		if err != nil {
			logger.Errorf("Failed to send individual FCM request: %v", err)
			lastErr = err
			continue
		}

		// 응답 처리
		var fcmResp FCMResponse
		if err := json.NewDecoder(resp.Body).Decode(&fcmResp); err != nil {
			logger.Errorf("Failed to decode individual FCM response: %v", err)
			resp.Body.Close()
			lastErr = err
			continue
		}
		resp.Body.Close()

		// 결과 확인
		if fcmResp.Success > 0 {
			successCount++
		} else {
			logger.Errorf("Individual FCM notification failed: %+v", fcmResp)
		}
	}

	logger.Infof("Bulk push notification sent: %d/%d successful", successCount, len(deviceTokens))

	if successCount == 0 && lastErr != nil {
		return lastErr
	}

	return nil
}

// getPriority 우선순위를 FCM 형식으로 변환
func (s *PushService) getPriority(priority string) string {
	switch priority {
	case "urgent", "high":
		return "high"
	case "medium", "low":
		return "normal"
	default:
		return "normal"
	}
}
