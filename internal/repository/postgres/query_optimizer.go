package postgres

import (
	"fmt"
	"skyclust/internal/domain"
	"strings"
	"time"

	"gorm.io/gorm"
)

// QueryOptimizer provides query optimization utilities
type QueryOptimizer struct {
	db *gorm.DB
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer(db *gorm.DB) *QueryOptimizer {
	return &QueryOptimizer{db: db}
}

// OptimizedUserQueries provides optimized user queries
type OptimizedUserQueries struct {
	db *gorm.DB
}

// NewOptimizedUserQueries creates optimized user queries
func NewOptimizedUserQueries(db *gorm.DB) *OptimizedUserQueries {
	return &OptimizedUserQueries{db: db}
}

// GetUsersWithPagination retrieves users with optimized pagination
func (q *OptimizedUserQueries) GetUsersWithPagination(limit, offset int, filters map[string]interface{}) ([]*domain.User, int64, error) {
	var users []*domain.User
	var total int64

	// Build base query
	query := q.db.Model(&domain.User{}).Where("deleted_at IS NULL")

	// Apply filters
	query = q.applyUserFilters(query, filters)

	// Get total count (optimized with index)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Apply pagination and ordering
	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}

	return users, total, nil
}

// applyUserFilters applies filters to the query
func (q *OptimizedUserQueries) applyUserFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	for key, value := range filters {
		switch key {
		case "is_active":
			if active, ok := value.(bool); ok {
				query = query.Where("is_active = ?", active)
			}
		case "email":
			if email, ok := value.(string); ok && email != "" {
				query = query.Where("email ILIKE ?", "%"+email+"%")
			}
		case "username":
			if username, ok := value.(string); ok && username != "" {
				query = query.Where("username ILIKE ?", "%"+username+"%")
			}
		case "created_after":
			if date, ok := value.(time.Time); ok {
				query = query.Where("created_at >= ?", date)
			}
		case "created_before":
			if date, ok := value.(time.Time); ok {
				query = query.Where("created_at <= ?", date)
			}
		}
	}
	return query
}

// GetUserByEmailOptimized retrieves user by email with optimized query
func (q *OptimizedUserQueries) GetUserByEmailOptimized(email string) (*domain.User, error) {
	var user domain.User
	if err := q.db.
		Select("id, username, email, is_active, created_at, updated_at").
		Where("email = ? AND deleted_at IS NULL", email).
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// GetUserByUsernameOptimized retrieves user by username with optimized query
func (q *OptimizedUserQueries) GetUserByUsernameOptimized(username string) (*domain.User, error) {
	var user domain.User
	if err := q.db.
		Select("id, username, email, is_active, created_at, updated_at").
		Where("username = ? AND deleted_at IS NULL", username).
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

// OptimizedCredentialQueries provides optimized credential queries
type OptimizedCredentialQueries struct {
	db *gorm.DB
}

// NewOptimizedCredentialQueries creates optimized credential queries
func NewOptimizedCredentialQueries(db *gorm.DB) *OptimizedCredentialQueries {
	return &OptimizedCredentialQueries{db: db}
}

// GetCredentialsByUserIDOptimized retrieves credentials by user ID with optimized query
func (q *OptimizedCredentialQueries) GetCredentialsByUserIDOptimized(userID string) ([]*domain.Credential, error) {
	var credentials []*domain.Credential
	if err := q.db.
		Select("id, user_id, provider, name, is_active, created_at, updated_at").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Find(&credentials).Error; err != nil {
		return nil, fmt.Errorf("failed to get credentials by user ID: %w", err)
	}
	return credentials, nil
}

// GetActiveCredentialsByProvider retrieves active credentials by provider
func (q *OptimizedCredentialQueries) GetActiveCredentialsByProvider(provider string) ([]*domain.Credential, error) {
	var credentials []*domain.Credential
	if err := q.db.
		Select("id, user_id, provider, name, is_active, created_at, updated_at").
		Where("provider = ? AND is_active = true AND deleted_at IS NULL", provider).
		Order("created_at DESC").
		Find(&credentials).Error; err != nil {
		return nil, fmt.Errorf("failed to get active credentials by provider: %w", err)
	}
	return credentials, nil
}

// OptimizedAuditLogQueries provides optimized audit log queries
type OptimizedAuditLogQueries struct {
	db *gorm.DB
}

// NewOptimizedAuditLogQueries creates optimized audit log queries
func NewOptimizedAuditLogQueries(db *gorm.DB) *OptimizedAuditLogQueries {
	return &OptimizedAuditLogQueries{db: db}
}

// GetAuditLogsByUserIDOptimized retrieves audit logs by user ID with optimized query
func (q *OptimizedAuditLogQueries) GetAuditLogsByUserIDOptimized(userID string, limit, offset int) ([]*domain.AuditLog, int64, error) {
	var logs []*domain.AuditLog
	var total int64

	// Count total logs
	if err := q.db.Model(&domain.AuditLog{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Get logs with pagination
	if err := q.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return logs, total, nil
}

// GetAuditLogsByActionOptimized retrieves audit logs by action with optimized query
func (q *OptimizedAuditLogQueries) GetAuditLogsByActionOptimized(action string, limit, offset int) ([]*domain.AuditLog, int64, error) {
	var logs []*domain.AuditLog
	var total int64

	// Count total logs
	if err := q.db.Model(&domain.AuditLog{}).Where("action = ?", action).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Get logs with pagination
	if err := q.db.
		Where("action = ?", action).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return logs, total, nil
}

// GetAuditLogsByDateRangeOptimized retrieves audit logs by date range with optimized query
func (q *OptimizedAuditLogQueries) GetAuditLogsByDateRangeOptimized(startDate, endDate time.Time, limit, offset int) ([]*domain.AuditLog, int64, error) {
	var logs []*domain.AuditLog
	var total int64

	// Count total logs
	if err := q.db.Model(&domain.AuditLog{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Get logs with pagination
	if err := q.db.
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return logs, total, nil
}

// BatchInsertAuditLogs inserts multiple audit logs in a single transaction
func (q *OptimizedAuditLogQueries) BatchInsertAuditLogs(logs []*domain.AuditLog) error {
	if len(logs) == 0 {
		return nil
	}

	// Use batch insert for better performance
	return q.db.Transaction(func(tx *gorm.DB) error {
		// Split into batches of 1000 for optimal performance
		batchSize := 1000
		for i := 0; i < len(logs); i += batchSize {
			end := i + batchSize
			if end > len(logs) {
				end = len(logs)
			}

			if err := tx.CreateInBatches(logs[i:end], batchSize).Error; err != nil {
				return fmt.Errorf("failed to batch insert audit logs: %w", err)
			}
		}
		return nil
	})
}

// CleanupOldAuditLogs removes old audit logs to maintain performance
func (q *OptimizedAuditLogQueries) CleanupOldAuditLogs(olderThan time.Time) (int64, error) {
	result := q.db.Where("created_at < ?", olderThan).Delete(&domain.AuditLog{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old audit logs: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// QueryBuilder provides a fluent query builder
type QueryBuilder struct {
	query *gorm.DB
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(db *gorm.DB, model interface{}) *QueryBuilder {
	return &QueryBuilder{
		query: db.Model(model),
	}
}

// Where adds a WHERE clause
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.query = qb.query.Where(condition, args...)
	return qb
}

// Order adds an ORDER BY clause
func (qb *QueryBuilder) Order(order string) *QueryBuilder {
	qb.query = qb.query.Order(order)
	return qb
}

// Limit adds a LIMIT clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.query = qb.query.Limit(limit)
	return qb
}

// Offset adds an OFFSET clause
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.query = qb.query.Offset(offset)
	return qb
}

// Select specifies columns to select
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.query = qb.query.Select(strings.Join(columns, ", "))
	return qb
}

// Build returns the final query
func (qb *QueryBuilder) Build() *gorm.DB {
	return qb.query
}
