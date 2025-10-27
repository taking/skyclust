package domain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ErrorCode represents a specific error type
type ErrorCode string

const (
	// Authentication errors
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden          ErrorCode = "FORBIDDEN"
	ErrCodeInvalidToken       ErrorCode = "INVALID_TOKEN"
	ErrCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"

	// Validation errors
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrCodeMissingField     ErrorCode = "MISSING_FIELD"
	ErrCodeBadRequest       ErrorCode = "BAD_REQUEST"

	// Resource errors
	ErrCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists     ErrorCode = "ALREADY_EXISTS"
	ErrCodeConflict          ErrorCode = "CONFLICT"
	ErrCodeResourceExhausted ErrorCode = "RESOURCE_EXHAUSTED"

	// System errors
	ErrCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout            ErrorCode = "TIMEOUT"
	ErrCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrCodeNetworkError       ErrorCode = "NETWORK_ERROR"

	// Cloud provider errors
	ErrCodeProviderError   ErrorCode = "PROVIDER_ERROR"
	ErrCodeProviderTimeout ErrorCode = "PROVIDER_TIMEOUT"
	ErrCodeProviderAuth    ErrorCode = "PROVIDER_AUTH_ERROR"
	ErrCodeProviderQuota   ErrorCode = "PROVIDER_QUOTA_EXCEEDED"

	// Plugin errors
	ErrCodePluginError      ErrorCode = "PLUGIN_ERROR"
	ErrCodePluginNotFound   ErrorCode = "PLUGIN_NOT_FOUND"
	ErrCodePluginLoadFailed ErrorCode = "PLUGIN_LOAD_FAILED"

	// Feature support errors
	ErrCodeNotSupported   ErrorCode = "NOT_SUPPORTED"
	ErrCodeNotImplemented ErrorCode = "NOT_IMPLEMENTED"
)

// DomainError represents a structured domain error
type DomainError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id,omitempty"`
	StatusCode int                    `json:"-"`
}

// Error implements the error interface
func (e *DomainError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewDomainError creates a new domain error
func NewDomainError(code ErrorCode, message string, statusCode int) *DomainError {
	return &DomainError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Timestamp:  time.Now(),
		Details:    make(map[string]interface{}),
	}
}

// WithDetails adds details to the error
func (e *DomainError) WithDetails(key string, value interface{}) *DomainError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithRequestID adds request ID to the error
func (e *DomainError) WithRequestID(requestID string) *DomainError {
	e.RequestID = requestID
	return e
}

// ToJSON converts the error to JSON
func (e *DomainError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Predefined domain errors
var (
	// User errors
	ErrUserNotFound       = NewDomainError(ErrCodeNotFound, "user not found", http.StatusNotFound)
	ErrUserAlreadyExists  = NewDomainError(ErrCodeAlreadyExists, "user already exists", http.StatusConflict)
	ErrInvalidCredentials = NewDomainError(ErrCodeInvalidCredentials, "invalid credentials", http.StatusUnauthorized)

	// Workspace errors
	ErrWorkspaceNotFound = NewDomainError(ErrCodeNotFound, "workspace not found", http.StatusNotFound)
	ErrWorkspaceExists   = NewDomainError(ErrCodeAlreadyExists, "workspace already exists", http.StatusConflict)

	// VM errors
	ErrVMNotFound      = NewDomainError(ErrCodeNotFound, "VM not found", http.StatusNotFound)
	ErrVMAlreadyExists = NewDomainError(ErrCodeAlreadyExists, "VM already exists", http.StatusConflict)

	// Credential errors
	ErrCredentialNotFound  = NewDomainError(ErrCodeNotFound, "credential not found", http.StatusNotFound)
	ErrNoActiveCredentials = NewDomainError(ErrCodeNotFound, "no active credentials found", http.StatusNotFound)

	// Audit log errors
	ErrAuditLogNotFound = NewDomainError(ErrCodeNotFound, "audit log not found", http.StatusNotFound)

	// Plugin errors
	ErrPluginNotFound  = NewDomainError(ErrCodePluginNotFound, "plugin not found", http.StatusNotFound)
	ErrPluginNotActive = NewDomainError(ErrCodePluginError, "plugin not active", http.StatusBadRequest)
	ErrInvalidProvider = NewDomainError(ErrCodeProviderError, "invalid provider", http.StatusBadRequest)

	// Cloud provider errors
	ErrProviderError   = NewDomainError(ErrCodeProviderError, "provider error", http.StatusBadGateway)
	ErrProviderTimeout = NewDomainError(ErrCodeProviderTimeout, "provider timeout", http.StatusGatewayTimeout)
	ErrProviderAuth    = NewDomainError(ErrCodeProviderAuth, "provider authentication error", http.StatusUnauthorized)
	ErrProviderQuota   = NewDomainError(ErrCodeProviderQuota, "provider quota exceeded", http.StatusTooManyRequests)

	// General errors
	ErrNotFound         = NewDomainError(ErrCodeNotFound, "resource not found", http.StatusNotFound)
	ErrValidationFailed = NewDomainError(ErrCodeValidationFailed, "validation failed", http.StatusBadRequest)
	ErrInternalError    = NewDomainError(ErrCodeInternalError, "internal server error", http.StatusInternalServerError)
)

// Helper functions
func IsDomainError(err error) bool {
	_, ok := err.(*DomainError)
	return ok
}

func GetDomainError(err error) *DomainError {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr
	}
	return ErrInternalError
}

func IsNotFoundError(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == ErrCodeNotFound
	}
	return false
}

func IsValidationError(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == ErrCodeValidationFailed
	}
	return false
}

func IsUnauthorizedError(err error) bool {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == ErrCodeUnauthorized || domainErr.Code == ErrCodeInvalidCredentials
	}
	return false
}
