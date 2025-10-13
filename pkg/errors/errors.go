package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Error represents a unified error structure
type Error struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Field       string                 `json:"field,omitempty"`
	Value       interface{}            `json:"value,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	HTTPStatus  int                    `json:"-"`
	Timestamp   time.Time              `json:"timestamp"`
	RequestID   string                 `json:"request_id,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
}

// ErrorCode represents a standardized error code
type ErrorCode string

// Standard error codes
const (
	// Authentication errors
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid       ErrorCode = "TOKEN_INVALID"
	ErrCodeSessionExpired     ErrorCode = "SESSION_EXPIRED"
	ErrCodeAccountLocked      ErrorCode = "ACCOUNT_LOCKED"
	ErrCodeAccountDisabled    ErrorCode = "ACCOUNT_DISABLED"
	ErrCodeEmailNotVerified   ErrorCode = "EMAIL_NOT_VERIFIED"

	// Authorization errors
	ErrCodeForbidden               ErrorCode = "FORBIDDEN"
	ErrCodeInsufficientPermissions ErrorCode = "INSUFFICIENT_PERMISSIONS"
	ErrCodeRoleRequired            ErrorCode = "ROLE_REQUIRED"
	ErrCodePermissionDenied        ErrorCode = "PERMISSION_DENIED"

	// Validation errors
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrCodeMissingField     ErrorCode = "MISSING_FIELD"
	ErrCodeInvalidFormat    ErrorCode = "INVALID_FORMAT"
	ErrCodeValueTooLong     ErrorCode = "VALUE_TOO_LONG"
	ErrCodeValueTooShort    ErrorCode = "VALUE_TOO_SHORT"
	ErrCodeInvalidEmail     ErrorCode = "INVALID_EMAIL"
	ErrCodeInvalidPassword  ErrorCode = "INVALID_PASSWORD"
	ErrCodeInvalidURL       ErrorCode = "INVALID_URL"
	ErrCodeInvalidUUID      ErrorCode = "INVALID_UUID"
	ErrCodeInvalidDate      ErrorCode = "INVALID_DATE"
	ErrCodeInvalidNumber    ErrorCode = "INVALID_NUMBER"
	ErrCodeInvalidBoolean   ErrorCode = "INVALID_BOOLEAN"
	ErrCodeInvalidEnum      ErrorCode = "INVALID_ENUM"

	// Resource errors
	ErrCodeNotFound             ErrorCode = "NOT_FOUND"
	ErrCodeResourceNotFound     ErrorCode = "RESOURCE_NOT_FOUND"
	ErrCodeUserNotFound         ErrorCode = "USER_NOT_FOUND"
	ErrCodeWorkspaceNotFound    ErrorCode = "WORKSPACE_NOT_FOUND"
	ErrCodeCredentialNotFound   ErrorCode = "CREDENTIAL_NOT_FOUND"
	ErrCodeProviderNotFound     ErrorCode = "PROVIDER_NOT_FOUND"
	ErrCodeInstanceNotFound     ErrorCode = "INSTANCE_NOT_FOUND"
	ErrCodeNotificationNotFound ErrorCode = "NOTIFICATION_NOT_FOUND"

	// Conflict errors
	ErrCodeConflict            ErrorCode = "CONFLICT"
	ErrCodeResourceExists      ErrorCode = "RESOURCE_EXISTS"
	ErrCodeDuplicateEmail      ErrorCode = "DUPLICATE_EMAIL"
	ErrCodeDuplicateUsername   ErrorCode = "DUPLICATE_USERNAME"
	ErrCodeDuplicateWorkspace  ErrorCode = "DUPLICATE_WORKSPACE"
	ErrCodeDuplicateCredential ErrorCode = "DUPLICATE_CREDENTIAL"
	ErrCodeResourceInUse       ErrorCode = "RESOURCE_IN_USE"
	ErrCodeDependencyExists    ErrorCode = "DEPENDENCY_EXISTS"

	// Rate limiting errors
	ErrCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrCodeTooManyRequests   ErrorCode = "TOO_MANY_REQUESTS"
	ErrCodeQuotaExceeded     ErrorCode = "QUOTA_EXCEEDED"
	ErrCodeBandwidthExceeded ErrorCode = "BANDWIDTH_EXCEEDED"

	// External service errors
	ErrCodeExternalServiceError ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrCodeProviderError        ErrorCode = "PROVIDER_ERROR"
	ErrCodeDatabaseError        ErrorCode = "DATABASE_ERROR"
	ErrCodeCacheError           ErrorCode = "CACHE_ERROR"
	ErrCodeQueueError           ErrorCode = "QUEUE_ERROR"
	ErrCodeFileSystemError      ErrorCode = "FILESYSTEM_ERROR"
	ErrCodeNetworkError         ErrorCode = "NETWORK_ERROR"
	ErrCodeTimeoutError         ErrorCode = "TIMEOUT_ERROR"

	// Business logic errors
	ErrCodeBusinessRuleViolation ErrorCode = "BUSINESS_RULE_VIOLATION"
	ErrCodeInvalidOperation      ErrorCode = "INVALID_OPERATION"
	ErrCodeOperationNotAllowed   ErrorCode = "OPERATION_NOT_ALLOWED"
	ErrCodeStateTransitionError  ErrorCode = "STATE_TRANSITION_ERROR"
	ErrCodeWorkflowError         ErrorCode = "WORKFLOW_ERROR"

	// System errors
	ErrCodeInternalError         ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable    ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeMaintenanceMode       ErrorCode = "MAINTENANCE_MODE"
	ErrCodeConfigurationError    ErrorCode = "CONFIGURATION_ERROR"
	ErrCodeDependencyUnavailable ErrorCode = "DEPENDENCY_UNAVAILABLE"
	ErrCodeResourceExhausted     ErrorCode = "RESOURCE_EXHAUSTED"
	ErrCodeDiskSpaceFull         ErrorCode = "DISK_SPACE_FULL"
	ErrCodeMemoryExhausted       ErrorCode = "MEMORY_EXHAUSTED"

	// Security errors
	ErrCodeSecurityViolation   ErrorCode = "SECURITY_VIOLATION"
	ErrCodeSuspiciousActivity  ErrorCode = "SUSPICIOUS_ACTIVITY"
	ErrCodeIPBlocked           ErrorCode = "IP_BLOCKED"
	ErrCodeUserBlocked         ErrorCode = "USER_BLOCKED"
	ErrCodeCSRFViolation       ErrorCode = "CSRF_VIOLATION"
	ErrCodeXSSAttempt          ErrorCode = "XSS_ATTEMPT"
	ErrCodeSQLInjectionAttempt ErrorCode = "SQL_INJECTION_ATTEMPT"

	// Data errors
	ErrCodeDataCorruption      ErrorCode = "DATA_CORRUPTION"
	ErrCodeDataInconsistency   ErrorCode = "DATA_INCONSISTENCY"
	ErrCodeDataValidationError ErrorCode = "DATA_VALIDATION_ERROR"
	ErrCodeDataIntegrityError  ErrorCode = "DATA_INTEGRITY_ERROR"
	ErrCodeDataMigrationError  ErrorCode = "DATA_MIGRATION_ERROR"

	// API errors
	ErrCodeInvalidAPIKey           ErrorCode = "INVALID_API_KEY"
	ErrCodeAPIKeyExpired           ErrorCode = "API_KEY_EXPIRED"
	ErrCodeAPIKeyRevoked           ErrorCode = "API_KEY_REVOKED"
	ErrCodeInvalidAPIVersion       ErrorCode = "INVALID_API_VERSION"
	ErrCodeDeprecatedAPI           ErrorCode = "DEPRECATED_API"
	ErrCodeAPINotImplemented       ErrorCode = "API_NOT_IMPLEMENTED"
	ErrCodeAPIMethodNotAllowed     ErrorCode = "API_METHOD_NOT_ALLOWED"
	ErrCodeAPIContentTypeError     ErrorCode = "API_CONTENT_TYPE_ERROR"
	ErrCodeAPIPayloadTooLarge      ErrorCode = "API_PAYLOAD_TOO_LARGE"
	ErrCodeAPIUnsupportedMediaType ErrorCode = "API_UNSUPPORTED_MEDIA_TYPE"

	// Cloud provider errors
	ErrCodeProviderTimeout ErrorCode = "PROVIDER_TIMEOUT"
	ErrCodeProviderAuth    ErrorCode = "PROVIDER_AUTH_ERROR"
	ErrCodeProviderQuota   ErrorCode = "PROVIDER_QUOTA_EXCEEDED"

	// Plugin errors
	ErrCodePluginError      ErrorCode = "PLUGIN_ERROR"
	ErrCodePluginNotFound   ErrorCode = "PLUGIN_NOT_FOUND"
	ErrCodePluginLoadFailed ErrorCode = "PLUGIN_LOAD_FAILED"
)

// ErrorDefinition represents an error definition with metadata
type ErrorDefinition struct {
	Code        ErrorCode `json:"code"`
	Message     string    `json:"message"`
	Description string    `json:"description"`
	HTTPStatus  int       `json:"http_status"`
	Category    string    `json:"category"`
	Severity    string    `json:"severity"`
	Retryable   bool      `json:"retryable"`
	UserFacing  bool      `json:"user_facing"`
}

// NewError creates a new error
func NewError(code ErrorCode, message string, httpStatus int) *Error {
	return &Error{
		Code:       string(code),
		Message:    message,
		HTTPStatus: httpStatus,
		Timestamp:  time.Now(),
		Details:    make(map[string]interface{}),
	}
}

// WithDetails adds details to the error
func (e *Error) WithDetails(key string, value interface{}) *Error {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithField adds a field to the error
func (e *Error) WithField(field string, value interface{}) *Error {
	e.Field = field
	e.Value = value
	return e
}

// WithSuggestions adds suggestions to the error
func (e *Error) WithSuggestions(suggestions []string) *Error {
	e.Suggestions = suggestions
	return e
}

// WithRequestID adds request ID to the error
func (e *Error) WithRequestID(requestID string) *Error {
	e.RequestID = requestID
	return e
}

// WithTraceInfo adds trace information to the error
func (e *Error) WithTraceInfo(traceID, spanID string) *Error {
	e.TraceID = traceID
	e.SpanID = spanID
	return e
}

// ToJSON converts the error to JSON
func (e *Error) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.Message
}

// GetErrorDefinition returns the error definition for a given code
func GetErrorDefinition(code ErrorCode) *ErrorDefinition {
	definitions := GetErrorDefinitions()
	if def, exists := definitions[code]; exists {
		return def
	}
	return &ErrorDefinition{
		Code:        code,
		Message:     "Unknown error",
		Description: "An unknown error occurred",
		HTTPStatus:  http.StatusInternalServerError,
		Category:    "system",
		Severity:    "error",
		Retryable:   false,
		UserFacing:  false,
	}
}

// GetErrorDefinitions returns all error definitions
func GetErrorDefinitions() map[ErrorCode]*ErrorDefinition {
	return map[ErrorCode]*ErrorDefinition{
		// Authentication errors
		ErrCodeUnauthorized: {
			Code:        ErrCodeUnauthorized,
			Message:     "Authentication required",
			Description: "The request requires authentication",
			HTTPStatus:  http.StatusUnauthorized,
			Category:    "authentication",
			Severity:    "warning",
			Retryable:   false,
			UserFacing:  true,
		},
		ErrCodeInvalidCredentials: {
			Code:        ErrCodeInvalidCredentials,
			Message:     "Invalid credentials",
			Description: "The provided credentials are invalid",
			HTTPStatus:  http.StatusUnauthorized,
			Category:    "authentication",
			Severity:    "warning",
			Retryable:   false,
			UserFacing:  true,
		},
		ErrCodeTokenExpired: {
			Code:        ErrCodeTokenExpired,
			Message:     "Token expired",
			Description: "The authentication token has expired",
			HTTPStatus:  http.StatusUnauthorized,
			Category:    "authentication",
			Severity:    "warning",
			Retryable:   false,
			UserFacing:  true,
		},

		// Authorization errors
		ErrCodeForbidden: {
			Code:        ErrCodeForbidden,
			Message:     "Access forbidden",
			Description: "The request is forbidden",
			HTTPStatus:  http.StatusForbidden,
			Category:    "authorization",
			Severity:    "warning",
			Retryable:   false,
			UserFacing:  true,
		},
		ErrCodeInsufficientPermissions: {
			Code:        ErrCodeInsufficientPermissions,
			Message:     "Insufficient permissions",
			Description: "The user does not have sufficient permissions",
			HTTPStatus:  http.StatusForbidden,
			Category:    "authorization",
			Severity:    "warning",
			Retryable:   false,
			UserFacing:  true,
		},

		// Validation errors
		ErrCodeValidationFailed: {
			Code:        ErrCodeValidationFailed,
			Message:     "Validation failed",
			Description: "The request data failed validation",
			HTTPStatus:  http.StatusUnprocessableEntity,
			Category:    "validation",
			Severity:    "warning",
			Retryable:   false,
			UserFacing:  true,
		},
		ErrCodeInvalidInput: {
			Code:        ErrCodeInvalidInput,
			Message:     "Invalid input",
			Description: "The provided input is invalid",
			HTTPStatus:  http.StatusBadRequest,
			Category:    "validation",
			Severity:    "warning",
			Retryable:   false,
			UserFacing:  true,
		},
		ErrCodeMissingField: {
			Code:        ErrCodeMissingField,
			Message:     "Missing required field",
			Description: "A required field is missing",
			HTTPStatus:  http.StatusBadRequest,
			Category:    "validation",
			Severity:    "warning",
			Retryable:   false,
			UserFacing:  true,
		},

		// Resource errors
		ErrCodeNotFound: {
			Code:        ErrCodeNotFound,
			Message:     "Resource not found",
			Description: "The requested resource was not found",
			HTTPStatus:  http.StatusNotFound,
			Category:    "resource",
			Severity:    "warning",
			Retryable:   false,
			UserFacing:  true,
		},
		ErrCodeUserNotFound: {
			Code:        ErrCodeUserNotFound,
			Message:     "User not found",
			Description: "The requested user was not found",
			HTTPStatus:  http.StatusNotFound,
			Category:    "resource",
			Severity:    "warning",
			Retryable:   false,
			UserFacing:  true,
		},
		ErrCodeWorkspaceNotFound: {
			Code:        ErrCodeWorkspaceNotFound,
			Message:     "Workspace not found",
			Description: "The requested workspace was not found",
			HTTPStatus:  http.StatusNotFound,
			Category:    "resource",
			Severity:    "warning",
			Retryable:   false,
			UserFacing:  true,
		},

		// Conflict errors
		ErrCodeConflict: {
			Code:        ErrCodeConflict,
			Message:     "Resource conflict",
			Description: "The request conflicts with the current state",
			HTTPStatus:  http.StatusConflict,
			Category:    "conflict",
			Severity:    "warning",
			Retryable:   false,
			UserFacing:  true,
		},
		ErrCodeResourceExists: {
			Code:        ErrCodeResourceExists,
			Message:     "Resource already exists",
			Description: "The resource already exists",
			HTTPStatus:  http.StatusConflict,
			Category:    "conflict",
			Severity:    "warning",
			Retryable:   false,
			UserFacing:  true,
		},
		ErrCodeDuplicateEmail: {
			Code:        ErrCodeDuplicateEmail,
			Message:     "Email already exists",
			Description: "An account with this email already exists",
			HTTPStatus:  http.StatusConflict,
			Category:    "conflict",
			Severity:    "warning",
			Retryable:   false,
			UserFacing:  true,
		},

		// Rate limiting errors
		ErrCodeRateLimitExceeded: {
			Code:        ErrCodeRateLimitExceeded,
			Message:     "Rate limit exceeded",
			Description: "The rate limit has been exceeded",
			HTTPStatus:  http.StatusTooManyRequests,
			Category:    "rate_limit",
			Severity:    "warning",
			Retryable:   true,
			UserFacing:  true,
		},

		// System errors
		ErrCodeInternalError: {
			Code:        ErrCodeInternalError,
			Message:     "Internal server error",
			Description: "An internal server error occurred",
			HTTPStatus:  http.StatusInternalServerError,
			Category:    "system",
			Severity:    "error",
			Retryable:   true,
			UserFacing:  false,
		},
		ErrCodeServiceUnavailable: {
			Code:        ErrCodeServiceUnavailable,
			Message:     "Service unavailable",
			Description: "The service is temporarily unavailable",
			HTTPStatus:  http.StatusServiceUnavailable,
			Category:    "system",
			Severity:    "error",
			Retryable:   true,
			UserFacing:  true,
		},
	}
}

// ErrorCodeFromHTTPStatus returns the most appropriate error code for an HTTP status
func ErrorCodeFromHTTPStatus(status int) ErrorCode {
	switch status {
	case http.StatusBadRequest:
		return ErrCodeInvalidInput
	case http.StatusUnauthorized:
		return ErrCodeUnauthorized
	case http.StatusForbidden:
		return ErrCodeForbidden
	case http.StatusNotFound:
		return ErrCodeNotFound
	case http.StatusConflict:
		return ErrCodeConflict
	case http.StatusUnprocessableEntity:
		return ErrCodeValidationFailed
	case http.StatusTooManyRequests:
		return ErrCodeRateLimitExceeded
	case http.StatusInternalServerError:
		return ErrCodeInternalError
	case http.StatusServiceUnavailable:
		return ErrCodeServiceUnavailable
	default:
		return ErrCodeInternalError
	}
}

// HTTPStatusFromErrorCode returns the HTTP status for an error code
func HTTPStatusFromErrorCode(code ErrorCode) int {
	def := GetErrorDefinition(code)
	return def.HTTPStatus
}

// IsRetryableError checks if an error code is retryable
func IsRetryableError(code ErrorCode) bool {
	def := GetErrorDefinition(code)
	return def.Retryable
}

// IsUserFacingError checks if an error code is user-facing
func IsUserFacingError(code ErrorCode) bool {
	def := GetErrorDefinition(code)
	return def.UserFacing
}

// GetErrorCategory returns the category for an error code
func GetErrorCategory(code ErrorCode) string {
	def := GetErrorDefinition(code)
	return def.Category
}

// GetErrorSeverity returns the severity for an error code
func GetErrorSeverity(code ErrorCode) string {
	def := GetErrorDefinition(code)
	return def.Severity
}

// FormatErrorCode formats an error code for display
func FormatErrorCode(code ErrorCode) string {
	return strings.ToLower(string(code))
}

// ParseErrorCode parses an error code from a string
func ParseErrorCode(s string) (ErrorCode, error) {
	code := ErrorCode(strings.ToUpper(s))
	if _, exists := GetErrorDefinitions()[code]; exists {
		return code, nil
	}
	return "", fmt.Errorf("invalid error code: %s", s)
}

// Legacy compatibility for existing domain errors

// DomainError represents a legacy domain error (for backward compatibility)
type DomainError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	StatusCode int                    `json:"status_code"`
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id,omitempty"`
}

// NewDomainError creates a new domain error (legacy compatibility)
func NewDomainError(code ErrorCode, message string, statusCode int) *DomainError {
	return &DomainError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Timestamp:  time.Now(),
		Details:    make(map[string]interface{}),
	}
}

// WithRequestID adds request ID to the domain error (legacy compatibility)
func (de *DomainError) WithRequestID(requestID string) *DomainError {
	de.RequestID = requestID
	return de
}

// WithDetails adds details to the domain error (legacy compatibility)
func (de *DomainError) WithDetails(key string, value interface{}) *DomainError {
	if de.Details == nil {
		de.Details = make(map[string]interface{})
	}
	de.Details[key] = value
	return de
}

// ToError converts a domain error to the new error format
func (de *DomainError) ToError() *Error {
	return &Error{
		Code:       string(de.Code),
		Message:    de.Message,
		Details:    de.Details,
		HTTPStatus: de.StatusCode,
		Timestamp:  de.Timestamp,
		RequestID:  de.RequestID,
	}
}

// Error implements the error interface (legacy compatibility)
func (de *DomainError) Error() string {
	return de.Message
}
