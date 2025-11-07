package cost_analysis

import (
	costanalysisservice "skyclust/internal/application/services/cost_analysis"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler handles cost analysis operations
type Handler struct {
	*handlers.BaseHandler
	costAnalysisService *costanalysisservice.Service
}

// NewHandler creates a new cost analysis handler
func NewHandler(costAnalysisService *costanalysisservice.Service) *Handler {
	return &Handler{
		BaseHandler:         handlers.NewBaseHandler("cost_analysis"),
		costAnalysisService: costAnalysisService,
	}
}

// GetCostSummary retrieves cost summary for a workspace
func (h *Handler) GetCostSummary(c *gin.Context) {
	handler := h.Compose(
		h.getCostSummaryHandler(),
		h.StandardCRUDDecorators("get_cost_summary")...,
	)

	handler(c)
}

// getCostSummaryHandler is the core business logic for getting cost summary
func (h *Handler) getCostSummaryHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		workspaceIDStr := c.Param("workspaceId")
		if workspaceIDStr == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Workspace ID is required", 400), "get_cost_summary")
			return
		}

		// Get period from query parameter (default: 30d)
		period := c.DefaultQuery("period", "30d")

		// Get resource types from query parameter (default: all)
		resourceTypes := c.DefaultQuery("resource_types", "all")

		// Get cost summary from service
		summary, err := h.costAnalysisService.GetCostSummary(c.Request.Context(), workspaceIDStr, period, resourceTypes)
		if err != nil {
			h.HandleError(c, err, "get_cost_summary")
			return
		}

		// Convert to response format
		response := gin.H{
			"workspace_id": workspaceIDStr,
			"total_cost":   summary.TotalCost,
			"currency":     summary.Currency,
			"period":       summary.Period,
			"start_date":   summary.StartDate,
			"end_date":     summary.EndDate,
			"by_provider":  summary.ByProvider,
		}

		// Include warnings if any
		if len(summary.Warnings) > 0 {
			response["warnings"] = summary.Warnings
		}

		h.OK(c, response, "Cost summary retrieved successfully")
	}
}

// GetCostPredictions retrieves cost predictions for a workspace
func (h *Handler) GetCostPredictions(c *gin.Context) {
	handler := h.Compose(
		h.getCostPredictionsHandler(),
		h.StandardCRUDDecorators("get_cost_predictions")...,
	)

	handler(c)
}

// getCostPredictionsHandler is the core business logic for getting cost predictions
func (h *Handler) getCostPredictionsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		if workspaceID == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Workspace ID is required", 400), "get_cost_predictions")
			return
		}

		// Get days from query parameter (default: 30)
		days := 30
		if daysStr := c.Query("days"); daysStr != "" {
			parsedDays, err := strconv.Atoi(daysStr)
			if err != nil || parsedDays < 1 || parsedDays > 365 {
				h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid days parameter (must be 1-365)", 400), "get_cost_predictions")
				return
			}
			days = parsedDays
		}

		// Get resource types from query parameter (default: all)
		resourceTypes := c.DefaultQuery("resource_types", "all")

		// Get cost predictions from service
		predictions, warnings, err := h.costAnalysisService.GetCostPredictions(c.Request.Context(), workspaceID, days, resourceTypes)
		if err != nil {
			h.HandleError(c, err, "get_cost_predictions")
			return
		}

		// Convert to response format
		response := gin.H{
			"workspace_id": workspaceID,
			"predictions":  predictions,
		}

		// Include warnings if any
		if len(warnings) > 0 {
			response["warnings"] = warnings
		}

		h.OK(c, response, "Cost predictions retrieved successfully")
	}
}

// GetBudgetAlerts retrieves budget alerts for a workspace
func (h *Handler) GetBudgetAlerts(c *gin.Context) {
	handler := h.Compose(
		h.getBudgetAlertsHandler(),
		h.StandardCRUDDecorators("get_budget_alerts")...,
	)

	handler(c)
}

// getBudgetAlertsHandler is the core business logic for getting budget alerts
func (h *Handler) getBudgetAlertsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		if workspaceID == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Workspace ID is required", 400), "get_budget_alerts")
			return
		}

		// Get budget limit from query parameter (optional)
		budgetLimit := 0.0
		if budgetLimitStr := c.Query("budget_limit"); budgetLimitStr != "" {
			parsedLimit, err := strconv.ParseFloat(budgetLimitStr, 64)
			if err != nil || parsedLimit < 0 {
				h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid budget_limit parameter", 400), "get_budget_alerts")
				return
			}
			budgetLimit = parsedLimit
		}

		var alerts []costanalysisservice.BudgetAlert
		var err error

		if budgetLimit > 0 {
			// Check budget alerts if limit is provided
			alerts, err = h.costAnalysisService.CheckBudgetAlerts(c.Request.Context(), workspaceID, budgetLimit)
			if err != nil {
				h.HandleError(c, err, "get_budget_alerts")
				return
			}
		} else {
			// Return empty alerts if no budget limit provided
			alerts = []costanalysisservice.BudgetAlert{}
		}

		// Convert to response format
		response := gin.H{
			"workspace_id": workspaceID,
			"alerts":       alerts,
		}

		h.OK(c, response, "Budget alerts retrieved successfully")
	}
}

// GetCostTrend retrieves cost trend for a workspace
func (h *Handler) GetCostTrend(c *gin.Context) {
	handler := h.Compose(
		h.getCostTrendHandler(),
		h.StandardCRUDDecorators("get_cost_trend")...,
	)

	handler(c)
}

// getCostTrendHandler is the core business logic for getting cost trend
func (h *Handler) getCostTrendHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		if workspaceID == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Workspace ID is required", 400), "get_cost_trend")
			return
		}

		// Get period from query parameter (default: 90d)
		period := c.DefaultQuery("period", "90d")

		// Get resource types from query parameter (default: all)
		resourceTypes := c.DefaultQuery("resource_types", "all")

		// Get cost summary to calculate trend
		summary, err := h.costAnalysisService.GetCostSummary(c.Request.Context(), workspaceID, period, resourceTypes)
		if err != nil {
			h.HandleError(c, err, "get_cost_trend")
			return
		}

		// Get cost trend from service
		trend, err := h.costAnalysisService.GetCostTrend(c.Request.Context(), workspaceID, period, resourceTypes)
		if err != nil {
			h.HandleError(c, err, "get_cost_trend")
			return
		}

		// Convert to response format
		response := gin.H{
			"workspace_id":      workspaceID,
			"trend":             trend.DailyCosts,
			"trend_direction":   trend.TrendDirection,
			"change_percentage": trend.ChangePercentage,
			"total_cost":        summary.TotalCost,
			"currency":          summary.Currency,
			"period":            summary.Period,
			"start_date":        summary.StartDate,
			"end_date":          summary.EndDate,
		}

		// Include warnings if any
		if len(trend.Warnings) > 0 {
			response["warnings"] = trend.Warnings
		}

		h.OK(c, response, "Cost trend retrieved successfully")
	}
}

// GetCostBreakdown retrieves cost breakdown for a workspace
func (h *Handler) GetCostBreakdown(c *gin.Context) {
	handler := h.Compose(
		h.getCostBreakdownHandler(),
		h.StandardCRUDDecorators("get_cost_breakdown")...,
	)

	handler(c)
}

// getCostBreakdownHandler is the core business logic for getting cost breakdown
func (h *Handler) getCostBreakdownHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		if workspaceID == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Workspace ID is required", 400), "get_cost_breakdown")
			return
		}

		// Get period from query parameter (default: 30d)
		period := c.DefaultQuery("period", "30d")

		// Get dimension from query parameter (default: service)
		dimension := c.DefaultQuery("dimension", "service")

		// Get resource types from query parameter (default: all)
		resourceTypes := c.DefaultQuery("resource_types", "all")

		// Get cost breakdown from service
		breakdown, warnings, err := h.costAnalysisService.GetCostBreakdown(c.Request.Context(), workspaceID, period, dimension, resourceTypes)
		if err != nil {
			h.HandleError(c, err, "get_cost_breakdown")
			return
		}

		// Convert to response format
		response := gin.H{
			"workspace_id": workspaceID,
			"breakdown":    breakdown,
		}

		// Include warnings if any
		if len(warnings) > 0 {
			response["warnings"] = warnings
		}

		h.OK(c, response, "Cost breakdown retrieved successfully")
	}
}

// GetCostComparison retrieves cost comparison for a workspace
func (h *Handler) GetCostComparison(c *gin.Context) {
	handler := h.Compose(
		h.getCostComparisonHandler(),
		h.StandardCRUDDecorators("get_cost_comparison")...,
	)

	handler(c)
}

// getCostComparisonHandler is the core business logic for getting cost comparison
func (h *Handler) getCostComparisonHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		if workspaceID == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Workspace ID is required", 400), "get_cost_comparison")
			return
		}

		// Get periods from query parameters
		currentPeriod := c.DefaultQuery("current_period", "30d")
		comparePeriod := c.DefaultQuery("compare_period", "30d")

		// Get cost comparison from service
		comparison, err := h.costAnalysisService.GetCostComparison(c.Request.Context(), workspaceID, currentPeriod, comparePeriod)
		if err != nil {
			h.HandleError(c, err, "get_cost_comparison")
			return
		}

		// Convert to response format
		response := gin.H{
			"workspace_id": workspaceID,
			"comparison":   comparison,
		}

		h.OK(c, response, "Cost comparison retrieved successfully")
	}
}
