package user

import (
	"time"

	"github.com/google/uuid"
)

// UserEvent represents domain events for user operations
type UserEvent struct {
	ID        uuid.UUID              `json:"id"`
	Type      string                 `json:"type"`
	UserID    uuid.UUID              `json:"user_id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// UserEventType defines the types of user events
type UserEventType string

const (
	UserCreatedEvent         UserEventType = "user.created"
	UserUpdatedEvent         UserEventType = "user.updated"
	UserActivatedEvent       UserEventType = "user.activated"
	UserDeactivatedEvent     UserEventType = "user.deactivated"
	UserPasswordChangedEvent UserEventType = "user.password_changed"
	UserDeletedEvent         UserEventType = "user.deleted"
)

// NewUserEvent creates a new user event
func NewUserEvent(eventType UserEventType, userID uuid.UUID, data map[string]interface{}) *UserEvent {
	return &UserEvent{
		ID:        uuid.New(),
		Type:      string(eventType),
		UserID:    userID,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// UserCreatedEventData represents data for user created event
type UserCreatedEventData struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// UserUpdatedEventData represents data for user updated event
type UserUpdatedEventData struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// UserPasswordChangedEventData represents data for password changed event
type UserPasswordChangedEventData struct {
	ChangedAt time.Time `json:"changed_at"`
}

// UserActivatedEventData represents data for user activated event
type UserActivatedEventData struct {
	ActivatedAt time.Time `json:"activated_at"`
}

// UserDeactivatedEventData represents data for user deactivated event
type UserDeactivatedEventData struct {
	DeactivatedAt time.Time `json:"deactivated_at"`
}

// UserDeletedEventData represents data for user deleted event
type UserDeletedEventData struct {
	DeletedAt time.Time `json:"deleted_at"`
}
