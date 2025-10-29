package handlers

import (
	"net/http"
	"skyclust/internal/domain"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// HandlerDecorator defines a decorator function for handlers
type HandlerDecorator func(next HandlerFunc) HandlerFunc

// HandlerFunc defines the standard handler function signature
type HandlerFunc func(c *gin.Context)

// HandlerFuncWithResult defines a handler function that returns a result
type HandlerFuncWithResult[T any] func(c *gin.Context) (T, error)

// HandlerFuncWithUserID defines a handler function that requires user ID
type HandlerFuncWithUserID[T any] func(c *gin.Context, userID uuid.UUID) (T, error)

// HandlerFuncWithRequest defines a handler function that requires a request object
type HandlerFuncWithRequest[T any, R any] func(c *gin.Context, req R) (T, error)

// WithRequestID adds request ID to context
func (h *BaseHandler) WithRequestID() HandlerDecorator {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *gin.Context) {
			requestID := c.GetHeader("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}
			c.Set("request_id", requestID)
			c.Header("X-Request-ID", requestID)
			next(c)
		}
	}
}

// WithPerformanceTracking adds performance tracking
func (h *BaseHandler) WithPerformanceTracking(operation string, expectedStatusCode int) HandlerDecorator {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *gin.Context) {
			start := time.Now()
			defer func() {
				duration := time.Since(start)
				h.LogInfo(c, "Performance tracking",
					zap.String("operation", operation),
					zap.Duration("duration", duration),
					zap.Int("status_code", c.Writer.Status()))
			}()
			next(c)
		}
	}
}

// WithAuthentication adds authentication check
func (h *BaseHandler) WithAuthentication() HandlerDecorator {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *gin.Context) {
			userID, err := h.GetUserIDFromToken(c)
			if err != nil {
				h.HandleError(c, err, "authentication")
				return
			}
			c.Set("user_id", userID)
			next(c)
		}
	}
}

// WithAuthorization adds authorization check
func (h *BaseHandler) WithAuthorization(requiredRole domain.Role) HandlerDecorator {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *gin.Context) {
			userRole, err := h.GetUserRoleFromToken(c)
			if err != nil {
				h.HandleError(c, err, "authorization")
				return
			}

			if !h.hasRequiredRole(userRole, requiredRole) {
				h.Forbidden(c, "Insufficient permissions")
				return
			}
			c.Set("user_role", userRole)
			next(c)
		}
	}
}

// WithValidation adds request validation
func (h *BaseHandler) WithValidation(req interface{}) HandlerDecorator {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *gin.Context) {
			if err := h.ValidateRequest(c, req); err != nil {
				h.HandleError(c, err, "validation")
				return
			}
			c.Set("validated_request", req)
			next(c)
		}
	}
}

// WithBusinessLogging adds business event logging
func (h *BaseHandler) WithBusinessLogging(operation string) HandlerDecorator {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *gin.Context) {
			userID := c.GetString("user_id")
			h.LogBusinessEvent(c, operation+"_started", userID, "", map[string]interface{}{
				"operation": operation,
			})
			next(c)
			h.LogBusinessEvent(c, operation+"_completed", userID, "", map[string]interface{}{
				"operation": operation,
			})
		}
	}
}

// WithAuditLogging adds audit event logging
func (h *BaseHandler) WithAuditLogging(operation string) HandlerDecorator {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *gin.Context) {
			userID := c.GetString("user_id")
			h.LogAuditEvent(c, operation, "resource", "", userID, map[string]interface{}{
				"operation": operation,
			})
			next(c)
		}
	}
}

// WithErrorHandling adds comprehensive error handling
func (h *BaseHandler) WithErrorHandling(operation string) HandlerDecorator {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *gin.Context) {
			defer func() {
				if r := recover(); r != nil {
					h.LogError(c, domain.NewDomainError(domain.ErrCodeInternalError, "Panic occurred", 500), "panic_recovery")
					h.InternalServerError(c, "An unexpected error occurred")
				}
			}()
			next(c)
		}
	}
}

// WithLogging wraps a handler with request logging
func (h *BaseHandler) WithLogging() HandlerDecorator {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *gin.Context) {
			h.LogInfo(c, "Request started",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("user_agent", c.GetHeader("User-Agent")),
				zap.String("client_ip", c.ClientIP()))
			next(c)
		}
	}
}

// Compose decorators into a single handler
func (h *BaseHandler) Compose(handler HandlerFunc, decorators ...HandlerDecorator) HandlerFunc {
	result := handler
	for _, decorator := range decorators {
		result = decorator(result)
	}
	return result
}

// hasRequiredRole checks if user has required role
func (h *BaseHandler) hasRequiredRole(userRole, requiredRole domain.Role) bool {
	// Admin has access to everything
	if userRole == domain.AdminRoleType {
		return true
	}

	// Check specific role requirements
	switch requiredRole {
	case domain.AdminRoleType:
		return userRole == domain.AdminRoleType
	case domain.UserRoleType:
		return true // All authenticated users
	default:
		return false
	}
}

// Standard decorator combinations for common patterns
func (h *BaseHandler) StandardCRUDDecorators(operation string) []HandlerDecorator {
	return []HandlerDecorator{
		h.WithRequestID(),
		h.WithLogging(),
		h.WithPerformanceTracking(operation, http.StatusOK),
		h.WithAuthentication(),
		h.WithBusinessLogging(operation),
		h.WithAuditLogging(operation),
		h.WithErrorHandling(operation),
	}
}

func (h *BaseHandler) AdminOnlyDecorators(operation string) []HandlerDecorator {
	return []HandlerDecorator{
		h.WithRequestID(),
		h.WithLogging(),
		h.WithPerformanceTracking(operation, http.StatusOK),
		h.WithAuthentication(),
		h.WithAuthorization(domain.AdminRoleType),
		h.WithBusinessLogging(operation),
		h.WithAuditLogging(operation),
		h.WithErrorHandling(operation),
	}
}

func (h *BaseHandler) PublicDecorators(operation string) []HandlerDecorator {
	return []HandlerDecorator{
		h.WithRequestID(),
		h.WithLogging(),
		h.WithPerformanceTracking(operation, http.StatusOK),
		h.WithErrorHandling(operation),
	}
}
