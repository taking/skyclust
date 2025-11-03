package export

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"skyclust/internal/domain"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

type Service struct {
	logger         *zap.Logger
	vmRepo         domain.VMRepository
	workspaceRepo  domain.WorkspaceRepository
	credentialRepo domain.CredentialRepository
	auditLogRepo   domain.AuditLogRepository
}

func NewService(
	logger *zap.Logger,
	vmRepo domain.VMRepository,
	workspaceRepo domain.WorkspaceRepository,
	credentialRepo domain.CredentialRepository,
	auditLogRepo domain.AuditLogRepository,
) *Service {
	return &Service{
		logger:         logger,
		vmRepo:         vmRepo,
		workspaceRepo:  workspaceRepo,
		credentialRepo: credentialRepo,
		auditLogRepo:   auditLogRepo,
	}
}

// ExportFormat represents supported export formats
type ExportFormat string

const (
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatJSON ExportFormat = "json"
	ExportFormatXLSX ExportFormat = "xlsx"
	ExportFormatPDF  ExportFormat = "pdf"
)

// ExportType represents types of data that can be exported
type ExportType string

const (
	ExportTypeVMs         ExportType = "vms"
	ExportTypeWorkspaces  ExportType = "workspaces"
	ExportTypeCredentials ExportType = "credentials"
	ExportTypeAuditLogs   ExportType = "audit_logs"
	ExportTypeCosts       ExportType = "costs"
)

// ExportRequest represents an export request
type ExportRequest struct {
	UserID         string                 `json:"user_id"`
	WorkspaceID    string                 `json:"workspace_id,omitempty"`
	Type           ExportType             `json:"type"`
	Format         ExportFormat           `json:"format"`
	Filters        map[string]interface{} `json:"filters,omitempty"`
	DateFrom       *time.Time             `json:"date_from,omitempty"`
	DateTo         *time.Time             `json:"date_to,omitempty"`
	IncludeDeleted bool                   `json:"include_deleted,omitempty"`
}

// ExportResult represents the result of an export operation
type ExportResult struct {
	ID          string       `json:"id"`
	UserID      string       `json:"user_id"`
	Type        ExportType   `json:"type"`
	Format      ExportFormat `json:"format"`
	Status      string       `json:"status"` // pending, processing, completed, failed
	FileSize    int64        `json:"file_size,omitempty"`
	DownloadURL string       `json:"download_url,omitempty"`
	Error       string       `json:"error,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
}

// ExportVMs exports VM data
func (s *Service) ExportVMs(ctx context.Context, req ExportRequest) (*ExportResult, error) {
	// Get VMs based on filters
	vms, err := s.getVMsForExport(ctx, req)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VMs: %v", err), 500)
	}

	// Convert to export format
	var data []byte
	var filename string
	_ = filename // Suppress unused variable warning

	switch req.Format {
	case ExportFormatCSV:
		data, filename, err = s.exportVMsToCSV(vms, req)
	case ExportFormatJSON:
		data, filename, err = s.exportVMsToJSON(vms, req)
	case ExportFormatXLSX:
		data, filename, err = s.exportVMsToXLSX(vms, req)
	case ExportFormatPDF:
		data, filename, err = s.exportVMsToPDF(vms, req)
	default:
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported format: %s", req.Format), 400)
	}

	_ = filename // Suppress unused variable warning

	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to export VMs: %v", err), 500)
	}

	// Create export result
	result := &ExportResult{
		ID:        fmt.Sprintf("export-%d", time.Now().Unix()),
		UserID:    req.UserID,
		Type:      req.Type,
		Format:    req.Format,
		Status:    "completed",
		FileSize:  int64(len(data)),
		CreatedAt: time.Now(),
	}

	// In a real implementation, you would save the file and return a download URL
	result.DownloadURL = fmt.Sprintf("/api/v1/exports/%s/download", result.ID)

	s.logger.Info("VMs exported successfully",
		zap.String("user_id", req.UserID),
		zap.String("format", string(req.Format)),
		zap.Int("count", len(vms)))

	return result, nil
}

// ExportWorkspaces exports workspace data
func (s *Service) ExportWorkspaces(ctx context.Context, req ExportRequest) (*ExportResult, error) {
	// Get workspaces based on filters
	workspaces, err := s.getWorkspacesForExport(ctx, req)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspaces: %v", err), 500)
	}

	// Convert to export format
	var data []byte

	switch req.Format {
	case ExportFormatCSV:
		data, _, err = s.exportWorkspacesToCSV(workspaces, req)
	case ExportFormatJSON:
		data, _, err = s.exportWorkspacesToJSON(workspaces, req)
	case ExportFormatXLSX:
		data, _, err = s.exportWorkspacesToXLSX(workspaces, req)
	case ExportFormatPDF:
		data, _, err = s.exportWorkspacesToPDF(workspaces, req)
	default:
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported format: %s", req.Format), 400)
	}

	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to export workspaces: %v", err), 500)
	}

	result := &ExportResult{
		ID:        fmt.Sprintf("export-%d", time.Now().Unix()),
		UserID:    req.UserID,
		Type:      req.Type,
		Format:    req.Format,
		Status:    "completed",
		FileSize:  int64(len(data)),
		CreatedAt: time.Now(),
	}

	result.DownloadURL = fmt.Sprintf("/api/v1/exports/%s/download", result.ID)

	s.logger.Info("Workspaces exported successfully",
		zap.String("user_id", req.UserID),
		zap.String("format", string(req.Format)),
		zap.Int("count", len(workspaces)))

	return result, nil
}

// ExportCredentials exports credential data
func (s *Service) ExportCredentials(ctx context.Context, req ExportRequest) (*ExportResult, error) {
	// Get credentials based on filters
	credentials, err := s.getCredentialsForExport(ctx, req)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get credentials: %v", err), 500)
	}

	// Convert to export format
	var data []byte

	switch req.Format {
	case ExportFormatCSV:
		data, _, err = s.exportCredentialsToCSV(credentials, req)
	case ExportFormatJSON:
		data, _, err = s.exportCredentialsToJSON(credentials, req)
	case ExportFormatXLSX:
		data, _, err = s.exportCredentialsToXLSX(credentials, req)
	case ExportFormatPDF:
		data, _, err = s.exportCredentialsToPDF(credentials, req)
	default:
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported format: %s", req.Format), 400)
	}

	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to export credentials: %v", err), 500)
	}

	result := &ExportResult{
		ID:        fmt.Sprintf("export-%d", time.Now().Unix()),
		UserID:    req.UserID,
		Type:      req.Type,
		Format:    req.Format,
		Status:    "completed",
		FileSize:  int64(len(data)),
		CreatedAt: time.Now(),
	}

	result.DownloadURL = fmt.Sprintf("/api/v1/exports/%s/download", result.ID)

	s.logger.Info("Credentials exported successfully",
		zap.String("user_id", req.UserID),
		zap.String("format", string(req.Format)),
		zap.Int("count", len(credentials)))

	return result, nil
}

// ExportAuditLogs exports audit log data
func (s *Service) ExportAuditLogs(ctx context.Context, req ExportRequest) (*ExportResult, error) {
	// Get audit logs based on filters
	auditLogs, err := s.getAuditLogsForExport(ctx, req)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get audit logs: %v", err), 500)
	}

	// Convert to export format
	var data []byte
	var filename string
	_ = filename // Suppress unused variable warning

	switch req.Format {
	case ExportFormatCSV:
		data, _, err = s.exportAuditLogsToCSV(auditLogs, req)
	case ExportFormatJSON:
		data, _, err = s.exportAuditLogsToJSON(auditLogs, req)
	case ExportFormatXLSX:
		data, _, err = s.exportAuditLogsToXLSX(auditLogs, req)
	case ExportFormatPDF:
		data, _, err = s.exportAuditLogsToPDF(auditLogs, req)
	default:
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported format: %s", req.Format), 400)
	}

	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to export audit logs: %v", err), 500)
	}

	result := &ExportResult{
		ID:        fmt.Sprintf("export-%d", time.Now().Unix()),
		UserID:    req.UserID,
		Type:      req.Type,
		Format:    req.Format,
		Status:    "completed",
		FileSize:  int64(len(data)),
		CreatedAt: time.Now(),
	}

	result.DownloadURL = fmt.Sprintf("/api/v1/exports/%s/download", result.ID)

	s.logger.Info("Audit logs exported successfully",
		zap.String("user_id", req.UserID),
		zap.String("format", string(req.Format)),
		zap.Int("count", len(auditLogs)))

	return result, nil
}

// GetExportStatus retrieves the status of an export
func (s *Service) GetExportStatus(ctx context.Context, exportID string) (*ExportResult, error) {
	// This would typically query a database
	// For now, return a mock result
	return &ExportResult{
		ID:          exportID,
		Status:      "completed",
		FileSize:    DefaultExportFileSize,
		DownloadURL: fmt.Sprintf("/api/v1/exports/%s/download", exportID),
		CreatedAt:   time.Now().Add(-5 * time.Minute),
		CompletedAt: &[]time.Time{time.Now().Add(-4 * time.Minute)}[0],
	}, nil
}

// Helper methods

func (s *Service) getVMsForExport(ctx context.Context, req ExportRequest) ([]*domain.VM, error) {
	if req.WorkspaceID != "" {
		return s.vmRepo.GetVMsByWorkspace(ctx, req.WorkspaceID)
	}

	// Get all VMs for user (this would require a user-specific query)
	return []*domain.VM{}, nil
}

func (s *Service) getWorkspacesForExport(ctx context.Context, req ExportRequest) ([]*domain.Workspace, error) {
	if req.WorkspaceID != "" {
		workspace, err := s.workspaceRepo.GetByID(ctx, req.WorkspaceID)
		if err != nil {
			return nil, err
		}
		if workspace == nil {
			return []*domain.Workspace{}, nil
		}
		return []*domain.Workspace{workspace}, nil
	}

	return s.workspaceRepo.GetUserWorkspaces(ctx, req.UserID)
}

func (s *Service) getCredentialsForExport(ctx context.Context, req ExportRequest) ([]*domain.Credential, error) {
	// This would typically query credentials for the user/workspace
	return []*domain.Credential{}, nil
}

func (s *Service) getAuditLogsForExport(ctx context.Context, req ExportRequest) ([]*domain.AuditLog, error) {
	// This would typically query audit logs with date filters
	return []*domain.AuditLog{}, nil
}

// CSV Export methods

func (s *Service) exportVMsToCSV(vms []*domain.VM, req ExportRequest) ([]byte, string, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"ID", "Name", "Provider", "Instance ID", "Status", "Type", "Region", "Image ID", "CPUs", "Memory (MB)", "Storage (GB)", "Created At", "Updated At"}
	if err := writer.Write(header); err != nil {
		return nil, "", err
	}

	// Write data
	for _, vm := range vms {
		record := []string{
			vm.ID,
			vm.Name,
			vm.Provider,
			vm.InstanceID,
			string(vm.Status),
			vm.Type,
			vm.Region,
			vm.ImageID,
			strconv.Itoa(vm.CPUs),
			strconv.Itoa(vm.Memory),
			strconv.Itoa(vm.Storage),
			vm.CreatedAt.Format(time.RFC3339),
			vm.UpdatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return nil, "", err
		}
	}

	writer.Flush()
	filename := fmt.Sprintf("vms_export_%s.csv", time.Now().Format("20060102_150405"))
	return []byte(buf.String()), filename, nil
}

func (s *Service) exportWorkspacesToCSV(workspaces []*domain.Workspace, req ExportRequest) ([]byte, string, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"ID", "Name", "Description", "Owner ID", "Is Active", "Created At", "Updated At"}
	if err := writer.Write(header); err != nil {
		return nil, "", err
	}

	// Write data
	for _, workspace := range workspaces {
		record := []string{
			workspace.ID,
			workspace.Name,
			workspace.Description,
			workspace.OwnerID,
			strconv.FormatBool(workspace.IsActive()),
			workspace.CreatedAt.Format(time.RFC3339),
			workspace.UpdatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return nil, "", err
		}
	}

	writer.Flush()
	filename := fmt.Sprintf("workspaces_export_%s.csv", time.Now().Format("20060102_150405"))
	return []byte(buf.String()), filename, nil
}

func (s *Service) exportCredentialsToCSV(credentials []*domain.Credential, req ExportRequest) ([]byte, string, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"ID", "Name", "Provider", "Type", "Is Active", "Created At", "Updated At"}
	if err := writer.Write(header); err != nil {
		return nil, "", err
	}

	// Write data
	for _, cred := range credentials {
		record := []string{
			cred.ID.String(),
			cred.Name,
			cred.Provider,
			"credential", // Default type since Type field doesn't exist
			strconv.FormatBool(cred.IsActive),
			cred.CreatedAt.Format(time.RFC3339),
			cred.UpdatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return nil, "", err
		}
	}

	writer.Flush()
	filename := fmt.Sprintf("credentials_export_%s.csv", time.Now().Format("20060102_150405"))
	return []byte(buf.String()), filename, nil
}

func (s *Service) exportAuditLogsToCSV(auditLogs []*domain.AuditLog, req ExportRequest) ([]byte, string, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"ID", "User ID", "Action", "Resource", "IP Address", "User Agent", "Created At"}
	if err := writer.Write(header); err != nil {
		return nil, "", err
	}

	// Write data
	for _, log := range auditLogs {
		record := []string{
			log.ID.String(),
			log.UserID.String(),
			log.Action,
			log.Resource,
			log.IPAddress,
			log.UserAgent,
			log.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return nil, "", err
		}
	}

	writer.Flush()
	filename := fmt.Sprintf("audit_logs_export_%s.csv", time.Now().Format("20060102_150405"))
	return []byte(buf.String()), filename, nil
}

// JSON Export methods

func (s *Service) exportVMsToJSON(vms []*domain.VM, req ExportRequest) ([]byte, string, error) {
	data, err := json.MarshalIndent(vms, "", "  ")
	if err != nil {
		return nil, "", err
	}
	filename := fmt.Sprintf("vms_export_%s.json", time.Now().Format("20060102_150405"))
	return data, filename, nil
}

func (s *Service) exportWorkspacesToJSON(workspaces []*domain.Workspace, req ExportRequest) ([]byte, string, error) {
	data, err := json.MarshalIndent(workspaces, "", "  ")
	if err != nil {
		return nil, "", err
	}
	filename := fmt.Sprintf("workspaces_export_%s.json", time.Now().Format("20060102_150405"))
	return data, filename, nil
}

func (s *Service) exportCredentialsToJSON(credentials []*domain.Credential, req ExportRequest) ([]byte, string, error) {
	data, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return nil, "", err
	}
	filename := fmt.Sprintf("credentials_export_%s.json", time.Now().Format("20060102_150405"))
	return data, filename, nil
}

func (s *Service) exportAuditLogsToJSON(auditLogs []*domain.AuditLog, req ExportRequest) ([]byte, string, error) {
	data, err := json.MarshalIndent(auditLogs, "", "  ")
	if err != nil {
		return nil, "", err
	}
	filename := fmt.Sprintf("audit_logs_export_%s.json", time.Now().Format("20060102_150405"))
	return data, filename, nil
}

// XLSX Export methods (placeholder implementations)

func (s *Service) exportVMsToXLSX(vms []*domain.VM, req ExportRequest) ([]byte, string, error) {
	// This would use a library like excelize to create XLSX files
	return []byte("XLSX export not implemented"), "vms_export.xlsx", nil
}

func (s *Service) exportWorkspacesToXLSX(workspaces []*domain.Workspace, req ExportRequest) ([]byte, string, error) {
	return []byte("XLSX export not implemented"), "workspaces_export.xlsx", nil
}

func (s *Service) exportCredentialsToXLSX(credentials []*domain.Credential, req ExportRequest) ([]byte, string, error) {
	return []byte("XLSX export not implemented"), "credentials_export.xlsx", nil
}

func (s *Service) exportAuditLogsToXLSX(auditLogs []*domain.AuditLog, req ExportRequest) ([]byte, string, error) {
	return []byte("XLSX export not implemented"), "audit_logs_export.xlsx", nil
}

// PDF Export methods (placeholder implementations)

func (s *Service) exportVMsToPDF(vms []*domain.VM, req ExportRequest) ([]byte, string, error) {
	// This would use a library like gofpdf to create PDF files
	return []byte("PDF export not implemented"), "vms_export.pdf", nil
}

func (s *Service) exportWorkspacesToPDF(workspaces []*domain.Workspace, req ExportRequest) ([]byte, string, error) {
	return []byte("PDF export not implemented"), "workspaces_export.pdf", nil
}

func (s *Service) exportCredentialsToPDF(credentials []*domain.Credential, req ExportRequest) ([]byte, string, error) {
	return []byte("PDF export not implemented"), "credentials_export.pdf", nil
}

func (s *Service) exportAuditLogsToPDF(auditLogs []*domain.AuditLog, req ExportRequest) ([]byte, string, error) {
	return []byte("PDF export not implemented"), "audit_logs_export.pdf", nil
}
