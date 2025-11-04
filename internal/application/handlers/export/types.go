package export

import "time"

// ExportRequest represents an export request
type ExportRequest struct {
	Type        string                 `json:"type" validate:"required,oneof=vms workspaces credentials audit_logs"`
	Format      string                 `json:"format" validate:"required,oneof=csv json xlsx"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	WorkspaceID string                 `json:"workspace_id,omitempty"`
	DateFrom    *time.Time             `json:"date_from,omitempty"`
	DateTo      *time.Time             `json:"date_to,omitempty"`
}

// ExportResponse represents an export response
type ExportResponse struct {
	ExportID string `json:"export_id"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

// ExportStatusResponse represents export status information
type ExportStatusResponse struct {
	ExportID    string     `json:"export_id"`
	Status      string     `json:"status"`
	Progress    int        `json:"progress"`
	FileURL     string     `json:"file_url,omitempty"`
	Error       string     `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// ExportHistoryResponse represents export history
type ExportHistoryResponse struct {
	Exports []*ExportStatusResponse `json:"exports"`
	Total   int64                   `json:"total"`
}

// SupportedFormatResponse represents a supported export format
type SupportedFormatResponse struct {
	Format      string `json:"format"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mime_type"`
}

// SupportedFormatsResponse represents supported export formats
type SupportedFormatsResponse struct {
	Formats []*SupportedFormatResponse `json:"formats"`
}
