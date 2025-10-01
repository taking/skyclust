package middleware

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"cmp/pkg/shared/errors"
	"cmp/pkg/shared/logger"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware provides centralized error handling
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Check if it's already an API error
			if apiErr, ok := err.Err.(*errors.APIError); ok {
				handleAPIError(c, apiErr)
				return
			}

			// Convert generic error to API error
			apiErr := convertToAPIError(err.Err)
			handleAPIError(c, apiErr)
		}
	}
}

// handleAPIError handles API errors and sends appropriate response
func handleAPIError(c *gin.Context, apiErr *errors.APIError) {
	// Add request ID if available
	if requestID := c.GetString("request_id"); requestID != "" {
		_ = apiErr.WithRequestID(requestID)
	}

	// Log the error
	logger.ErrorWithAPIError("Request failed", apiErr)

	// Set response headers
	c.Header("Content-Type", "application/json")

	// Send error response
	response := errors.ErrorResponse{
		Error: *apiErr,
	}

	c.JSON(apiErr.StatusCode, response)
}

// convertToAPIError converts a generic error to an API error
func convertToAPIError(err error) *errors.APIError {
	// Check for common error patterns
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "not found"):
		return errors.NewNotFoundError("Resource not found")
	case strings.Contains(errStr, "unauthorized"):
		return errors.NewUnauthorizedError("Unauthorized access")
	case strings.Contains(errStr, "forbidden"):
		return errors.NewForbiddenError("Access forbidden")
	case strings.Contains(errStr, "validation"):
		return errors.NewValidationError("Validation failed")
	case strings.Contains(errStr, "timeout"):
		return errors.NewAPIError(errors.ErrCodeTimeout, "Request timeout", http.StatusRequestTimeout)
	case strings.Contains(errStr, "database"):
		return errors.NewAPIError(errors.ErrCodeDatabaseError, "Database error", http.StatusInternalServerError)
	case strings.Contains(errStr, "network"):
		return errors.NewAPIError(errors.ErrCodeNetworkError, "Network error", http.StatusBadGateway)
	case strings.Contains(errStr, "provider"):
		return errors.NewProviderError("Cloud provider error")
	case strings.Contains(errStr, "plugin"):
		return errors.NewPluginError("Plugin error")
	default:
		return errors.NewInternalError("Internal server error")
	}
}

// RecoveryMiddleware provides panic recovery
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic
				stack := make([]byte, 4096)
				length := runtime.Stack(stack, false)
				logger.Errorf("Panic recovered: %v\nStack: %s", r, string(stack[:length]))

				// Create API error for panic
				apiErr := errors.NewInternalError("Internal server error")
				_ = apiErr.WithDetails("panic", r)

				// Add request ID if available
				if requestID := c.GetString("request_id"); requestID != "" {
					_ = apiErr.WithRequestID(requestID)
				}

				// Send error response
				response := errors.ErrorResponse{
					Error: *apiErr,
				}

				c.JSON(http.StatusInternalServerError, response)
				c.Abort()
			}
		}()

		c.Next()
	}
}

// RequestIDMiddleware adds request ID to context
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate a simple request ID
			requestID = generateRequestID()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// LoggingMiddleware provides request/response logging
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Get request ID
		requestID := c.GetString("request_id")

		// Log request
		logger.WithRequestID(requestID).Infof(
			"Request started: %s %s from %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
		)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log response
		logger.WithRequestID(requestID).Infof(
			"Request completed: %s %s - %d in %v",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
		)
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
