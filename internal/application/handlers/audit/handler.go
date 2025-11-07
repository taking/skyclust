package audit

import (
	"fmt"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler: 감사 로그 관리 작업을 처리하는 핸들러
type Handler struct {
	*handlers.BaseHandler
	auditLogService domain.AuditLogService
}

// NewHandler: 새로운 감사 로그 핸들러를 생성합니다
func NewHandler(auditLogService domain.AuditLogService) *Handler {
	return &Handler{
		BaseHandler:     handlers.NewBaseHandler("audit"),
		auditLogService: auditLogService,
	}
}

// GetAuditLogs: 필터링을 포함한 감사 로그를 조회합니다
// 쿼리 파라미터 지원: aggregate (stats), format (summary), 필터링 파라미터
func (h *Handler) GetAuditLogs(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "get_audit_logs", 200)

	// Check for aggregate query (stats)
	aggregate := c.Query("aggregate")
	if aggregate == "stats" {
		h.GetAuditStats(c)
		return
	}

	// Check for format query (summary)
	format := c.Query("format")
	if format == "summary" {
		h.GetAuditLogSummary(c)
		return
	}

	// Log operation start
	h.LogInfo(c, "Getting audit logs",
		zap.String("operation", "get_audit_logs"))

	// Parse and validate query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	userID := c.Query("user_id")
	action := c.Query("action")
	resource := c.Query("resource")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

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

	var startTime *time.Time
	if startTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = &parsed
		} else {
			h.LogWarn(c, "Invalid start_time format, expected RFC3339",
				zap.String("start_time", startTimeStr),
				zap.Error(err))
		}
	}

	var endTime *time.Time
	if endTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = &parsed
		} else {
			h.LogWarn(c, "Invalid end_time format, expected RFC3339",
				zap.String("end_time", endTimeStr),
				zap.Error(err))
		}
	}

	// Validate that at least one filter is provided
	if userIDUUID == nil && action == "" && resource == "" && startTime == nil && endTime == nil {
		h.LogWarn(c, "No filters provided for audit log query")
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "at least one filter is required. Provide one of: user_id, action, resource, or start_time/end_time", 400), "get_audit_logs")
		return
	}

	// Validate date range: if one date is provided, both should be provided
	if (startTime != nil && endTime == nil) || (startTime == nil && endTime != nil) {
		h.LogWarn(c, "Both start_time and end_time must be provided together")
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "both start_time and end_time must be provided together", 400), "get_audit_logs")
		return
	}

	filters := domain.AuditLogFilters{
		Limit:     limit,
		UserID:    userIDUUID,
		Action:    action,
		Resource:  resource,
		StartTime: startTime,
		EndTime:   endTime,
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

// GetAuditStats: 감사 로그 통계를 조회합니다
// GET /audit-logs?aggregate=stats로 호출
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

// GetAuditLogSummary: 감사 로그 요약을 조회합니다
// GET /audit-logs?format=summary로 호출
func (h *Handler) GetAuditLogSummary(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "get_audit_summary", 200)

	// Log operation start
	h.LogInfo(c, "Getting audit log summary",
		zap.String("operation", "get_audit_summary"))

	// Parse date range parameters (optional)
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	// Default to last 30 days if no date range provided
	startTime := time.Now().AddDate(0, 0, -30)
	endTime := time.Now()

	if startTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = parsed
		} else {
			h.LogWarn(c, "Invalid start_time format, expected RFC3339, using default (30 days ago)",
				zap.String("start_time", startTimeStr),
				zap.Error(err))
		}
	}

	if endTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = parsed
		} else {
			h.LogWarn(c, "Invalid end_time format, expected RFC3339, using default (now)",
				zap.String("end_time", endTimeStr),
				zap.Error(err))
		}
	}

	// Validate date range: start_time should be before end_time
	if startTime.After(endTime) {
		h.LogWarn(c, "start_time is after end_time, swapping values")
		startTime, endTime = endTime, startTime
	}

	// Log business event
	h.LogBusinessEvent(c, "audit_summary_requested", "", "", map[string]interface{}{
		"operation":   "get_audit_summary",
		"start_time":  startTime,
		"end_time":    endTime,
		"period_days": int(endTime.Sub(startTime).Hours() / 24),
	})

	summary, err := h.auditLogService.GetAuditLogSummary(startTime, endTime)
	if err != nil {
		h.LogError(c, err, "Failed to get audit log summary")
		h.HandleError(c, err, "get_audit_summary")
		return
	}

	// Log successful operation
	h.LogInfo(c, "Audit log summary retrieved successfully",
		zap.Int64("total_events", summary.TotalEvents),
		zap.Int64("unique_users", summary.UniqueUsers),
		zap.Int64("security_events", summary.SecurityEvents),
		zap.Int64("error_events", summary.ErrorEvents))

	h.OK(c, summary, "Audit log summary retrieved successfully")
}

// ExportAuditLogs: 감사 로그를 내보냅니다
func (h *Handler) ExportAuditLogs(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "export_audit_logs", 200)

	// Log operation start
	h.LogInfo(c, "Exporting audit logs",
		zap.String("operation", "export_audit_logs"))

	// Parse format (default: json)
	format := c.DefaultQuery("format", "json")
	if format != "json" && format != "csv" && format != "xlsx" {
		h.BadRequest(c, "Invalid format. Supported formats: json, csv")
		return
	}

	// Parse filters
	filters := h.parseExportFilters(c)

	// Validate that at least one filter is provided (optional validation, but helpful)
	if filters.UserID == nil && filters.Action == "" && filters.Resource == "" && filters.StartTime == nil && filters.EndTime == nil {
		// Allow export without filters, but log it
		h.LogInfo(c, "Exporting audit logs without filters",
			zap.String("operation", "export_audit_logs"),
			zap.String("format", format))
	}

	// Log business event
	h.LogBusinessEvent(c, "audit_logs_export_requested", "", "", map[string]interface{}{
		"operation": "export_audit_logs",
		"format":    format,
		"filters":   filters,
	})

	// Export audit logs
	data, err := h.auditLogService.ExportAuditLogs(filters, format)
	if err != nil {
		h.HandleError(c, err, "export_audit_logs")
		return
	}

	// Set response headers for file download
	c.Header("Content-Type", h.getContentType(format))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=audit_logs_%s.%s", time.Now().Format("20060102_150405"), format))

	// Return file data
	c.Data(200, h.getContentType(format), data)
}

// parseExportFilters parses export filters from query parameters
func (h *Handler) parseExportFilters(c *gin.Context) domain.AuditLogFilters {
	filters := domain.AuditLogFilters{}

	// Parse user ID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			filters.UserID = &userID
		}
	}

	// Parse action
	filters.Action = c.Query("action")

	// Parse resource
	filters.Resource = c.Query("resource")

	// Parse date range
	if startStr := c.Query("start_time"); startStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startStr); err == nil {
			filters.StartTime = &startTime
		} else {
			// Log parsing error but don't fail - will be validated in service
			filters.StartTime = nil
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endStr); err == nil {
			filters.EndTime = &endTime
		} else {
			// Log parsing error but don't fail - will be validated in service
			filters.EndTime = nil
		}
	}

	// Parse pagination
	limitStr := c.DefaultQuery("limit", "1000")
	if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
		filters.Limit = limit
	} else {
		filters.Limit = 1000 // Default limit for export
	}

	pageStr := c.DefaultQuery("page", "1")
	if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
		filters.Page = page
	} else {
		filters.Page = 1
	}

	return filters
}

// getContentType returns the Content-Type header for the given format
func (h *Handler) getContentType(format string) string {
	switch format {
	case "csv":
		return "text/csv"
	case "json":
		return "application/json"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}

// CleanupAuditLogs: 오래된 감사 로그를 정리합니다
func (h *Handler) CleanupAuditLogs(c *gin.Context) {
	// Start performance tracking
	defer h.TrackRequest(c, "cleanup_audit_logs", 200)

	// Log operation start
	h.LogInfo(c, "Cleaning up audit logs",
		zap.String("operation", "cleanup_audit_logs"))

	// Parse retention days from query parameter or body
	retentionDaysStr := c.DefaultQuery("retention_days", "90")
	retentionDays, err := strconv.Atoi(retentionDaysStr)
	if err != nil || retentionDays <= 0 {
		h.BadRequest(c, "Invalid retention_days parameter. Must be a positive integer greater than 0.")
		return
	}

	// Log business event
	h.LogBusinessEvent(c, "audit_logs_cleanup_requested", "", "", map[string]interface{}{
		"operation":      "cleanup_audit_logs",
		"retention_days": retentionDays,
	})

	// Cleanup audit logs
	deletedCount, err := h.auditLogService.CleanupAuditLogs(retentionDays)
	if err != nil {
		h.HandleError(c, err, "cleanup_audit_logs")
		return
	}

	h.OK(c, gin.H{
		"message":        "Audit logs cleaned up successfully",
		"retention_days": retentionDays,
		"deleted_count":  deletedCount,
	}, "Cleanup completed")
}
