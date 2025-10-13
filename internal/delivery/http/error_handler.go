package http

import (
	"skyclust/internal/domain"
	"encoding/json"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ErrorHandler provides centralized error handling for HTTP layer
type ErrorHandler struct {
	logger *zap.Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *zap.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// HandleError handles errors and sends appropriate HTTP responses
func (h *ErrorHandler) HandleError(c *gin.Context, err error) {
	// Extract request ID
	requestID := c.GetString("request_id")
	if requestID == "" {
		requestID = uuid.New().String()
		c.Set("request_id", requestID)
	}

	// Convert error to domain error if needed
	domainErr := h.convertToDomainError(err)
	domainErr = domainErr.WithRequestID(requestID)

	// Log the error with structured logging
	h.logError(c, domainErr)

	// Send standardized error response
	h.sendErrorResponse(c, domainErr)
}

// HandlePanic handles panics and sends appropriate HTTP responses
func (h *ErrorHandler) HandlePanic(c *gin.Context, r interface{}) {
	// Extract request ID
	requestID := c.GetString("request_id")
	if requestID == "" {
		requestID = uuid.New().String()
		c.Set("request_id", requestID)
	}

	// Create domain error for panic
	domainErr := domain.NewDomainError(
		domain.ErrCodeInternalError,
		"Internal server error",
		http.StatusInternalServerError,
	).WithRequestID(requestID).WithDetails("panic", r)

	// Get stack trace
	stack := make([]byte, 4096)
	length := runtime.Stack(stack, false)
	domainErr = domainErr.WithDetails("stack", string(stack[:length]))

	// Log the panic with structured logging
	h.logError(c, domainErr)

	// Send standardized error response
	h.sendErrorResponse(c, domainErr)
}

// convertToDomainError converts a generic error to a domain error
func (h *ErrorHandler) convertToDomainError(err error) *domain.DomainError {
	// Check if it's already a domain error
	if domainErr, ok := err.(*domain.DomainError); ok {
		return domainErr
	}

	// Convert based on error patterns
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "not found"):
		return domain.ErrNotFound
	case strings.Contains(errStr, "unauthorized"):
		return domain.NewDomainError(domain.ErrCodeUnauthorized, "Unauthorized access", http.StatusUnauthorized)
	case strings.Contains(errStr, "forbidden"):
		return domain.NewDomainError(domain.ErrCodeForbidden, "Access forbidden", http.StatusForbidden)
	case strings.Contains(errStr, "validation"):
		return domain.ErrValidationFailed
	case strings.Contains(errStr, "timeout"):
		return domain.NewDomainError(domain.ErrCodeTimeout, "Request timeout", http.StatusRequestTimeout)
	case strings.Contains(errStr, "database"):
		return domain.NewDomainError(domain.ErrCodeDatabaseError, "Database error", http.StatusInternalServerError)
	case strings.Contains(errStr, "network"):
		return domain.NewDomainError(domain.ErrCodeNetworkError, "Network error", http.StatusBadGateway)
	case strings.Contains(errStr, "provider"):
		return domain.ErrProviderError
	case strings.Contains(errStr, "plugin"):
		return domain.NewDomainError(domain.ErrCodePluginError, "Plugin error", http.StatusInternalServerError)
	default:
		return domain.ErrInternalError
	}
}

// logError logs the error with structured information
func (h *ErrorHandler) logError(c *gin.Context, domainErr *domain.DomainError) {
	// Prepare structured log data
	logData := map[string]interface{}{
		"error_code":    domainErr.Code,
		"error_message": domainErr.Message,
		"status_code":   domainErr.StatusCode,
		"request_id":    domainErr.RequestID,
		"method":        c.Request.Method,
		"path":          c.Request.URL.Path,
		"user_agent":    c.Request.UserAgent(),
		"client_ip":     c.ClientIP(),
		"timestamp":     domainErr.Timestamp,
	}

	// Add error details if available
	if len(domainErr.Details) > 0 {
		logData["error_details"] = domainErr.Details
	}

	// Add user information if available
	if userID := c.GetString("user_id"); userID != "" {
		logData["user_id"] = userID
	}

	// Log based on severity
	switch domainErr.StatusCode {
	case http.StatusInternalServerError, http.StatusBadGateway:
		h.logger.Error("Server error occurred", zap.Any("data", logData))
	case http.StatusUnauthorized, http.StatusForbidden:
		h.logger.Warn("Authentication/authorization error", zap.Any("data", logData))
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		h.logger.Warn("Client error occurred", zap.Any("data", logData))
	default:
		h.logger.Info("Request error occurred", zap.Any("data", logData))
	}
}

// sendErrorResponse sends a standardized error response
func (h *ErrorHandler) sendErrorResponse(c *gin.Context, domainErr *domain.DomainError) {
	// Prepare error response
	errorResponse := ErrorResponseData{
		Success:   false,
		Error:     domainErr.Message,
		Code:      string(domainErr.Code),
		RequestID: domainErr.RequestID,
		Timestamp: domainErr.Timestamp,
	}

	// Add details if available (but filter sensitive information)
	if len(domainErr.Details) > 0 {
		filteredDetails := h.filterSensitiveDetails(domainErr.Details)
		if len(filteredDetails) > 0 {
			errorResponse.Details = filteredDetails
		}
	}

	// Set response headers
	c.Header("Content-Type", "application/json")
	c.Header("X-Request-ID", domainErr.RequestID)

	// Send response
	c.JSON(domainErr.StatusCode, errorResponse)
}

// filterSensitiveDetails removes sensitive information from error details
func (h *ErrorHandler) filterSensitiveDetails(details map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{})

	sensitiveKeys := []string{
		"password", "token", "secret", "key", "credential",
		"authorization", "auth", "private", "sensitive",
	}

	for key, value := range details {
		keyLower := strings.ToLower(key)
		isSensitive := false

		for _, sensitiveKey := range sensitiveKeys {
			if strings.Contains(keyLower, sensitiveKey) {
				isSensitive = true
				break
			}
		}

		if !isSensitive {
			filtered[key] = value
		} else {
			filtered[key] = "[REDACTED]"
		}
	}

	return filtered
}

// ErrorResponseData represents error response data
type ErrorResponseData struct {
	Success   bool                   `json:"success"`
	Error     string                 `json:"error"`
	Code      string                 `json:"code"`
	RequestID string                 `json:"request_id,omitempty"`
	Timestamp time.Time              `json:"timestamp,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// ToJSON converts the error response to JSON
func (e *ErrorResponseData) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Middleware functions

// ErrorHandlerMiddleware provides centralized error handling middleware
func (h *ErrorHandler) ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			h.HandleError(c, err.Err)
		}
	}
}

// RecoveryMiddleware provides panic recovery middleware
func (h *ErrorHandler) RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				h.HandlePanic(c, r)
				c.Abort()
			}
		}()

		c.Next()
	}
}

// RequestIDMiddleware adds request ID to context
func (h *ErrorHandler) RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// LoggingMiddleware provides request/response logging middleware
func (h *ErrorHandler) LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Get request ID
		requestID := c.GetString("request_id")

		// Log request
		h.logger.Info("Request started",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log response
		h.logger.Info("Request completed",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.String("duration", duration.String()),
		)
	}
}
