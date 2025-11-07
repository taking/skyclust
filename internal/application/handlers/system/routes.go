package system

import (
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up system monitoring routes
func SetupRoutes(router *gin.RouterGroup, monitoringService interface{}) {
	systemHandler := NewHandler(monitoringService)

	// Health check endpoint
	router.GET("/health", systemHandler.HealthCheck)

	// System metrics endpoint
	router.GET("/metrics", systemHandler.GetSystemMetrics)

	// System alerts endpoint
	router.GET("/alerts", systemHandler.GetSystemAlerts)

	// Legacy endpoints for backward compatibility
	router.GET("/status", systemHandler.GetSystemStatus)
}

// SetupRoutesWithUserService sets up system monitoring routes with user service (for initialization check)
func SetupRoutesWithUserService(router *gin.RouterGroup, monitoringService interface{}, userService domain.UserService) {
	systemHandler := NewHandlerWithUserService(monitoringService, userService)

	// Health check endpoint
	router.GET("/health", systemHandler.HealthCheck)

	// System metrics endpoint
	router.GET("/metrics", systemHandler.GetSystemMetrics)

	// System alerts endpoint
	router.GET("/alerts", systemHandler.GetSystemAlerts)

	// Legacy endpoints for backward compatibility
	router.GET("/status", systemHandler.GetSystemStatus)

	// System initialization status endpoint (public)
	router.GET("/initialized", systemHandler.GetInitializationStatus)
}
