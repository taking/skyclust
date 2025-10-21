package system

import (
	"skyclust/pkg/logger"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up system management routes
func SetupRoutes(router *gin.RouterGroup, logger *logger.Logger) {
	systemHandler := NewHandler(logger)

	// System status and health
	router.GET("/status", systemHandler.GetSystemStatus) // GET /api/v1/admin/system/status
	router.GET("/health", systemHandler.GetSystemHealth) // GET /api/v1/admin/system/health

	// Configuration management
	router.GET("/config", systemHandler.GetSystemConfig)    // GET /api/v1/admin/system/config
	router.PUT("/config", systemHandler.UpdateSystemConfig) // PUT /api/v1/admin/system/config

	// Logs and monitoring
	router.GET("/logs", systemHandler.GetSystemLogs) // GET /api/v1/admin/system/logs

	// System operations
	router.POST("/restart", systemHandler.RestartSystem) // POST /api/v1/admin/system/restart
}
