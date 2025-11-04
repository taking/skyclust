package system

import (
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
