package logging

import (
	"time"

	"go.uber.org/zap"
)

// LogLevelGuidelines defines when to use each log level
// This ensures consistent log level usage across the codebase
var LogLevelGuidelines = struct {
	// Debug: Detailed information for debugging, typically only enabled during development
	// Examples: Function entry/exit, variable values, detailed state information
	Debug string

	// Info: General informational messages about application flow
	// Examples: Successful operations, state changes, business events
	Info string

	// Warn: Warning messages for potentially harmful situations
	// Examples: Deprecated features, fallback to default values, recoverable errors
	Warn string

	// Error: Error messages for error events that might still allow the application to continue
	// Examples: Failed operations, validation errors, external service failures
	Error string

	// Fatal: Critical errors that cause the application to abort
	// Examples: Database connection failures, critical configuration errors
	Fatal string
}{
	Debug: "Use for detailed debugging information, typically only in development",
	Info:  "Use for general informational messages about successful operations",
	Warn:  "Use for warning messages about potentially harmful situations",
	Error: "Use for error events that might still allow the application to continue",
	Fatal: "Use for critical errors that cause the application to abort",
}

// StandardizedLogging provides helper methods for consistent structured logging
type StandardizedLogging struct {
	logger *zap.Logger
}

// NewStandardizedLogging creates a new standardized logging instance
func NewStandardizedLogging(logger *zap.Logger) *StandardizedLogging {
	return &StandardizedLogging{
		logger: logger,
	}
}

// LogOperationSuccess logs a successful operation with standardized fields
func (sl *StandardizedLogging) LogOperationSuccess(operation string, fields ...zap.Field) {
	allFields := []zap.Field{WithOperation(operation)}
	allFields = append(allFields, fields...)
	sl.logger.Info("Operation completed successfully", allFields...)
}

// LogOperationError logs a failed operation with standardized fields
func (sl *StandardizedLogging) LogOperationError(operation string, err error, fields ...zap.Field) {
	allFields := []zap.Field{
		WithOperation(operation),
		zap.Error(err),
	}
	allFields = append(allFields, fields...)
	sl.logger.Error("Operation failed", allFields...)
}

// LogOperationWarning logs a warning during an operation with standardized fields
func (sl *StandardizedLogging) LogOperationWarning(operation string, message string, fields ...zap.Field) {
	allFields := []zap.Field{
		WithOperation(operation),
		zap.String(FieldMessage, message),
	}
	allFields = append(allFields, fields...)
	sl.logger.Warn("Operation warning", allFields...)
}

// LogOperationDebug logs debug information for an operation with standardized fields
func (sl *StandardizedLogging) LogOperationDebug(operation string, message string, fields ...zap.Field) {
	allFields := []zap.Field{
		WithOperation(operation),
		zap.String(FieldMessage, message),
	}
	allFields = append(allFields, fields...)
	sl.logger.Debug("Operation debug", allFields...)
}

// Additional helper functions for common field combinations

// WithUsername adds a username field (for user-related operations)
func WithUsername(username string) zap.Field {
	return zap.String("username", username)
}

// WithEmail adds an email field (for user-related operations)
func WithEmail(email string) zap.Field {
	return zap.String("email", email)
}

// WithClusterName adds a cluster_name field (for Kubernetes operations)
func WithClusterName(clusterName string) zap.Field {
	return zap.String("cluster_name", clusterName)
}

// WithVPCID adds a vpc_id field (for network operations)
func WithVPCID(vpcID string) zap.Field {
	return zap.String("vpc_id", vpcID)
}

// WithSubnetID adds a subnet_id field (for network operations)
func WithSubnetID(subnetID string) zap.Field {
	return zap.String("subnet_id", subnetID)
}

// WithSecurityGroupID adds a security_group_id field (for network operations)
func WithSecurityGroupID(securityGroupID string) zap.Field {
	return zap.String("security_group_id", securityGroupID)
}

// WithRegion adds a region field (for cloud provider operations)
func WithRegion(region string) zap.Field {
	return zap.String("region", region)
}

// WithAvailabilityZone adds an availability_zone field (for cloud provider operations)
func WithAvailabilityZone(zone string) zap.Field {
	return zap.String("availability_zone", zone)
}

// WithCount adds a count field (for list operations)
func WithCount(count int) zap.Field {
	return zap.Int("count", count)
}

// WithLimit adds a limit field (for pagination)
func WithLimit(limit int) zap.Field {
	return zap.Int("limit", limit)
}

// WithOffset adds an offset field (for pagination)
func WithOffset(offset int) zap.Field {
	return zap.Int("offset", offset)
}

// WithDuration adds a duration field (for performance metrics)
func WithDuration(duration interface{}) zap.Field {
	switch v := duration.(type) {
	case int64:
		return zap.Int64(FieldDuration, v)
	case time.Duration:
		return zap.Duration(FieldDuration, v)
	default:
		return zap.Any(FieldDuration, duration)
	}
}

// WithError adds an error field (standardized error logging)
func WithError(err error) zap.Field {
	return zap.Error(err)
}
