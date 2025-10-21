package export

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up export routes
func SetupRoutes(router *gin.RouterGroup) {
	exportHandler := NewHandler()

	// Export data
	router.POST("", exportHandler.ExportData)

	// Get supported formats and types
	router.GET("/formats", exportHandler.GetSupportedFormats)

	// Export history
	router.GET("/history", exportHandler.GetExportHistory)

	// Export status and download
	router.GET("/:id/status", exportHandler.GetExportStatus)
	router.GET("/:id/download", exportHandler.DownloadExport)
}
