package audit

import "time"

// AuditLogResponse represents an audit log in API responses
type AuditLogResponse struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resource_id,omitempty"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Success    bool                   `json:"success"`
	CreatedAt  time.Time              `json:"created_at"`
}

// AuditLogListResponse represents a list of audit logs
type AuditLogListResponse struct {
	Logs  []*AuditLogResponse `json:"logs"`
	Total int64               `json:"total"`
}

// AuditLogFilters represents filters for audit log queries
type AuditLogFilters struct {
	UserID    string    `json:"user_id,omitempty"`
	Action    string    `json:"action,omitempty"`
	Resource  string    `json:"resource,omitempty"`
	Success   *bool     `json:"success,omitempty"`
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate   time.Time `json:"end_date,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"`
}

// AuditStatsResponse represents audit statistics
type AuditStatsResponse struct {
	TotalActions      int64               `json:"total_actions"`
	SuccessfulActions int64               `json:"successful_actions"`
	FailedActions     int64               `json:"failed_actions"`
	TopActions        []*ActionCount      `json:"top_actions"`
	TopUsers          []*UserActionCount  `json:"top_users"`
	RecentActivity    []*AuditLogResponse `json:"recent_activity"`
}

// ActionCount represents action count statistics
type ActionCount struct {
	Action string `json:"action"`
	Count  int64  `json:"count"`
}

// UserActionCount represents user action count statistics
type UserActionCount struct {
	UserID string `json:"user_id"`
	Count  int64  `json:"count"`
}
