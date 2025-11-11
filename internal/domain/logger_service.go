package domain

import (
	"context"
)

// LoggerService defines the interface for logging operations
// This abstraction allows the application layer to use logging without
// depending on specific logging implementations (e.g., zap, logrus)
type LoggerService interface {
	// Debug logs a debug message
	Debug(ctx context.Context, message string, fields ...LogField)

	// Info logs an info message
	Info(ctx context.Context, message string, fields ...LogField)

	// Warn logs a warning message
	Warn(ctx context.Context, message string, fields ...LogField)

	// Error logs an error message
	Error(ctx context.Context, message string, err error, fields ...LogField)

	// Fatal logs a fatal message and exits
	Fatal(ctx context.Context, message string, err error, fields ...LogField)
}

// LogField represents a single log field
type LogField struct {
	Key   string
	Value interface{}
}

// NewLogField creates a new log field
func NewLogField(key string, value interface{}) LogField {
	return LogField{Key: key, Value: value}
}
