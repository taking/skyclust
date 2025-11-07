package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuditLogService defines the interface for audit log business logic
type AuditLogService interface {
	LogAction(userID uuid.UUID, action, resource string, details map[string]interface{}) error
	GetUserLogs(userID uuid.UUID, limit, offset int) ([]*AuditLog, error)
	GetLogsByAction(action string, limit, offset int) ([]*AuditLog, error)
	GetLogsByDateRange(start, end time.Time, limit, offset int) ([]*AuditLog, error)
	CleanupOldLogs(retentionDays int) error

	// Admin-specific methods
	GetAuditLogs(filters AuditLogFilters) ([]*AuditLog, int64, error)
	GetAuditLogByID(id uuid.UUID) (*AuditLog, error)
	GetAuditStats(filters AuditStatsFilters) (*AuditStats, error)
	ExportAuditLogs(filters AuditLogFilters, format string) ([]byte, error)
	CleanupAuditLogs(retentionDays int) (int64, error)
	GetAuditLogSummary(startTime, endTime time.Time) (*AuditLogSummary, error)
}
