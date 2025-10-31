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

// GetByID retrieves an audit log by ID
func (r *auditLogRepository) GetByID(id uuid.UUID) (*domain.AuditLog, error) {
	var log domain.AuditLog
	if err := r.db.Where("id = ?", id).First(&log).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrAuditLogNotFound
		}
		logger.Errorf("Failed to get audit log by ID: %v", err)
		return nil, err
	}
	return &log, nil
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

// DeleteOldLogs deletes audit logs older than the specified time and returns the count of deleted logs
func (r *auditLogRepository) DeleteOldLogs(before time.Time) (int64, error) {
	result := r.db.Where("created_at < ?", before).Delete(&domain.AuditLog{})
	if result.Error != nil {
		logger.Errorf("Failed to delete old audit logs: %v", result.Error)
		return 0, result.Error
	}

	deletedCount := result.RowsAffected
	if deletedCount > 0 {
		logger.Infof("Deleted %d old audit logs (older than %v)", deletedCount, before)
	}

	return deletedCount, nil
}

// GetTotalCount returns the total number of audit logs
func (r *auditLogRepository) GetTotalCount(filters domain.AuditStatsFilters) (int64, error) {
	var count int64
	query := r.db.Model(&domain.AuditLog{})

	// Apply date range filter if provided
	if filters.StartTime != nil && filters.EndTime != nil {
		query = query.Where("created_at BETWEEN ? AND ?", *filters.StartTime, *filters.EndTime)
	}

	if err := query.Count(&count).Error; err != nil {
		logger.Errorf("Failed to count total audit logs: %v", err)
		return 0, err
	}

	return count, nil
}

// GetUniqueUsersCount returns the count of unique users in audit logs
func (r *auditLogRepository) GetUniqueUsersCount(filters domain.AuditStatsFilters) (int64, error) {
	var count int64
	query := "SELECT COUNT(DISTINCT user_id) FROM audit_logs"
	var args []interface{}

	// Apply date range filter if provided
	if filters.StartTime != nil && filters.EndTime != nil {
		query += " WHERE created_at BETWEEN ? AND ?"
		args = append(args, *filters.StartTime, *filters.EndTime)
	}

	if err := r.db.Raw(query, args...).Scan(&count).Error; err != nil {
		logger.Errorf("Failed to count unique users: %v", err)
		return 0, err
	}

	return count, nil
}

// GetTopActions returns the top actions by count
func (r *auditLogRepository) GetTopActions(filters domain.AuditStatsFilters, limit int) ([]map[string]interface{}, error) {
	type Result struct {
		Action string `gorm:"column:action"`
		Count  int64  `gorm:"column:count"`
	}

	var results []Result
	query := r.db.Model(&domain.AuditLog{}).
		Select("action, COUNT(*) as count").
		Group("action").
		Order("count DESC")

	// Apply date range filter if provided
	if filters.StartTime != nil && filters.EndTime != nil {
		query = query.Where("created_at BETWEEN ? AND ?", *filters.StartTime, *filters.EndTime)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Scan(&results).Error; err != nil {
		logger.Errorf("Failed to get top actions: %v", err)
		return nil, err
	}

	topActions := make([]map[string]interface{}, len(results))
	for i, result := range results {
		topActions[i] = map[string]interface{}{
			"action": result.Action,
			"count":  result.Count,
		}
	}

	return topActions, nil
}

// GetTopResources returns the top resources by count
func (r *auditLogRepository) GetTopResources(filters domain.AuditStatsFilters, limit int) ([]map[string]interface{}, error) {
	type Result struct {
		Resource string `gorm:"column:resource"`
		Count    int64  `gorm:"column:count"`
	}

	var results []Result
	query := r.db.Model(&domain.AuditLog{}).
		Select("resource, COUNT(*) as count").
		Where("resource != ''").
		Group("resource").
		Order("count DESC")

	// Apply date range filter if provided
	if filters.StartTime != nil && filters.EndTime != nil {
		query = query.Where("created_at BETWEEN ? AND ?", *filters.StartTime, *filters.EndTime)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Scan(&results).Error; err != nil {
		logger.Errorf("Failed to get top resources: %v", err)
		return nil, err
	}

	topResources := make([]map[string]interface{}, len(results))
	for i, result := range results {
		topResources[i] = map[string]interface{}{
			"resource": result.Resource,
			"count":     result.Count,
		}
	}

	return topResources, nil
}

// GetEventsByDay returns the count of events grouped by day
func (r *auditLogRepository) GetEventsByDay(filters domain.AuditStatsFilters) ([]map[string]interface{}, error) {
	type Result struct {
		Date  string `gorm:"column:date"`
		Count int64  `gorm:"column:count"`
	}

	var results []Result

	// Default to last 30 days if no date range provided
	startTime := time.Now().AddDate(0, 0, -30)
	endTime := time.Now()
	if filters.StartTime != nil {
		startTime = *filters.StartTime
	}
	if filters.EndTime != nil {
		endTime = *filters.EndTime
	}

	query := r.db.Model(&domain.AuditLog{}).
		Select("DATE(created_at)::text as date, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Group("DATE(created_at)").
		Order("DATE(created_at) DESC")

	if err := query.Scan(&results).Error; err != nil {
		logger.Errorf("Failed to get events by day: %v", err)
		return nil, err
	}

	eventsByDay := make([]map[string]interface{}, len(results))
	for i, result := range results {
		eventsByDay[i] = map[string]interface{}{
			"date":  result.Date,
			"count": result.Count,
		}
	}

	return eventsByDay, nil
}

// GetMostActiveUser returns the most active user ID and their event count
func (r *auditLogRepository) GetMostActiveUser(startTime, endTime time.Time) (uuid.UUID, int64, error) {
	type Result struct {
		UserID uuid.UUID `gorm:"column:user_id"`
		Count  int64     `gorm:"column:count"`
	}

	var result Result
	query := r.db.Model(&domain.AuditLog{}).
		Select("user_id, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Group("user_id").
		Order("count DESC").
		Limit(1)

	if err := query.Scan(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return uuid.Nil, 0, nil
		}
		logger.Errorf("Failed to get most active user: %v", err)
		return uuid.Nil, 0, err
	}

	return result.UserID, result.Count, nil
}

// GetSecurityEventsCount returns the count of security-related events
func (r *auditLogRepository) GetSecurityEventsCount(startTime, endTime time.Time) (int64, error) {
	var count int64
	// Security-related actions: password_change, credential_create, credential_update, credential_delete, user_delete
	securityActions := []string{
		domain.ActionPasswordChange,
		domain.ActionCredentialCreate,
		domain.ActionCredentialUpdate,
		domain.ActionCredentialDelete,
		domain.ActionUserDelete,
	}

	query := r.db.Model(&domain.AuditLog{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Where("action IN ?", securityActions)

	if err := query.Count(&count).Error; err != nil {
		logger.Errorf("Failed to count security events: %v", err)
		return 0, err
	}

	return count, nil
}

// GetErrorEventsCount returns the count of error events
// Error events are identified by checking if details JSONB contains error-related fields
func (r *auditLogRepository) GetErrorEventsCount(startTime, endTime time.Time) (int64, error) {
	var count int64
	// Check for error-related indicators in action or details
	// Actions that typically indicate errors: failed login attempts, etc.
	// For now, we'll check if action contains "error" or if details contains error fields
	// This is a simplified approach - in production, you might want to check details JSONB more carefully
	
	// Count logs where action contains "error" or similar patterns
	query := r.db.Model(&domain.AuditLog{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Where("(action LIKE '%error%' OR action LIKE '%fail%' OR details::text LIKE '%\"error\"%' OR details::text LIKE '%\"failed\"%')")

	if err := query.Count(&count).Error; err != nil {
		logger.Errorf("Failed to count error events: %v", err)
		return 0, err
	}

	return count, nil
}
