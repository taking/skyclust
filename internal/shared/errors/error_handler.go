package errors

import (
	"fmt"
	"runtime"
	"skyclust/internal/domain"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ErrorLevel represents the severity level of an error
type ErrorLevel int

const (
	ErrorLevelInfo ErrorLevel = iota
	ErrorLevelWarning
	ErrorLevelError
	ErrorLevelCritical
)

// ErrorCategory represents the category of an error
type ErrorCategory int

const (
	ErrorCategoryValidation ErrorCategory = iota
	ErrorCategoryAuthentication
	ErrorCategoryAuthorization
	ErrorCategoryBusiness
	ErrorCategoryInfrastructure
	ErrorCategoryExternal
	ErrorCategorySystem
)

// UnifiedError represents a comprehensive error structure
type UnifiedError struct {
	ID          string                 `json:"id"`
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Level       ErrorLevel             `json:"level"`
	Category    ErrorCategory          `json:"category"`
	Timestamp   time.Time              `json:"timestamp"`
	StackTrace  []string               `json:"stack_trace,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	UserID      *uuid.UUID             `json:"user_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Operation   string                 `json:"operation,omitempty"`
	Recoverable bool                   `json:"recoverable"`
	Cause       error                  `json:"-"`
}

// Error implements the error interface
func (e *UnifiedError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying cause
func (e *UnifiedError) Unwrap() error {
	return e.Cause
}

// WithContext adds context information to the error
func (e *UnifiedError) WithContext(key string, value interface{}) *UnifiedError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithUserID sets the user ID for the error
func (e *UnifiedError) WithUserID(userID uuid.UUID) *UnifiedError {
	e.UserID = &userID
	return e
}

// WithRequestID sets the request ID for the error
func (e *UnifiedError) WithRequestID(requestID string) *UnifiedError {
	e.RequestID = requestID
	return e
}

// WithOperation sets the operation name for the error
func (e *UnifiedError) WithOperation(operation string) *UnifiedError {
	e.Operation = operation
	return e
}

// WithStackTrace captures the current stack trace
func (e *UnifiedError) WithStackTrace() *UnifiedError {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	e.StackTrace = strings.Split(string(buf[:n]), "\n")
	return e
}

// IsRecoverable checks if the error is recoverable
func (e *UnifiedError) IsRecoverable() bool {
	return e.Recoverable
}

// ErrorHandler provides unified error handling functionality
type ErrorHandler struct {
	logger interface{} // Will be replaced with actual logger interface
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger interface{}) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// CreateError creates a new unified error
func (eh *ErrorHandler) CreateError(code, message string, level ErrorLevel, category ErrorCategory) *UnifiedError {
	return &UnifiedError{
		ID:          uuid.New().String(),
		Code:        code,
		Message:     message,
		Level:       level,
		Category:    category,
		Timestamp:   time.Now(),
		Recoverable: level < ErrorLevelCritical,
	}
}

// CreateValidationError creates a validation error
func (eh *ErrorHandler) CreateValidationError(message string, details map[string]interface{}) *UnifiedError {
	return eh.CreateError("VALIDATION_ERROR", message, ErrorLevelWarning, ErrorCategoryValidation).
		WithDetails(details)
}

// CreateAuthenticationError creates an authentication error
func (eh *ErrorHandler) CreateAuthenticationError(message string) *UnifiedError {
	return eh.CreateError("AUTHENTICATION_ERROR", message, ErrorLevelError, ErrorCategoryAuthentication)
}

// CreateAuthorizationError creates an authorization error
func (eh *ErrorHandler) CreateAuthorizationError(message string) *UnifiedError {
	return eh.CreateError("AUTHORIZATION_ERROR", message, ErrorLevelError, ErrorCategoryAuthorization)
}

// CreateBusinessError creates a business logic error
func (eh *ErrorHandler) CreateBusinessError(message string, details map[string]interface{}) *UnifiedError {
	return eh.CreateError("BUSINESS_ERROR", message, ErrorLevelError, ErrorCategoryBusiness).
		WithDetails(details)
}

// CreateInfrastructureError creates an infrastructure error
func (eh *ErrorHandler) CreateInfrastructureError(message string, cause error) *UnifiedError {
	return eh.CreateError("INFRASTRUCTURE_ERROR", message, ErrorLevelError, ErrorCategoryInfrastructure).
		WithCause(cause)
}

// CreateExternalError creates an external service error
func (eh *ErrorHandler) CreateExternalError(message string, cause error) *UnifiedError {
	return eh.CreateError("EXTERNAL_ERROR", message, ErrorLevelError, ErrorCategoryExternal).
		WithCause(cause)
}

// CreateSystemError creates a system error
func (eh *ErrorHandler) CreateSystemError(message string, cause error) *UnifiedError {
	return eh.CreateError("SYSTEM_ERROR", message, ErrorLevelCritical, ErrorCategorySystem).
		WithCause(cause).
		WithStackTrace()
}

// WithDetails adds details to the error
func (e *UnifiedError) WithDetails(details map[string]interface{}) *UnifiedError {
	e.Details = details
	return e
}

// WithCause sets the underlying cause
func (e *UnifiedError) WithCause(cause error) *UnifiedError {
	e.Cause = cause
	return e
}

// HandleError processes and handles an error
func (eh *ErrorHandler) HandleError(err error, context map[string]interface{}) *UnifiedError {
	if err == nil {
		return nil
	}

	// If it's already a UnifiedError, enhance it with context
	if unifiedErr, ok := err.(*UnifiedError); ok {
		for k, v := range context {
			_ = unifiedErr.WithContext(k, v)
		}
		return unifiedErr
	}

	// Convert domain errors to unified errors
	if domainErr, ok := err.(*domain.DomainError); ok {
		unifiedErr := eh.CreateError(string(domainErr.Code), domainErr.Message, ErrorLevelError, ErrorCategoryBusiness)
		if domainErr.Details != nil {
			_ = unifiedErr.WithDetails(domainErr.Details)
		}
		for k, v := range context {
			_ = unifiedErr.WithContext(k, v)
		}
		return unifiedErr
	}

	// Create a generic system error for unknown errors
	unifiedErr := eh.CreateSystemError("Unknown error occurred", err)
	for k, v := range context {
		_ = unifiedErr.WithContext(k, v)
	}
	return unifiedErr
}

// DefaultErrorHandler provides default error handling functionality
type DefaultErrorHandler struct {
	*ErrorHandler
}

// NewDefaultErrorHandler creates a new default error handler
func NewDefaultErrorHandler(logger interface{}) *DefaultErrorHandler {
	return &DefaultErrorHandler{
		ErrorHandler: NewErrorHandler(logger),
	}
}

// Predefined error codes
const (
	ErrCodeValidationFailed     = "VALIDATION_FAILED"
	ErrCodeAuthenticationFailed = "AUTHENTICATION_FAILED"
	ErrCodeAuthorizationFailed  = "AUTHORIZATION_FAILED"
	ErrCodeResourceNotFound     = "RESOURCE_NOT_FOUND"
	ErrCodeResourceConflict     = "RESOURCE_CONFLICT"
	ErrCodeInternalError        = "INTERNAL_ERROR"
	ErrCodeExternalServiceError = "EXTERNAL_SERVICE_ERROR"
	ErrCodeDatabaseError        = "DATABASE_ERROR"
	ErrCodeNetworkError         = "NETWORK_ERROR"
	ErrCodeTimeoutError         = "TIMEOUT_ERROR"
)

// ErrorWrapper provides utility functions for error handling
type ErrorWrapper struct {
	handler *ErrorHandler
}

// NewErrorWrapper creates a new error wrapper
func NewErrorWrapper(handler *ErrorHandler) *ErrorWrapper {
	return &ErrorWrapper{
		handler: handler,
	}
}

// Wrap wraps an error with additional context
func (ew *ErrorWrapper) Wrap(err error, message string, context map[string]interface{}) *UnifiedError {
	if err == nil {
		return nil
	}

	unifiedErr := ew.handler.HandleError(err, context)
	unifiedErr.Message = message
	return unifiedErr
}

// WrapWithOperation wraps an error with operation context
func (ew *ErrorWrapper) WrapWithOperation(err error, operation string, context map[string]interface{}) *UnifiedError {
	if err == nil {
		return nil
	}

	unifiedErr := ew.handler.HandleError(err, context)
	_ = unifiedErr.WithOperation(operation)
	return unifiedErr
}

// WrapWithUser wraps an error with user context
func (ew *ErrorWrapper) WrapWithUser(err error, userID uuid.UUID, context map[string]interface{}) *UnifiedError {
	if err == nil {
		return nil
	}

	unifiedErr := ew.handler.HandleError(err, context)
	_ = unifiedErr.WithUserID(userID)
	return unifiedErr
}

// String methods for enums
func (el ErrorLevel) String() string {
	switch el {
	case ErrorLevelInfo:
		return "info"
	case ErrorLevelWarning:
		return "warning"
	case ErrorLevelError:
		return "error"
	case ErrorLevelCritical:
		return "critical"
	default:
		return "unknown"
	}
}

func (ec ErrorCategory) String() string {
	switch ec {
	case ErrorCategoryValidation:
		return "validation"
	case ErrorCategoryAuthentication:
		return "authentication"
	case ErrorCategoryAuthorization:
		return "authorization"
	case ErrorCategoryBusiness:
		return "business"
	case ErrorCategoryInfrastructure:
		return "infrastructure"
	case ErrorCategoryExternal:
		return "external"
	case ErrorCategorySystem:
		return "system"
	default:
		return "unknown"
	}
}
