package monitoring

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Metrics holds all application metrics
type Metrics struct {
	requestDuration metric.Float64Histogram
	requestCount    metric.Int64Counter
	errorCount      metric.Int64Counter
	activeUsers     metric.Int64UpDownCounter
	dbConnections   metric.Int64Gauge
	cacheHits       metric.Int64Counter
	cacheMisses     metric.Int64Counter
}

// NewMetrics creates a new metrics instance
func NewMetrics() (*Metrics, error) {
	meter := otel.Meter("cmp-server")

	requestDuration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	requestCount, err := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		return nil, err
	}

	errorCount, err := meter.Int64Counter(
		"http_errors_total",
		metric.WithDescription("Total number of HTTP errors"),
	)
	if err != nil {
		return nil, err
	}

	activeUsers, err := meter.Int64UpDownCounter(
		"active_users",
		metric.WithDescription("Number of active users"),
	)
	if err != nil {
		return nil, err
	}

	dbConnections, err := meter.Int64Gauge(
		"database_connections",
		metric.WithDescription("Number of database connections"),
	)
	if err != nil {
		return nil, err
	}

	cacheHits, err := meter.Int64Counter(
		"cache_hits_total",
		metric.WithDescription("Total number of cache hits"),
	)
	if err != nil {
		return nil, err
	}

	cacheMisses, err := meter.Int64Counter(
		"cache_misses_total",
		metric.WithDescription("Total number of cache misses"),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		requestDuration: requestDuration,
		requestCount:    requestCount,
		errorCount:      errorCount,
		activeUsers:     activeUsers,
		dbConnections:   dbConnections,
		cacheHits:       cacheHits,
		cacheMisses:     cacheMisses,
	}, nil
}

// RecordRequestDuration records HTTP request duration
func (m *Metrics) RecordRequestDuration(method, path string, statusCode int, duration time.Duration) {
	m.requestDuration.Record(
		context.Background(),
		duration.Seconds(),
		metric.WithAttributes(
			attribute.String("method", method),
			attribute.String("path", path),
			attribute.Int("status_code", statusCode),
		),
	)
}

// IncrementRequestCount increments HTTP request count
func (m *Metrics) IncrementRequestCount(method, path string, statusCode int) {
	m.requestCount.Add(
		context.Background(),
		1,
		metric.WithAttributes(
			attribute.String("method", method),
			attribute.String("path", path),
			attribute.Int("status_code", statusCode),
		),
	)
}

// IncrementErrorCount increments HTTP error count
func (m *Metrics) IncrementErrorCount(method, path string, errorType string) {
	m.errorCount.Add(
		context.Background(),
		1,
		metric.WithAttributes(
			attribute.String("method", method),
			attribute.String("path", path),
			attribute.String("error_type", errorType),
		),
	)
}

// SetActiveUsers sets the number of active users
func (m *Metrics) SetActiveUsers(count int64) {
	m.activeUsers.Add(
		context.Background(),
		count,
		metric.WithAttributes(
			attribute.String("status", "active"),
		),
	)
}

// SetDBConnections sets the number of database connections
func (m *Metrics) SetDBConnections(count int64) {
	m.dbConnections.Record(
		context.Background(),
		count,
		metric.WithAttributes(
			attribute.String("state", "open"),
		),
	)
}

// IncrementCacheHits increments cache hits
func (m *Metrics) IncrementCacheHits(cacheType string) {
	m.cacheHits.Add(
		context.Background(),
		1,
		metric.WithAttributes(
			attribute.String("cache_type", cacheType),
		),
	)
}

// IncrementCacheMisses increments cache misses
func (m *Metrics) IncrementCacheMisses(cacheType string) {
	m.cacheMisses.Add(
		context.Background(),
		1,
		metric.WithAttributes(
			attribute.String("cache_type", cacheType),
		),
	)
}
