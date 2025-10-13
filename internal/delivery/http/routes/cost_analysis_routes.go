package routes

import (
	"skyclust/internal/delivery/http"
	"github.com/gin-gonic/gin"
)

func SetupCostAnalysisRoutes(router *gin.RouterGroup, costAnalysisHandler *http.CostAnalysisHandler) {
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
