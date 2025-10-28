package system

import (
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
)

// Handler handles system monitoring operations
type Handler struct {
	*handlers.BaseHandler
	monitoringService interface{}
}

// NewHandler creates a new system handler
func NewHandler(monitoringService interface{}) *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("system"),
		monitoringService: monitoringService,
	}
}

// HealthCheck provides comprehensive health check endpoint
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

// GetSystemMetrics returns system performance metrics
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

// GetSystemAlerts returns current alert status
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

// GetSystemStatus returns the current system status (detailed health check)
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

// GetSystemHealth returns detailed health information (alias for HealthCheck)
func (h *Handler) GetSystemHealth(c *gin.Context) {
	h.HealthCheck(c)
}
