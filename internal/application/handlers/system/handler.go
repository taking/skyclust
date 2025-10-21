package system

import (
	"net/http"
	"runtime"
	"strconv"
	"time"

	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/responses"
	"skyclust/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Handler handles system management operations
type Handler struct {
	*handlers.BaseHandler
}

// NewHandler creates a new system handler
func NewHandler(logger *logger.Logger) *Handler {
	return &Handler{
		BaseHandler: handlers.NewBaseHandler("system"),
	}
}

// GetSystemStatus returns the current system status
func (h *Handler) GetSystemStatus(c *gin.Context) {
	// Check admin permission
	if !h.checkAdminPermission(c) {
		return
	}
	// Get system metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Get actual uptime (assuming process start time is available)
	uptime := time.Since(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) // This should be replaced with actual process start time

	status := gin.H{
		"status":      "healthy",
		"timestamp":   time.Now().Format(time.RFC3339),
		"uptime":      uptime.String(),
		"version":     "1.0.0", // This should come from config
		"environment": h.getEnvironment(),
		"metrics": gin.H{
			"memory": gin.H{
				"alloc_mb":       bToMb(m.Alloc),
				"total_alloc_mb": bToMb(m.TotalAlloc),
				"sys_mb":         bToMb(m.Sys),
				"num_gc":         m.NumGC,
			},
			"goroutines": runtime.NumGoroutine(),
		},
	}

	responses.OK(c, status, "System status retrieved successfully")
}

// GetSystemHealth returns detailed health information
func (h *Handler) GetSystemHealth(c *gin.Context) {
	// Check admin permission
	if !h.checkAdminPermission(c) {
		return
	}
	// TODO: Implement comprehensive health checks
	health := gin.H{
		"status": "healthy",
		"checks": gin.H{
			"database": "healthy",
			"redis":    "healthy",
			"plugins":  "healthy",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	responses.OK(c, health, "System health retrieved successfully")
}

// GetSystemConfig returns current system configuration
func (h *Handler) GetSystemConfig(c *gin.Context) {
	// Check admin permission
	if !h.checkAdminPermission(c) {
		return
	}
	// TODO: Return actual configuration
	config := gin.H{
		"server": gin.H{
			"port": 8080,
			"host": "localhost",
		},
		"database": gin.H{
			"host": "localhost",
			"port": 5432,
		},
		"redis": gin.H{
			"host": "localhost",
			"port": 6379,
		},
	}

	responses.OK(c, config, "System configuration retrieved successfully")
}

// UpdateSystemConfig updates system configuration
func (h *Handler) UpdateSystemConfig(c *gin.Context) {
	// Check admin permission
	if !h.checkAdminPermission(c) {
		return
	}
	var req gin.H
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.BadRequest(c, "Invalid request body")
		return
	}

	// TODO: Implement configuration update
	responses.OK(c, gin.H{
		"message": "Configuration updated successfully",
		"config":  req,
	}, "System configuration updated successfully")
}

// GetSystemLogs retrieves system logs
func (h *Handler) GetSystemLogs(c *gin.Context) {
	// Check admin permission
	if !h.checkAdminPermission(c) {
		return
	}
	// Parse query parameters
	level := c.DefaultQuery("level", "info")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	// Limit maximum logs to prevent memory issues
	if limit > 1000 {
		limit = 1000
	}

	// Use limit in the response
	_ = limit // Ensure limit is used

	// TODO: Implement actual log retrieval
	// This would typically involve:
	// 1. Reading from log files
	// 2. Filtering by level and time
	// 3. Paginating results

	logs := []gin.H{
		{
			"timestamp": time.Now().Format(time.RFC3339),
			"level":     level,
			"message":   "Sample log entry",
			"source":    "system",
		},
	}

	responses.OK(c, gin.H{
		"logs":  logs,
		"level": level,
		"limit": limit,
	}, "System logs retrieved successfully")
}

// RestartSystem restarts the system
func (h *Handler) RestartSystem(c *gin.Context) {
	// Check admin permission
	if !h.checkAdminPermission(c) {
		return
	}
	// TODO: Implement system restart
	// This would typically involve:
	// 1. Graceful shutdown
	// 2. Process restart
	// 3. Health checks

	responses.OK(c, gin.H{
		"message": "System restart initiated",
		"status":  "restarting",
	}, "System restart initiated successfully")
}

// Helper functions
func (h *Handler) getEnvironment() string {
	// TODO: Get actual environment from config
	return "development"
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// checkAdminPermission checks if the current user has admin permission
func (h *Handler) checkAdminPermission(c *gin.Context) bool {
	// Get current user role from token for authorization
	userRole, err := h.GetUserRoleFromToken(c)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			responses.DomainError(c, domainErr)
		} else {
			responses.InternalServerError(c, "Failed to get user role from token")
		}
		return false
	}

	// Check if user has admin role
	if userRole != domain.AdminRoleType {
		responses.DomainError(c, domain.NewDomainError(
			domain.ErrCodeForbidden,
			"Insufficient permissions - admin role required",
			http.StatusForbidden,
		))
		return false
	}

	return true
}
