package http

import (
	"strconv"
	"time"

	"skyclust/internal/domain"
	"skyclust/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuditHandler handles audit log management operations
type AuditHandler struct {
	auditService domain.AuditLogService
	logger       *logger.Logger
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService domain.AuditLogService, logger *logger.Logger) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
		logger:       logger,
	}
}

// GetAuditLogs retrieves audit logs with filtering and pagination
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	userID := c.Query("user_id")
	action := c.Query("action")
	resource := c.Query("resource")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}

	// Parse dates
	var startTime, endTime *time.Time
	if startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			startTime = &t
		}
	}
	if endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			endTime = &t
		}
	}

	// Parse user ID
	var userUUID *uuid.UUID
	if userID != "" {
		if id, err := uuid.Parse(userID); err == nil {
			userUUID = &id
		}
	}

	// Get audit logs
	logs, total, err := h.auditService.GetAuditLogs(domain.AuditLogFilters{
		UserID:    userUUID,
		Action:    action,
		Resource:  resource,
		StartTime: startTime,
		EndTime:   endTime,
		Page:      page,
		Limit:     limit,
	})
	if err != nil {
		h.logger.Errorf("Failed to get audit logs: %v", err)
		InternalServerErrorResponse(c, "Failed to retrieve audit logs")
		return
	}

	// Calculate pagination info
	totalPages := (total + int64(limit) - 1) / int64(limit)
	hasNext := int64(page) < totalPages
	hasPrev := page > 1

	// Format response
	logList := make([]gin.H, len(logs))
	for i, log := range logs {
		logList[i] = gin.H{
			"id":         log.ID,
			"user_id":    log.UserID,
			"action":     log.Action,
			"resource":   log.Resource,
			"details":    log.Details,
			"ip_address": log.IPAddress,
			"user_agent": log.UserAgent,
			"created_at": log.CreatedAt,
		}
	}

	OKResponse(c, gin.H{
		"logs": logList,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
		"filters": gin.H{
			"user_id":    userID,
			"action":     action,
			"resource":   resource,
			"start_date": startDate,
			"end_date":   endDate,
		},
	}, "Audit logs retrieved successfully")
}

// GetAuditLog retrieves a specific audit log by ID
func (h *AuditHandler) GetAuditLog(c *gin.Context) {
	logIDStr := c.Param("id")
	logID, err := uuid.Parse(logIDStr)
	if err != nil {
		BadRequestResponse(c, "Invalid audit log ID")
		return
	}

	log, err := h.auditService.GetAuditLogByID(logID)
	if err != nil {
		if err == domain.ErrAuditLogNotFound {
			NotFoundResponse(c, "Audit log not found")
			return
		}
		h.logger.Errorf("Failed to get audit log %s: %v", logID, err)
		InternalServerErrorResponse(c, "Failed to retrieve audit log")
		return
	}

	OKResponse(c, gin.H{
		"audit_log": gin.H{
			"id":         log.ID,
			"user_id":    log.UserID,
			"action":     log.Action,
			"resource":   log.Resource,
			"details":    log.Details,
			"ip_address": log.IPAddress,
			"user_agent": log.UserAgent,
			"created_at": log.CreatedAt,
		},
	}, "Audit log retrieved successfully")
}

// GetAuditStats retrieves audit log statistics
func (h *AuditHandler) GetAuditStats(c *gin.Context) {
	// Parse query parameters
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	// Parse dates
	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		BadRequestResponse(c, "Invalid start_date format")
		return
	}
	endTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		BadRequestResponse(c, "Invalid end_date format")
		return
	}

	// Get statistics
	stats, err := h.auditService.GetAuditStats(domain.AuditStatsFilters{
		StartTime: &startTime,
		EndTime:   &endTime,
	})
	if err != nil {
		h.logger.Errorf("Failed to get audit stats: %v", err)
		InternalServerErrorResponse(c, "Failed to retrieve audit statistics")
		return
	}

	OKResponse(c, gin.H{
		"stats": gin.H{
			"total_events":  stats.TotalEvents,
			"unique_users":  stats.UniqueUsers,
			"top_actions":   stats.TopActions,
			"top_resources": stats.TopResources,
			"events_by_day": stats.EventsByDay,
			"period": gin.H{
				"start_date": startDate,
				"end_date":   endDate,
			},
		},
	}, "Audit statistics retrieved successfully")
}

// ExportAuditLogs exports audit logs in various formats
func (h *AuditHandler) ExportAuditLogs(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	userID := c.Query("user_id")
	action := c.Query("action")
	resource := c.Query("resource")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Validate format
	validFormats := []string{"json", "csv", "xlsx"}
	validFormat := false
	for _, f := range validFormats {
		if format == f {
			validFormat = true
			break
		}
	}
	if !validFormat {
		BadRequestResponse(c, "Invalid format. Supported formats: json, csv, xlsx")
		return
	}

	// Parse filters
	var userUUID *uuid.UUID
	if userID != "" {
		if id, err := uuid.Parse(userID); err == nil {
			userUUID = &id
		}
	}

	var startTime, endTime *time.Time
	if startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			startTime = &t
		}
	}
	if endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			endTime = &t
		}
	}

	// Export audit logs
	exportData, err := h.auditService.ExportAuditLogs(domain.AuditLogFilters{
		UserID:    userUUID,
		Action:    action,
		Resource:  resource,
		StartTime: startTime,
		EndTime:   endTime,
	}, format)
	if err != nil {
		h.logger.Errorf("Failed to export audit logs: %v", err)
		InternalServerErrorResponse(c, "Failed to export audit logs")
		return
	}

	// Set appropriate headers based on format
	switch format {
	case "json":
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", "attachment; filename=audit_logs.json")
	case "csv":
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=audit_logs.csv")
	case "xlsx":
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename=audit_logs.xlsx")
	}

	c.Data(200, "", exportData)
}

// CleanupAuditLogs removes old audit logs based on retention policy
func (h *AuditHandler) CleanupAuditLogs(c *gin.Context) {
	var req struct {
		RetentionDays int `json:"retention_days" binding:"required,min=1,max=3650"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "Invalid request body")
		return
	}

	// Cleanup old audit logs
	deletedCount, err := h.auditService.CleanupAuditLogs(req.RetentionDays)
	if err != nil {
		h.logger.Errorf("Failed to cleanup audit logs: %v", err)
		InternalServerErrorResponse(c, "Failed to cleanup audit logs")
		return
	}

	OKResponse(c, gin.H{
		"deleted_count":  deletedCount,
		"retention_days": req.RetentionDays,
		"message":        "Audit logs cleanup completed successfully",
	}, "Audit logs cleanup completed")
}

// GetAuditLogSummary retrieves a summary of audit log activities
func (h *AuditHandler) GetAuditLogSummary(c *gin.Context) {
	// Get summary for the last 7 days
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -7)

	summary, err := h.auditService.GetAuditLogSummary(startTime, endTime)
	if err != nil {
		h.logger.Errorf("Failed to get audit log summary: %v", err)
		InternalServerErrorResponse(c, "Failed to retrieve audit log summary")
		return
	}

	OKResponse(c, gin.H{
		"summary": gin.H{
			"period": gin.H{
				"start_time": startTime.Format(time.RFC3339),
				"end_time":   endTime.Format(time.RFC3339),
			},
			"total_events":     summary.TotalEvents,
			"unique_users":     summary.UniqueUsers,
			"most_active_user": summary.MostActiveUser,
			"top_actions":      summary.TopActions,
			"security_events":  summary.SecurityEvents,
			"error_events":     summary.ErrorEvents,
		},
	}, "Audit log summary retrieved successfully")
}
