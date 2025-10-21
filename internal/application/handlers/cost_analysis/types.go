package cost_analysis

import "time"

// CostAnalysisRequest represents a cost analysis request
type CostAnalysisRequest struct {
	Provider     string                 `json:"provider" validate:"required"`
	Region       string                 `json:"region,omitempty"`
	InstanceType string                 `json:"instance_type,omitempty"`
	TimeRange    TimeRange              `json:"time_range" validate:"required"`
	Filters      map[string]interface{} `json:"filters,omitempty"`
}

// TimeRange represents a time range for analysis
type TimeRange struct {
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required"`
}

// CostAnalysisResponse represents cost analysis results
type CostAnalysisResponse struct {
	TotalCost       float64           `json:"total_cost"`
	Currency        string            `json:"currency"`
	Breakdown       []*CostBreakdown  `json:"breakdown"`
	Trends          []*CostTrend      `json:"trends"`
	Recommendations []*Recommendation `json:"recommendations"`
	GeneratedAt     time.Time         `json:"generated_at"`
}

// CostBreakdown represents cost breakdown by category
type CostBreakdown struct {
	Category string  `json:"category"`
	Cost     float64 `json:"cost"`
	Percent  float64 `json:"percent"`
}

// CostTrend represents cost trend over time
type CostTrend struct {
	Date time.Time `json:"date"`
	Cost float64   `json:"cost"`
}

// Recommendation represents a cost optimization recommendation
type Recommendation struct {
	Type        string  `json:"type"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Savings     float64 `json:"savings"`
	Priority    string  `json:"priority"`
}

// CostReportResponse represents a cost report
type CostReportResponse struct {
	ReportID    string                `json:"report_id"`
	Status      string                `json:"status"`
	Data        *CostAnalysisResponse `json:"data,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
	CompletedAt *time.Time            `json:"completed_at,omitempty"`
}
