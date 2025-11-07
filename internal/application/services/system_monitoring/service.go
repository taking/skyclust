package system_monitoring

import (
	"context"
	"runtime"
	"time"

	serviceconstants "skyclust/internal/application/services"
	"skyclust/pkg/cache"
	"skyclust/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Service: 시스템 모니터링 및 헬스 체크를 처리하는 서비스
type Service struct {
	logger       *zap.Logger
	config       *config.Config
	cache        cache.Cache
	startTime    time.Time
	requestCount int64
	errorCount   int64
}

// NewService: 새로운 시스템 모니터링 서비스를 생성합니다
func NewService(
	logger *zap.Logger,
	config *config.Config,
	cache cache.Cache,
) *Service {
	return &Service{
		logger:       logger,
		config:       config,
		cache:        cache,
		startTime:    time.Now(),
		requestCount: 0,
		errorCount:   0,
	}
}

// IncrementRequestCount: 요청 카운트를 증가시킵니다
func (s *Service) IncrementRequestCount() {
	s.requestCount++
}

// IncrementErrorCount: 에러 카운트를 증가시킵니다
func (s *Service) IncrementErrorCount() {
	s.errorCount++
}

// GetHealthStatus: 포괄적인 헬스 상태를 반환합니다
func (s *Service) GetHealthStatus() gin.H {
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

// GetSystemMetrics: 시스템 성능 메트릭을 반환합니다
func (s *Service) GetSystemMetrics() gin.H {
	return gin.H{
		"memory_usage": s.getMemoryMetrics(),
		"performance":  s.getPerformanceMetrics(),
	}
}

// GetAlerts: 현재 알림 상태를 반환합니다
func (s *Service) GetAlerts() gin.H {
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

// getServiceStatus: 서비스 상태를 조회합니다
func (s *Service) getServiceStatus() gin.H {
	services := gin.H{
		"database": s.checkDatabaseStatus(),
		"redis":    s.checkRedisStatus(),
		"auth":     s.checkAuthServiceStatus(),
	}
	return services
}

// isAllServicesHealthy: 모든 서비스가 정상인지 확인합니다
func (s *Service) isAllServicesHealthy(services gin.H) bool {
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

// checkDatabaseStatus: 데이터베이스 연결 상태를 확인합니다
func (s *Service) checkDatabaseStatus() gin.H {
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

// checkRedisStatus: Redis 연결 상태를 확인합니다
func (s *Service) checkRedisStatus() gin.H {
	if s.cache == nil {
		// Even when not configured, return consistent structure with 0 response time
		return gin.H{
			"healthy":          false,
			"status":           "not_configured",
			"response_time_ms": 0.0,
			"note":             "Cache not available",
		}
	}

	// Try to get Redis client if cache is RedisService (for health check)
	var redisClient *redis.Client
	if redisService, ok := s.cache.(*cache.RedisService); ok {
		redisClient = redisService.GetClient()
	}

	if redisClient == nil {
		// For non-Redis cache (e.g., MemoryCache), use Exists as health check
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		start := time.Now()
		_, err := s.cache.Exists(ctx, "health_check")
		responseTime := time.Since(start)

		if err != nil {
			s.logger.Warn("Cache health check failed", zap.Error(err))
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

	// For Redis, use PING command for health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	err := redisClient.Ping(ctx).Err()
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

// checkAuthServiceStatus: 인증 서비스 상태를 확인합니다
func (s *Service) checkAuthServiceStatus() gin.H {
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

// isDebugMode: 디버그 모드가 활성화되어 있는지 반환합니다
func (s *Service) isDebugMode() bool {
	if s.config == nil {
		return true // Default to debug mode when config is nil
	}
	return s.config.Server.Host == "0.0.0.0"
}

// getEnvironment: 현재 환경을 반환합니다
func (s *Service) getEnvironment() string {
	if s.config == nil {
		return "development"
	}
	if s.config.Server.Host == "0.0.0.0" {
		return "production"
	}
	return "development"
}

// getDependenciesStatus: 외부 의존성 상태를 확인합니다
func (s *Service) getDependenciesStatus() gin.H {
	return gin.H{
		"postgres": s.checkPostgresDependency(),
		"redis":    s.checkRedisDependency(),
	}
}

// isAllDependenciesHealthy: 모든 의존성이 정상인지 확인합니다
func (s *Service) isAllDependenciesHealthy(dependencies gin.H) bool {
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

// getMemoryMetrics: 메모리 사용량 메트릭을 반환합니다
func (s *Service) getMemoryMetrics() gin.H {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return gin.H{
		"alloc_mb":       float64(m.Alloc) / serviceconstants.BytesPerMB,
		"total_alloc_mb": float64(m.TotalAlloc) / serviceconstants.BytesPerMB,
		"sys_mb":         float64(m.Sys) / serviceconstants.BytesPerMB,
		"num_gc":         m.NumGC,
		"heap_alloc_mb":  float64(m.HeapAlloc) / serviceconstants.BytesPerMB,
		"heap_sys_mb":    float64(m.HeapSys) / serviceconstants.BytesPerMB,
		"stack_inuse_mb": float64(m.StackInuse) / serviceconstants.BytesPerMB,
		"stack_sys_mb":   float64(m.StackSys) / serviceconstants.BytesPerMB,
	}
}

// getPerformanceMetrics: 성능 메트릭을 반환합니다
func (s *Service) getPerformanceMetrics() gin.H {
	return gin.H{
		"goroutines":    runtime.NumGoroutine(),
		"cpu_cores":     runtime.NumCPU(),
		"request_count": s.requestCount,
		"error_count":   s.errorCount,
		"error_rate":    s.calculateErrorRate(),
		"rps":           s.calculateRequestsPerSecond(),
	}
}

// calculateErrorRate: 에러율을 계산합니다
func (s *Service) calculateErrorRate() float64 {
	if s.requestCount == 0 {
		return 0.0
	}
	return float64(s.errorCount) / float64(s.requestCount) * serviceconstants.PercentageBase
}

// calculateRequestsPerSecond: 초당 요청 수를 계산합니다
func (s *Service) calculateRequestsPerSecond() float64 {
	uptimeSeconds := time.Since(s.startTime).Seconds()
	if uptimeSeconds <= 0 {
		return 0.0
	}
	return float64(s.requestCount) / uptimeSeconds
}

// getAlertStatus: 알림 상태를 반환합니다
func (s *Service) getAlertStatus(services, dependencies, metrics gin.H) gin.H {
	alerts := s.collectAllAlerts(services, dependencies, metrics)

	return gin.H{
		"enabled":    len(alerts) > 0,
		"count":      len(alerts),
		"alerts":     alerts,
		"thresholds": s.getAlertThresholds(),
	}
}

// collectAllAlerts: 모든 알림을 수집합니다
func (s *Service) collectAllAlerts(services, dependencies, metrics gin.H) []gin.H {
	var alerts []gin.H

	alerts = append(alerts, s.checkServiceAlerts(services)...)
	alerts = append(alerts, s.checkDependencyAlerts(dependencies)...)
	alerts = append(alerts, s.checkPerformanceAlerts(metrics)...)
	alerts = append(alerts, s.checkMemoryAlerts(metrics)...)

	return alerts
}

// checkServiceAlerts: 서비스 알림을 확인합니다
func (s *Service) checkServiceAlerts(services gin.H) []gin.H {
	if s.isAllServicesHealthy(services) {
		return nil
	}

	return []gin.H{{
		"type":    "service",
		"level":   "warning",
		"message": "One or more services are unhealthy",
	}}
}

// checkDependencyAlerts: 의존성 알림을 확인합니다
func (s *Service) checkDependencyAlerts(dependencies gin.H) []gin.H {
	if s.isAllDependenciesHealthy(dependencies) {
		return nil
	}

	return []gin.H{{
		"type":    "dependency",
		"level":   "critical",
		"message": "External dependencies are failing",
	}}
}

// checkPerformanceAlerts: 성능 알림을 확인합니다
func (s *Service) checkPerformanceAlerts(metrics gin.H) []gin.H {
	metricsMap, ok := metrics["performance"].(gin.H)
	if !ok {
		return nil
	}

	errorRate, exists := metricsMap["error_rate"]
	if !exists {
		return nil
	}

	rate, ok := errorRate.(float64)
	if !ok || rate <= serviceconstants.HighErrorRateThreshold {
		return nil
	}

	return []gin.H{{
		"type":    "performance",
		"level":   "warning",
		"message": "High error rate detected",
		"value":   rate,
	}}
}

// checkMemoryAlerts: 메모리 알림을 확인합니다
func (s *Service) checkMemoryAlerts(metrics gin.H) []gin.H {
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

// checkHeapAllocationAlert: 힙 할당 알림을 확인합니다
func (s *Service) checkHeapAllocationAlert(metricsMap gin.H) gin.H {
	allocMB, exists := metricsMap["alloc_mb"]
	if !exists {
		return nil
	}

	alloc, ok := allocMB.(float64)
	if !ok || alloc <= serviceconstants.MemoryWarningThreshold {
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

// checkSystemMemoryAlert: 시스템 메모리 알림을 확인합니다
func (s *Service) checkSystemMemoryAlert(metricsMap gin.H) gin.H {
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

// checkStackMemoryAlert: 스택 메모리 알림을 확인합니다
func (s *Service) checkStackMemoryAlert(metricsMap gin.H) gin.H {
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

// getAlertThresholds: 알림 임계값을 반환합니다
func (s *Service) getAlertThresholds() gin.H {
	return gin.H{
		"error_rate":       serviceconstants.HighErrorRateThreshold,
		"memory_mb":        serviceconstants.MemoryWarningThreshold,
		"system_memory_mb": serviceconstants.MemoryWarningThreshold * 2,
		"stack_memory_mb":  50.0,
		"uptime_hours":     24.0,
	}
}

// checkPostgresDependency: PostgreSQL 의존성 상태를 확인합니다
func (s *Service) checkPostgresDependency() gin.H {
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

// checkRedisDependency: Redis 의존성 상태를 확인합니다
func (s *Service) checkRedisDependency() gin.H {
	// Use the same logic as checkRedisStatus
	return s.checkRedisStatus()
}
