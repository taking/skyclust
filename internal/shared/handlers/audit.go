package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuditLogger provides audit logging functionality
type AuditLogger struct {
	logger *zap.Logger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *zap.Logger) *AuditLogger {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &AuditLogger{
		logger: logger,
	}
}

// LogAction logs a user action for audit purposes
func (al *AuditLogger) LogAction(c *gin.Context, action string, resource string, resourceID string, userID string, success bool, details map[string]interface{}) {
	auditData := map[string]interface{}{
		"action":      action,
		"resource":    resource,
		"resource_id": resourceID,
		"user_id":     userID,
		"success":     success,
		"ip":          c.ClientIP(),
		"user_agent":  c.Request.UserAgent(),
		"timestamp":   time.Now(),
		"request_id":  c.GetHeader("X-Request-ID"),
		"path":        c.Request.URL.Path,
		"method":      c.Request.Method,
	}

	// Add custom details
	for key, value := range details {
		auditData[key] = value
	}

	al.logger.Info("Audit action", zap.Any("audit", auditData))
}

// LogSecurityAction logs security-related actions
func (al *AuditLogger) LogSecurityAction(c *gin.Context, action string, userID string, success bool, details map[string]interface{}) {
	securityData := map[string]interface{}{
		"action":     action,
		"user_id":    userID,
		"success":    success,
		"ip":         c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
		"timestamp":  time.Now(),
		"request_id": c.GetHeader("X-Request-ID"),
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
	}

	// Add custom details
	for key, value := range details {
		securityData[key] = value
	}

	al.logger.Warn("Security action", zap.Any("security", securityData))
}

// LogDataAccess logs data access events
func (al *AuditLogger) LogDataAccess(c *gin.Context, dataType string, operation string, userID string, resourceID string, details map[string]interface{}) {
	accessData := map[string]interface{}{
		"data_type":   dataType,
		"operation":   operation,
		"user_id":     userID,
		"resource_id": resourceID,
		"ip":          c.ClientIP(),
		"user_agent":  c.Request.UserAgent(),
		"timestamp":   time.Now(),
		"request_id":  c.GetHeader("X-Request-ID"),
		"path":        c.Request.URL.Path,
		"method":      c.Request.Method,
	}

	// Add custom details
	for key, value := range details {
		accessData[key] = value
	}

	al.logger.Info("Data access", zap.Any("access", accessData))
}

// LogSystemEvent logs system-level events
func (al *AuditLogger) LogSystemEvent(c *gin.Context, event string, component string, level string, details map[string]interface{}) {
	systemData := map[string]interface{}{
		"event":      event,
		"component":  component,
		"level":      level,
		"ip":         c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
		"timestamp":  time.Now(),
		"request_id": c.GetHeader("X-Request-ID"),
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
	}

	// Add custom details
	for key, value := range details {
		systemData[key] = value
	}

	al.logger.Info("System event", zap.Any("system", systemData))
}

// AuditContext provides audit context for handlers
type AuditContext struct {
	UserID     string
	Action     string
	Resource   string
	ResourceID string
	Details    map[string]interface{}
}

// NewAuditContext creates a new audit context
func NewAuditContext(userID, action, resource, resourceID string) *AuditContext {
	return &AuditContext{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    make(map[string]interface{}),
	}
}

// WithDetail adds a detail to the audit context
func (ac *AuditContext) WithDetail(key string, value interface{}) *AuditContext {
	ac.Details[key] = value
	return ac
}

// WithDetails adds multiple details to the audit context
func (ac *AuditContext) WithDetails(details map[string]interface{}) *AuditContext {
	for key, value := range details {
		ac.Details[key] = value
	}
	return ac
}

// LogSuccess logs a successful audit event
func (ac *AuditContext) LogSuccess(c *gin.Context, auditLogger *AuditLogger) {
	auditLogger.LogAction(c, ac.Action, ac.Resource, ac.ResourceID, ac.UserID, true, ac.Details)
}

// LogFailure logs a failed audit event
func (ac *AuditContext) LogFailure(c *gin.Context, auditLogger *AuditLogger, errorMsg string) {
	ac.Details["error"] = errorMsg
	auditLogger.LogAction(c, ac.Action, ac.Resource, ac.ResourceID, ac.UserID, false, ac.Details)
}

// Global convenience functions for audit logging
func LogAuditAction(c *gin.Context, action string, resource string, resourceID string, userID string, success bool, details map[string]interface{}) {
	NewAuditLogger(nil).LogAction(c, action, resource, resourceID, userID, success, details)
}

func LogAuditSecurity(c *gin.Context, action string, userID string, success bool, details map[string]interface{}) {
	NewAuditLogger(nil).LogSecurityAction(c, action, userID, success, details)
}

func LogAuditDataAccess(c *gin.Context, dataType string, operation string, userID string, resourceID string, details map[string]interface{}) {
	NewAuditLogger(nil).LogDataAccess(c, dataType, operation, userID, resourceID, details)
}

func LogAuditSystemEvent(c *gin.Context, event string, component string, level string, details map[string]interface{}) {
	NewAuditLogger(nil).LogSystemEvent(c, event, component, level, details)
}

// AuditMiddleware provides middleware for automatic audit logging
func AuditMiddleware(auditLogger *AuditLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log audit event
		userID := ""
		if userIDInterface, exists := c.Get("user_id"); exists {
			userID = userIDInterface.(string)
		}

		auditLogger.LogAction(c, c.Request.Method, c.Request.URL.Path, "", userID, c.Writer.Status() < 400, map[string]interface{}{
			"duration": time.Since(start).String(),
			"status":   c.Writer.Status(),
		})
	}
}
