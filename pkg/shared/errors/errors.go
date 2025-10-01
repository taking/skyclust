package errors

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
)

// APIError represents a structured API error
type APIError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id,omitempty"`
	StatusCode int                    `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewAPIError creates a new API error
func NewAPIError(code ErrorCode, message string, statusCode int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Timestamp:  time.Now(),
		Details:    make(map[string]interface{}),
	}
}

// WithDetails adds details to the error
func (e *APIError) WithDetails(key string, value interface{}) *APIError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithRequestID adds request ID to the error
func (e *APIError) WithRequestID(requestID string) *APIError {
	e.RequestID = requestID
	return e
}

// ToJSON converts the error to JSON
func (e *APIError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Predefined error constructors
func NewUnauthorizedError(message string) *APIError {
	return NewAPIError(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func NewForbiddenError(message string) *APIError {
	return NewAPIError(ErrCodeForbidden, message, http.StatusForbidden)
}

func NewNotFoundError(message string) *APIError {
	return NewAPIError(ErrCodeNotFound, message, http.StatusNotFound)
}

func NewValidationError(message string) *APIError {
	return NewAPIError(ErrCodeValidationFailed, message, http.StatusBadRequest)
}

func NewInternalError(message string) *APIError {
	return NewAPIError(ErrCodeInternalError, message, http.StatusInternalServerError)
}

func NewServiceUnavailableError(message string) *APIError {
	return NewAPIError(ErrCodeServiceUnavailable, message, http.StatusServiceUnavailable)
}

func NewProviderError(message string) *APIError {
	return NewAPIError(ErrCodeProviderError, message, http.StatusBadGateway)
}

func NewPluginError(message string) *APIError {
	return NewAPIError(ErrCodePluginError, message, http.StatusInternalServerError)
}

// ErrorResponse represents the error response structure
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// WrapError wraps a standard error into an API error
func WrapError(err error, code ErrorCode, message string, statusCode int) *APIError {
	apiErr := NewAPIError(code, message, statusCode)
	_ = apiErr.WithDetails("original_error", err.Error())
	return apiErr
}

// IsAPIError checks if an error is an API error
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

// GetAPIError extracts API error from error
func GetAPIError(err error) *APIError {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}
	return NewInternalError("Internal server error")
}

// Error codes mapping to HTTP status codes
var ErrorCodeToStatus = map[ErrorCode]int{
	ErrCodeUnauthorized:       http.StatusUnauthorized,
	ErrCodeForbidden:          http.StatusForbidden,
	ErrCodeInvalidToken:       http.StatusUnauthorized,
	ErrCodeTokenExpired:       http.StatusUnauthorized,
	ErrCodeInvalidCredentials: http.StatusUnauthorized,
	ErrCodeValidationFailed:   http.StatusBadRequest,
	ErrCodeInvalidInput:       http.StatusBadRequest,
	ErrCodeMissingField:       http.StatusBadRequest,
	ErrCodeNotFound:           http.StatusNotFound,
	ErrCodeAlreadyExists:      http.StatusConflict,
	ErrCodeConflict:           http.StatusConflict,
	ErrCodeResourceExhausted:  http.StatusTooManyRequests,
	ErrCodeInternalError:      http.StatusInternalServerError,
	ErrCodeServiceUnavailable: http.StatusServiceUnavailable,
	ErrCodeTimeout:            http.StatusRequestTimeout,
	ErrCodeDatabaseError:      http.StatusInternalServerError,
	ErrCodeNetworkError:       http.StatusBadGateway,
	ErrCodeProviderError:      http.StatusBadGateway,
	ErrCodeProviderTimeout:    http.StatusGatewayTimeout,
	ErrCodeProviderAuth:       http.StatusUnauthorized,
	ErrCodeProviderQuota:      http.StatusTooManyRequests,
	ErrCodePluginError:        http.StatusInternalServerError,
	ErrCodePluginNotFound:     http.StatusNotFound,
	ErrCodePluginLoadFailed:   http.StatusInternalServerError,
}
