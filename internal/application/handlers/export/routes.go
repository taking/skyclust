package export

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up export routes
func SetupRoutes(router *gin.RouterGroup) {
	exportHandler := NewHandler()
	SetupRoutesWithHandler(router, exportHandler)
}

// SetupRoutesWithHandler sets up export routes with a provided handler
func SetupRoutesWithHandler(router *gin.RouterGroup, handler *Handler) {
	// Export data
	router.POST("", handler.ExportData)

	// Get supported formats and types
	router.GET("/formats", handler.GetSupportedFormats)

	// Export history
	router.GET("/history", handler.GetExportHistory)

	// Export status and file download (RESTful)
	router.GET("/:id", handler.GetExportStatus)        // GET /exports/:id (status)
	router.GET("/:id/file", handler.GetExportFile)      // GET /exports/:id/file (download file)
}
