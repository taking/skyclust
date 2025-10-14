package cost_analysis

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up cost analysis routes
func SetupRoutes(router *gin.RouterGroup) {
	costAnalysisHandler := NewHandler()

	cost := router.Group("/cost-analysis")
	{
		// Workspace-specific cost analysis
		cost.GET("/workspaces/:workspaceId/summary", costAnalysisHandler.GetCostSummary)
		cost.GET("/workspaces/:workspaceId/predictions", costAnalysisHandler.GetCostPredictions)
		cost.GET("/workspaces/:workspaceId/budget-alerts", costAnalysisHandler.GetBudgetAlerts)
		cost.GET("/workspaces/:workspaceId/trend", costAnalysisHandler.GetCostTrend)
		cost.GET("/workspaces/:workspaceId/breakdown", costAnalysisHandler.GetCostBreakdown)
		cost.GET("/workspaces/:workspaceId/comparison", costAnalysisHandler.GetCostComparison)
	}
}
