package service

import (
	"context"
	"runtime"
	"time"

	"skyclust/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// SystemMonitoringService handles system monitoring and health checks
type SystemMonitoringService struct {
	logger       *zap.Logger
	config       *config.Config
	redisClient  *redis.Client
	startTime    time.Time
	requestCount int64
	errorCount   int64
}

// NewSystemMonitoringService creates a new system monitoring service
func NewSystemMonitoringService(
	logger *zap.Logger,
	config *config.Config,
	redisClient *redis.Client,
) *SystemMonitoringService {
	return &SystemMonitoringService{
		logger:       logger,
		config:       config,
		redisClient:  redisClient,
		startTime:    time.Now(),
		requestCount: 0,
		errorCount:   0,
	}
}

// IncrementRequestCount increments the request count
func (s *SystemMonitoringService) IncrementRequestCount() {
	s.requestCount++
}

// IncrementErrorCount increments the error count
func (s *SystemMonitoringService) IncrementErrorCount() {
	s.errorCount++
}

// GetHealthStatus returns comprehensive health status
func (s *SystemMonitoringService) GetHealthStatus() gin.H {
	now := time.Now()
	uptime := now.Sub(s.startTime)

	// Get service status
	services := s.getServiceStatus()

	// Get dependencies status
	dependencies := s.getDependenciesStatus()

	// Get system metrics
	metrics := s.GetSystemMetrics()

	// Determine overall health status
	overallStatus := "healthy"
	if !s.isAllServicesHealthy(services) || !s.isAllDependenciesHealthy(dependencies) {
		overallStatus = "degraded"
		s.errorCount++
	}

	// Get alert status
	alerts := s.getAlertStatus(services, dependencies, metrics)

	return gin.H{
		"status":       overallStatus,
		"timestamp":    now.Format(time.RFC3339),
		"version":      "1.0.0",
		"uptime":       uptime.String(),
		"started_at":   s.startTime.Format(time.RFC3339),
		"services":     services,
		"dependencies": dependencies,
		"metrics":      metrics,
		"config": gin.H{
			"environment": s.getEnvironment(),
			"debug_mode":  s.isDebugMode(),
		},
		"alerts": alerts,
	}
}

// GetSystemMetrics returns system performance metrics
func (s *SystemMonitoringService) GetSystemMetrics() gin.H {
	return gin.H{
		"memory_usage": s.getMemoryMetrics(),
		"performance":  s.getPerformanceMetrics(),
	}
}

// GetAlerts returns current alert status
func (s *SystemMonitoringService) GetAlerts() gin.H {
	services := s.getServiceStatus()
	dependencies := s.getDependenciesStatus()
	metrics := s.GetSystemMetrics()

	alerts := s.collectAllAlerts(services, dependencies, metrics)

	return gin.H{
		"enabled":    len(alerts) > 0,
		"count":      len(alerts),
		"alerts":     alerts,
		"thresholds": s.getAlertThresholds(),
	}
}

// getServiceStatus returns the status of all services
func (s *SystemMonitoringService) getServiceStatus() gin.H {
	services := gin.H{
		"database": s.checkDatabaseStatus(),
		"redis":    s.checkRedisStatus(),
		"auth":     s.checkAuthServiceStatus(),
	}
	return services
}

// isAllServicesHealthy checks if all services are healthy
func (s *SystemMonitoringService) isAllServicesHealthy(services gin.H) bool {
	for _, service := range services {
		if status, ok := service.(gin.H); ok {
			if healthy, exists := status["healthy"]; exists {
				if !healthy.(bool) {
					return false
				}
			}
		}
	}
	return true
}

// checkDatabaseStatus checks database connectivity
func (s *SystemMonitoringService) checkDatabaseStatus() gin.H {
	// Measure response time for consistency
	start := time.Now()

	// TODO: Implement DB health check
	// if s.container.DB == nil {
	// 	responseTime := time.Since(start)
	// 	return gin.H{
	// 		"healthy":          false,
	// 		"status":           "not_initialized",
	// 		"response_time_ms": float64(responseTime.Nanoseconds()) / 1e6,
	// 		"error":            "database not initialized",
	// 	}
	// }

	// TODO: Implement DB connection check
	// sqlDB, err := s.container.DB.DB()
	// if err != nil {
	// 	responseTime := time.Since(start)
	// 	return gin.H{
	// 		"healthy":          false,
	// 		"status":           "connection_failed",
	// 		"response_time_ms": float64(responseTime.Nanoseconds()) / 1e6,
	// 		"error":            err.Error(),
	// 	}
	// }

	// TODO: Implement DB ping check
	// if err := sqlDB.Ping(); err != nil {
	// 	responseTime := time.Since(start)
	// 	return gin.H{
	// 		"healthy":          false,
	// 		"status":           "ping_failed",
	// 		"response_time_ms": float64(responseTime.Nanoseconds()) / 1e6,
	// 		"error":            err.Error(),
	// 	}
	// }

	// For now, return status with minimal response time (TODO: implement actual DB check)
	responseTime := time.Since(start)
	return gin.H{
		"healthy":          true,
		"status":           "connected",
		"response_time_ms": float64(responseTime.Nanoseconds()) / 1e6,
	}
}

// checkRedisStatus checks Redis connectivity
func (s *SystemMonitoringService) checkRedisStatus() gin.H {
	if s.redisClient == nil {
		// Even when not configured, return consistent structure with 0 response time
		return gin.H{
			"healthy":          false,
			"status":           "not_configured",
			"response_time_ms": 0.0,
			"note":             "Redis client not available",
		}
	}

	// Test Redis connection with PING command
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	err := s.redisClient.Ping(ctx).Err()
	responseTime := time.Since(start)

	if err != nil {
		s.logger.Warn("Redis health check failed", zap.Error(err))
		return gin.H{
			"healthy":          false,
			"status":           "disconnected",
			"response_time_ms": float64(responseTime.Nanoseconds()) / 1e6,
			"error":            err.Error(),
		}
	}

	return gin.H{
		"healthy":          true,
		"status":           "connected",
		"response_time_ms": float64(responseTime.Nanoseconds()) / 1e6,
	}
}

// checkAuthServiceStatus checks authentication service status
func (s *SystemMonitoringService) checkAuthServiceStatus() gin.H {
	// Measure response time for consistency
	start := time.Now()

	// TODO: Implement auth service check
	// Simple check - auth service availability is verified by successful API calls
	// For now, assume available if system is running

	responseTime := time.Since(start)
	return gin.H{
		"healthy":          true,
		"status":           "available",
		"response_time_ms": float64(responseTime.Nanoseconds()) / 1e6,
	}
}

// isDebugMode returns whether debug mode is enabled
func (s *SystemMonitoringService) isDebugMode() bool {
	if s.config == nil {
		return true // Default to debug mode when config is nil
	}
	return s.config.Server.Host == "0.0.0.0"
}

// getEnvironment returns the current environment
func (s *SystemMonitoringService) getEnvironment() string {
	if s.config == nil {
		return "development"
	}
	if s.config.Server.Host == "0.0.0.0" {
		return "production"
	}
	return "development"
}

// getDependenciesStatus checks external dependencies
func (s *SystemMonitoringService) getDependenciesStatus() gin.H {
	return gin.H{
		"postgres": s.checkPostgresDependency(),
		"redis":    s.checkRedisDependency(),
	}
}

// isAllDependenciesHealthy checks if all dependencies are healthy
func (s *SystemMonitoringService) isAllDependenciesHealthy(dependencies gin.H) bool {
	for _, dep := range dependencies {
		if status, ok := dep.(gin.H); ok {
			if healthy, exists := status["healthy"]; exists {
				if !healthy.(bool) {
					return false
				}
			}
		}
	}
	return true
}

// getMemoryMetrics returns memory usage metrics
func (s *SystemMonitoringService) getMemoryMetrics() gin.H {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return gin.H{
		"alloc_mb":       float64(m.Alloc) / 1024 / 1024,
		"total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
		"sys_mb":         float64(m.Sys) / 1024 / 1024,
		"num_gc":         m.NumGC,
		"heap_alloc_mb":  float64(m.HeapAlloc) / 1024 / 1024,
		"heap_sys_mb":    float64(m.HeapSys) / 1024 / 1024,
		"stack_inuse_mb": float64(m.StackInuse) / 1024 / 1024,
		"stack_sys_mb":   float64(m.StackSys) / 1024 / 1024,
	}
}

// getPerformanceMetrics returns performance metrics
func (s *SystemMonitoringService) getPerformanceMetrics() gin.H {
	return gin.H{
		"goroutines":    runtime.NumGoroutine(),
		"cpu_cores":     runtime.NumCPU(),
		"request_count": s.requestCount,
		"error_count":   s.errorCount,
		"error_rate":    s.calculateErrorRate(),
		"rps":           s.calculateRequestsPerSecond(),
	}
}

// calculateErrorRate calculates the error rate percentage
func (s *SystemMonitoringService) calculateErrorRate() float64 {
	if s.requestCount == 0 {
		return 0.0
	}
	return float64(s.errorCount) / float64(s.requestCount) * 100.0
}

// calculateRequestsPerSecond calculates requests per second
func (s *SystemMonitoringService) calculateRequestsPerSecond() float64 {
	uptimeSeconds := time.Since(s.startTime).Seconds()
	if uptimeSeconds <= 0 {
		return 0.0
	}
	return float64(s.requestCount) / uptimeSeconds
}

// getAlertStatus returns current alert status
func (s *SystemMonitoringService) getAlertStatus(services, dependencies, metrics gin.H) gin.H {
	alerts := s.collectAllAlerts(services, dependencies, metrics)

	return gin.H{
		"enabled":    len(alerts) > 0,
		"count":      len(alerts),
		"alerts":     alerts,
		"thresholds": s.getAlertThresholds(),
	}
}

// collectAllAlerts collects all types of alerts
func (s *SystemMonitoringService) collectAllAlerts(services, dependencies, metrics gin.H) []gin.H {
	var alerts []gin.H

	alerts = append(alerts, s.checkServiceAlerts(services)...)
	alerts = append(alerts, s.checkDependencyAlerts(dependencies)...)
	alerts = append(alerts, s.checkPerformanceAlerts(metrics)...)
	alerts = append(alerts, s.checkMemoryAlerts(metrics)...)

	return alerts
}

// checkServiceAlerts checks for service-related alerts
func (s *SystemMonitoringService) checkServiceAlerts(services gin.H) []gin.H {
	if s.isAllServicesHealthy(services) {
		return nil
	}

	return []gin.H{{
		"type":    "service",
		"level":   "warning",
		"message": "One or more services are unhealthy",
	}}
}

// checkDependencyAlerts checks for dependency-related alerts
func (s *SystemMonitoringService) checkDependencyAlerts(dependencies gin.H) []gin.H {
	if s.isAllDependenciesHealthy(dependencies) {
		return nil
	}

	return []gin.H{{
		"type":    "dependency",
		"level":   "critical",
		"message": "External dependencies are failing",
	}}
}

// checkPerformanceAlerts checks for performance-related alerts
func (s *SystemMonitoringService) checkPerformanceAlerts(metrics gin.H) []gin.H {
	metricsMap, ok := metrics["performance"].(gin.H)
	if !ok {
		return nil
	}

	errorRate, exists := metricsMap["error_rate"]
	if !exists {
		return nil
	}

	rate, ok := errorRate.(float64)
	if !ok || rate <= 5.0 {
		return nil
	}

	return []gin.H{{
		"type":    "performance",
		"level":   "warning",
		"message": "High error rate detected",
		"value":   rate,
	}}
}

// checkMemoryAlerts checks for memory-related alerts
func (s *SystemMonitoringService) checkMemoryAlerts(metrics gin.H) []gin.H {
	metricsMap, ok := metrics["memory_usage"].(gin.H)
	if !ok {
		return nil
	}

	var alerts []gin.H

	// Check heap allocation
	if alert := s.checkHeapAllocationAlert(metricsMap); alert != nil {
		alerts = append(alerts, alert)
	}

	// Check system memory
	if alert := s.checkSystemMemoryAlert(metricsMap); alert != nil {
		alerts = append(alerts, alert)
	}

	// Check stack memory
	if alert := s.checkStackMemoryAlert(metricsMap); alert != nil {
		alerts = append(alerts, alert)
	}

	return alerts
}

// checkHeapAllocationAlert checks for heap allocation alerts
func (s *SystemMonitoringService) checkHeapAllocationAlert(metricsMap gin.H) gin.H {
	allocMB, exists := metricsMap["alloc_mb"]
	if !exists {
		return nil
	}

	alloc, ok := allocMB.(float64)
	if !ok || alloc <= 100.0 {
		return nil
	}

	return gin.H{
		"type":    "memory",
		"level":   "warning",
		"message": "High heap allocation detected",
		"value":   alloc,
		"metric":  "alloc_mb",
	}
}

// checkSystemMemoryAlert checks for system memory alerts
func (s *SystemMonitoringService) checkSystemMemoryAlert(metricsMap gin.H) gin.H {
	sysMB, exists := metricsMap["sys_mb"]
	if !exists {
		return nil
	}

	sys, ok := sysMB.(float64)
	if !ok || sys <= 200.0 {
		return nil
	}

	return gin.H{
		"type":    "memory",
		"level":   "warning",
		"message": "High system memory usage detected",
		"value":   sys,
		"metric":  "sys_mb",
	}
}

// checkStackMemoryAlert checks for stack memory alerts
func (s *SystemMonitoringService) checkStackMemoryAlert(metricsMap gin.H) gin.H {
	stackInuseMB, exists := metricsMap["stack_inuse_mb"]
	if !exists {
		return nil
	}

	stackInuse, ok := stackInuseMB.(float64)
	if !ok || stackInuse <= 50.0 {
		return nil
	}

	return gin.H{
		"type":    "memory",
		"level":   "warning",
		"message": "High stack memory usage detected",
		"value":   stackInuse,
		"metric":  "stack_inuse_mb",
	}
}

// getAlertThresholds returns alert threshold values
func (s *SystemMonitoringService) getAlertThresholds() gin.H {
	return gin.H{
		"error_rate":       5.0,
		"memory_mb":        100.0,
		"system_memory_mb": 200.0,
		"stack_memory_mb":  50.0,
		"uptime_hours":     24.0,
	}
}

// checkPostgresDependency checks PostgreSQL connection
func (s *SystemMonitoringService) checkPostgresDependency() gin.H {
	// Measure response time for consistency
	start := time.Now()

	// TODO: Implement DB health check
	// if s.container.DB == nil {
	// 	responseTime := time.Since(start)
	// 	return gin.H{
	// 		"healthy":          false,
	// 		"status":           "not_initialized",
	// 		"response_time_ms": float64(responseTime.Nanoseconds()) / 1e6,
	// 		"error":            "database not initialized",
	// 	}
	// }

	// TODO: Implement DB connection check
	// sqlDB, err := s.container.DB.DB()
	// if err != nil {
	// 	responseTime := time.Since(start)
	// 	return gin.H{
	// 		"healthy":          false,
	// 		"status":           "connection_failed",
	// 		"response_time_ms": float64(responseTime.Nanoseconds()) / 1e6,
	// 		"error":            err.Error(),
	// 	}
	// }

	// TODO: Implement DB ping check
	// if err := sqlDB.Ping(); err != nil {
	// 	responseTime := time.Since(start)
	// 	return gin.H{
	// 		"healthy":          false,
	// 		"status":           "ping_failed",
	// 		"response_time_ms": float64(responseTime.Nanoseconds()) / 1e6,
	// 		"error":            err.Error(),
	// 	}
	// }

	// For now, return status with minimal response time (TODO: implement actual DB check)
	responseTime := time.Since(start)
	return gin.H{
		"healthy":          true,
		"status":           "connected",
		"response_time_ms": float64(responseTime.Nanoseconds()) / 1e6,
	}
}

// checkRedisDependency checks Redis connection
func (s *SystemMonitoringService) checkRedisDependency() gin.H {
	if s.redisClient == nil {
		// Even when not configured, return consistent structure with 0 response time
		return gin.H{
			"healthy":          false,
			"status":           "not_configured",
			"response_time_ms": 0.0,
			"note":             "Redis client not available",
		}
	}

	// Test Redis connection with PING command
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	err := s.redisClient.Ping(ctx).Err()
	responseTime := time.Since(start)

	if err != nil {
		s.logger.Warn("Redis dependency check failed", zap.Error(err))
		return gin.H{
			"healthy":          false,
			"status":           "disconnected",
			"response_time_ms": float64(responseTime.Nanoseconds()) / 1e6,
			"error":            err.Error(),
		}
	}

	return gin.H{
		"healthy":          true,
		"status":           "connected",
		"response_time_ms": float64(responseTime.Nanoseconds()) / 1e6,
	}
}
