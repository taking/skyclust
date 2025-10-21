package services

// ExportService defines the interface for export operations
type ExportService interface {
	// ExportData exports data in the specified format
	ExportData(userID string, dataType string, format string) (interface{}, error)

	// GetExportJob retrieves an export job by ID
	GetExportJob(id string) (interface{}, error)

	// GetExportJobs retrieves export jobs for a user
	GetExportJobs(userID string, limit, offset int) ([]interface{}, error)

	// GetExportJobStatus retrieves the status of an export job
	GetExportJobStatus(id string) (string, error)

	// DownloadExport downloads the result of an export job
	DownloadExport(id string) ([]byte, error)

	// DeleteExportJob deletes an export job
	DeleteExportJob(id string) error

	// GetSupportedFormats returns supported export formats
	GetSupportedFormats() ([]string, error)

	// GetSupportedDataTypes returns supported data types for export
	GetSupportedDataTypes() ([]string, error)
}
