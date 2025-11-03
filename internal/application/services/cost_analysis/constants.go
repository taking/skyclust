package cost_analysis

import "time"

// Cost Analysis Service Constants
// These constants are specific to cost analysis and prediction operations

// Cost calculation constants
const (
	// DefaultCostPredictionMargin is the default margin (20%) for cost predictions
	DefaultCostPredictionMargin = 0.2

	// MinCostPredictionConfidence is the minimum confidence level (10%) for predictions
	MinCostPredictionConfidence = 0.1

	// CostVarianceNormalizationBase is used for variance normalization in predictions
	CostVarianceNormalizationBase = 100.0
)

// Historical data period for predictions
const (
	// CostPredictionHistoricalDays is the number of days of historical data used for predictions
	CostPredictionHistoricalDays = 30
)

// Cache configuration
const (
	// CostAnalysisCacheTTL is the time-to-live for cost analysis cache entries
	CostAnalysisCacheTTL = 15 * time.Minute
)
