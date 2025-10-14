package routes

import (
	httpDelivery "skyclust/internal/delivery/http"
	"skyclust/internal/domain"
	"skyclust/internal/plugin"

	"github.com/gin-gonic/gin"
)

// SetupProviderRoutes sets up cloud provider routes
func SetupProviderRoutes(router *gin.RouterGroup, pluginManager *plugin.Manager, auditLogRepo domain.AuditLogRepository) {
	providerHandler := httpDelivery.NewProviderHandler(pluginManager, auditLogRepo)

	router.GET("", providerHandler.GetProviders)
	router.GET("/:name", providerHandler.GetProvider)
	router.GET("/:name/instances", providerHandler.GetInstances)
	router.GET("/:name/instances/:id", providerHandler.GetInstance)
	router.POST("/:name/instances", providerHandler.CreateInstance)
	router.DELETE("/:name/instances/:id", providerHandler.DeleteInstance)
	router.GET("/:name/regions", providerHandler.GetRegions)
	router.GET("/:name/cost-estimates", providerHandler.GetCostEstimates)
	router.POST("/:name/cost-estimates", providerHandler.CreateCostEstimate)
}
