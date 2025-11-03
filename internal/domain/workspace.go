package domain

import (
	"time"
)

// Workspace represents a workspace in the system
type Workspace struct {
	ID          string                 `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name        string                 `json:"name" gorm:"uniqueIndex;not null"`
	Description string                 `json:"description" gorm:"type:text"`
	OwnerID     string                 `json:"owner_id" gorm:"not null;type:uuid"`
	Settings    map[string]interface{} `json:"settings" gorm:"type:jsonb"`
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time             `json:"-" gorm:"index"`
	Active      bool                   `json:"is_active" gorm:"default:true"`
}

// TableName specifies the table name for Workspace
func (Workspace) TableName() string {
	return "workspaces"
}

// Business methods for Workspace entity

// IsActive checks if the workspace is active
func (w *Workspace) IsActive() bool {
	return w.Active
}

// Activate activates the workspace
func (w *Workspace) Activate() {
	w.Active = true
	w.UpdatedAt = time.Now()
}

// Deactivate deactivates the workspace
func (w *Workspace) Deactivate() {
	w.Active = false
	w.UpdatedAt = time.Now()
}

// UpdateInfo updates workspace information
func (w *Workspace) UpdateInfo(name, description string) error {
	if name == "" {
		return NewDomainError(ErrCodeValidationFailed, "workspace name cannot be empty", 400)
	}

	w.Name = name
	w.Description = description
	w.UpdatedAt = time.Now()
	return nil
}

// SetSetting sets a workspace setting
func (w *Workspace) SetSetting(key string, value interface{}) {
	if w.Settings == nil {
		w.Settings = make(map[string]interface{})
	}
	w.Settings[key] = value
	w.UpdatedAt = time.Now()
}

// GetSetting gets a workspace setting
func (w *Workspace) GetSetting(key string) (interface{}, bool) {
	if w.Settings == nil {
		return nil, false
	}
	value, exists := w.Settings[key]
	return value, exists
}

// RemoveSetting removes a workspace setting
func (w *Workspace) RemoveSetting(key string) {
	if w.Settings != nil {
		delete(w.Settings, key)
		w.UpdatedAt = time.Now()
	}
}

// CanUserAccess checks if a user can access this workspace
func (w *Workspace) CanUserAccess(userID string, userRole Role) bool {
	// Admins can access any workspace
	if userRole == AdminRoleType {
		return true
	}
	// Owners can access their own workspaces
	return w.OwnerID == userID
}

// IsOwner checks if a user is the owner of this workspace
func (w *Workspace) IsOwner(userID string) bool {
	return w.OwnerID == userID
}

// GetDisplayName returns the display name for the workspace
func (w *Workspace) GetDisplayName() string {
	return w.Name
}

// WorkspaceUser represents a user in a workspace
type WorkspaceUser struct {
	UserID      string     `json:"user_id" gorm:"primaryKey;type:uuid"`
	WorkspaceID string     `json:"workspace_id" gorm:"primaryKey;type:uuid"`
	Role        string     `json:"role" gorm:"not null;default:member"`
	JoinedAt    time.Time  `json:"joined_at" gorm:"autoCreateTime"`
	DeletedAt   *time.Time `json:"-" gorm:"index"`
}

// TableName specifies the table name for WorkspaceUser
func (WorkspaceUser) TableName() string {
	return "workspace_users"
}

// CreateWorkspaceRequest represents the request to create a new workspace
type CreateWorkspaceRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
	OwnerID     string `json:"owner_id,omitempty"` // Set by handler from token, not by client
}

// UpdateWorkspaceRequest represents the request to update a workspace
type UpdateWorkspaceRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// Validate performs validation on the CreateWorkspaceRequest
func (r *CreateWorkspaceRequest) Validate() error {
	if len(r.Name) < 3 || len(r.Name) > 100 {
		return NewDomainError(ErrCodeValidationFailed, "name must be between 3 and 100 characters", 400)
	}
	if len(r.Description) > 500 {
		return NewDomainError(ErrCodeValidationFailed, "description must be less than 500 characters", 400)
	}
	if len(r.OwnerID) == 0 {
		return NewDomainError(ErrCodeValidationFailed, "owner_id is required", 400)
	}
	return nil
}

// Validate performs validation on the UpdateWorkspaceRequest
func (r *UpdateWorkspaceRequest) Validate() error {
	if r.Name != nil && (len(*r.Name) < 3 || len(*r.Name) > 100) {
		return NewDomainError(ErrCodeValidationFailed, "name must be between 3 and 100 characters", 400)
	}
	if r.Description != nil && len(*r.Description) > 500 {
		return NewDomainError(ErrCodeValidationFailed, "description must be less than 500 characters", 400)
	}
	return nil
}
