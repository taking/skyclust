package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuditLogRepository defines the interface for audit log data operations
type AuditLogRepository interface {
	Create(log *AuditLog) error
	GetByID(id uuid.UUID) (*AuditLog, error)
	GetByUserID(userID uuid.UUID, limit, offset int) ([]*AuditLog, error)
	GetByAction(action string, limit, offset int) ([]*AuditLog, error)
	GetByDateRange(start, end time.Time, limit, offset int) ([]*AuditLog, error)
	CountByUserID(userID uuid.UUID) (int64, error)
	CountByAction(action string) (int64, error)
	DeleteOldLogs(before time.Time) (int64, error)
	// Statistics methods
	GetTotalCount(filters AuditStatsFilters) (int64, error)
	GetUniqueUsersCount(filters AuditStatsFilters) (int64, error)
	GetTopActions(filters AuditStatsFilters, limit int) ([]map[string]interface{}, error)
	GetTopResources(filters AuditStatsFilters, limit int) ([]map[string]interface{}, error)
	GetEventsByDay(filters AuditStatsFilters) ([]map[string]interface{}, error)
	// Summary methods
	GetMostActiveUser(startTime, endTime time.Time) (uuid.UUID, int64, error)
	GetSecurityEventsCount(startTime, endTime time.Time) (int64, error)
	GetErrorEventsCount(startTime, endTime time.Time) (int64, error)
}

