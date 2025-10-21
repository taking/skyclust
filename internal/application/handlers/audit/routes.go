package audit

import (
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up audit log management routes
func SetupRoutes(router *gin.RouterGroup, auditLogService domain.AuditLogService) {
	auditHandler := NewHandler(auditLogService)

	// Audit log management
	router.GET("/", auditHandler.GetAuditLogs)              // GET /api/v1/admin/audit
	router.GET("/stats", auditHandler.GetAuditStats)        // GET /api/v1/admin/audit/stats
	router.GET("/summary", auditHandler.GetAuditLogSummary) // GET /api/v1/admin/audit/summary
	router.GET("/:id", auditHandler.GetAuditLog)            // GET /api/v1/admin/audit/:id
	router.GET("/export", auditHandler.ExportAuditLogs)     // GET /api/v1/admin/audit/export
	router.POST("/cleanup", auditHandler.CleanupAuditLogs)  // POST /api/v1/admin/audit/cleanup
}
