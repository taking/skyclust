package system

import "time"

// SystemStatusResponse represents system status information
type SystemStatusResponse struct {
	Status    string                   `json:"status"`
	Timestamp time.Time                `json:"timestamp"`
	Version   string                   `json:"version"`
	Uptime    string                   `json:"uptime"`
	Services  map[string]ServiceStatus `json:"services"`
	Metrics   SystemMetrics            `json:"metrics"`
}

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Healthy bool   `json:"healthy"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

// SystemMetrics represents system performance metrics
type SystemMetrics struct {
	MemoryUsage MemoryUsage    `json:"memory_usage"`
	Performance Performance    `json:"performance"`
	Database    DatabaseStatus `json:"database"`
}

// MemoryUsage represents memory usage metrics
type MemoryUsage struct {
	AllocMB      float64 `json:"alloc_mb"`
	TotalAllocMB float64 `json:"total_alloc_mb"`
	SysMB        float64 `json:"sys_mb"`
	NumGC        uint32  `json:"num_gc"`
}

// Performance represents performance metrics
type Performance struct {
	Goroutines   int     `json:"goroutines"`
	CPUCores     int     `json:"cpu_cores"`
	RequestCount int64   `json:"request_count"`
	ErrorCount   int64   `json:"error_count"`
	ErrorRate    float64 `json:"error_rate"`
	RPS          float64 `json:"rps"`
}

// DatabaseStatus represents database status
type DatabaseStatus struct {
	Healthy      bool    `json:"healthy"`
	Status       string  `json:"status"`
	ResponseTime float64 `json:"response_time_ms,omitempty"`
}

// HealthCheckResponse represents a health check response
type HealthCheckResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}
