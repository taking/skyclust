package common

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceTracker tracks performance metrics for requests
type PerformanceTracker struct {
	startTime time.Time
	operation string
	context   context.Context
}

// NewPerformanceTracker creates a new performance tracker
func NewPerformanceTracker(operation string) *PerformanceTracker {
	return &PerformanceTracker{
		startTime: time.Now(),
		operation: operation,
		context:   context.Background(),
	}
}

// WithContext sets the context for the performance tracker
func (pt *PerformanceTracker) WithContext(ctx context.Context) *PerformanceTracker {
	pt.context = ctx
	return pt
}

// Finish completes the performance tracking and returns the duration
func (pt *PerformanceTracker) Finish() time.Duration {
	return time.Since(pt.startTime)
}

// TrackRequest tracks the performance of an HTTP request
func (pt *PerformanceTracker) TrackRequest(c *gin.Context, userID string, statusCode int) {
	duration := pt.Finish()

	// Log performance metrics
	LogBusinessEvent(c, pt.operation, userID, "", map[string]interface{}{
		"duration":    duration.String(),
		"status_code": statusCode,
	})
}

// CacheConfig defines cache configuration
type CacheConfig struct {
	TTL     time.Duration
	MaxSize int
	Enabled bool
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		TTL:     5 * time.Minute,
		MaxSize: 1000,
		Enabled: true,
	}
}

// PaginationOptimizer optimizes pagination queries
type PaginationOptimizer struct {
	DefaultLimit int
	MaxLimit     int
}

// NewPaginationOptimizer creates a new pagination optimizer
func NewPaginationOptimizer() *PaginationOptimizer {
	return &PaginationOptimizer{
		DefaultLimit: 10,
		MaxLimit:     100,
	}
}

// OptimizePagination optimizes pagination parameters
func (po *PaginationOptimizer) OptimizePagination(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = po.DefaultLimit
	}
	if limit > po.MaxLimit {
		limit = po.MaxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// MemoryOptimizer provides memory optimization utilities
type MemoryOptimizer struct {
	MaxMemoryUsage int64 // in bytes
	GCThreshold    float64
}

// NewMemoryOptimizer creates a new memory optimizer
func NewMemoryOptimizer() *MemoryOptimizer {
	return &MemoryOptimizer{
		MaxMemoryUsage: 100 * 1024 * 1024, // 100MB
		GCThreshold:    0.8,               // 80%
	}
}

// CheckMemoryUsage checks if memory usage is within limits
func (mo *MemoryOptimizer) CheckMemoryUsage() bool {
	// This would typically use runtime.MemStats
	// For now, return true
	return true
}

// OptimizeMemory optimizes memory usage
func (mo *MemoryOptimizer) OptimizeMemory() {
	// Add memory optimization logic here
	// This could include garbage collection hints, etc.
}

// ResponseOptimizer optimizes HTTP responses
type ResponseOptimizer struct {
	CompressionEnabled bool
	MaxResponseSize    int64
}

// NewResponseOptimizer creates a new response optimizer
func NewResponseOptimizer() *ResponseOptimizer {
	return &ResponseOptimizer{
		CompressionEnabled: true,
		MaxResponseSize:    10 * 1024 * 1024, // 10MB
	}
}

// OptimizeResponse optimizes HTTP responses
func (ro *ResponseOptimizer) OptimizeResponse(c *gin.Context, data interface{}) {
	// Add response optimization logic here
	// This could include compression, caching headers, etc.

	if ro.CompressionEnabled {
		c.Header("Content-Encoding", "gzip")
	}

	// Set cache headers
	c.Header("Cache-Control", "public, max-age=300")
	c.Header("ETag", generateETag(data))
}

// generateETag generates an ETag for the response
func generateETag(data interface{}) string {
	// Simple ETag generation - in production, use proper hashing
	return "etag-" + time.Now().Format("20060102150405")
}
