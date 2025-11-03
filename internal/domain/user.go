package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Username     string     `json:"username" gorm:"not null;size:50"` // Not unique - multiple users can have same username
	Email        string     `json:"email" gorm:"uniqueIndex;not null;size:100"`
	PasswordHash string     `json:"-" gorm:"column:password_hash;not null;size:255"`
	OIDCProvider string     `json:"oidc_provider,omitempty" gorm:"size:20"` // google, github, azure
	OIDCSubject  string     `json:"oidc_subject,omitempty" gorm:"size:100"` // OIDC subject ID
	Active       bool       `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    *time.Time `json:"-" gorm:"index"`

	// Relationships
	// Note: Credentials are now workspace-based, not user-based
	AuditLogs []AuditLog `json:"audit_logs,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	UserRoles []UserRole `json:"user_roles,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// Business methods for User entity

// IsActive checks if the user is active
func (u *User) IsActive() bool {
	return u.Active
}

// Activate activates the user
func (u *User) Activate() {
	u.Active = true
	u.UpdatedAt = time.Now()
}

// Deactivate deactivates the user
func (u *User) Deactivate() {
	u.Active = false
	u.UpdatedAt = time.Now()
}

// UpdateProfile updates user profile information
func (u *User) UpdateProfile(username, email string) error {
	if username == "" {
		return NewDomainError(ErrCodeValidationFailed, "username cannot be empty", 400)
	}
	if email == "" {
		return NewDomainError(ErrCodeValidationFailed, "email cannot be empty", 400)
	}

	u.Username = username
	u.Email = email
	u.UpdatedAt = time.Now()
	return nil
}

// SetPasswordHash sets the password hash
func (u *User) SetPasswordHash(hash string) {
	u.PasswordHash = hash
	u.UpdatedAt = time.Now()
}

// CanAccessResource checks if user can access a resource
func (u *User) CanAccessResource(resourceUserID uuid.UUID, userRole Role) bool {
	return u.ID == resourceUserID || userRole == AdminRoleType
}

// IsAdmin checks if user has admin role
func (u *User) IsAdmin(userRoles []Role) bool {
	for _, role := range userRoles {
		if role == AdminRoleType {
			return true
		}
	}
	return false
}

// GetDisplayName returns the display name for the user
func (u *User) GetDisplayName() string {
	if u.Username != "" {
		return u.Username
	}
	return u.Email
}

// IsOIDCUser checks if this is an OIDC user
func (u *User) IsOIDCUser() bool {
	return u.OIDCProvider != "" && u.OIDCSubject != ""
}

// SetOIDCInfo sets OIDC provider information
func (u *User) SetOIDCInfo(provider, subject string) {
	u.OIDCProvider = provider
	u.OIDCSubject = subject
	u.UpdatedAt = time.Now()
}

// UserFilters represents filters for user queries
type UserFilters struct {
	Search string
	Role   string
	Status string
	Page   int
	Limit  int
}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers    int64
	ActiveUsers   int64
	InactiveUsers int64
	NewUsersToday int64
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// Validate performs validation on the CreateUserRequest
func (r *CreateUserRequest) Validate() error {
	if len(r.Username) < 3 || len(r.Username) > 50 {
		return NewDomainError(ErrCodeValidationFailed, "username must be between 3 and 50 characters", 400)
	}
	if len(r.Email) == 0 {
		return NewDomainError(ErrCodeValidationFailed, "email is required", 400)
	}
	if len(r.Password) < 8 {
		return NewDomainError(ErrCodeValidationFailed, "password must be at least 8 characters", 400)
	}
	return nil
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=8"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// Validate performs validation on the UpdateUserRequest
func (r *UpdateUserRequest) Validate() error {
	if r.Username != nil && (len(*r.Username) < 3 || len(*r.Username) > 50) {
		return NewDomainError(ErrCodeValidationFailed, "username must be between 3 and 50 characters", 400)
	}
	if r.Email != nil && len(*r.Email) == 0 {
		return NewDomainError(ErrCodeValidationFailed, "email cannot be empty", 400)
	}
	if r.Password != nil && len(*r.Password) < 8 {
		return NewDomainError(ErrCodeValidationFailed, "password must be at least 8 characters", 400)
	}
	return nil
}

// OIDCLoginRequest represents the request for OIDC login
type OIDCLoginRequest struct {
	Provider string `json:"provider" validate:"required,oneof=google github azure"`
	Code     string `json:"code" validate:"required"`
	State    string `json:"state" validate:"required"`
}
