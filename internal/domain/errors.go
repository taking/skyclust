package domain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ErrorCode: 특정 오류 타입을 나타내는 타입
type ErrorCode string

const (
	// 인증 관련 오류
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden          ErrorCode = "FORBIDDEN"
	ErrCodeInvalidToken       ErrorCode = "INVALID_TOKEN"
	ErrCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"

	// 유효성 검증 오류
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrCodeMissingField     ErrorCode = "MISSING_FIELD"
	ErrCodeBadRequest       ErrorCode = "BAD_REQUEST"

	// 리소스 관련 오류
	ErrCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists     ErrorCode = "ALREADY_EXISTS"
	ErrCodeConflict          ErrorCode = "CONFLICT"
	ErrCodeResourceExhausted ErrorCode = "RESOURCE_EXHAUSTED"

	// 시스템 오류
	ErrCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout            ErrorCode = "TIMEOUT"
	ErrCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrCodeNetworkError       ErrorCode = "NETWORK_ERROR"

	// 클라우드 제공자 오류
	ErrCodeProviderError   ErrorCode = "PROVIDER_ERROR"
	ErrCodeProviderTimeout ErrorCode = "PROVIDER_TIMEOUT"
	ErrCodeProviderAuth    ErrorCode = "PROVIDER_AUTH_ERROR"
	ErrCodeProviderQuota   ErrorCode = "PROVIDER_QUOTA_EXCEEDED"

	// 플러그인 오류
	ErrCodePluginError      ErrorCode = "PLUGIN_ERROR"
	ErrCodePluginNotFound   ErrorCode = "PLUGIN_NOT_FOUND"
	ErrCodePluginLoadFailed ErrorCode = "PLUGIN_LOAD_FAILED"

	// 기능 지원 오류
	ErrCodeNotSupported   ErrorCode = "NOT_SUPPORTED"
	ErrCodeNotImplemented ErrorCode = "NOT_IMPLEMENTED"
)

// DomainError: 구조화된 도메인 오류를 나타내는 타입
type DomainError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id,omitempty"`
	StatusCode int                    `json:"-"`
}

// Error: error 인터페이스를 구현합니다
func (e *DomainError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewDomainError: 새로운 도메인 오류를 생성합니다
func NewDomainError(code ErrorCode, message string, statusCode int) *DomainError {
	return &DomainError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Timestamp:  time.Now(),
		Details:    make(map[string]interface{}),
	}
}

// WithDetails: 오류에 상세 정보를 추가합니다
func (e *DomainError) WithDetails(key string, value interface{}) *DomainError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithRequestID: 오류에 요청 ID를 추가합니다
func (e *DomainError) WithRequestID(requestID string) *DomainError {
	e.RequestID = requestID
	return e
}

// ToJSON: 오류를 JSON으로 변환합니다
func (e *DomainError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// 미리 정의된 도메인 오류들
var (
	// 사용자 관련 오류
	ErrUserNotFound       = NewDomainError(ErrCodeNotFound, "user not found", http.StatusNotFound)
	ErrUserAlreadyExists  = NewDomainError(ErrCodeAlreadyExists, "user already exists", http.StatusConflict)
	ErrInvalidCredentials = NewDomainError(ErrCodeInvalidCredentials, "invalid credentials", http.StatusUnauthorized)

	// 워크스페이스 관련 오류
	ErrWorkspaceNotFound = NewDomainError(ErrCodeNotFound, "workspace not found", http.StatusNotFound)
	ErrWorkspaceExists   = NewDomainError(ErrCodeAlreadyExists, "workspace already exists", http.StatusConflict)

	// VM 관련 오류
	ErrVMNotFound      = NewDomainError(ErrCodeNotFound, "VM not found", http.StatusNotFound)
	ErrVMAlreadyExists = NewDomainError(ErrCodeAlreadyExists, "VM already exists", http.StatusConflict)

	// 자격증명 관련 오류
	ErrCredentialNotFound  = NewDomainError(ErrCodeNotFound, "credential not found", http.StatusNotFound)
	ErrNoActiveCredentials = NewDomainError(ErrCodeNotFound, "no active credentials found", http.StatusNotFound)

	// 감사 로그 관련 오류
	ErrAuditLogNotFound = NewDomainError(ErrCodeNotFound, "audit log not found", http.StatusNotFound)

	// 플러그인 관련 오류
	ErrPluginNotFound  = NewDomainError(ErrCodePluginNotFound, "plugin not found", http.StatusNotFound)
	ErrPluginNotActive = NewDomainError(ErrCodePluginError, "plugin not active", http.StatusBadRequest)
	ErrInvalidProvider = NewDomainError(ErrCodeProviderError, "invalid provider", http.StatusBadRequest)

	// 클라우드 제공자 관련 오류
	ErrProviderError   = NewDomainError(ErrCodeProviderError, "provider error", http.StatusBadGateway)
	ErrProviderTimeout = NewDomainError(ErrCodeProviderTimeout, "provider timeout", http.StatusGatewayTimeout)
	ErrProviderAuth    = NewDomainError(ErrCodeProviderAuth, "provider authentication error", http.StatusUnauthorized)
	ErrProviderQuota   = NewDomainError(ErrCodeProviderQuota, "provider quota exceeded", http.StatusTooManyRequests)

	// 일반 오류
	ErrNotFound         = NewDomainError(ErrCodeNotFound, "resource not found", http.StatusNotFound)
	ErrValidationFailed = NewDomainError(ErrCodeValidationFailed, "validation failed", http.StatusBadRequest)
	ErrInternalError    = NewDomainError(ErrCodeInternalError, "internal server error", http.StatusInternalServerError)
)

// IsDomainError: 오류가 도메인 오류인지 확인합니다
func IsDomainError(err error) bool {
	_, ok := err.(*DomainError)
	return ok
}

// GetDomainError: 오류에서 도메인 오류를 추출합니다
func GetDomainError(err error) *DomainError {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr
	}
	return ErrInternalError
}

// IsNotFoundError: 오류가 리소스를 찾을 수 없는 오류인지 확인합니다
func IsNotFoundError(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == ErrCodeNotFound
	}
	return false
}

// IsValidationError: 오류가 유효성 검증 오류인지 확인합니다
func IsValidationError(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == ErrCodeValidationFailed
	}
	return false
}

// IsUnauthorizedError: 오류가 인증 오류인지 확인합니다
func IsUnauthorizedError(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == ErrCodeUnauthorized || domainErr.Code == ErrCodeInvalidCredentials
	}
	return false
}
