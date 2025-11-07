package logging

import (
	"time"

	"skyclust/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logging field name constants for consistency
// All logging fields should use snake_case for consistency
const (
	// Common fields
	FieldOperation    = "operation"
	FieldUserID       = "user_id"
	FieldResourceID   = "resource_id"
	FieldWorkspaceID  = "workspace_id"
	FieldCredentialID = "credential_id"
	FieldProvider     = "provider"
	FieldAction       = "action"
	FieldResource     = "resource"
	FieldEventType    = "event_type"

	// Request context fields
	FieldPath       = "path"
	FieldMethod     = "method"
	FieldUserAgent  = "user_agent"
	FieldClientIP   = "client_ip"
	FieldIP         = "ip"
	FieldRequestID  = "request_id"
	FieldMessage    = "message"
	FieldDuration   = "duration"
	FieldStatusCode = "status_code"
	FieldSuccess    = "success"
)

// RequestLogger provides structured logging for HTTP requests
type RequestLogger struct {
	logger *zap.Logger
}

// NewRequestLogger creates a new request logger
func NewRequestLogger(loggerInstance *zap.Logger) *RequestLogger {
	if loggerInstance == nil {
		loggerInstance = logger.DefaultLogger.GetLogger()
	}
	return &RequestLogger{
		logger: loggerInstance,
	}
}

// LogRequest logs an incoming HTTP request
func (rl *RequestLogger) LogRequest(c *gin.Context, userID string, action string, duration time.Duration, statusCode int) {
	fields := []zapcore.Field{
		zap.String(FieldMethod, c.Request.Method),
		zap.String(FieldPath, c.Request.URL.Path),
		zap.String(FieldUserID, userID),
		zap.String(FieldAction, action),
		zap.Duration(FieldDuration, duration),
		zap.Int(FieldStatusCode, statusCode),
		zap.String(FieldRequestID, c.GetHeader("X-Request-ID")),
		zap.String(FieldUserAgent, c.Request.UserAgent()),
		zap.String(FieldClientIP, c.ClientIP()),
	}

	if statusCode >= 400 {
		rl.logger.Error("HTTP request failed", fields...)
	} else {
		rl.logger.Info("HTTP request completed", fields...)
	}
}

// LogError logs an error with context
func (rl *RequestLogger) LogError(c *gin.Context, err error, message string, fields ...zap.Field) {
	errorFields := []zapcore.Field{
		zap.Error(err),
		zap.String(FieldMessage, message),
		zap.String(FieldRequestID, c.GetHeader("X-Request-ID")),
		zap.String(FieldPath, c.Request.URL.Path),
		zap.String(FieldMethod, c.Request.Method),
	}

	errorFields = append(errorFields, fields...)

	rl.logger.Error("Request error", errorFields...)
}

// LogSecurityEvent logs security-related events
func (rl *RequestLogger) LogSecurityEvent(c *gin.Context, eventType string, userID string, details map[string]interface{}) {
	fields := []zapcore.Field{
		zap.String(FieldEventType, eventType),
		zap.String(FieldUserID, userID),
		zap.String(FieldIP, c.ClientIP()),
		zap.String(FieldUserAgent, c.Request.UserAgent()),
		zap.String(FieldRequestID, c.GetHeader("X-Request-ID")),
		zap.String(FieldPath, c.Request.URL.Path),
		zap.String(FieldMethod, c.Request.Method),
	}

	// Add custom details
	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}

	rl.logger.Warn("Security event", fields...)
}

// LogBusinessEvent logs business logic events
func (rl *RequestLogger) LogBusinessEvent(c *gin.Context, eventType string, userID string, resourceID string, details map[string]interface{}) {
	fields := []zapcore.Field{
		zap.String(FieldEventType, eventType),
		zap.String(FieldUserID, userID),
		zap.String(FieldResourceID, resourceID),
	}

	// Safely get request information
	if c.Request != nil && c.Request.URL != nil {
		fields = append(fields, zap.String(FieldPath, c.Request.URL.Path))
	}
	if c.Request != nil {
		fields = append(fields, zap.String(FieldMethod, c.Request.Method))
	}

	// Safely get request ID
	if requestID := c.GetString(FieldRequestID); requestID != "" {
		fields = append(fields, zap.String(FieldRequestID, requestID))
	}

	// Add custom details
	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}

	rl.logger.Info("Business event", fields...)
}

// LogPerformance logs performance metrics
func (rl *RequestLogger) LogPerformance(c *gin.Context, operation string, duration time.Duration, metrics map[string]interface{}) {
	fields := []zapcore.Field{
		zap.String(FieldOperation, operation),
		zap.Duration(FieldDuration, duration),
		zap.String(FieldRequestID, c.GetHeader("X-Request-ID")),
		zap.String(FieldPath, c.Request.URL.Path),
		zap.String(FieldMethod, c.Request.Method),
	}

	// Add performance metrics
	for key, value := range metrics {
		fields = append(fields, zap.Any(key, value))
	}

	rl.logger.Info("Performance metric", fields...)
}

// LogAudit logs audit events
func (rl *RequestLogger) LogAudit(c *gin.Context, action string, resource string, userID string, success bool, details map[string]interface{}) {
	fields := []zapcore.Field{
		zap.String(FieldAction, action),
		zap.String(FieldResource, resource),
		zap.String(FieldUserID, userID),
		zap.Bool(FieldSuccess, success),
		zap.String(FieldIP, c.ClientIP()),
		zap.String(FieldUserAgent, c.Request.UserAgent()),
		zap.String(FieldRequestID, c.GetHeader("X-Request-ID")),
		zap.String(FieldPath, c.Request.URL.Path),
		zap.String(FieldMethod, c.Request.Method),
	}

	// Add custom details
	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}

	rl.logger.Info("Audit event", fields...)
}

// Standard logging helpers that ensure consistent field naming

// WithOperation adds an operation field
func WithOperation(operation string) zap.Field {
	return zap.String(FieldOperation, operation)
}

// WithUserID adds a user_id field
func WithUserID(userID string) zap.Field {
	return zap.String(FieldUserID, userID)
}

// WithResourceID adds a resource_id field
func WithResourceID(resourceID string) zap.Field {
	return zap.String(FieldResourceID, resourceID)
}

// WithWorkspaceID adds a workspace_id field
func WithWorkspaceID(workspaceID string) zap.Field {
	return zap.String(FieldWorkspaceID, workspaceID)
}

// WithCredentialID adds a credential_id field
func WithCredentialID(credentialID string) zap.Field {
	return zap.String(FieldCredentialID, credentialID)
}

// WithProvider adds a provider field
func WithProvider(provider string) zap.Field {
	return zap.String(FieldProvider, provider)
}

// WithAction adds an action field
func WithAction(action string) zap.Field {
	return zap.String(FieldAction, action)
}

// WithResource adds a resource field
func WithResource(resource string) zap.Field {
	return zap.String(FieldResource, resource)
}

// WithEventType adds an event_type field
func WithEventType(eventType string) zap.Field {
	return zap.String(FieldEventType, eventType)
}

// Global convenience functions
func LogRequest(c *gin.Context, userID string, action string, duration time.Duration, statusCode int) {
	NewRequestLogger(nil).LogRequest(c, userID, action, duration, statusCode)
}

func LogError(c *gin.Context, err error, message string, fields ...zap.Field) {
	NewRequestLogger(nil).LogError(c, err, message, fields...)
}

func LogSecurityEvent(c *gin.Context, eventType string, userID string, details map[string]interface{}) {
	NewRequestLogger(nil).LogSecurityEvent(c, eventType, userID, details)
}

func LogBusinessEvent(c *gin.Context, eventType string, userID string, resourceID string, details map[string]interface{}) {
	NewRequestLogger(nil).LogBusinessEvent(c, eventType, userID, resourceID, details)
}

func LogAudit(c *gin.Context, action string, resource string, userID string, success bool, details map[string]interface{}) {
	NewRequestLogger(nil).LogAudit(c, action, resource, userID, success, details)
}
