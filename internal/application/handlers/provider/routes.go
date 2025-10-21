package provider

import (
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up cloud provider routes
func SetupRoutes(router *gin.RouterGroup, providerManager interface{}, auditLogRepo domain.AuditLogRepository) {
	providerHandler := NewHandler(providerManager, auditLogRepo)

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
