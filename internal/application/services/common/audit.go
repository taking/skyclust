package common

import (
	"context"
	"fmt"
	"skyclust/internal/domain"
	"time"

	"github.com/google/uuid"
	"skyclust/pkg/logger"
)

// LogAction logs an audit action
// userID는 파라미터로 받거나 context에서 추출할 수 있습니다.
// userID가 nil이면 context에서 추출을 시도합니다.
// IPAddress와 UserAgent는 context에서 자동 추출하거나 선택적 파라미터로 받을 수 있습니다.
func LogAction(
	ctx context.Context,
	auditLogRepo domain.AuditLogRepository,
	userID *uuid.UUID,
	action string,
	resource string,
	details map[string]interface{},
) {
	LogActionWithContext(ctx, auditLogRepo, userID, action, resource, details, "", "")
}

// LogActionWithContext logs an audit action with optional IP address and user agent
// userID는 파라미터로 받거나 context에서 추출할 수 있습니다.
// ipAddress와 userAgent가 빈 문자열이면 context에서 추출을 시도합니다.
func LogActionWithContext(
	ctx context.Context,
	auditLogRepo domain.AuditLogRepository,
	userID *uuid.UUID,
	action string,
	resource string,
	details map[string]interface{},
	ipAddress string,
	userAgent string,
) {
	// userID가 nil이면 context에서 추출 시도
	if userID == nil {
		if id := getUserIDFromContext(ctx); id != nil {
			userID = id
		} else {
			// userID를 찾을 수 없으면 로그만 남기고 감사로그는 기록하지 않음
			logger.Warn("Cannot log audit action: userID not found in context or parameters")
			return
		}
	}

	// IPAddress가 빈 문자열이면 context에서 추출 시도
	if ipAddress == "" {
		ipAddress = getIPAddressFromContext(ctx)
	}

	// UserAgent가 빈 문자열이면 context에서 추출 시도
	if userAgent == "" {
		userAgent = getUserAgentFromContext(ctx)
	}

	// Details가 nil이면 빈 맵으로 초기화
	if details == nil {
		details = make(map[string]interface{})
	}

	auditLog := &domain.AuditLog{
		ID:        uuid.New(),
		UserID:    *userID,
		Action:    action,
		Resource:  resource,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   domain.JSONBMap(details),
		CreatedAt: time.Now(),
	}

	if err := auditLogRepo.Create(auditLog); err != nil {
		logger.Error(fmt.Sprintf("Failed to create audit log: %v - action: %s, resource: %s, userID: %s", err, action, resource, userID.String()))
		return
	}

	logger.Debug(fmt.Sprintf("Audit log created successfully: action=%s, resource=%s, userID=%s", action, resource, userID.String()))
}

// getUserIDFromContext attempts to extract userID from context
// Context에 userID가 저장되어 있는 경우를 처리합니다.
func getUserIDFromContext(ctx context.Context) *uuid.UUID {
	// Gin context에서 user_id를 가져오는 경우
	if userIDValue := ctx.Value("user_id"); userIDValue != nil {
		if userIDStr, ok := userIDValue.(string); ok {
			if userID, err := uuid.Parse(userIDStr); err == nil {
				return &userID
			}
		}
		if userID, ok := userIDValue.(uuid.UUID); ok {
			return &userID
		}
	}

	// 다른 키로 저장된 경우
	if userIDValue := ctx.Value("userID"); userIDValue != nil {
		if userIDStr, ok := userIDValue.(string); ok {
			if userID, err := uuid.Parse(userIDStr); err == nil {
				return &userID
			}
		}
		if userID, ok := userIDValue.(uuid.UUID); ok {
			return &userID
		}
	}

	return nil
}

// getIPAddressFromContext attempts to extract IP address from context
// Context에 IP address가 저장되어 있는 경우를 처리합니다.
func getIPAddressFromContext(ctx context.Context) string {
	// Gin context에서 client_ip를 가져오는 경우
	if ipValue := ctx.Value("client_ip"); ipValue != nil {
		if ipStr, ok := ipValue.(string); ok && ipStr != "" {
			return ipStr
		}
	}

	// 다른 키로 저장된 경우
	if ipValue := ctx.Value("ip_address"); ipValue != nil {
		if ipStr, ok := ipValue.(string); ok && ipStr != "" {
			return ipStr
		}
	}

	if ipValue := ctx.Value("ip"); ipValue != nil {
		if ipStr, ok := ipValue.(string); ok && ipStr != "" {
			return ipStr
		}
	}

	// 기본값: localhost
	return "127.0.0.1"
}

// getUserAgentFromContext attempts to extract user agent from context
// Context에 user agent가 저장되어 있는 경우를 처리합니다.
func getUserAgentFromContext(ctx context.Context) string {
	// Gin context에서 user_agent를 가져오는 경우
	if uaValue := ctx.Value("user_agent"); uaValue != nil {
		if uaStr, ok := uaValue.(string); ok && uaStr != "" {
			return uaStr
		}
	}

	// 다른 키로 저장된 경우
	if uaValue := ctx.Value("userAgent"); uaValue != nil {
		if uaStr, ok := uaValue.(string); ok && uaStr != "" {
			return uaStr
		}
	}

	return ""
}
