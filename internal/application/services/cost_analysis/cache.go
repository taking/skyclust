package cost_analysis

import (
	"context"
	"fmt"
	"time"

	"skyclust/pkg/logger"
)

// Cache key helpers
const (
	cachePrefixCostSummary     = "cost_summary"
	cachePrefixCostPredictions = "cost_predictions"
	cachePrefixCostBreakdown   = "cost_breakdown"
)

// buildCostSummaryCacheKey builds a cache key for cost summary
// Format: cost_summary:{workspace_id}:{period}:{resource_types}
func buildCostSummaryCacheKey(workspaceID, period, resourceTypes string) string {
	return fmt.Sprintf("%s:%s:%s:%s", cachePrefixCostSummary, workspaceID, period, resourceTypes)
}

// buildCostPredictionsCacheKey builds a cache key for cost predictions
// Format: cost_predictions:{workspace_id}:{days}:{resource_types}
func buildCostPredictionsCacheKey(workspaceID string, days int, resourceTypes string) string {
	return fmt.Sprintf("%s:%s:%d:%s", cachePrefixCostPredictions, workspaceID, days, resourceTypes)
}

// buildCostBreakdownCacheKey builds a cache key for cost breakdown
// Format: cost_breakdown:{workspace_id}:{period}:{dimension}:{resource_types}
func buildCostBreakdownCacheKey(workspaceID, period, dimension, resourceTypes string) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s", cachePrefixCostBreakdown, workspaceID, period, dimension, resourceTypes)
}

// getFromCache retrieves value from cache if available
func (s *Service) getFromCache(ctx context.Context, cacheKey string, result interface{}) (bool, error) {
	if s.cache == nil {
		return false, nil
	}

	err := s.cache.Get(ctx, cacheKey, result)
	if err == nil {
		return true, nil
	}

	return false, nil
}

// setCache sets value to cache (non-blocking)
func (s *Service) setCache(ctx context.Context, cacheKey string, value interface{}, ttl time.Duration) {
	if s.cache == nil {
		return
	}

	if err := s.cache.Set(ctx, cacheKey, value, ttl); err != nil {
		logger.Warnf("Failed to cache %s: %v", cacheKey, err)
	}
}
