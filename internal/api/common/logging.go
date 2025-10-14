package common

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// RequestLogger provides structured logging for HTTP requests
type RequestLogger struct {
	logger *zap.Logger
}

// NewRequestLogger creates a new request logger
func NewRequestLogger(logger *zap.Logger) *RequestLogger {
	return &RequestLogger{
		logger: logger,
	}
}

// LogRequest logs an incoming HTTP request
func (rl *RequestLogger) LogRequest(c *gin.Context, userID string, action string, duration time.Duration, statusCode int) {
	fields := []zapcore.Field{
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("user_id", userID),
		zap.String("action", action),
		zap.Duration("duration", duration),
		zap.Int("status_code", statusCode),
		zap.String("request_id", getRequestID(c)),
		zap.String("user_agent", c.Request.UserAgent()),
		zap.String("ip", c.ClientIP()),
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
		zap.String("message", message),
		zap.String("request_id", getRequestID(c)),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	}

	errorFields = append(errorFields, fields...)

	rl.logger.Error("Request error", errorFields...)
}

// LogSecurityEvent logs security-related events
func (rl *RequestLogger) LogSecurityEvent(c *gin.Context, eventType string, userID string, details map[string]interface{}) {
	fields := []zapcore.Field{
		zap.String("event_type", eventType),
		zap.String("user_id", userID),
		zap.String("ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
		zap.String("request_id", getRequestID(c)),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
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
		zap.String("event_type", eventType),
		zap.String("user_id", userID),
		zap.String("resource_id", resourceID),
		zap.String("request_id", getRequestID(c)),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
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
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.String("request_id", getRequestID(c)),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
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
		zap.String("action", action),
		zap.String("resource", resource),
		zap.String("user_id", userID),
		zap.Bool("success", success),
		zap.String("ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
		zap.String("request_id", getRequestID(c)),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	}

	// Add custom details
	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}

	rl.logger.Info("Audit event", fields...)
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
