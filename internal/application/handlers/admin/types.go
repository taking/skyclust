package admin

import "time"

// AdminUserResponse represents a user in admin API responses
type AdminUserResponse struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	IsActive     bool       `json:"is_active"`
	OIDCProvider string     `json:"oidc_provider,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

// AdminUserListResponse represents a list of users for admin
type AdminUserListResponse struct {
	Users []*AdminUserResponse `json:"users"`
	Total int64                `json:"total"`
}

// AdminUserStatsResponse represents user statistics for admin
type AdminUserStatsResponse struct {
	TotalUsers    int64 `json:"total_users"`
	ActiveUsers   int64 `json:"active_users"`
	InactiveUsers int64 `json:"inactive_users"`
	NewUsersToday int64 `json:"new_users_today"`
}

// AdminSystemStatsResponse represents system statistics for admin
type AdminSystemStatsResponse struct {
	Users          AdminUserStatsResponse    `json:"users"`
	SystemHealth   SystemHealthResponse      `json:"system_health"`
	RecentActivity []*RecentActivityResponse `json:"recent_activity"`
}

// SystemHealthResponse represents system health information
type SystemHealthResponse struct {
	Status   string                   `json:"status"`
	Services map[string]ServiceHealth `json:"services"`
	Alerts   []*AlertResponse         `json:"alerts"`
	Uptime   string                   `json:"uptime"`
}

// ServiceHealth represents the health of a service
type ServiceHealth struct {
	Status  string  `json:"status"`
	Healthy bool    `json:"healthy"`
	Latency float64 `json:"latency_ms,omitempty"`
}

// AlertResponse represents a system alert
type AlertResponse struct {
	Type    string      `json:"type"`
	Level   string      `json:"level"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// RecentActivityResponse represents recent system activity
type RecentActivityResponse struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	UserID    string    `json:"user_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
