package notification

import "time"

// NotificationResponse represents a notification in API responses
type NotificationResponse struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status"`
	Data      map[string]interface{} `json:"data,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
}

// NotificationListResponse represents a list of notifications
type NotificationListResponse struct {
	Notifications []*NotificationResponse `json:"notifications"`
	Total         int64                   `json:"total"`
	UnreadCount   int64                   `json:"unread_count"`
}

// MarkAsReadRequest represents a mark as read request
type MarkAsReadRequest struct {
	NotificationIDs []string `json:"notification_ids" validate:"required,min=1"`
}

// MarkAllAsReadResponse represents a mark all as read response
type MarkAllAsReadResponse struct {
	MarkedCount int64 `json:"marked_count"`
}

// NotificationPreferencesResponse represents user notification preferences
type NotificationPreferencesResponse struct {
	EmailEnabled    bool `json:"email_enabled"`
	PushEnabled     bool `json:"push_enabled"`
	BrowserEnabled  bool `json:"browser_enabled"`
	WorkspaceEvents bool `json:"workspace_events"`
	SecurityAlerts  bool `json:"security_alerts"`
	SystemUpdates   bool `json:"system_updates"`
}

// UpdatePreferencesRequest represents an update preferences request
type UpdatePreferencesRequest struct {
	EmailEnabled    *bool `json:"email_enabled,omitempty"`
	PushEnabled     *bool `json:"push_enabled,omitempty"`
	BrowserEnabled  *bool `json:"browser_enabled,omitempty"`
	WorkspaceEvents *bool `json:"workspace_events,omitempty"`
	SecurityAlerts  *bool `json:"security_alerts,omitempty"`
	SystemUpdates   *bool `json:"system_updates,omitempty"`
}
