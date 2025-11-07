package system

import (
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
)

// Handler: 시스템 모니터링 작업을 처리하는 핸들러
type Handler struct {
	*handlers.BaseHandler
	monitoringService interface{}
	userService       domain.UserService
}

// NewHandler: 새로운 시스템 핸들러를 생성합니다
func NewHandler(monitoringService interface{}) *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("system"),
		monitoringService: monitoringService,
	}
}

// NewHandlerWithUserService: 사용자 서비스를 포함한 시스템 핸들러를 생성합니다
func NewHandlerWithUserService(monitoringService interface{}, userService domain.UserService) *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("system"),
		monitoringService: monitoringService,
		userService:       userService,
	}
}

// HealthCheck: 종합적인 헬스 체크 엔드포인트를 제공합니다
func (h *Handler) HealthCheck(c *gin.Context) {
	// Type assertion to access monitoring service methods
	monitoringService := h.monitoringService.(interface {
		IncrementRequestCount()
		GetHealthStatus() gin.H
	})

	// Increment request count for monitoring
	monitoringService.IncrementRequestCount()

	// Get health status
	healthStatus := monitoringService.GetHealthStatus()

	h.OK(c, healthStatus, "Health check completed successfully")
}

// GetSystemMetrics: 시스템 성능 메트릭을 반환합니다
func (h *Handler) GetSystemMetrics(c *gin.Context) {
	// Type assertion to access monitoring service methods
	monitoringService := h.monitoringService.(interface {
		IncrementRequestCount()
		GetSystemMetrics() gin.H
	})

	// Increment request count for monitoring
	monitoringService.IncrementRequestCount()

	// Get system metrics
	metrics := monitoringService.GetSystemMetrics()

	h.OK(c, metrics, "System metrics retrieved successfully")
}

// GetSystemAlerts: 현재 알림 상태를 반환합니다
func (h *Handler) GetSystemAlerts(c *gin.Context) {
	// Type assertion to access monitoring service methods
	monitoringService := h.monitoringService.(interface {
		IncrementRequestCount()
		GetAlerts() gin.H
	})

	// Increment request count for monitoring
	monitoringService.IncrementRequestCount()

	// Get alerts
	alerts := monitoringService.GetAlerts()

	h.OK(c, alerts, "System alerts retrieved successfully")
}

// GetSystemStatus: 현재 시스템 상태를 반환합니다 (상세 헬스 체크)
func (h *Handler) GetSystemStatus(c *gin.Context) {
	// Type assertion to access monitoring service methods
	monitoringService := h.monitoringService.(interface {
		IncrementRequestCount()
		GetHealthStatus() gin.H
	})

	// Increment request count for monitoring
	monitoringService.IncrementRequestCount()

	// Get detailed health status (same as GetHealthStatus)
	status := monitoringService.GetHealthStatus()

	h.OK(c, status, "System status retrieved successfully")
}

// GetSystemHealth: 상세한 헬스 정보를 반환합니다 (HealthCheck의 별칭)
func (h *Handler) GetSystemHealth(c *gin.Context) {
	h.HealthCheck(c)
}

// GetInitializationStatus 시스템 초기화 상태 확인 (공개 엔드포인트)
// GET /api/v1/system/initialized
func (h *Handler) GetInitializationStatus(c *gin.Context) {
	// UserService가 없으면 에러 반환
	if h.userService == nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeServiceUnavailable, "User service is not available", 503), "get_initialization_status")
		return
	}

	// 사용자 수 조회
	userCount, err := h.userService.GetUserCount()
	if err != nil {
		h.HandleError(c, err, "get_initialization_status")
		return
	}

	// 초기화 상태 반환 (사용자가 1명 이상 있으면 초기화됨)
	h.OK(c, gin.H{
		"initialized": userCount > 0,
		"user_count":  userCount,
	}, "Initialization status retrieved successfully")
}
