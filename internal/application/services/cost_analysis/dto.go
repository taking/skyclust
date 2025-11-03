package cost_analysis

import "time"

// CostData represents cost information for a specific period
type CostData struct {
	Date         time.Time `json:"date"`
	Amount       float64   `json:"amount"`
	Currency     string    `json:"currency"`
	Service      string    `json:"service"`
	ResourceID   string    `json:"resource_id"`
	ResourceType string    `json:"resource_type"`
	Provider     string    `json:"provider"`
	Region       string    `json:"region"`
	WorkspaceID  string    `json:"workspace_id"`
}

// CostWarning represents a warning message about cost calculation
type CostWarning struct {
	Code         string `json:"code"`                    // warning code (e.g., "API_PERMISSION_DENIED", "API_NOT_ENABLED")
	Message      string `json:"message"`                 // user-friendly warning message
	Provider     string `json:"provider,omitempty"`      // provider name if applicable (aws, gcp)
	ResourceType string `json:"resource_type,omitempty"` // resource type if applicable (vm, cluster)
}

// CostSummary represents simplified cost data
type CostSummary struct {
	TotalCost  float64            `json:"total_cost"`
	Currency   string             `json:"currency"`
	Period     string             `json:"period"`
	StartDate  time.Time          `json:"start_date"`
	EndDate    time.Time          `json:"end_date"`
	ByProvider map[string]float64 `json:"by_provider"`
	Warnings   []CostWarning      `json:"warnings,omitempty"` // warnings about cost calculation issues
}

// CostPrediction represents future cost predictions
type CostPrediction struct {
	Date       time.Time `json:"date"`
	Predicted  float64   `json:"predicted"`
	Confidence float64   `json:"confidence"` // 0-1
	LowerBound float64   `json:"lower_bound"`
	UpperBound float64   `json:"upper_bound"`
}

// BudgetAlert represents budget threshold alerts
type BudgetAlert struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	BudgetLimit float64   `json:"budget_limit"`
	CurrentCost float64   `json:"current_cost"`
	Percentage  float64   `json:"percentage"`
	AlertLevel  string    `json:"alert_level"` // "warning", "critical"
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}

// CostTrend represents cost trend data
type CostTrend struct {
	DailyCosts       []DailyCostData `json:"daily_costs"`
	TrendDirection   string          `json:"trend_direction"` // "increasing", "decreasing", "stable"
	ChangePercentage float64         `json:"change_percentage"`
	Warnings         []CostWarning   `json:"warnings,omitempty"` // warnings about cost calculation issues
}

// DailyCostData represents daily cost data
type DailyCostData struct {
	Date   time.Time `json:"date"`
	Amount float64   `json:"amount"`
}

// CostBreakdown represents cost breakdown by dimension
type CostBreakdown map[string]CategoryBreakdown

// CategoryBreakdown represents breakdown for a category
type CategoryBreakdown struct {
	Cost       float64            `json:"cost"`
	Percentage float64            `json:"percentage"`
	Services   map[string]float64 `json:"services,omitempty"`
}

// CostComparison represents cost comparison between periods
type CostComparison struct {
	CurrentPeriod  PeriodComparison `json:"current_period"`
	PreviousPeriod PeriodComparison `json:"previous_period"`
	Comparison     ComparisonData   `json:"comparison"`
}

// PeriodComparison represents cost data for a period
type PeriodComparison struct {
	Period string  `json:"period"`
	Cost   float64 `json:"cost"`
}

// ComparisonData represents comparison metrics
type ComparisonData struct {
	CostChange       float64 `json:"cost_change"`
	PercentageChange float64 `json:"percentage_change"`
}
