package export

import (
	"net/http"

	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/readability"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles export operations using improved patterns
type Handler struct {
	*handlers.BaseHandler
	readabilityHelper *readability.ReadabilityHelper
}

// NewHandler creates a new export handler
func NewHandler() *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("export"),
		readabilityHelper: readability.NewReadabilityHelper(),
	}
}

// ExportData handles data export requests using decorator pattern
func (h *Handler) ExportData(c *gin.Context) {
	var req gin.H

	handler := h.Compose(
		h.exportDataHandler(req),
		h.StandardCRUDDecorators("export_data")...,
	)

	handler(c)
}

// exportDataHandler is the core business logic for exporting data
func (h *Handler) exportDataHandler(req gin.H) handlers.HandlerFunc {
	return func(c *gin.Context) {
		req = h.extractValidatedRequest(c)
		userID := h.extractUserID(c)

		h.logExportDataAttempt(c, userID, req)

		// TODO: Implement export functionality
		exportID := uuid.New().String()

		h.logExportDataSuccess(c, userID, exportID)
		h.OK(c, gin.H{
			"export_id": exportID,
			"status":    "processing",
			"message":   "Export initiated",
		}, "Export initiated successfully")
	}
}

// GetSupportedFormats returns supported export formats using decorator pattern
func (h *Handler) GetSupportedFormats(c *gin.Context) {
	handler := h.Compose(
		h.getSupportedFormatsHandler(),
		h.StandardCRUDDecorators("get_supported_formats")...,
	)

	handler(c)
}

// getSupportedFormatsHandler is the core business logic for getting supported formats
func (h *Handler) getSupportedFormatsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		h.logSupportedFormatsRequest(c)

		formats := h.getSupportedFormats()

		h.OK(c, gin.H{
			"formats": formats,
		}, "Supported formats retrieved successfully")
	}
}

// GetExportHistory retrieves export history using decorator pattern
func (h *Handler) GetExportHistory(c *gin.Context) {
	handler := h.Compose(
		h.getExportHistoryHandler(),
		h.StandardCRUDDecorators("get_export_history")...,
	)

	handler(c)
}

// getExportHistoryHandler is the core business logic for getting export history
func (h *Handler) getExportHistoryHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID := h.extractUserID(c)

		h.logExportHistoryRequest(c, userID)

		// TODO: Implement export history retrieval
		history := h.getSampleExportHistory()

		h.OK(c, gin.H{
			"exports": history,
		}, "Export history retrieved successfully")
	}
}

// GetExportStatus retrieves export status using decorator pattern
func (h *Handler) GetExportStatus(c *gin.Context) {
	handler := h.Compose(
		h.getExportStatusHandler(),
		h.StandardCRUDDecorators("get_export_status")...,
	)

	handler(c)
}

// getExportStatusHandler is the core business logic for getting export status
func (h *Handler) getExportStatusHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID := h.extractUserID(c)
		exportID := h.parseExportID(c)

		if exportID == "" {
			return
		}

		h.logExportStatusRequest(c, userID, exportID)

		// TODO: Implement status retrieval
		h.OK(c, gin.H{
			"id":       exportID,
			"status":   "completed",
			"progress": 100,
		}, "Export status retrieved successfully")
	}
}

// DownloadExport handles export download using decorator pattern
func (h *Handler) DownloadExport(c *gin.Context) {
	handler := h.Compose(
		h.downloadExportHandler(),
		h.StandardCRUDDecorators("download_export")...,
	)

	handler(c)
}

// downloadExportHandler is the core business logic for downloading export
func (h *Handler) downloadExportHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID := h.extractUserID(c)
		exportID := h.parseExportID(c)

		if exportID == "" {
			return
		}

		h.logDownloadExportRequest(c, userID, exportID)

		// TODO: Implement download functionality
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", "attachment; filename=export.csv")
		c.String(http.StatusOK, "Sample export data")
	}
}

// Helper methods for better readability

func (h *Handler) extractUserID(c *gin.Context) uuid.UUID {
	userID, exists := c.Get("user_id")
	if !exists {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), "extract_user_id")
		return uuid.Nil
	}
	return userID.(uuid.UUID)
}

func (h *Handler) extractValidatedRequest(c *gin.Context) gin.H {
	var req gin.H
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid request body", 400), "extract_validated_request")
		return gin.H{}
	}
	return req
}

func (h *Handler) parseExportID(c *gin.Context) string {
	exportID := c.Param("id")
	if exportID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Export ID is required", 400), "parse_export_id")
		return ""
	}
	return exportID
}

func (h *Handler) getSupportedFormats() []gin.H {
	return []gin.H{
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
	}
}

func (h *Handler) getSampleExportHistory() []gin.H {
	return []gin.H{
		{
			"id":           uuid.New().String(),
			"type":         "users",
			"format":       "csv",
			"status":       "completed",
			"created_at":   "2024-01-01T00:00:00Z",
			"completed_at": "2024-01-01T00:01:00Z",
		},
	}
}

// Logging helper methods

func (h *Handler) logExportDataAttempt(c *gin.Context, userID uuid.UUID, req gin.H) {
	h.LogBusinessEvent(c, "data_export_attempted", userID.String(), "", map[string]interface{}{
		"operation": "export_data",
		"request":   req,
	})
}

func (h *Handler) logExportDataSuccess(c *gin.Context, userID uuid.UUID, exportID string) {
	h.LogBusinessEvent(c, "data_export_initiated", userID.String(), exportID, map[string]interface{}{
		"operation": "export_data",
		"export_id": exportID,
	})
}

func (h *Handler) logSupportedFormatsRequest(c *gin.Context) {
	h.LogBusinessEvent(c, "supported_formats_requested", "", "", map[string]interface{}{
		"operation": "get_supported_formats",
	})
}

func (h *Handler) logExportHistoryRequest(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "export_history_requested", userID.String(), "", map[string]interface{}{
		"operation": "get_export_history",
	})
}

func (h *Handler) logExportStatusRequest(c *gin.Context, userID uuid.UUID, exportID string) {
	h.LogBusinessEvent(c, "export_status_requested", userID.String(), exportID, map[string]interface{}{
		"export_id": exportID,
	})
}

func (h *Handler) logDownloadExportRequest(c *gin.Context, userID uuid.UUID, exportID string) {
	h.LogBusinessEvent(c, "export_download_requested", userID.String(), exportID, map[string]interface{}{
		"export_id": exportID,
	})
}
