package routes

import (
	"skyclust/internal/delivery/http"

	"github.com/gin-gonic/gin"
)

// SetupSystemRoutes sets up system management routes
func SetupSystemRoutes(router *gin.RouterGroup, systemHandler *http.SystemHandler) {
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
