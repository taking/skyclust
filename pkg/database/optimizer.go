package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cmp/pkg/shared/logger"
)

// IndexInfo represents information about a database index
type IndexInfo struct {
	TableName  string
	IndexName  string
	ColumnName string
	IndexType  string
	IsUnique   bool
	IsPrimary  bool
	Size       int64
	Usage      int64
	LastUsed   time.Time
}

// QueryStats represents query performance statistics
type QueryStats struct {
	Query        string
	AvgDuration  time.Duration
	TotalCalls   int64
	SlowQueries  int64
	LastExecuted time.Time
}

// DatabaseOptimizer provides database optimization tools
type DatabaseOptimizer struct {
	db *sql.DB
}

// NewDatabaseOptimizer creates a new database optimizer
func NewDatabaseOptimizer(db *sql.DB) *DatabaseOptimizer {
	return &DatabaseOptimizer{db: db}
}

// AnalyzeIndexes analyzes database indexes and provides optimization recommendations
func (do *DatabaseOptimizer) AnalyzeIndexes(ctx context.Context) ([]IndexRecommendation, error) {
	// Get all indexes
	indexes, err := do.getIndexes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get indexes: %w", err)
	}

	// Get index usage statistics
	usage, err := do.getIndexUsage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get index usage: %w", err)
	}

	// Analyze and generate recommendations
	var recommendations []IndexRecommendation

	// Check for unused indexes
	for _, index := range indexes {
		if usage[index.IndexName] == 0 && !index.IsPrimary {
			recommendations = append(recommendations, IndexRecommendation{
				Type:        "unused_index",
				Severity:    "medium",
				Table:       index.TableName,
				Index:       index.IndexName,
				Description: fmt.Sprintf("Index %s on table %s is not being used", index.IndexName, index.TableName),
				Action:      fmt.Sprintf("Consider dropping index %s", index.IndexName),
			})
		}
	}

	// Check for missing indexes on foreign keys
	missingFKIndexes, err := do.findMissingFKIndexes(ctx)
	if err != nil {
		logger.Warnf("Failed to check for missing FK indexes: %v", err)
	} else {
		recommendations = append(recommendations, missingFKIndexes...)
	}

	// Check for duplicate indexes
	duplicates, err := do.findDuplicateIndexes(ctx)
	if err != nil {
		logger.Warnf("Failed to check for duplicate indexes: %v", err)
	} else {
		recommendations = append(recommendations, duplicates...)
	}

	return recommendations, nil
}

// IndexRecommendation represents a database optimization recommendation
type IndexRecommendation struct {
	Type        string
	Severity    string // low, medium, high, critical
	Table       string
	Index       string
	Column      string
	Description string
	Action      string
	SQL         string
}

// getIndexes retrieves all database indexes
func (do *DatabaseOptimizer) getIndexes(ctx context.Context) ([]IndexInfo, error) {
	query := `
		SELECT 
			t.relname as table_name,
			i.relname as index_name,
			a.attname as column_name,
			am.amname as index_type,
			i.relisunique as is_unique,
			indisprimary as is_primary,
			pg_relation_size(i.oid) as size
		FROM pg_class t
		JOIN pg_index ix ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		JOIN pg_am am ON i.relam = am.oid
		WHERE t.relkind = 'r'
		AND t.relname NOT LIKE 'pg_%'
		ORDER BY t.relname, i.relname
	`

	rows, err := do.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []IndexInfo
	for rows.Next() {
		var index IndexInfo
		err := rows.Scan(
			&index.TableName,
			&index.IndexName,
			&index.ColumnName,
			&index.IndexType,
			&index.IsUnique,
			&index.IsPrimary,
			&index.Size,
		)
		if err != nil {
			return nil, err
		}
		indexes = append(indexes, index)
	}

	return indexes, nil
}

// getIndexUsage retrieves index usage statistics
func (do *DatabaseOptimizer) getIndexUsage(ctx context.Context) (map[string]int64, error) {
	query := `
		SELECT 
			schemaname,
			tablename,
			indexname,
			idx_scan as usage_count
		FROM pg_stat_user_indexes
		WHERE schemaname = 'public'
	`

	rows, err := do.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usage := make(map[string]int64)
	for rows.Next() {
		var schema, table, index string
		var usageCount int64
		err := rows.Scan(&schema, &table, &index, &usageCount)
		if err != nil {
			return nil, err
		}
		usage[index] = usageCount
	}

	return usage, nil
}

// findMissingFKIndexes finds foreign keys without indexes
func (do *DatabaseOptimizer) findMissingFKIndexes(ctx context.Context) ([]IndexRecommendation, error) {
	query := `
		SELECT 
			tc.table_name,
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
		AND tc.table_schema = 'public'
	`

	rows, err := do.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recommendations []IndexRecommendation
	for rows.Next() {
		var table, column, foreignTable, foreignColumn string
		err := rows.Scan(&table, &column, &foreignTable, &foreignColumn)
		if err != nil {
			return nil, err
		}

		// Check if index exists for this foreign key
		indexExists, err := do.checkIndexExists(ctx, table, column)
		if err != nil {
			logger.Warnf("Failed to check index for %s.%s: %v", table, column, err)
			continue
		}

		if !indexExists {
			recommendations = append(recommendations, IndexRecommendation{
				Type:        "missing_fk_index",
				Severity:    "high",
				Table:       table,
				Column:      column,
				Description: fmt.Sprintf("Foreign key %s.%s references %s.%s but has no index", table, column, foreignTable, foreignColumn),
				Action:      fmt.Sprintf("Create index on %s.%s", table, column),
				SQL:         fmt.Sprintf("CREATE INDEX idx_%s_%s ON %s (%s)", table, column, table, column),
			})
		}
	}

	return recommendations, nil
}

// findDuplicateIndexes finds duplicate or redundant indexes
func (do *DatabaseOptimizer) findDuplicateIndexes(ctx context.Context) ([]IndexRecommendation, error) {
	query := `
		SELECT 
			t.relname as table_name,
			i1.relname as index1,
			i2.relname as index2,
			array_agg(a1.attname ORDER BY a1.attnum) as columns1,
			array_agg(a2.attname ORDER BY a2.attnum) as columns2
		FROM pg_class t
		JOIN pg_index ix1 ON t.oid = ix1.indrelid
		JOIN pg_class i1 ON i1.oid = ix1.indexrelid
		JOIN pg_attribute a1 ON a1.attrelid = t.oid AND a1.attnum = ANY(ix1.indkey)
		JOIN pg_index ix2 ON t.oid = ix2.indrelid
		JOIN pg_class i2 ON i2.oid = ix2.indexrelid
		JOIN pg_attribute a2 ON a2.attrelid = t.oid AND a2.attnum = ANY(ix2.indkey)
		WHERE t.relkind = 'r'
		AND t.relname NOT LIKE 'pg_%'
		AND i1.oid < i2.oid
		AND array_agg(a1.attname ORDER BY a1.attnum) = array_agg(a2.attname ORDER BY a2.attnum)
		GROUP BY t.relname, i1.relname, i2.relname
	`

	rows, err := do.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recommendations []IndexRecommendation
	for rows.Next() {
		var table, index1, index2 string
		var columns1, columns2 []string
		err := rows.Scan(&table, &index1, &index2, &columns1, &columns2)
		if err != nil {
			return nil, err
		}

		recommendations = append(recommendations, IndexRecommendation{
			Type:        "duplicate_index",
			Severity:    "medium",
			Table:       table,
			Index:       index1,
			Description: fmt.Sprintf("Indexes %s and %s on table %s have identical columns", index1, index2, table),
			Action:      fmt.Sprintf("Consider dropping one of the duplicate indexes: %s or %s", index1, index2),
		})
	}

	return recommendations, nil
}

// checkIndexExists checks if an index exists for a table column
func (do *DatabaseOptimizer) checkIndexExists(ctx context.Context, table, column string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM pg_indexes
		WHERE tablename = $1
		AND indexdef LIKE '%' || $2 || '%'
	`

	var count int
	err := do.db.QueryRowContext(ctx, query, table, column).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// AnalyzeSlowQueries analyzes slow queries
func (do *DatabaseOptimizer) AnalyzeSlowQueries(ctx context.Context) ([]QueryStats, error) {
	query := `
		SELECT 
			query,
			mean_exec_time as avg_duration_ms,
			calls as total_calls,
			rows as rows_returned,
			last_exec_time as last_executed
		FROM pg_stat_statements
		WHERE mean_exec_time > 1000  -- Queries taking more than 1 second
		ORDER BY mean_exec_time DESC
		LIMIT 20
	`

	rows, err := do.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []QueryStats
	for rows.Next() {
		var stat QueryStats
		var avgDurationMs float64
		var lastExecuted sql.NullTime

		err := rows.Scan(
			&stat.Query,
			&avgDurationMs,
			&stat.TotalCalls,
			&stat.SlowQueries,
			&lastExecuted,
		)
		if err != nil {
			return nil, err
		}

		stat.AvgDuration = time.Duration(avgDurationMs) * time.Millisecond
		if lastExecuted.Valid {
			stat.LastExecuted = lastExecuted.Time
		}

		stats = append(stats, stat)
	}

	return stats, nil
}

// GetTableStats returns table statistics
func (do *DatabaseOptimizer) GetTableStats(ctx context.Context) (map[string]TableStats, error) {
	query := `
		SELECT 
			schemaname,
			tablename,
			n_tup_ins as inserts,
			n_tup_upd as updates,
			n_tup_del as deletes,
			n_live_tup as live_tuples,
			n_dead_tup as dead_tuples,
			last_vacuum,
			last_autovacuum,
			last_analyze,
			last_autoanalyze
		FROM pg_stat_user_tables
		WHERE schemaname = 'public'
	`

	rows, err := do.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]TableStats)
	for rows.Next() {
		var tableName string
		var stat TableStats
		var lastVacuum, lastAutovacuum, lastAnalyze, lastAutoanalyze sql.NullTime

		err := rows.Scan(
			&tableName,
			&stat.Inserts,
			&stat.Updates,
			&stat.Deletes,
			&stat.LiveTuples,
			&stat.DeadTuples,
			&lastVacuum,
			&lastAutovacuum,
			&lastAnalyze,
			&lastAutoanalyze,
		)
		if err != nil {
			return nil, err
		}

		if lastVacuum.Valid {
			stat.LastVacuum = lastVacuum.Time
		}
		if lastAutovacuum.Valid {
			stat.LastAutovacuum = lastAutovacuum.Time
		}
		if lastAnalyze.Valid {
			stat.LastAnalyze = lastAnalyze.Time
		}
		if lastAutoanalyze.Valid {
			stat.LastAutoanalyze = lastAutoanalyze.Time
		}

		stats[tableName] = stat
	}

	return stats, nil
}

// TableStats represents table statistics
type TableStats struct {
	Inserts         int64
	Updates         int64
	Deletes         int64
	LiveTuples      int64
	DeadTuples      int64
	LastVacuum      time.Time
	LastAutovacuum  time.Time
	LastAnalyze     time.Time
	LastAutoanalyze time.Time
}

// OptimizeDatabase runs database optimization
func (do *DatabaseOptimizer) OptimizeDatabase(ctx context.Context) error {
	logger.Info("Starting database optimization...")

	// Analyze tables
	logger.Info("Analyzing tables...")
	_, err := do.db.ExecContext(ctx, "ANALYZE")
	if err != nil {
		return fmt.Errorf("failed to analyze tables: %w", err)
	}

	// Vacuum tables
	logger.Info("Vacuuming tables...")
	_, err = do.db.ExecContext(ctx, "VACUUM")
	if err != nil {
		return fmt.Errorf("failed to vacuum tables: %w", err)
	}

	// Reindex if needed
	logger.Info("Checking for reindex needs...")
	stats, err := do.GetTableStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get table stats: %w", err)
	}

	for tableName, stat := range stats {
		// Reindex if dead tuples are more than 20% of live tuples
		if stat.DeadTuples > 0 && float64(stat.DeadTuples)/float64(stat.LiveTuples) > 0.2 {
			logger.Infof("Reindexing table %s (dead tuples: %d, live tuples: %d)", tableName, stat.DeadTuples, stat.LiveTuples)
			_, err := do.db.ExecContext(ctx, fmt.Sprintf("REINDEX TABLE %s", tableName))
			if err != nil {
				logger.Warnf("Failed to reindex table %s: %v", tableName, err)
			}
		}
	}

	logger.Info("Database optimization completed")
	return nil
}
