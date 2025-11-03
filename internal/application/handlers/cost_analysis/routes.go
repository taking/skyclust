package cost_analysis

import (
	costanalysisservice "skyclust/internal/application/services/cost_analysis"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up cost analysis routes
func SetupRoutes(router *gin.RouterGroup, costAnalysisService *costanalysisservice.Service) {
	costAnalysisHandler := NewHandler(costAnalysisService)

	// Workspace-specific cost analysis
	router.GET("/workspaces/:workspaceId/summary", costAnalysisHandler.GetCostSummary)
	router.GET("/workspaces/:workspaceId/predictions", costAnalysisHandler.GetCostPredictions)
	router.GET("/workspaces/:workspaceId/budget-alerts", costAnalysisHandler.GetBudgetAlerts)
	router.GET("/workspaces/:workspaceId/trend", costAnalysisHandler.GetCostTrend)
	router.GET("/workspaces/:workspaceId/breakdown", costAnalysisHandler.GetCostBreakdown)
	router.GET("/workspaces/:workspaceId/comparison", costAnalysisHandler.GetCostComparison)
}
