package export

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up export routes
func SetupRoutes(router *gin.RouterGroup) {
	exportHandler := NewHandler()

	exports := router.Group("/exports")
	{
		// Export data
		exports.POST("", exportHandler.ExportData)

		// Get supported formats and types
		exports.GET("/formats", exportHandler.GetSupportedFormats)

		// Export history
		exports.GET("/history", exportHandler.GetExportHistory)

		// Export status and download
		exports.GET("/:id/status", exportHandler.GetExportStatus)
		exports.GET("/:id/download", exportHandler.DownloadExport)
	}
}
