package user

import (
	"fmt"
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
	// Credentials []Credential `json:"credentials,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	// AuditLogs   []AuditLog   `json:"audit_logs,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	UserRoles []UserRole `json:"user_roles,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// UserRole represents a user role
type UserRole struct {
	ID     uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID uuid.UUID `json:"user_id" gorm:"not null;type:uuid"`
	Role   string    `json:"role" gorm:"not null;size:20"`
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
		return fmt.Errorf("username cannot be empty")
	}
	if email == "" {
		return fmt.Errorf("email cannot be empty")
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
func (u *User) CanAccessResource(resourceUserID uuid.UUID, userRole string) bool {
	return u.ID == resourceUserID || userRole == "admin"
}

// IsAdmin checks if user has admin role
func (u *User) IsAdmin(userRoles []string) bool {
	for _, role := range userRoles {
		if role == "admin" {
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
