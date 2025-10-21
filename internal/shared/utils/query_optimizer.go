package utils

import (
	"context"
	"time"

	"gorm.io/gorm"
	"skyclust/internal/domain"
)

// QueryOptimizer provides database query optimization utilities
type QueryOptimizer struct {
	db *gorm.DB
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer(db *gorm.DB) *QueryOptimizer {
	return &QueryOptimizer{
		db: db,
	}
}

// OptimizeUserQuery optimizes user-related queries
func (qo *QueryOptimizer) OptimizeUserQuery(ctx context.Context, userID string, includes []string) *gorm.DB {
	query := qo.db.WithContext(ctx).Model(&domain.User{})

	// Add preloading for related entities
	for _, include := range includes {
		switch include {
		case "credentials":
			query = query.Preload("Credentials")
		case "audit_logs":
			query = query.Preload("AuditLogs")
		case "workspaces":
			query = query.Preload("Workspaces")
		}
	}

	return query
}

// OptimizeWorkspaceQuery optimizes workspace-related queries
func (qo *QueryOptimizer) OptimizeWorkspaceQuery(ctx context.Context, userID string, includes []string) *gorm.DB {
	query := qo.db.WithContext(ctx).Model(&domain.Workspace{})

	// Add preloading for related entities
	for _, include := range includes {
		switch include {
		case "users":
			query = query.Preload("Users")
		case "credentials":
			query = query.Preload("Credentials")
		}
	}

	return query
}

// OptimizeNotificationQuery optimizes notification queries
func (qo *QueryOptimizer) OptimizeNotificationQuery(ctx context.Context, userID string, filters map[string]interface{}) *gorm.DB {
	query := qo.db.WithContext(ctx).Model(&domain.Notification{})

	// Apply filters efficiently
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if unreadOnly, ok := filters["unread_only"].(bool); ok && unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	if category, ok := filters["category"].(string); ok && category != "" {
		query = query.Where("category = ?", category)
	}

	if priority, ok := filters["priority"].(string); ok && priority != "" {
		query = query.Where("priority = ?", priority)
	}

	// Add ordering
	query = query.Order("created_at DESC")

	return query
}

// OptimizeAuditLogQuery optimizes audit log queries
func (qo *QueryOptimizer) OptimizeAuditLogQuery(ctx context.Context, filters map[string]interface{}) *gorm.DB {
	query := qo.db.WithContext(ctx).Model(&domain.AuditLog{})

	// Apply filters efficiently
	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if action, ok := filters["action"].(string); ok && action != "" {
		query = query.Where("action = ?", action)
	}

	if resource, ok := filters["resource"].(string); ok && resource != "" {
		query = query.Where("resource = ?", resource)
	}

	if startTime, ok := filters["start_time"].(time.Time); ok {
		query = query.Where("created_at >= ?", startTime)
	}

	if endTime, ok := filters["end_time"].(time.Time); ok {
		query = query.Where("created_at <= ?", endTime)
	}

	// Add ordering
	query = query.Order("created_at DESC")

	return query
}

// BatchQueryOptimizer provides batch query optimization
type BatchQueryOptimizer struct {
	db *gorm.DB
}

// NewBatchQueryOptimizer creates a new batch query optimizer
func NewBatchQueryOptimizer(db *gorm.DB) *BatchQueryOptimizer {
	return &BatchQueryOptimizer{
		db: db,
	}
}

// OptimizeBatchUserQuery optimizes batch user queries
func (bqo *BatchQueryOptimizer) OptimizeBatchUserQuery(ctx context.Context, userIDs []string) ([]*domain.User, error) {
	var users []*domain.User

	// Use IN clause for batch queries
	err := bqo.db.WithContext(ctx).
		Where("id IN ?", userIDs).
		Preload("Credentials").
		Preload("AuditLogs").
		Find(&users).Error

	return users, err
}

// OptimizeBatchWorkspaceQuery optimizes batch workspace queries
func (bqo *BatchQueryOptimizer) OptimizeBatchWorkspaceQuery(ctx context.Context, workspaceIDs []string) ([]*domain.Workspace, error) {
	var workspaces []*domain.Workspace

	err := bqo.db.WithContext(ctx).
		Where("id IN ?", workspaceIDs).
		Preload("Users").
		Find(&workspaces).Error

	return workspaces, err
}

// QueryCache provides query result caching
type QueryCache struct {
	cache map[string]interface{}
	ttl   time.Duration
}

// NewQueryCache creates a new query cache
func NewQueryCache(ttl time.Duration) *QueryCache {
	return &QueryCache{
		cache: make(map[string]interface{}),
		ttl:   ttl,
	}
}

// Get retrieves a cached query result
func (qc *QueryCache) Get(key string) (interface{}, bool) {
	// In a real implementation, this would use Redis or similar
	// For now, return false to indicate cache miss
	return nil, false
}

// Set stores a query result in cache
func (qc *QueryCache) Set(key string, value interface{}) {
	// In a real implementation, this would use Redis or similar
	// For now, do nothing
}

// QueryMetrics tracks query performance metrics
type QueryMetrics struct {
	QueryCount  int64
	TotalTime   time.Duration
	AverageTime time.Duration
	SlowQueries int64
	ErrorCount  int64
}

// QueryProfiler provides query profiling capabilities
type QueryProfiler struct {
	metrics *QueryMetrics
}

// NewQueryProfiler creates a new query profiler
func NewQueryProfiler() *QueryProfiler {
	return &QueryProfiler{
		metrics: &QueryMetrics{},
	}
}

// ProfileQuery profiles a database query
func (qp *QueryProfiler) ProfileQuery(queryName string, fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	// Update metrics
	qp.metrics.QueryCount++
	qp.metrics.TotalTime += duration
	qp.metrics.AverageTime = qp.metrics.TotalTime / time.Duration(qp.metrics.QueryCount)

	// Track slow queries (over 100ms)
	if duration > 100*time.Millisecond {
		qp.metrics.SlowQueries++
	}

	if err != nil {
		qp.metrics.ErrorCount++
	}

	return err
}

// GetMetrics returns current query metrics
func (qp *QueryProfiler) GetMetrics() *QueryMetrics {
	return qp.metrics
}

// QueryBuilder provides a fluent interface for building queries
type QueryBuilder struct {
	db      *gorm.DB
	query   *gorm.DB
	context context.Context
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(db *gorm.DB, ctx context.Context) *QueryBuilder {
	return &QueryBuilder{
		db:      db,
		query:   db.WithContext(ctx),
		context: ctx,
	}
}

// Select specifies the columns to select
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.query = qb.query.Select(columns)
	return qb
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

// Preload adds a preload clause
func (qb *QueryBuilder) Preload(associations ...string) *QueryBuilder {
	for _, assoc := range associations {
		qb.query = qb.query.Preload(assoc)
	}
	return qb
}

// Execute executes the query
func (qb *QueryBuilder) Execute(dest interface{}) error {
	return qb.query.Find(dest).Error
}

// Count returns the count of records
func (qb *QueryBuilder) Count(count *int64) error {
	return qb.query.Count(count).Error
}
