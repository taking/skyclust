package usecase

import (
	"skyclust/internal/domain"
	"time"

	"github.com/google/uuid"
)

// auditLogService implements the audit log business logic
type auditLogService struct {
	auditLogRepo domain.AuditLogRepository
}

// NewAuditLogService creates a new audit log service
func NewAuditLogService(auditLogRepo domain.AuditLogRepository) domain.AuditLogService {
	return &auditLogService{
		auditLogRepo: auditLogRepo,
	}
}

// LogAction logs a user action
func (s *auditLogService) LogAction(userID uuid.UUID, action, resource string, details map[string]interface{}) error {
	log := &domain.AuditLog{
		UserID:   userID,
		Action:   action,
		Resource: resource,
		Details:  details,
	}
	return s.auditLogRepo.Create(log)
}

// GetUserLogs retrieves audit log entries for a user with pagination
func (s *auditLogService) GetUserLogs(userID uuid.UUID, limit, offset int) ([]*domain.AuditLog, error) {
	return s.auditLogRepo.GetByUserID(userID, limit, offset)
}

// GetLogsByAction retrieves audit log entries by action with pagination
func (s *auditLogService) GetLogsByAction(action string, limit, offset int) ([]*domain.AuditLog, error) {
	return s.auditLogRepo.GetByAction(action, limit, offset)
}

// GetLogsByDateRange retrieves audit log entries by date range with pagination
func (s *auditLogService) GetLogsByDateRange(start, end time.Time, limit, offset int) ([]*domain.AuditLog, error) {
	return s.auditLogRepo.GetByDateRange(start, end, limit, offset)
}

// CleanupOldLogs removes old audit log entries
func (s *auditLogService) CleanupOldLogs(retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	return s.auditLogRepo.DeleteOldLogs(cutoffDate)
}
