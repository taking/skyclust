package audit

import (
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
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
	// Start performance tracking
	defer h.TrackRequest(c, "get_audit_logs", 200)

	// Log operation start
	h.LogInfo(c, "Getting audit logs",
		zap.String("operation", "get_audit_logs"))

	// Parse and validate query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	userID := c.Query("user_id")
	action := c.Query("action")
	resource := c.Query("resource")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		h.LogWarn(c, "Invalid limit parameter, using default",
			zap.String("limit", limitStr),
			zap.Int("default_limit", 50))
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		h.LogWarn(c, "Invalid offset parameter, using default",
			zap.String("offset", offsetStr),
			zap.Int("default_offset", 0))
		offset = 0
	}

	var userIDUUID *uuid.UUID
	if userID != "" {
		if parsed, err := uuid.Parse(userID); err == nil {
			userIDUUID = &parsed
		} else {
			h.LogWarn(c, "Invalid user ID format",
				zap.String("user_id", userID),
				zap.Error(err))
		}
	}

	filters := domain.AuditLogFilters{
		Limit:    limit,
		UserID:   userIDUUID,
		Action:   action,
		Resource: resource,
	}

	// Log business event
	h.LogBusinessEvent(c, "audit_logs_requested", "", "", map[string]interface{}{
		"limit":    limit,
		"offset":   offset,
		"user_id":  userID,
		"action":   action,
		"resource": resource,
	})

	logs, total, err := h.auditLogService.GetAuditLogs(filters)
	if err != nil {
		h.LogError(c, err, "Failed to get audit logs")
		h.HandleError(c, err, "get_audit_logs")
		return
	}

	// Log successful operation
	h.LogInfo(c, "Audit logs retrieved successfully",
		zap.Int("logs_count", len(logs)),
		zap.Int64("total", total))

	h.OK(c, gin.H{
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
	// Start performance tracking
	defer h.TrackRequest(c, "get_audit_log", 200)

	logIDStr := c.Param("id")

	// Log operation start
	h.LogInfo(c, "Getting specific audit log",
		zap.String("operation", "get_audit_log"),
		zap.String("log_id", logIDStr))

	logID, err := uuid.Parse(logIDStr)
	if err != nil {
		h.LogWarn(c, "Invalid audit log ID format",
			zap.String("log_id", logIDStr),
			zap.Error(err))
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid audit log ID format", 400), "get_audit_log")
		return
	}

	// Log business event
	h.LogBusinessEvent(c, "audit_log_requested", "", logID.String(), map[string]interface{}{
		"log_id": logID.String(),
	})

	log, err := h.auditLogService.GetAuditLogByID(logID)
	if err != nil {
		h.LogError(c, err, "Failed to get audit log")
		h.HandleError(c, err, "get_audit_log")
		return
	}

	// Log successful operation
	h.LogInfo(c, "Audit log retrieved successfully",
		zap.String("log_id", logID.String()))

	h.OK(c, log, "Audit log retrieved successfully")
}

// GetAuditStats retrieves audit log statistics
func (h *Handler) GetAuditStats(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "get_audit_stats", 200)

	// Log operation start
	h.LogInfo(c, "Getting audit statistics",
		zap.String("operation", "get_audit_stats"))

	// Log business event
	h.LogBusinessEvent(c, "audit_stats_requested", "", "", map[string]interface{}{
		"operation": "get_audit_stats",
	})

	stats, err := h.auditLogService.GetAuditStats(domain.AuditStatsFilters{})
	if err != nil {
		h.LogError(c, err, "Failed to get audit statistics")
		h.HandleError(c, err, "get_audit_stats")
		return
	}

	// Log successful operation
	h.LogInfo(c, "Audit statistics retrieved successfully")

	h.OK(c, stats, "Audit statistics retrieved successfully")
}

// GetAuditLogSummary retrieves audit log summary
func (h *Handler) GetAuditLogSummary(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "get_audit_summary", 200)

	// Log operation start
	h.LogInfo(c, "Getting audit log summary",
		zap.String("operation", "get_audit_summary"))

	// Log business event
	h.LogBusinessEvent(c, "audit_summary_requested", "", "", map[string]interface{}{
		"operation": "get_audit_summary",
		"period":    "30_days",
	})

	summary, err := h.auditLogService.GetAuditLogSummary(time.Now().AddDate(0, 0, -30), time.Now())
	if err != nil {
		h.LogError(c, err, "Failed to get audit log summary")
		h.HandleError(c, err, "get_audit_summary")
		return
	}

	// Log successful operation
	h.LogInfo(c, "Audit log summary retrieved successfully")

	h.OK(c, summary, "Audit log summary retrieved successfully")
}

// ExportAuditLogs exports audit logs
func (h *Handler) ExportAuditLogs(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "export_audit_logs", 200)

	// Log operation start
	h.LogInfo(c, "Exporting audit logs",
		zap.String("operation", "export_audit_logs"))

	// Log business event
	h.LogBusinessEvent(c, "audit_logs_export_requested", "", "", map[string]interface{}{
		"operation": "export_audit_logs",
	})

	// TODO: Implement export functionality
	h.LogWarn(c, "Export functionality not implemented yet")

	h.OK(c, gin.H{
		"message": "Export functionality not implemented yet",
	}, "Export initiated")
}

// CleanupAuditLogs cleans up old audit logs
func (h *Handler) CleanupAuditLogs(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "cleanup_audit_logs", 200)

	// Log operation start
	h.LogInfo(c, "Cleaning up audit logs",
		zap.String("operation", "cleanup_audit_logs"))

	// Log business event
	h.LogBusinessEvent(c, "audit_logs_cleanup_requested", "", "", map[string]interface{}{
		"operation": "cleanup_audit_logs",
	})

	// TODO: Implement cleanup functionality
	h.LogWarn(c, "Cleanup functionality not implemented yet")

	h.OK(c, gin.H{
		"message": "Cleanup functionality not implemented yet",
	}, "Cleanup initiated")
}
