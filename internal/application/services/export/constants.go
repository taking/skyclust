package export

// Export Service Constants
// These constants are specific to export operations

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

// Export operation constants
const (
	// MaxExportRecords is the maximum number of records that can be exported in a single request
	MaxExportRecords = 10000

	// DefaultExportFileSize is the default file size for placeholder exports (1MB)
	DefaultExportFileSize = 1024 * 1024
)

