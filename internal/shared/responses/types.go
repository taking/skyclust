package responses

import "time"

// APIResponse represents a standardized API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Error     *Error      `json:"error,omitempty"`
	Code      string      `json:"code,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp time.Time   `json:"timestamp,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
}

// Meta represents pagination and metadata information
type Meta struct {
	Page       int   `json:"page,omitempty"`
	Limit      int   `json:"limit,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
	HasNext    bool  `json:"has_next,omitempty"`
	HasPrev    bool  `json:"has_prev,omitempty"`
}

// Error represents a unified error structure
type Error struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Field       string                 `json:"field,omitempty"`
	Value       interface{}            `json:"value,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	RequestID   string                 `json:"request_id,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Success   bool        `json:"success"`
	Error     string      `json:"error"`
	Code      string      `json:"code"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Details   interface{} `json:"details,omitempty"`
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page  int `json:"page,omitempty" form:"page" validate:"min=1"`
	Limit int `json:"limit,omitempty" form:"limit" validate:"min=1,max=100"`
}

// PaginationResponse represents pagination information in responses
type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// SortRequest represents sorting parameters
type SortRequest struct {
	SortBy    string `json:"sort_by,omitempty" form:"sort_by"`
	SortOrder string `json:"sort_order,omitempty" form:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// FilterRequest represents filtering parameters
type FilterRequest struct {
	Search  string                 `json:"search,omitempty" form:"search"`
	Filters map[string]interface{} `json:"filters,omitempty"`
}
