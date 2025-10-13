package http

import (
	"skyclust/internal/usecase"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ExportHandler struct {
	logger        *zap.Logger
	exportService *usecase.ExportService
}

func NewExportHandler(logger *zap.Logger, exportService *usecase.ExportService) *ExportHandler {
	return &ExportHandler{
		logger:        logger,
		exportService: exportService,
	}
}

// ExportData exports data in the specified format
func (h *ExportHandler) ExportData(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var request usecase.ExportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		BadRequestResponse(c, "Invalid request body")
		return
	}

	request.UserID = userID.(string)

	// Validate request
	if err := h.validateExportRequest(&request); err != nil {
		BadRequestResponse(c, err.Error())
		return
	}

	// Export data based on type
	var result *usecase.ExportResult
	var err error

	switch request.Type {
	case usecase.ExportTypeVMs:
		result, err = h.exportService.ExportVMs(c.Request.Context(), request)
	case usecase.ExportTypeWorkspaces:
		result, err = h.exportService.ExportWorkspaces(c.Request.Context(), request)
	case usecase.ExportTypeCredentials:
		result, err = h.exportService.ExportCredentials(c.Request.Context(), request)
	case usecase.ExportTypeAuditLogs:
		result, err = h.exportService.ExportAuditLogs(c.Request.Context(), request)
	case usecase.ExportTypeCosts:
		BadRequestResponse(c, "Cost export not implemented yet")
		return
	default:
		BadRequestResponse(c, "Invalid export type")
		return
	}

	if err != nil {
		h.logger.Error("Failed to export data",
			zap.String("user_id", userID.(string)),
			zap.String("type", string(request.Type)),
			zap.String("format", string(request.Format)),
			zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to export data", "")
		return
	}

	SuccessResponse(c, http.StatusOK, result, "Data exported successfully")
}

// GetExportStatus retrieves the status of an export
func (h *ExportHandler) GetExportStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	exportID := c.Param("id")
	if exportID == "" {
		BadRequestResponse(c, "Export ID is required")
		return
	}

	result, err := h.exportService.GetExportStatus(c.Request.Context(), exportID)
	if err != nil {
		h.logger.Error("Failed to get export status",
			zap.String("user_id", userID.(string)),
			zap.String("export_id", exportID),
			zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to get export status", "")
		return
	}

	SuccessResponse(c, http.StatusOK, result, "Export status retrieved successfully")
}

// DownloadExport downloads an exported file
func (h *ExportHandler) DownloadExport(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	exportID := c.Param("id")
	if exportID == "" {
		BadRequestResponse(c, "Export ID is required")
		return
	}

	// Get export status
	result, err := h.exportService.GetExportStatus(c.Request.Context(), exportID)
	if err != nil {
		h.logger.Error("Failed to get export status for download",
			zap.String("user_id", userID.(string)),
			zap.String("export_id", exportID),
			zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to get export status", "")
		return
	}

	if result.Status != "completed" {
		BadRequestResponse(c, "Export is not ready for download")
		return
	}

	// In a real implementation, you would serve the actual file
	// For now, return a mock response
	c.Header("Content-Disposition", "attachment; filename=export.csv")
	c.Header("Content-Type", "text/csv")
	c.String(http.StatusOK, "Mock export data for export ID: %s", exportID)
}

// GetExportHistory retrieves export history for the user
func (h *ExportHandler) GetExportHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedResponse(c, "User not authenticated")
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		BadRequestResponse(c, "Invalid limit parameter. Must be between 1 and 100")
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		BadRequestResponse(c, "Invalid offset parameter. Must be >= 0")
		return
	}

	// This would typically query export history from a database
	// For now, return mock data
	history := []usecase.ExportResult{
		{
			ID:          "export-1",
			UserID:      userID.(string),
			Type:        usecase.ExportTypeVMs,
			Format:      usecase.ExportFormatCSV,
			Status:      "completed",
			FileSize:    1024,
			DownloadURL: "/api/v1/exports/export-1/download",
		},
		{
			ID:          "export-2",
			UserID:      userID.(string),
			Type:        usecase.ExportTypeWorkspaces,
			Format:      usecase.ExportFormatJSON,
			Status:      "completed",
			FileSize:    2048,
			DownloadURL: "/api/v1/exports/export-2/download",
		},
	}

	SuccessResponse(c, http.StatusOK, map[string]interface{}{
		"exports": history,
		"total":   len(history),
		"limit":   limit,
		"offset":  offset,
	}, "Export history retrieved successfully")
}

// GetSupportedFormats returns supported export formats
func (h *ExportHandler) GetSupportedFormats(c *gin.Context) {
	formats := map[string]interface{}{
		"formats": []map[string]interface{}{
			{
				"format":      "csv",
				"name":        "CSV",
				"description": "Comma-separated values",
				"mime_type":   "text/csv",
			},
			{
				"format":      "json",
				"name":        "JSON",
				"description": "JavaScript Object Notation",
				"mime_type":   "application/json",
			},
			{
				"format":      "xlsx",
				"name":        "Excel",
				"description": "Microsoft Excel format",
				"mime_type":   "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			},
			{
				"format":      "pdf",
				"name":        "PDF",
				"description": "Portable Document Format",
				"mime_type":   "application/pdf",
			},
		},
		"types": []map[string]interface{}{
			{
				"type":        "vms",
				"name":        "Virtual Machines",
				"description": "Export VM instances and their configurations",
			},
			{
				"type":        "workspaces",
				"name":        "Workspaces",
				"description": "Export workspace information and settings",
			},
			{
				"type":        "credentials",
				"name":        "Credentials",
				"description": "Export cloud provider credentials (encrypted)",
			},
			{
				"type":        "audit_logs",
				"name":        "Audit Logs",
				"description": "Export system audit logs and activity history",
			},
			{
				"type":        "costs",
				"name":        "Cost Data",
				"description": "Export cost analysis and billing information",
			},
		},
	}

	SuccessResponse(c, http.StatusOK, formats, "Supported formats retrieved successfully")
}

// Helper methods

func (h *ExportHandler) validateExportRequest(req *usecase.ExportRequest) error {
	if req.Type == "" {
		return fmt.Errorf("export type is required")
	}
	if req.Format == "" {
		return fmt.Errorf("export format is required")
	}

	// Validate export type
	validTypes := map[usecase.ExportType]bool{
		usecase.ExportTypeVMs:         true,
		usecase.ExportTypeWorkspaces:  true,
		usecase.ExportTypeCredentials: true,
		usecase.ExportTypeAuditLogs:   true,
		usecase.ExportTypeCosts:       true,
	}
	if !validTypes[req.Type] {
		return fmt.Errorf("invalid export type: %s", req.Type)
	}

	// Validate export format
	validFormats := map[usecase.ExportFormat]bool{
		usecase.ExportFormatCSV:  true,
		usecase.ExportFormatJSON: true,
		usecase.ExportFormatXLSX: true,
		usecase.ExportFormatPDF:  true,
	}
	if !validFormats[req.Format] {
		return fmt.Errorf("invalid export format: %s", req.Format)
	}

	return nil
}
