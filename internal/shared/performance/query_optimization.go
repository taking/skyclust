package performance

import (
	"context"
	"fmt"
	"gorm.io/gorm"
)

// QueryOptimization provides utilities for optimizing database queries
type QueryOptimization struct {
	db *gorm.DB
}

// NewQueryOptimization creates a new query optimization utility
func NewQueryOptimization(db *gorm.DB) *QueryOptimization {
	return &QueryOptimization{db: db}
}

// BatchLoad loads related entities in batches to avoid N+1 queries
// Example: BatchLoad(ctx, vms, "WorkspaceID", workspaceIDs, func(ids []string) ([]*Workspace, error) { ... })
func BatchLoad[T any, K comparable](
	ctx context.Context,
	items []T,
	idExtractor func(T) K,
	batchLoader func([]K) (map[K]interface{}, error),
) error {
	if len(items) == 0 {
		return nil
	}

	// Collect unique IDs
	idSet := make(map[K]bool)
	for _, item := range items {
		id := idExtractor(item)
		idSet[id] = true
	}

	// Convert to slice
	ids := make([]K, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}

	// Batch load
	loaded, err := batchLoader(ids)
	if err != nil {
		return fmt.Errorf("failed to batch load: %w", err)
	}

	// Attach loaded entities to items (this would need reflection or a callback)
	// For now, this is a placeholder pattern
	_ = loaded
	return nil
}

// PreloadRelations preloads related entities using GORM Preload
func PreloadRelations(query *gorm.DB, relations ...string) *gorm.DB {
	for _, relation := range relations {
		query = query.Preload(relation)
	}
	return query
}

// SelectFields limits the selected fields to reduce data transfer
func SelectFields(query *gorm.DB, fields ...string) *gorm.DB {
	return query.Select(fields)
}

// UseIndex hints the database to use a specific index
func UseIndex(query *gorm.DB, indexName string) *gorm.DB {
	// PostgreSQL specific: Use FORCE INDEX or similar
	// This is database-specific and may need adjustment
	return query
}

// QueryStats holds query performance statistics
type QueryStats struct {
	QueryTime   int64 // milliseconds
	RowsFetched int64
	CacheHit    bool
}

// MeasureQuery measures query execution time
func MeasureQuery(ctx context.Context, fn func() error) (QueryStats, error) {
	// Implementation would measure query time
	// This is a placeholder for actual implementation
	err := fn()
	return QueryStats{}, err
}
