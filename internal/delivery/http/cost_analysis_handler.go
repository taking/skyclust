package http

import (
	"net/http"
	"skyclust/internal/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CostAnalysisHandler struct {
	logger              *zap.Logger
	costAnalysisService *usecase.CostAnalysisService
}

func NewCostAnalysisHandler(logger *zap.Logger, costAnalysisService *usecase.CostAnalysisService) *CostAnalysisHandler {
	return &CostAnalysisHandler{
		logger:              logger,
		costAnalysisService: costAnalysisService,
	}
}

// GetCostSummary retrieves cost summary for a workspace
func (h *CostAnalysisHandler) GetCostSummary(c *gin.Context) {
	workspaceID := c.Param("workspaceId")
	period := c.DefaultQuery("period", "30d")

	summary, err := h.costAnalysisService.GetCostSummary(c.Request.Context(), workspaceID, period)
	if err != nil {
		h.logger.Error("Failed to get cost summary",
			zap.String("workspace_id", workspaceID),
			zap.String("period", period),
			zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to get cost summary", "")
		return
	}

	SuccessResponse(c, http.StatusOK, summary, "Cost summary retrieved successfully")
}

// GetCostPredictions retrieves cost predictions for a workspace
func (h *CostAnalysisHandler) GetCostPredictions(c *gin.Context) {
	workspaceID := c.Param("workspaceId")
	daysStr := c.DefaultQuery("days", "30")

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 || days > 365 {
		BadRequestResponse(c, "Invalid days parameter. Must be between 1 and 365")
		return
	}

	predictions, err := h.costAnalysisService.GetCostPredictions(c.Request.Context(), workspaceID, days)
	if err != nil {
		h.logger.Error("Failed to get cost predictions",
			zap.String("workspace_id", workspaceID),
			zap.Int("days", days),
			zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to get cost predictions", "")
		return
	}

	SuccessResponse(c, http.StatusOK, predictions, "Cost predictions retrieved successfully")
}

// GetBudgetAlerts retrieves budget alerts for a workspace
func (h *CostAnalysisHandler) GetBudgetAlerts(c *gin.Context) {
	workspaceID := c.Param("workspaceId")
	budgetLimitStr := c.Query("budget_limit")

	if budgetLimitStr == "" {
		BadRequestResponse(c, "Budget limit is required")
		return
	}

	budgetLimit, err := strconv.ParseFloat(budgetLimitStr, 64)
	if err != nil || budgetLimit <= 0 {
		BadRequestResponse(c, "Invalid budget limit. Must be a positive number")
		return
	}

	alerts, err := h.costAnalysisService.CheckBudgetAlerts(c.Request.Context(), workspaceID, budgetLimit)
	if err != nil {
		h.logger.Error("Failed to get budget alerts",
			zap.String("workspace_id", workspaceID),
			zap.Float64("budget_limit", budgetLimit),
			zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to get budget alerts", "")
		return
	}

	SuccessResponse(c, http.StatusOK, alerts, "Budget alerts retrieved successfully")
}

// GetCostTrend retrieves cost trend analysis for a workspace
func (h *CostAnalysisHandler) GetCostTrend(c *gin.Context) {
	workspaceID := c.Param("workspaceId")
	period := c.DefaultQuery("period", "90d")

	summary, err := h.costAnalysisService.GetCostSummary(c.Request.Context(), workspaceID, period)
	if err != nil {
		h.logger.Error("Failed to get cost trend",
			zap.String("workspace_id", workspaceID),
			zap.String("period", period),
			zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to get cost trend", "")
		return
	}

	// Create trend response
	trendResponse := map[string]interface{}{
		"total_cost": summary.TotalCost,
		"currency":   summary.Currency,
		"period":     summary.Period,
		"start_date": summary.StartDate,
		"end_date":   summary.EndDate,
	}

	SuccessResponse(c, http.StatusOK, trendResponse, "Cost trend retrieved successfully")
}

// GetCostBreakdown retrieves cost breakdown by various dimensions
func (h *CostAnalysisHandler) GetCostBreakdown(c *gin.Context) {
	workspaceID := c.Param("workspaceId")
	period := c.DefaultQuery("period", "30d")
	dimension := c.DefaultQuery("dimension", "service") // service, provider, region, workspace

	summary, err := h.costAnalysisService.GetCostSummary(c.Request.Context(), workspaceID, period)
	if err != nil {
		h.logger.Error("Failed to get cost breakdown",
			zap.String("workspace_id", workspaceID),
			zap.String("period", period),
			zap.String("dimension", dimension),
			zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to get cost breakdown", "")
		return
	}

	var breakdown map[string]float64
	switch dimension {
	case "provider":
		breakdown = summary.ByProvider
	default:
		BadRequestResponse(c, "Invalid dimension. Must be one of: provider")
		return
	}

	// Convert to array format for easier charting
	var breakdownArray []map[string]interface{}
	for key, value := range breakdown {
		percentage := (value / summary.TotalCost) * 100
		breakdownArray = append(breakdownArray, map[string]interface{}{
			"name":       key,
			"value":      value,
			"percentage": percentage,
		})
	}

	response := map[string]interface{}{
		"dimension":  dimension,
		"total_cost": summary.TotalCost,
		"currency":   summary.Currency,
		"period":     summary.Period,
		"breakdown":  breakdownArray,
	}

	SuccessResponse(c, http.StatusOK, response, "Cost breakdown retrieved successfully")
}

// GetCostComparison compares costs between different periods
func (h *CostAnalysisHandler) GetCostComparison(c *gin.Context) {
	workspaceID := c.Param("workspaceId")
	currentPeriod := c.DefaultQuery("current_period", "30d")
	comparePeriod := c.DefaultQuery("compare_period", "30d")

	// Get current period costs
	currentSummary, err := h.costAnalysisService.GetCostSummary(c.Request.Context(), workspaceID, currentPeriod)
	if err != nil {
		h.logger.Error("Failed to get current period costs",
			zap.String("workspace_id", workspaceID),
			zap.String("current_period", currentPeriod),
			zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to get current period costs", "")
		return
	}

	// Get comparison period costs
	compareSummary, err := h.costAnalysisService.GetCostSummary(c.Request.Context(), workspaceID, comparePeriod)
	if err != nil {
		h.logger.Error("Failed to get comparison period costs",
			zap.String("workspace_id", workspaceID),
			zap.String("compare_period", comparePeriod),
			zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, "Failed to get comparison period costs", "")
		return
	}

	// Calculate comparison metrics
	var costChange float64
	var percentageChange float64

	if compareSummary.TotalCost > 0 {
		costChange = currentSummary.TotalCost - compareSummary.TotalCost
		percentageChange = (costChange / compareSummary.TotalCost) * 100
	}

	response := map[string]interface{}{
		"current_period": map[string]interface{}{
			"period":     currentSummary.Period,
			"total_cost": currentSummary.TotalCost,
			"start_date": currentSummary.StartDate,
			"end_date":   currentSummary.EndDate,
		},
		"compare_period": map[string]interface{}{
			"period":     compareSummary.Period,
			"total_cost": compareSummary.TotalCost,
			"start_date": compareSummary.StartDate,
			"end_date":   compareSummary.EndDate,
		},
		"comparison": map[string]interface{}{
			"cost_change":       costChange,
			"percentage_change": percentageChange,
		},
	}

	SuccessResponse(c, http.StatusOK, response, "Cost comparison retrieved successfully")
}
