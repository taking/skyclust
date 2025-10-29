package logging

import (
	"skyclust/internal/shared/errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// LogCategory represents the category of the log entry
type LogCategory int

const (
	LogCategorySystem LogCategory = iota
	LogCategoryBusiness
	LogCategoryAudit
	LogCategorySecurity
	LogCategoryPerformance
	LogCategoryError
)

// LogEntry represents a structured log entry
type LogEntry struct {
	ID         string                 `json:"id"`
	Level      LogLevel               `json:"level"`
	Category   LogCategory            `json:"category"`
	Message    string                 `json:"message"`
	Timestamp  time.Time              `json:"timestamp"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	UserID     *uuid.UUID             `json:"user_id,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	Operation  string                 `json:"operation,omitempty"`
	Duration   *time.Duration         `json:"duration,omitempty"`
	Error      *errors.UnifiedError   `json:"error,omitempty"`
	StackTrace []string               `json:"stack_trace,omitempty"`
}

// UnifiedLogger provides comprehensive logging functionality
type UnifiedLogger struct {
	logger *zap.Logger
	config *LoggerConfig
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level           LogLevel `json:"level"`
	Development     bool     `json:"development"`
	Structured      bool     `json:"structured"`
	EnableCaller    bool     `json:"enable_caller"`
	EnableStack     bool     `json:"enable_stack"`
	EnableTimestamp bool     `json:"enable_timestamp"`
	EnableRequestID bool     `json:"enable_request_id"`
	EnableUserID    bool     `json:"enable_user_id"`
	EnableOperation bool     `json:"enable_operation"`
	EnableDuration  bool     `json:"enable_duration"`
}

// NewUnifiedLogger creates a new unified logger
func NewUnifiedLogger(config *LoggerConfig) (*UnifiedLogger, error) {
	if config == nil {
		config = DefaultLoggerConfig()
	}

	var zapConfig zap.Config
	if config.Development {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	// Set log level
	switch config.Level {
	case LogLevelDebug:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case LogLevelInfo:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case LogLevelWarn:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case LogLevelError:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case LogLevelFatal:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Configure encoder
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncoderConfig.LevelKey = "level"
	zapConfig.EncoderConfig.MessageKey = "message"
	zapConfig.EncoderConfig.CallerKey = "caller"

	// Enable/disable caller
	if !config.EnableCaller {
		zapConfig.DisableCaller = true
	}

	// Enable/disable stack
	if !config.EnableStack {
		zapConfig.DisableStacktrace = true
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return &UnifiedLogger{
		logger: logger,
		config: config,
	}, nil
}

// DefaultLoggerConfig returns default logger configuration
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:           LogLevelInfo,
		Development:     false,
		Structured:      true,
		EnableCaller:    true,
		EnableStack:     true,
		EnableTimestamp: true,
		EnableRequestID: true,
		EnableUserID:    true,
		EnableOperation: true,
		EnableDuration:  true,
	}
}

// LogInfo logs an info level message
func (ul *UnifiedLogger) LogInfo(message string, fields ...zap.Field) {
	ul.logger.Info(message, fields...)
}

// LogWarn logs a warning level message
func (ul *UnifiedLogger) LogWarn(message string, fields ...zap.Field) {
	ul.logger.Warn(message, fields...)
}

// LogError logs an error level message
func (ul *UnifiedLogger) LogError(message string, fields ...zap.Field) {
	ul.logger.Error(message, fields...)
}

// LogDebug logs a debug level message
func (ul *UnifiedLogger) LogDebug(message string, fields ...zap.Field) {
	ul.logger.Debug(message, fields...)
}

// LogFatal logs a fatal level message
func (ul *UnifiedLogger) LogFatal(message string, fields ...zap.Field) {
	ul.logger.Fatal(message, fields...)
}

// LogBusinessEvent logs a business event
func (ul *UnifiedLogger) LogBusinessEvent(eventType, userID, resourceID string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", eventType),
		zap.String("category", "business"),
	}

	if userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	if resourceID != "" {
		fields = append(fields, zap.String("resource_id", resourceID))
	}

	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}

	ul.logger.Info("Business event", fields...)
}

// LogAuditEvent logs an audit event
func (ul *UnifiedLogger) LogAuditEvent(eventType, resourceType, resourceID, userID string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", eventType),
		zap.String("resource_type", resourceType),
		zap.String("category", "audit"),
	}

	if resourceID != "" {
		fields = append(fields, zap.String("resource_id", resourceID))
	}

	if userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}

	ul.logger.Info("Audit event", fields...)
}

// LogSecurityEvent logs a security event
func (ul *UnifiedLogger) LogSecurityEvent(eventType, userID string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", eventType),
		zap.String("category", "security"),
	}

	if userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}

	ul.logger.Warn("Security event", fields...)
}

// LogPerformanceEvent logs a performance event
func (ul *UnifiedLogger) LogPerformanceEvent(operation string, duration time.Duration, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.String("category", "performance"),
	}

	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}

	ul.logger.Info("Performance event", fields...)
}

// LogErrorEvent logs an error event
func (ul *UnifiedLogger) LogErrorEvent(err *errors.UnifiedError, context map[string]interface{}) {
	fields := []zap.Field{
		zap.String("error_id", err.ID),
		zap.String("error_code", err.Code),
		zap.String("error_message", err.Message),
		zap.String("error_level", err.Level.String()),
		zap.String("error_category", err.Category.String()),
		zap.String("category", "error"),
	}

	if err.UserID != nil {
		fields = append(fields, zap.String("user_id", err.UserID.String()))
	}

	if err.RequestID != "" {
		fields = append(fields, zap.String("request_id", err.RequestID))
	}

	if err.Operation != "" {
		fields = append(fields, zap.String("operation", err.Operation))
	}

	if err.Details != nil {
		for k, v := range err.Details {
			fields = append(fields, zap.Any("error_detail_"+k, v))
		}
	}

	for k, v := range context {
		fields = append(fields, zap.Any("context_"+k, v))
	}

	ul.logger.Error("Error event", fields...)
}

// LoggerManager manages multiple loggers
type LoggerManager struct {
	loggers       map[string]*UnifiedLogger
	defaultLogger *UnifiedLogger
}

// NewLoggerManager creates a new logger manager
func NewLoggerManager() *LoggerManager {
	defaultLogger, _ := NewUnifiedLogger(DefaultLoggerConfig())
	return &LoggerManager{
		loggers:       make(map[string]*UnifiedLogger),
		defaultLogger: defaultLogger,
	}
}

// GetLogger returns a logger by name
func (lm *LoggerManager) GetLogger(name string) *UnifiedLogger {
	if logger, exists := lm.loggers[name]; exists {
		return logger
	}
	return lm.defaultLogger
}

// AddLogger adds a logger to the manager
func (lm *LoggerManager) AddLogger(name string, logger *UnifiedLogger) {
	lm.loggers[name] = logger
}

// ContextLogger provides context-aware logging
type ContextLogger struct {
	*UnifiedLogger
	context map[string]interface{}
}

// NewContextLogger creates a new context logger
func NewContextLogger(logger *UnifiedLogger, context map[string]interface{}) *ContextLogger {
	return &ContextLogger{
		UnifiedLogger: logger,
		context:       context,
	}
}

// WithContext adds context to the logger
func (cl *ContextLogger) WithContext(key string, value interface{}) *ContextLogger {
	if cl.context == nil {
		cl.context = make(map[string]interface{})
	}
	cl.context[key] = value
	return cl
}

// WithUserID adds user ID to the context
func (cl *ContextLogger) WithUserID(userID uuid.UUID) *ContextLogger {
	return cl.WithContext("user_id", userID.String())
}

// WithRequestID adds request ID to the context
func (cl *ContextLogger) WithRequestID(requestID string) *ContextLogger {
	return cl.WithContext("request_id", requestID)
}

// WithOperation adds operation to the context
func (cl *ContextLogger) WithOperation(operation string) *ContextLogger {
	return cl.WithContext("operation", operation)
}

// LogInfo logs an info message with context
func (cl *ContextLogger) LogInfo(message string, fields ...zap.Field) {
	allFields := cl.buildFields(fields...)
	cl.UnifiedLogger.LogInfo(message, allFields...)
}

// LogWarn logs a warning message with context
func (cl *ContextLogger) LogWarn(message string, fields ...zap.Field) {
	allFields := cl.buildFields(fields...)
	cl.UnifiedLogger.LogWarn(message, allFields...)
}

// LogError logs an error message with context
func (cl *ContextLogger) LogError(message string, fields ...zap.Field) {
	allFields := cl.buildFields(fields...)
	cl.UnifiedLogger.LogError(message, allFields...)
}

// buildFields builds fields with context
func (cl *ContextLogger) buildFields(fields ...zap.Field) []zap.Field {
	allFields := make([]zap.Field, 0, len(fields)+len(cl.context))
	allFields = append(allFields, fields...)

	for k, v := range cl.context {
		allFields = append(allFields, zap.Any(k, v))
	}

	return allFields
}

func (ll LogLevel) String() string {
	switch ll {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	case LogLevelFatal:
		return "fatal"
	default:
		return "unknown"
	}
}

func (lc LogCategory) String() string {
	switch lc {
	case LogCategorySystem:
		return "system"
	case LogCategoryBusiness:
		return "business"
	case LogCategoryAudit:
		return "audit"
	case LogCategorySecurity:
		return "security"
	case LogCategoryPerformance:
		return "performance"
	case LogCategoryError:
		return "error"
	default:
		return "unknown"
	}
}
