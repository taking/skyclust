package export

import (
	"context"
	"fmt"
	"skyclust/internal/domain"
	service "skyclust/internal/application/services"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/readability"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ExportStorage stores export data and status in memory
type ExportStorage struct {
	mu    sync.RWMutex
	store map[string]*ExportStorageItem
}

// ExportStorageItem represents a stored export
type ExportStorageItem struct {
	ID          string
	UserID      string
	Type        string
	Format      string
	Status      string // pending, processing, completed, failed
	Data        []byte
	FileName    string
	FileSize    int64
	Error       string
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// NewExportStorage creates a new in-memory export storage
func NewExportStorage() *ExportStorage {
	return &ExportStorage{
		store: make(map[string]*ExportStorageItem),
	}
}

// Store stores an export item
func (s *ExportStorage) Store(item *ExportStorageItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[item.ID] = item
}

// Get retrieves an export item by ID
func (s *ExportStorage) Get(id string) (*ExportStorageItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, exists := s.store[id]
	return item, exists
}

// GetByUserID retrieves all exports for a user
func (s *ExportStorage) GetByUserID(userID string) []*ExportStorageItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var items []*ExportStorageItem
	for _, item := range s.store {
		if item.UserID == userID {
			items = append(items, item)
		}
	}
	return items
}

// Delete removes an export item
func (s *ExportStorage) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, id)
}

// Handler handles export operations using improved patterns
type Handler struct {
	*handlers.BaseHandler
	readabilityHelper *readability.ReadabilityHelper
	exportService     *service.ExportService
	exportStorage     *ExportStorage
}

// NewHandler creates a new export handler
func NewHandler() *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("export"),
		readabilityHelper: readability.NewReadabilityHelper(),
		exportStorage:     NewExportStorage(),
	}
}

// SetExportService sets the export service (for dependency injection)
func (h *Handler) SetExportService(exportService *service.ExportService) {
	h.exportService = exportService
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
		var exportReq ExportRequest
		if err := h.ValidateRequest(c, &exportReq); err != nil {
			h.HandleError(c, err, "export_data")
			return
		}

		userID := h.extractUserID(c)
		if userID == uuid.Nil {
			return
		}

		// Validate export request
		if exportReq.Type == "" {
			h.BadRequest(c, "type is required")
			return
		}
		if exportReq.Format == "" {
			h.BadRequest(c, "format is required")
			return
		}

		h.logExportDataAttempt(c, userID, gin.H{
			"type":   exportReq.Type,
			"format": exportReq.Format,
		})

		// Create export ID
		exportID := uuid.New().String()

		// Store initial export status
		storageItem := &ExportStorageItem{
			ID:        exportID,
			UserID:    userID.String(),
			Type:      exportReq.Type,
			Format:    string(exportReq.Format),
			Status:    "processing",
			CreatedAt: time.Now(),
		}
		h.exportStorage.Store(storageItem)

		// Process export asynchronously
		go func() {
			ctx := context.Background()
			serviceReq := service.ExportRequest{
				UserID:      userID.String(),
				WorkspaceID: exportReq.WorkspaceID,
				Type:        service.ExportType(exportReq.Type),
				Format:      service.ExportFormat(exportReq.Format),
				Filters:     exportReq.Filters,
			}

			if exportReq.DateFrom != nil {
				serviceReq.DateFrom = exportReq.DateFrom
			}
			if exportReq.DateTo != nil {
				serviceReq.DateTo = exportReq.DateTo
			}

			// Perform export based on type
			var exportData []byte
			var result *service.ExportResult
			var err error

			if h.exportService == nil {
				err = fmt.Errorf("export service not initialized")
			} else {
				// Call export service methods
				// Note: ExportService methods create data internally, but don't return it
				// For now, we'll store the result without actual data
				// In a real implementation, ExportService would return both result and data
				switch serviceReq.Type {
				case service.ExportTypeVMs:
					result, err = h.exportService.ExportVMs(ctx, serviceReq)
				case service.ExportTypeWorkspaces:
					result, err = h.exportService.ExportWorkspaces(ctx, serviceReq)
				case service.ExportTypeCredentials:
					result, err = h.exportService.ExportCredentials(ctx, serviceReq)
				case service.ExportTypeAuditLogs:
					result, err = h.exportService.ExportAuditLogs(ctx, serviceReq)
				default:
					err = fmt.Errorf("unsupported export type: %s", exportReq.Type)
				}

				// For now, we'll store a placeholder for data
				// TODO: Modify ExportService to return actual data
				if err == nil && result != nil {
					// Create placeholder data
					exportData = []byte(fmt.Sprintf("Export data for %s in %s format (size: %d bytes)", exportReq.Type, exportReq.Format, result.FileSize))
				}
			}

			// Update storage with result
			now := time.Now()
			item, exists := h.exportStorage.Get(exportID)
			if exists {
				if err != nil {
					item.Status = "failed"
					item.Error = err.Error()
				} else if result != nil {
					item.Status = "completed"
					item.FileSize = result.FileSize
					item.Data = exportData
					item.CompletedAt = &now
					item.FileName = fmt.Sprintf("%s_export_%s.%s", exportReq.Type, now.Format("20060102_150405"), exportReq.Format)
				}
				h.exportStorage.Store(item)
			}
		}()

		h.logExportDataSuccess(c, userID, exportID)
		h.OK(c, gin.H{
			"id":      exportID,
			"status":  "processing",
			"message": "Export initiated successfully",
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
		if userID == uuid.Nil {
			return
		}

		exportID := h.parseExportID(c)
		if exportID == "" {
			return
		}

		h.logExportStatusRequest(c, userID, exportID)

		// Retrieve export status from storage
		item, exists := h.exportStorage.Get(exportID)
		if !exists {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "Export not found", 404), "get_export_status")
			return
		}

		// Verify ownership
		if item.UserID != userID.String() {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeForbidden, "You don't have access to this export", 403), "get_export_status")
			return
		}

		// Calculate progress
		progress := 0
		if item.Status == "completed" {
			progress = 100
		} else if item.Status == "failed" {
			progress = 0
		} else if item.Status == "processing" {
			progress = 50 // Assume 50% when processing
		}

		response := gin.H{
			"id":         item.ID,
			"user_id":    item.UserID,
			"type":       item.Type,
			"format":     item.Format,
			"status":     item.Status,
			"progress":   progress,
			"file_size":  item.FileSize,
			"created_at": item.CreatedAt,
		}

		if item.CompletedAt != nil {
			response["completed_at"] = item.CompletedAt
		}
		if item.Error != "" {
			response["error"] = item.Error
		}
		if item.Status == "completed" {
			response["download_url"] = fmt.Sprintf("/api/v1/exports/%s/download", item.ID)
		}

		h.OK(c, response, "Export status retrieved successfully")
	}
}

// GetExportFile handles export file download (RESTful: GET /exports/:id/file)
// This replaces GET /exports/:id/download
func (h *Handler) GetExportFile(c *gin.Context) {
	handler := h.Compose(
		h.getExportFileHandler(),
		h.StandardCRUDDecorators("get_export_file")...,
	)

	handler(c)
}

// getExportFileHandler is the core business logic for getting export file
func (h *Handler) getExportFileHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID := h.extractUserID(c)
		if userID == uuid.Nil {
			return
		}

		exportID := h.parseExportID(c)
		if exportID == "" {
			return
		}

		h.logGetExportFileRequest(c, userID, exportID)

		// Retrieve export from storage
		item, exists := h.exportStorage.Get(exportID)
		if !exists {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "Export not found", 404), "get_export_file")
			return
		}

		// Verify ownership
		if item.UserID != userID.String() {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeForbidden, "You don't have access to this export", 403), "get_export_file")
			return
		}

		// Check if export is completed
		if item.Status != "completed" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("Export is not ready for download. Current status: %s", item.Status), 400), "get_export_file")
			return
		}

		// Regenerate export data if not stored (for now, we'll regenerate it)
		if len(item.Data) == 0 {
			// In a real implementation, you would retrieve the actual export data
			// For now, we'll return an error indicating data is not available
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeInternalError, "Export data not available. Please regenerate the export.", 500), "get_export_file")
			return
		}

		// Set content type based on format
		contentType := "application/octet-stream"
		switch item.Format {
		case "json":
			contentType = "application/json"
		case "csv":
			contentType = "text/csv; charset=utf-8"
		case "xlsx":
			contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		}

		// Set headers for file download
		filename := item.FileName
		if filename == "" {
			filename = fmt.Sprintf("export_%s.%s", exportID, item.Format)
		}

		c.Header("Content-Type", contentType)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Length", fmt.Sprintf("%d", len(item.Data)))

		// Send file data
		c.Data(200, contentType, item.Data)
	}
}

// Helper methods for better readability

func (h *Handler) extractUserID(c *gin.Context) uuid.UUID {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "User not authenticated", 401), "extract_user_id")
		return uuid.Nil
	}
	
	// Convert to uuid.UUID (handle both string and uuid.UUID types)
	switch v := userIDValue.(type) {
	case uuid.UUID:
		return v
	case string:
		parsedUserID, err := uuid.Parse(v)
		if err != nil {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "Invalid user ID format", 401), "extract_user_id")
			return uuid.Nil
		}
		return parsedUserID
	default:
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeUnauthorized, "Invalid user ID type", 401), "extract_user_id")
		return uuid.Nil
	}
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

func (h *Handler) logGetExportFileRequest(c *gin.Context, userID uuid.UUID, exportID string) {
	h.LogBusinessEvent(c, "export_file_requested", userID.String(), exportID, map[string]interface{}{
		"export_id": exportID,
	})
}
