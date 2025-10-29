package consistency

import (
	"skyclust/internal/shared/errors"
	"skyclust/internal/shared/validation"

	"github.com/gin-gonic/gin"
)

// UnifiedHandler provides unified handler functionality
type UnifiedHandler struct {
	consistencyManager *UnifiedConsistencyManager
	config             *HandlerConfig
}

// HandlerConfig holds handler configuration
type HandlerConfig struct {
	EnableRequestID           bool `json:"enable_request_id"`
	EnablePerformanceTracking bool `json:"enable_performance_tracking"`
	EnableAuthentication      bool `json:"enable_authentication"`
	EnableAuthorization       bool `json:"enable_authorization"`
	EnableValidation          bool `json:"enable_validation"`
	EnableBusinessLogging     bool `json:"enable_business_logging"`
	EnableAuditLogging        bool `json:"enable_audit_logging"`
	EnableErrorHandling       bool `json:"enable_error_handling"`
}

// HandlerFunc defines the standard handler function signature
type HandlerFunc func(c *gin.Context)

// HandlerFuncWithRequest defines a handler function that requires a request object
type HandlerFuncWithRequest func(c *gin.Context, req interface{}) (interface{}, error)

// HandlerFuncWithUserID defines a handler function that requires a user ID
type HandlerFuncWithUserID func(c *gin.Context, userID string) (interface{}, error)

// NewUnifiedHandler creates a new unified handler
func NewUnifiedHandler(consistencyManager *UnifiedConsistencyManager, config *HandlerConfig) *UnifiedHandler {
	if config == nil {
		config = DefaultHandlerConfig()
	}

	return &UnifiedHandler{
		consistencyManager: consistencyManager,
		config:             config,
	}
}

// DefaultHandlerConfig returns default handler configuration
func DefaultHandlerConfig() *HandlerConfig {
	return &HandlerConfig{
		EnableRequestID:           true,
		EnablePerformanceTracking: true,
		EnableAuthentication:      true,
		EnableAuthorization:       true,
		EnableValidation:          true,
		EnableBusinessLogging:     true,
		EnableAuditLogging:        true,
		EnableErrorHandling:       true,
	}
}

// Wrap wraps a handler with unified functionality
func (uh *UnifiedHandler) Wrap(handler HandlerFunc) HandlerFunc {
	return func(c *gin.Context) {
		// Add request ID
		if uh.config.EnableRequestID {
			requestID := c.GetHeader("X-Request-ID")
			if requestID == "" {
				requestID = "req-" + generateID()
			}
			c.Set("request_id", requestID)
			c.Header("X-Request-ID", requestID)
		}

		// Track performance
		if uh.config.EnablePerformanceTracking {
			startTime := getCurrentTime()
			defer func() {
				duration := getCurrentTime() - startTime
				uh.consistencyManager.TrackRequest("operation", duration, map[string]interface{}{
					"operation": "handler",
				})
			}()
		}

		// Execute handler
		handler(c)
	}
}

// WrapWithRequest wraps a handler that requires request validation
func (uh *UnifiedHandler) WrapWithRequest(handler HandlerFuncWithRequest, req interface{}) HandlerFunc {
	return func(c *gin.Context) {
		// Validate request
		if uh.config.EnableValidation {
			result := uh.consistencyManager.ValidateEntity("request", req)
			if !result.Valid {
				uh.handleValidationError(c, result)
				return
			}
		}

		// Execute handler
		result, err := handler(c, req)
		if err != nil {
			uh.handleError(c, err)
			return
		}

		// Send success response
		uh.sendSuccessResponse(c, result)
	}
}

// WrapWithUserID wraps a handler that requires user ID
func (uh *UnifiedHandler) WrapWithUserID(handler HandlerFuncWithUserID) HandlerFunc {
	return func(c *gin.Context) {
		// Extract user ID
		userID, exists := c.Get("user_id")
		if !exists {
			uh.handleError(c, uh.consistencyManager.CreateAuthenticationError("User not authenticated"))
			return
		}

		// Execute handler
		result, err := handler(c, userID.(string))
		if err != nil {
			uh.handleError(c, err)
			return
		}

		// Send success response
		uh.sendSuccessResponse(c, result)
	}
}

// WrapWithAudit wraps a handler with audit logging
func (uh *UnifiedHandler) WrapWithAudit(handler HandlerFunc, operation string) HandlerFunc {
	return func(c *gin.Context) {
		// Log audit event
		if uh.config.EnableAuditLogging {
			userID := c.GetString("user_id")
			uh.consistencyManager.LogAuditEvent(operation, "resource", "", userID, map[string]interface{}{
				"operation": operation,
			})
		}

		// Execute handler
		handler(c)
	}
}

// handleError handles errors consistently
func (uh *UnifiedHandler) handleError(c *gin.Context, err error) {
	if uh.config.EnableErrorHandling {
		unifiedErr := uh.consistencyManager.HandleError(err, map[string]interface{}{
			"request_id": c.GetString("request_id"),
			"user_id":    c.GetString("user_id"),
		})

		if unifiedErr != nil {
			c.JSON(getStatusCode(unifiedErr), map[string]interface{}{
				"success": false,
				"error":   unifiedErr,
			})
		}
	}
}

// handleValidationError handles validation errors
func (uh *UnifiedHandler) handleValidationError(c *gin.Context, result *validation.ValidationResult) {
	if uh.config.EnableErrorHandling {
		validationErr := uh.consistencyManager.CreateValidationError("Validation failed", map[string]interface{}{
			"errors": result.Errors,
		})

		c.JSON(400, map[string]interface{}{
			"success": false,
			"error":   validationErr,
		})
	}
}

// sendSuccessResponse sends a success response
func (uh *UnifiedHandler) sendSuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(200, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// Helper functions
func generateID() string {
	// Simple ID generation - in production, use proper UUID
	return "req-123456"
}

func getCurrentTime() int64 {
	// Simple time tracking - in production, use proper time
	return 1234567890
}

func getStatusCode(err *errors.UnifiedError) int {
	switch err.Category {
	case errors.ErrorCategoryValidation:
		return 400
	case errors.ErrorCategoryAuthentication:
		return 401
	case errors.ErrorCategoryAuthorization:
		return 403
	case errors.ErrorCategoryBusiness:
		return 400
	case errors.ErrorCategoryInfrastructure:
		return 500
	case errors.ErrorCategoryExternal:
		return 502
	case errors.ErrorCategorySystem:
		return 500
	default:
		return 500
	}
}
