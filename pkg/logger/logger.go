package logger

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger provides comprehensive logging functionality
type Logger struct {
	logger *zap.Logger
	config *LoggerConfig
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level            string   `json:"level"`
	Development      bool     `json:"development"`
	Structured       bool     `json:"structured"`
	EnableCaller     bool     `json:"enable_caller"`
	EnableStack      bool     `json:"enable_stack"`
	OutputPaths      []string `json:"output_paths"`
	ErrorOutputPaths []string `json:"error_output_paths"`
}

// NewLogger creates a new logger
func NewLogger(config *LoggerConfig) (*Logger, error) {
	if config == nil {
		config = GetDefaultLoggerConfig()
	}

	var zapConfig zap.Config
	if config.Development {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	// Set log level
	switch config.Level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Configure output
	if len(config.OutputPaths) > 0 {
		zapConfig.OutputPaths = config.OutputPaths
	}
	if len(config.ErrorOutputPaths) > 0 {
		zapConfig.ErrorOutputPaths = config.ErrorOutputPaths
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

	return &Logger{
		logger: logger,
		config: config,
	}, nil
}

// GetDefaultLoggerConfig returns default logger configuration
func GetDefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:            "info",
		Development:      false,
		Structured:       true,
		EnableCaller:     true,
		EnableStack:      true,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

// WithContext creates a logger with context information
func (l *Logger) WithContext(ctx context.Context) *zap.Logger {
	// Extract trace information from context
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		spanCtx := span.SpanContext()
		return l.logger.With(
			zap.String("trace_id", spanCtx.TraceID().String()),
			zap.String("span_id", spanCtx.SpanID().String()),
		)
	}
	return l.logger
}

// WithRequestID creates a logger with request ID
func (l *Logger) WithRequestID(requestID string) *zap.Logger {
	return l.logger.With(zap.String("request_id", requestID))
}

// WithUser creates a logger with user information
func (l *Logger) WithUser(userID, username string) *zap.Logger {
	return l.logger.With(
		zap.String("user_id", userID),
		zap.String("username", username),
	)
}

// WithFields creates a logger with multiple fields
func (l *Logger) WithFields(fields map[string]interface{}) *zap.Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	return l.logger.With(zapFields...)
}

// WithError creates a logger with error information
func (l *Logger) WithError(err error) *zap.Logger {
	return l.logger.With(zap.Error(err))
}

// LogHTTPRequest logs an HTTP request
func (l *Logger) LogHTTPRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration, requestID string) {
	logger := l.WithContext(ctx).With(
		zap.String("type", "http_request"),
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
		zap.String("request_id", requestID),
	)

	if statusCode >= 400 {
		logger.Warn("HTTP request completed with error")
	} else {
		logger.Info("HTTP request completed")
	}
}

// LogDatabaseQuery logs a database query
func (l *Logger) LogDatabaseQuery(ctx context.Context, operation, table string, duration time.Duration, success bool, err error) {
	logger := l.WithContext(ctx).With(
		zap.String("type", "database_query"),
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Duration("duration", duration),
		zap.Bool("success", success),
	)

	if err != nil {
		logger.With(zap.Error(err)).Error("Database query failed")
	} else {
		logger.Debug("Database query completed")
	}
}

// LogCacheOperation logs a cache operation
func (l *Logger) LogCacheOperation(ctx context.Context, operation, key string, duration time.Duration, success bool, err error) {
	logger := l.WithContext(ctx).With(
		zap.String("type", "cache_operation"),
		zap.String("operation", operation),
		zap.String("key", key),
		zap.Duration("duration", duration),
		zap.Bool("success", success),
	)

	if err != nil {
		logger.With(zap.Error(err)).Error("Cache operation failed")
	} else {
		logger.Debug("Cache operation completed")
	}
}

// LogBusinessEvent logs a business event
func (l *Logger) LogBusinessEvent(ctx context.Context, event, userID string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("type", "business_event"),
		zap.String("event", event),
		zap.String("user_id", userID),
	}

	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}

	l.WithContext(ctx).With(fields...).Info("Business event occurred")
}

// LogSecurityEvent logs a security event
func (l *Logger) LogSecurityEvent(ctx context.Context, event, userID, ip string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("type", "security_event"),
		zap.String("event", event),
		zap.String("user_id", userID),
		zap.String("ip_address", ip),
	}

	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}

	l.WithContext(ctx).With(fields...).Warn("Security event occurred")
}

// LogSystemEvent logs a system event
func (l *Logger) LogSystemEvent(ctx context.Context, event string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("type", "system_event"),
		zap.String("event", event),
	}

	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}

	l.WithContext(ctx).With(fields...).Info("System event occurred")
}

// LogPluginEvent logs a plugin event
func (l *Logger) LogPluginEvent(ctx context.Context, pluginName, event string, duration time.Duration, success bool, err error) {
	logger := l.WithContext(ctx).With(
		zap.String("type", "plugin_event"),
		zap.String("plugin_name", pluginName),
		zap.String("event", event),
		zap.Duration("duration", duration),
		zap.Bool("success", success),
	)

	if err != nil {
		logger.With(zap.Error(err)).Error("Plugin event failed")
	} else {
		logger.Info("Plugin event completed")
	}
}

// LogPerformanceMetrics logs performance metrics
func (l *Logger) LogPerformanceMetrics(ctx context.Context, metrics map[string]interface{}) {
	fields := make([]zap.Field, 0, len(metrics))
	for key, value := range metrics {
		fields = append(fields, zap.Any(key, value))
	}
	l.WithContext(ctx).With(fields...).Info("Performance metrics collected")
}

// LogError logs an error with context
func (l *Logger) LogError(ctx context.Context, err error, message string, fields ...zap.Field) {
	l.WithContext(ctx).With(zap.Error(err)).With(fields...).Error(message)
}

// LogWarning logs a warning with context
func (l *Logger) LogWarning(ctx context.Context, message string, fields ...zap.Field) {
	l.WithContext(ctx).With(fields...).Warn(message)
}

// LogInfo logs an info message with context
func (l *Logger) LogInfo(ctx context.Context, message string, fields ...zap.Field) {
	l.WithContext(ctx).With(fields...).Info(message)
}

// LogDebug logs a debug message with context
func (l *Logger) LogDebug(ctx context.Context, message string, fields ...zap.Field) {
	l.WithContext(ctx).With(fields...).Debug(message)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	// Ignore sync errors for stderr as it's a known issue on some systems
	_ = l.logger.Sync()
	return nil
}

// Close closes the logger
func (l *Logger) Close() error {
	// Ignore sync errors for stderr as it's a known issue on some systems
	_ = l.logger.Sync()
	return nil
}

// GetLogger returns the underlying zap logger
func (l *Logger) GetLogger() *zap.Logger {
	return l.logger
}

// GetConfig returns the logger configuration
func (l *Logger) GetConfig() *LoggerConfig {
	return l.config
}

// Legacy compatibility methods for existing code

// WithFieldLegacy adds a field to the logger (legacy compatibility)
func (l *Logger) WithFieldLegacy(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With(zap.Any(key, value)),
		config: l.config,
	}
}

// WithFieldsLegacy adds multiple fields to the logger (legacy compatibility)
func (l *Logger) WithFieldsLegacy(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &Logger{
		logger: l.logger.With(zapFields...),
		config: l.config,
	}
}

// WithRequestIDLegacy adds request ID to the logger (legacy compatibility)
func (l *Logger) WithRequestIDLegacy(requestID string) *Logger {
	return l.WithFieldLegacy("request_id", requestID)
}

// WithErrorLegacy adds error to the logger (legacy compatibility)
func (l *Logger) WithErrorLegacy(err error) *Logger {
	return &Logger{
		logger: l.logger.With(zap.Error(err)),
		config: l.config,
	}
}

// Basic logging methods (legacy compatibility)
func (l *Logger) Debug(message string) {
	l.logger.Debug(message)
}

func (l *Logger) Info(message string) {
	l.logger.Info(message)
}

func (l *Logger) Warn(message string) {
	l.logger.Warn(message)
}

func (l *Logger) Error(message string) {
	l.logger.Error(message)
}

func (l *Logger) Fatal(message string) {
	l.logger.Fatal(message)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Sugar().Debugf(format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Sugar().Infof(format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Sugar().Warnf(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Sugar().Errorf(format, args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Sugar().Fatalf(format, args...)
}

// Global logger instance for backward compatibility
var (
	DefaultLogger *Logger
)

// Initialize the default logger
func init() {
	var err error
	DefaultLogger, err = NewLogger(GetDefaultLoggerConfig())
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
}

// SetDefaultLogger sets the default logger
func SetDefaultLogger(logger *Logger) {
	DefaultLogger = logger
}

// Global logging functions for backward compatibility
func Debug(message string) {
	DefaultLogger.Debug(message)
}

func Info(message string) {
	DefaultLogger.Info(message)
}

func Warn(message string) {
	DefaultLogger.Warn(message)
}

func Error(message string) {
	DefaultLogger.Error(message)
}

func Fatal(message string) {
	DefaultLogger.Fatal(message)
}

func Debugf(format string, args ...interface{}) {
	DefaultLogger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	DefaultLogger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	DefaultLogger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	DefaultLogger.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	DefaultLogger.Fatalf(format, args...)
}

// WithRequestID returns a logger with request ID context (legacy compatibility)
func WithRequestID(requestID string) *Logger {
	if requestID == "" {
		return DefaultLogger
	}
	return DefaultLogger.WithFieldLegacy("request_id", requestID)
}

// WithError returns a logger with error context (legacy compatibility)
func WithError(err error) *Logger {
	return DefaultLogger.WithErrorLegacy(err)
}
