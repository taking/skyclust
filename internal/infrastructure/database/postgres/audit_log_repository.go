package postgres

import (
	"skyclust/internal/domain"
	"skyclust/pkg/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// auditLogRepository implements the AuditLogRepository interface
type auditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *gorm.DB) domain.AuditLogRepository {
	return &auditLogRepository{db: db}
}

// Create creates a new audit log entry
func (r *auditLogRepository) Create(log *domain.AuditLog) error {
	if err := r.db.Create(log).Error; err != nil {
		logger.Errorf("Failed to create audit log: %v", err)
		return err
	}
	return nil
}

// GetByUserID retrieves audit logs for a user with pagination
func (r *auditLogRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	query := r.db.Where("user_id = ?", userID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&logs).Error; err != nil {
		logger.Errorf("Failed to get audit logs by user ID: %v", err)
		return nil, err
	}
	return logs, nil
}

// GetByAction retrieves audit logs by action with pagination
func (r *auditLogRepository) GetByAction(action string, limit, offset int) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	query := r.db.Where("action = ?", action).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&logs).Error; err != nil {
		logger.Errorf("Failed to get audit logs by action: %v", err)
		return nil, err
	}
	return logs, nil
}

// GetByDateRange retrieves audit logs within a date range with pagination
func (r *auditLogRepository) GetByDateRange(start, end time.Time, limit, offset int) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	query := r.db.Where("created_at BETWEEN ? AND ?", start, end).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&logs).Error; err != nil {
		logger.Errorf("Failed to get audit logs by date range: %v", err)
		return nil, err
	}
	return logs, nil
}

// CountByUserID returns the total number of audit logs for a user
func (r *auditLogRepository) CountByUserID(userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.Model(&domain.AuditLog{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		logger.Errorf("Failed to count audit logs by user ID: %v", err)
		return 0, err
	}
	return count, nil
}

// CountByAction returns the total number of audit logs for an action
func (r *auditLogRepository) CountByAction(action string) (int64, error) {
	var count int64
	if err := r.db.Model(&domain.AuditLog{}).Where("action = ?", action).Count(&count).Error; err != nil {
		logger.Errorf("Failed to count audit logs by action: %v", err)
		return 0, err
	}
	return count, nil
}

// DeleteOldLogs deletes audit logs older than the specified time
func (r *auditLogRepository) DeleteOldLogs(before time.Time) error {
	if err := r.db.Where("created_at < ?", before).Delete(&domain.AuditLog{}).Error; err != nil {
		logger.Errorf("Failed to delete old audit logs: %v", err)
		return err
	}
	return nil
}
