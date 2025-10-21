package audit

import (
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/responses"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles audit log management operations
type Handler struct {
	*handlers.BaseHandler
	auditLogService domain.AuditLogService
}

// NewHandler creates a new audit handler
func NewHandler(auditLogService domain.AuditLogService) *Handler {
	return &Handler{
		BaseHandler:     handlers.NewBaseHandler("audit"),
		auditLogService: auditLogService,
	}
}

// GetAuditLogs retrieves audit logs with filtering
func (h *Handler) GetAuditLogs(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	userID := c.Query("user_id")
	action := c.Query("action")
	resource := c.Query("resource")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var userIDUUID *uuid.UUID
	if userID != "" {
		if parsed, err := uuid.Parse(userID); err == nil {
			userIDUUID = &parsed
		}
	}

	filters := domain.AuditLogFilters{
		Limit:    limit,
		UserID:   userIDUUID,
		Action:   action,
		Resource: resource,
	}

	logs, total, err := h.auditLogService.GetAuditLogs(filters)
	if err != nil {
		responses.InternalServerError(c, "Failed to get audit logs")
		return
	}

	responses.OK(c, gin.H{
		"logs": logs,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	}, "Audit logs retrieved successfully")
}

// GetAuditLog retrieves a specific audit log
func (h *Handler) GetAuditLog(c *gin.Context) {
	logIDStr := c.Param("id")
	logID, err := uuid.Parse(logIDStr)
	if err != nil {
		responses.BadRequest(c, "Invalid audit log ID format")
		return
	}

	log, err := h.auditLogService.GetAuditLogByID(logID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			responses.NotFound(c, "Audit log not found")
			return
		}
		responses.InternalServerError(c, "Failed to get audit log")
		return
	}

	responses.OK(c, log, "Audit log retrieved successfully")
}

// GetAuditStats retrieves audit log statistics
func (h *Handler) GetAuditStats(c *gin.Context) {
	stats, err := h.auditLogService.GetAuditStats(domain.AuditStatsFilters{})
	if err != nil {
		responses.InternalServerError(c, "Failed to get audit statistics")
		return
	}

	responses.OK(c, stats, "Audit statistics retrieved successfully")
}

// GetAuditLogSummary retrieves audit log summary
func (h *Handler) GetAuditLogSummary(c *gin.Context) {
	summary, err := h.auditLogService.GetAuditLogSummary(time.Now().AddDate(0, 0, -30), time.Now())
	if err != nil {
		responses.InternalServerError(c, "Failed to get audit log summary")
		return
	}

	responses.OK(c, summary, "Audit log summary retrieved successfully")
}

// ExportAuditLogs exports audit logs
func (h *Handler) ExportAuditLogs(c *gin.Context) {
	// TODO: Implement export functionality
	responses.OK(c, gin.H{
		"message": "Export functionality not implemented yet",
	}, "Export initiated")
}

// CleanupAuditLogs cleans up old audit logs
func (h *Handler) CleanupAuditLogs(c *gin.Context) {
	// TODO: Implement cleanup functionality
	responses.OK(c, gin.H{
		"message": "Cleanup functionality not implemented yet",
	}, "Cleanup initiated")
}
