package adapters

import (
	"context"

	"go.uber.org/zap"
	"skyclust/internal/domain"
)

// LoggerAdapter adapts zap.Logger to domain.LoggerService
type LoggerAdapter struct {
	logger *zap.Logger
}

// NewLoggerAdapter creates a new logger adapter
func NewLoggerAdapter(logger *zap.Logger) domain.LoggerService {
	return &LoggerAdapter{logger: logger}
}

// Debug logs a debug message
func (a *LoggerAdapter) Debug(ctx context.Context, message string, fields ...domain.LogField) {
	zapFields := a.convertFields(fields)
	a.logger.Debug(message, zapFields...)
}

// Info logs an info message
func (a *LoggerAdapter) Info(ctx context.Context, message string, fields ...domain.LogField) {
	zapFields := a.convertFields(fields)
	a.logger.Info(message, zapFields...)
}

// Warn logs a warning message
func (a *LoggerAdapter) Warn(ctx context.Context, message string, fields ...domain.LogField) {
	zapFields := a.convertFields(fields)
	a.logger.Warn(message, zapFields...)
}

// Error logs an error message
func (a *LoggerAdapter) Error(ctx context.Context, message string, err error, fields ...domain.LogField) {
	zapFields := a.convertFields(fields)
	zapFields = append(zapFields, zap.Error(err))
	a.logger.Error(message, zapFields...)
}

// Fatal logs a fatal message and exits
func (a *LoggerAdapter) Fatal(ctx context.Context, message string, err error, fields ...domain.LogField) {
	zapFields := a.convertFields(fields)
	zapFields = append(zapFields, zap.Error(err))
	a.logger.Fatal(message, zapFields...)
}

// convertFields converts domain.LogField to zap.Field
func (a *LoggerAdapter) convertFields(fields []domain.LogField) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}
	return zapFields
}
