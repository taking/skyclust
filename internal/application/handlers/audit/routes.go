package audit

import (
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up audit log management routes
func SetupRoutes(router *gin.RouterGroup, auditLogService domain.AuditLogService) {
	auditHandler := NewHandler(auditLogService)

	// Audit log management (RESTful)
	// Base path: /api/v1/admin/audit-logs
	router.GET("", auditHandler.GetAuditLogs)          // GET /api/v1/admin/audit-logs (with query params: aggregate=stats, format=summary)
	router.GET("/:id", auditHandler.GetAuditLog)        // GET /api/v1/admin/audit-logs/:id
	router.GET("/export", auditHandler.ExportAuditLogs) // GET /api/v1/admin/audit-logs/export
	router.DELETE("", auditHandler.CleanupAuditLogs)    // DELETE /api/v1/admin/audit-logs?retention_days=90
}
