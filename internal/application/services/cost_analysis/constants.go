package cost_analysis

import "time"

// Cost Analysis Service Constants
// These constants are specific to cost analysis and prediction operations

// HTTP Status Code Constants
const (
	HTTPStatusBadRequest          = 400
	HTTPStatusInternalServerError = 500
)

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

// Date format constants
const (
	// DateFormatISO is the ISO 8601 date format used for AWS Cost Explorer API
	DateFormatISO = "2006-01-02"
)

// AWS configuration constants
const (
	// AWSDefaultRegion is the default AWS region for Cost Explorer API
	AWSDefaultRegion = "us-east-1"
)

// Resource type constants
const (
	ResourceTypeVM      = "vm"
	ResourceTypeCluster = "cluster"
)

// Currency constants
const (
	CurrencyUSD = "USD"
)

// Provider constants
const (
	ProviderAWS   = "aws"
	ProviderGCP   = "gcp"
	ProviderAzure = "azure"
)

// Service name constants by provider
const (
	ServiceNameAWSEC2               = "EC2"
	ServiceNameGCPCompute           = "Compute"
	ServiceNameAzureVirtualMachines = "Virtual Machines"
	ServiceNameDefaultCompute       = "compute"
)

// VM pricing estimation constants
const (
	// BaseHourlyRate is the base hourly rate ($0.05 per hour) for VM cost estimation
	BaseHourlyRate = 0.05

	// CPUMultiplierPerCore is the cost multiplier per CPU core ($0.01 per core per hour)
	CPUMultiplierPerCore = 0.01

	// MemoryMultiplierPerGB is the cost multiplier per GB of memory ($0.005 per GB per hour)
	MemoryMultiplierPerGB = 0.005

	// StorageMultiplierPerGB is the cost multiplier per GB of storage ($0.0001 per GB per hour)
	StorageMultiplierPerGB = 0.0001

	// Provider multipliers
	ProviderMultiplierAWS   = 1.0
	ProviderMultiplierGCP   = 0.9
	ProviderMultiplierAzure = 1.1
)

// Trend calculation constants
const (
	// TrendPercentageThreshold is the percentage change threshold (5%) for determining trend direction
	TrendPercentageThreshold = 5.0

	// TrendDirection constants
	TrendDirectionStable     = "stable"
	TrendDirectionIncreasing = "increasing"
	TrendDirectionDecreasing = "decreasing"
)

// Array index constants for AWS Cost Explorer results
const (
	// AWS Cost Explorer group keys indices
	GroupKeyIndexService = 0
	GroupKeyIndexRegion  = 1
)

// Minimum data requirements
const (
	// MinDataPointsForPrediction is the minimum number of data points required for prediction
	MinDataPointsForPrediction = 2

	// MinDataPointsForVariance is the minimum number of data points required for variance calculation
	MinDataPointsForVariance = 2
)
