package notification

import "time"

// Notification represents a notification message (service-level DTO)
// Note: This is different from domain.Notification which is the persistence model
type Notification struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	WorkspaceID string                 `json:"workspace_id,omitempty"`
	Type        NotificationType       `json:"type"`
	Priority    NotificationPriority   `json:"priority"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Channels    []NotificationChannel  `json:"channels"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Read        bool                   `json:"read"`
	CreatedAt   time.Time              `json:"created_at"`
	ReadAt      *time.Time             `json:"read_at,omitempty"`
}

// NotificationTemplate represents a notification template
type NotificationTemplate struct {
	ID        string                `json:"id"`
	Name      string                `json:"name"`
	Type      NotificationType      `json:"type"`
	Priority  NotificationPriority  `json:"priority"`
	Title     string                `json:"title"`
	Message   string                `json:"message"`
	Channels  []NotificationChannel `json:"channels"`
	Variables []string              `json:"variables"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// NotificationPreferences represents user notification preferences (service-level DTO)
// Note: This is different from domain.NotificationPreferences which is the persistence model
type NotificationPreferences struct {
	UserID      string                                     `json:"user_id"`
	Email       bool                                       `json:"email"`
	Browser     bool                                       `json:"browser"`
	SMS         bool                                       `json:"sms"`
	InApp       bool                                       `json:"in_app"`
	Webhook     bool                                       `json:"webhook"`
	Preferences map[NotificationType]bool                  `json:"preferences"`
	Channels    map[NotificationType][]NotificationChannel `json:"channels"`
}
