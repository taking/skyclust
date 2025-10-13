package routes

import (
	"skyclust/internal/delivery/http"
	"github.com/gin-gonic/gin"
)

func SetupExportRoutes(router *gin.RouterGroup, exportHandler *http.ExportHandler) {
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
