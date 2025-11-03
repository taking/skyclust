package service

// Common Service Constants
// These constants are shared across multiple services in the application layer

// File Upload Constants
const (
	// MaxMultipartFormSize is the maximum size for multipart form data (32MB)
	MaxMultipartFormSize = 32 << 20
)

// Memory and Performance Constants
const (
	// BytesPerMB is the number of bytes in a megabyte
	BytesPerMB = 1024 * 1024

	// Percentage calculation
	PercentageBase = 100.0 // Base for percentage calculations

	// Error rate thresholds
	HighErrorRateThreshold = 5.0   // 5% error rate threshold
	MemoryWarningThreshold = 100.0 // 100MB memory warning threshold
)
