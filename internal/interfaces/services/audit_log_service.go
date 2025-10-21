package services

import (
	"skyclust/internal/domain"
	"time"
)

// AuditLogService defines the interface for audit log operations
type AuditLogService interface {
	// CreateAuditLog creates a new audit log entry
	CreateAuditLog(auditLog *domain.AuditLog) error

	// GetAuditLog retrieves an audit log by ID
	GetAuditLog(id string) (*domain.AuditLog, error)

	// GetAuditLogs retrieves audit logs with filtering
	GetAuditLogs(userID string, action string, startDate, endDate time.Time, limit, offset int) ([]*domain.AuditLog, error)

	// GetAuditLogsByUser retrieves audit logs for a specific user
	GetAuditLogsByUser(userID string, limit, offset int) ([]*domain.AuditLog, error)

	// GetAuditLogsByAction retrieves audit logs for a specific action
	GetAuditLogsByAction(action string, limit, offset int) ([]*domain.AuditLog, error)

	// GetAuditLogsByDateRange retrieves audit logs within a date range
	GetAuditLogsByDateRange(startDate, endDate time.Time, limit, offset int) ([]*domain.AuditLog, error)

	// DeleteOldAuditLogs deletes audit logs older than specified date
	DeleteOldAuditLogs(beforeDate time.Time) error

	// GetAuditLogStats returns statistics about audit logs
	GetAuditLogStats() (map[string]interface{}, error)
}
