package services

import "time"

// CostAnalysisService defines the interface for cost analysis operations
type CostAnalysisService interface {
	// AnalyzeCosts performs cost analysis for a workspace
	AnalyzeCosts(workspaceID string, startDate, endDate time.Time) (interface{}, error)

	// GetCostTrends retrieves cost trends over time
	GetCostTrends(workspaceID string, period string) ([]interface{}, error)

	// GetCostBreakdown retrieves cost breakdown by category
	GetCostBreakdown(workspaceID string, startDate, endDate time.Time) (interface{}, error)

	// GetCostRecommendations retrieves cost optimization recommendations
	GetCostRecommendations(workspaceID string) ([]interface{}, error)

	// GetCostAlerts retrieves cost alerts for a workspace
	GetCostAlerts(workspaceID string) ([]interface{}, error)

	// SetCostAlert creates a new cost alert
	SetCostAlert(workspaceID string, alert interface{}) error

	// UpdateCostAlert updates an existing cost alert
	UpdateCostAlert(alertID string, alert interface{}) error

	// DeleteCostAlert deletes a cost alert
	DeleteCostAlert(alertID string) error

	// GetCostReport generates a cost report
	GetCostReport(workspaceID string, startDate, endDate time.Time, format string) (interface{}, error)

	// GetCostMetrics retrieves cost metrics for a workspace
	GetCostMetrics(workspaceID string) (interface{}, error)
}
