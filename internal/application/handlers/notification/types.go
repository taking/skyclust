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

// UpdateNotificationRequest represents an update notification request (RESTful)
type UpdateNotificationRequest struct {
	Read *bool `json:"read,omitempty"` // true to mark as read, false to mark as unread
}

// UpdateNotificationsRequest represents a bulk update notifications request (RESTful)
type UpdateNotificationsRequest struct {
	Read            *bool     `json:"read,omitempty"`             // true to mark as read, false to mark as unread
	NotificationIDs []string  `json:"notification_ids,omitempty"` // Optional: specific notification IDs to update. If not provided, updates all.
}

// UpdateNotificationsResponse represents a bulk update notifications response
type UpdateNotificationsResponse struct {
	UpdatedCount int64 `json:"updated_count"`
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
