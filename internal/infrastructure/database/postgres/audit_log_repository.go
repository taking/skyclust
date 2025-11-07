package postgres

import (
	"skyclust/internal/domain"
	"skyclust/pkg/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// auditLogRepository: domain.AuditLogRepository 인터페이스 구현체
type auditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository: 새로운 AuditLogRepository를 생성합니다
func NewAuditLogRepository(db *gorm.DB) domain.AuditLogRepository {
	return &auditLogRepository{db: db}
}

// Create: 새로운 감사 로그 항목을 생성합니다
func (r *auditLogRepository) Create(log *domain.AuditLog) error {
	if err := r.db.Create(log).Error; err != nil {
		logger.Errorf("Failed to create audit log: %v", err)
		return err
	}
	return nil
}

// GetByID: ID로 감사 로그를 조회합니다
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

// GetByUserID: 사용자의 감사 로그를 페이지네이션과 함께 조회합니다
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

// GetByAction: 액션으로 감사 로그를 페이지네이션과 함께 조회합니다
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

// GetByDateRange: 날짜 범위 내의 감사 로그를 페이지네이션과 함께 조회합니다
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

// CountByUserID: 사용자의 감사 로그 총 개수를 반환합니다
func (r *auditLogRepository) CountByUserID(userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.Model(&domain.AuditLog{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		logger.Errorf("Failed to count audit logs by user ID: %v", err)
		return 0, err
	}
	return count, nil
}

// CountByAction: 액션별 감사 로그 총 개수를 반환합니다
func (r *auditLogRepository) CountByAction(action string) (int64, error) {
	var count int64
	if err := r.db.Model(&domain.AuditLog{}).Where("action = ?", action).Count(&count).Error; err != nil {
		logger.Errorf("Failed to count audit logs by action: %v", err)
		return 0, err
	}
	return count, nil
}

// DeleteOldLogs: 지정된 시간보다 오래된 감사 로그를 삭제하고 삭제된 로그 수를 반환합니다
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

// GetTotalCount: 감사 로그의 총 개수를 반환합니다
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

// GetUniqueUsersCount: 감사 로그의 고유 사용자 수를 반환합니다
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

// GetTopActions: 개수별 상위 액션 목록을 반환합니다
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

// GetTopResources: 개수별 상위 리소스 목록을 반환합니다
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

// GetEventsByDay: 일별로 그룹화된 이벤트 개수를 반환합니다
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

// GetMostActiveUser: 가장 활성화된 사용자 ID와 이벤트 개수를 반환합니다
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

// GetSecurityEventsCount: 보안 관련 이벤트 개수를 반환합니다
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

// GetErrorEventsCount: 에러 이벤트 개수를 반환합니다
// 에러 이벤트는 details JSONB에 에러 관련 필드가 포함되어 있는지 확인하여 식별합니다
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
