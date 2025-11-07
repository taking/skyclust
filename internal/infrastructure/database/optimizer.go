package database

import (
	"fmt"
	"skyclust/internal/domain"
	"time"

	"gorm.io/gorm"
)

// DatabaseOptimizer handles database performance optimizations
type DatabaseOptimizer struct {
	db *gorm.DB
}

// NewDatabaseOptimizer creates a new database optimizer
func NewDatabaseOptimizer(db *gorm.DB) *DatabaseOptimizer {
	return &DatabaseOptimizer{
		db: db,
	}
}

// CreateIndexes creates optimized indexes for better query performance
func (o *DatabaseOptimizer) CreateIndexes() error {
	// User table indexes
	if err := o.createUserIndexes(); err != nil {
		return fmt.Errorf("failed to create user indexes: %w", err)
	}

	// Workspace table indexes
	if err := o.createWorkspaceIndexes(); err != nil {
		return fmt.Errorf("failed to create workspace indexes: %w", err)
	}

	// Credential table indexes
	if err := o.createCredentialIndexes(); err != nil {
		return fmt.Errorf("failed to create credential indexes: %w", err)
	}

	// AuditLog table indexes
	if err := o.createAuditLogIndexes(); err != nil {
		return fmt.Errorf("failed to create audit log indexes: %w", err)
	}

	return nil
}

// createUserIndexes creates indexes for the users table
func (o *DatabaseOptimizer) createUserIndexes() error {
	indexes := []string{
		// Primary lookup indexes
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_username ON users(username)",
		// Temporarily disabled until schema is aligned: idx_users_active

		// Composite indexes for common queries
		// Temporarily disabled until schema is aligned: idx_users_active_created
		// Temporarily disabled until schema is aligned: idx_users_email_active
	}

	for _, indexSQL := range indexes {
		if err := o.db.Exec(indexSQL).Error; err != nil {
			return fmt.Errorf("failed to create user index: %w", err)
		}
	}

	return nil
}

// createWorkspaceIndexes creates indexes for the workspaces table
func (o *DatabaseOptimizer) createWorkspaceIndexes() error {
	indexes := []string{
		// Primary lookup indexes
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_workspaces_name ON workspaces(name)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_workspaces_owner_id ON workspaces(owner_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_workspaces_active ON workspaces(is_active)",

		// Composite indexes for common queries
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_workspaces_owner_active ON workspaces(owner_id, is_active)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_workspaces_name_owner ON workspaces(name, owner_id)",
	}

	for _, indexSQL := range indexes {
		if err := o.db.Exec(indexSQL).Error; err != nil {
			return fmt.Errorf("failed to create workspace index: %w", err)
		}
	}

	return nil
}

// createCredentialIndexes creates indexes for the credentials table
func (o *DatabaseOptimizer) createCredentialIndexes() error {
	indexes := []string{
		// Primary lookup indexes
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_credentials_user_id ON credentials(user_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_credentials_provider ON credentials(provider)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_credentials_active ON credentials(is_active)",

		// Composite indexes for common queries
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_credentials_user_provider ON credentials(user_id, provider)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_credentials_user_active ON credentials(user_id, is_active)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_credentials_provider_active ON credentials(provider, is_active)",
	}

	for _, indexSQL := range indexes {
		if err := o.db.Exec(indexSQL).Error; err != nil {
			return fmt.Errorf("failed to create credential index: %w", err)
		}
	}

	return nil
}

// createAuditLogIndexes creates indexes for the audit_logs table
func (o *DatabaseOptimizer) createAuditLogIndexes() error {
	indexes := []string{
		// Primary lookup indexes
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_action ON audit_logs(action)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at)",

		// Composite indexes for common queries
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_user_action ON audit_logs(user_id, action)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_user_created ON audit_logs(user_id, created_at)",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_action_created ON audit_logs(action, created_at)",

		// Time-based partitioning support
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_logs_created_at_desc ON audit_logs(created_at DESC)",
	}

	for _, indexSQL := range indexes {
		if err := o.db.Exec(indexSQL).Error; err != nil {
			return fmt.Errorf("failed to create audit log index: %w", err)
		}
	}

	return nil
}

// OptimizeQueries performs query optimization tasks
func (o *DatabaseOptimizer) OptimizeQueries() error {
	// Update table statistics
	if err := o.updateTableStatistics(); err != nil {
		return fmt.Errorf("failed to update table statistics: %w", err)
	}

	// Analyze query performance
	if err := o.analyzeQueryPerformance(); err != nil {
		return fmt.Errorf("failed to analyze query performance: %w", err)
	}

	return nil
}

// updateTableStatistics updates table statistics for better query planning
func (o *DatabaseOptimizer) updateTableStatistics() error {
	tables := []string{"users", "workspaces", "credentials", "audit_logs"}

	for _, table := range tables {
		if err := o.db.Exec(fmt.Sprintf("ANALYZE %s", table)).Error; err != nil {
			return fmt.Errorf("failed to analyze table %s: %w", table, err)
		}
	}

	return nil
}

// analyzeQueryPerformance analyzes and reports on query performance
func (o *DatabaseOptimizer) analyzeQueryPerformance() error {
	// This would typically involve analyzing slow query logs
	// For now, we'll just ensure the database is ready for performance monitoring
	return nil
}

// GetDatabaseStats returns database performance statistics
func (o *DatabaseOptimizer) GetDatabaseStats() (*DatabaseStats, error) {
	var stats DatabaseStats

	// Get connection pool stats
	sqlDB, err := o.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats.MaxOpenConns = sqlDB.Stats().MaxOpenConnections
	stats.OpenConns = sqlDB.Stats().OpenConnections
	stats.InUse = sqlDB.Stats().InUse
	stats.Idle = sqlDB.Stats().Idle
	stats.WaitCount = sqlDB.Stats().WaitCount
	stats.WaitDuration = sqlDB.Stats().WaitDuration

	// Get table sizes
	if err := o.getTableSizes(&stats); err != nil {
		return nil, fmt.Errorf("failed to get table sizes: %w", err)
	}

	// Get index usage stats
	if err := o.getIndexUsageStats(&stats); err != nil {
		return nil, fmt.Errorf("failed to get index usage stats: %w", err)
	}

	return &stats, nil
}

// getTableSizes gets the size of each table
func (o *DatabaseOptimizer) getTableSizes(stats *DatabaseStats) error {
	var results []struct {
		TableName string `json:"table_name"`
		Size      int64  `json:"size"`
	}

	if err := o.db.Raw(`
		SELECT 
			tablename as table_name,
			pg_total_relation_size(schemaname||'.'||tablename) as size
		FROM pg_tables 
		WHERE schemaname = 'public'
		ORDER BY size DESC
	`).Scan(&results).Error; err != nil {
		return err
	}

	stats.TableSizes = make(map[string]int64)
	for _, result := range results {
		stats.TableSizes[result.TableName] = result.Size
	}

	return nil
}

// getIndexUsageStats gets index usage statistics
func (o *DatabaseOptimizer) getIndexUsageStats(stats *DatabaseStats) error {
	var results []struct {
		IndexName string `json:"index_name"`
		TableName string `json:"table_name"`
		IndexSize int64  `json:"index_size"`
		Usage     int64  `json:"usage"`
	}

	if err := o.db.Raw(`
		SELECT 
			i.indexname as index_name,
			i.tablename as table_name,
			pg_relation_size(i.indexrelid) as index_size,
			s.idx_tup_read as usage
		FROM pg_indexes i
		LEFT JOIN pg_stat_user_indexes s ON i.indexname = s.indexrelname
		WHERE i.schemaname = 'public'
		ORDER BY index_size DESC
	`).Scan(&results).Error; err != nil {
		return err
	}

	stats.IndexUsage = make(map[string]int64)
	for _, result := range results {
		stats.IndexUsage[result.IndexName] = result.Usage
	}

	return nil
}

// DatabaseStats represents database performance statistics
type DatabaseStats struct {
	MaxOpenConns int              `json:"max_open_conns"`
	OpenConns    int              `json:"open_conns"`
	InUse        int              `json:"in_use"`
	Idle         int              `json:"idle"`
	WaitCount    int64            `json:"wait_count"`
	WaitDuration time.Duration    `json:"wait_duration"`
	TableSizes   map[string]int64 `json:"table_sizes"`
	IndexUsage   map[string]int64 `json:"index_usage"`
}

// CleanupOldData removes old data to maintain performance
func (o *DatabaseOptimizer) CleanupOldData() error {
	// Clean up old audit logs (older than 90 days)
	if err := o.db.Where("created_at < ?", time.Now().AddDate(0, 0, -90)).Delete(&domain.AuditLog{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup old audit logs: %w", err)
	}

	// Note: Soft delete has been removed, so no cleanup of soft-deleted records is needed
	// Records are now permanently deleted when Delete() is called

	return nil
}
