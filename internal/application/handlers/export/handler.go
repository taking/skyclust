package export

import (
	"context"
	"fmt"
	"skyclust/internal/domain"
	exportservice "skyclust/internal/application/services/export"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/readability"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ExportStorage: 메모리에 내보내기 데이터와 상태를 저장하는 구조체
type ExportStorage struct {
	mu    sync.RWMutex
	store map[string]*ExportStorageItem
}

// ExportStorageItem: 저장된 내보내기 항목을 나타내는 구조체
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

// NewExportStorage: 새로운 메모리 내보내기 저장소를 생성합니다
func NewExportStorage() *ExportStorage {
	return &ExportStorage{
		store: make(map[string]*ExportStorageItem),
	}
}

// Store: 내보내기 항목을 저장합니다
func (s *ExportStorage) Store(item *ExportStorageItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[item.ID] = item
}

// Get: ID로 내보내기 항목을 조회합니다
func (s *ExportStorage) Get(id string) (*ExportStorageItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, exists := s.store[id]
	return item, exists
}

// GetByUserID: 사용자의 모든 내보내기를 조회합니다
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

// Delete: 내보내기 항목을 제거합니다
func (s *ExportStorage) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, id)
}

// Handler: 내보내기 작업을 처리하는 핸들러
type Handler struct {
	*handlers.BaseHandler
	readabilityHelper *readability.ReadabilityHelper
	exportService     *exportservice.Service
	exportStorage     *ExportStorage
}

// NewHandler: 새로운 내보내기 핸들러를 생성합니다
func NewHandler() *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("export"),
		readabilityHelper: readability.NewReadabilityHelper(),
		exportStorage:     NewExportStorage(),
	}
}

// SetExportService: 내보내기 서비스를 설정합니다 (의존성 주입용)
func (h *Handler) SetExportService(exportService *exportservice.Service) {
	h.exportService = exportService
}

// ExportData: 데이터 내보내기 요청을 처리합니다 (데코레이터 패턴 사용)
func (h *Handler) ExportData(c *gin.Context) {
	var req gin.H

	handler := h.Compose(
		h.exportDataHandler(req),
		h.StandardCRUDDecorators("export_data")...,
	)

	handler(c)
}

// exportDataHandler: 데이터 내보내기의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) exportDataHandler(req gin.H) handlers.HandlerFunc {
	return func(c *gin.Context) {
		var exportReq ExportRequest
		if err := h.ValidateRequest(c, &exportReq); err != nil {
			h.HandleError(c, err, "export_data")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "export_data")
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
			serviceReq := map[string]interface{}{
				"user_id":      userID.String(),
				"workspace_id": exportReq.WorkspaceID,
				"type":          exportReq.Type,
				"format":        exportReq.Format,
				"filters":       exportReq.Filters,
			}

			if exportReq.DateFrom != nil {
				serviceReq["date_from"] = exportReq.DateFrom
			}
			if exportReq.DateTo != nil {
				serviceReq["date_to"] = exportReq.DateTo
			}

			// Convert map to ExportRequest
			exportType := exportservice.ExportType(exportReq.Type)
			exportFormat := exportservice.ExportFormat(exportReq.Format)
			serviceExportReq := exportservice.ExportRequest{
				UserID:      serviceReq["user_id"].(string),
				WorkspaceID: serviceReq["workspace_id"].(string),
				Type:        exportType,
				Format:      exportFormat,
				Filters:     exportReq.Filters,
			}
			if dateFrom, ok := serviceReq["date_from"].(*time.Time); ok {
				serviceExportReq.DateFrom = dateFrom
			}
			if dateTo, ok := serviceReq["date_to"].(*time.Time); ok {
				serviceExportReq.DateTo = dateTo
			}

			// Perform export based on type
			var exportData []byte
			var result *exportservice.ExportResult
			var err error

			if h.exportService == nil {
				err = fmt.Errorf("export service not initialized")
			} else {
				// Call export service methods
				switch exportType {
				case exportservice.ExportTypeVMs:
					result, err = h.exportService.ExportVMs(ctx, serviceExportReq)
				case exportservice.ExportTypeWorkspaces:
					result, err = h.exportService.ExportWorkspaces(ctx, serviceExportReq)
				case exportservice.ExportTypeCredentials:
					result, err = h.exportService.ExportCredentials(ctx, serviceExportReq)
				case exportservice.ExportTypeAuditLogs:
					result, err = h.exportService.ExportAuditLogs(ctx, serviceExportReq)
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

// GetSupportedFormats: 지원되는 내보내기 형식을 반환합니다 (데코레이터 패턴 사용)
func (h *Handler) GetSupportedFormats(c *gin.Context) {
	handler := h.Compose(
		h.getSupportedFormatsHandler(),
		h.StandardCRUDDecorators("get_supported_formats")...,
	)

	handler(c)
}

// getSupportedFormatsHandler: 지원되는 형식 조회의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) getSupportedFormatsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		h.logSupportedFormatsRequest(c)

		formats := h.getSupportedFormats()

		h.OK(c, gin.H{
			"formats": formats,
		}, "Supported formats retrieved successfully")
	}
}

// GetExportHistory: 내보내기 이력을 조회합니다 (데코레이터 패턴 사용)
func (h *Handler) GetExportHistory(c *gin.Context) {
	handler := h.Compose(
		h.getExportHistoryHandler(),
		h.StandardCRUDDecorators("get_export_history")...,
	)

	handler(c)
}

// getExportHistoryHandler: 내보내기 이력 조회의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) getExportHistoryHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_export_history")
			return
		}

		h.logExportHistoryRequest(c, userID)

		// TODO: Implement export history retrieval
		history := h.getSampleExportHistory()

		h.OK(c, gin.H{
			"exports": history,
		}, "Export history retrieved successfully")
	}
}

// GetExportStatus: 내보내기 상태를 조회합니다 (데코레이터 패턴 사용)
func (h *Handler) GetExportStatus(c *gin.Context) {
	handler := h.Compose(
		h.getExportStatusHandler(),
		h.StandardCRUDDecorators("get_export_status")...,
	)

	handler(c)
}

// getExportStatusHandler: 내보내기 상태 조회의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) getExportStatusHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "export_data")
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

// GetExportFile: 내보내기 파일 다운로드를 처리합니다 (RESTful: GET /exports/:id/file)
// GET /exports/:id/download를 대체합니다
func (h *Handler) GetExportFile(c *gin.Context) {
	handler := h.Compose(
		h.getExportFileHandler(),
		h.StandardCRUDDecorators("get_export_file")...,
	)

	handler(c)
}

// getExportFileHandler: 내보내기 파일 다운로드의 핵심 비즈니스 로직을 처리합니다
func (h *Handler) getExportFileHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "export_data")
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

// parseExportID: 요청에서 내보내기 ID를 파싱합니다
func (h *Handler) parseExportID(c *gin.Context) string {
	exportID := c.Param("id")
	if exportID == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Export ID is required", 400), "parse_export_id")
		return ""
	}
	return exportID
}

// getSupportedFormats: 지원되는 내보내기 형식 목록을 반환합니다
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

// getSampleExportHistory: 샘플 내보내기 이력을 반환합니다
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

// logExportDataAttempt: 데이터 내보내기 시도 로그를 기록합니다
func (h *Handler) logExportDataAttempt(c *gin.Context, userID uuid.UUID, req gin.H) {
	h.LogBusinessEvent(c, "data_export_attempted", userID.String(), "", map[string]interface{}{
		"operation": "export_data",
		"request":   req,
	})
}

// logExportDataSuccess: 데이터 내보내기 성공 로그를 기록합니다
func (h *Handler) logExportDataSuccess(c *gin.Context, userID uuid.UUID, exportID string) {
	h.LogBusinessEvent(c, "data_export_initiated", userID.String(), exportID, map[string]interface{}{
		"operation": "export_data",
		"export_id": exportID,
	})
}

// logSupportedFormatsRequest: 지원 형식 조회 요청 로그를 기록합니다
func (h *Handler) logSupportedFormatsRequest(c *gin.Context) {
	h.LogBusinessEvent(c, "supported_formats_requested", "", "", map[string]interface{}{
		"operation": "get_supported_formats",
	})
}

// logExportHistoryRequest: 내보내기 이력 조회 요청 로그를 기록합니다
func (h *Handler) logExportHistoryRequest(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "export_history_requested", userID.String(), "", map[string]interface{}{
		"operation": "get_export_history",
	})
}

// logExportStatusRequest: 내보내기 상태 조회 요청 로그를 기록합니다
func (h *Handler) logExportStatusRequest(c *gin.Context, userID uuid.UUID, exportID string) {
	h.LogBusinessEvent(c, "export_status_requested", userID.String(), exportID, map[string]interface{}{
		"export_id": exportID,
	})
}

// logGetExportFileRequest: 내보내기 파일 다운로드 요청 로그를 기록합니다
func (h *Handler) logGetExportFileRequest(c *gin.Context, userID uuid.UUID, exportID string) {
	h.LogBusinessEvent(c, "export_file_requested", userID.String(), exportID, map[string]interface{}{
		"export_id": exportID,
	})
}
