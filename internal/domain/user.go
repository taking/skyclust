package domain

import (
	"context"
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
	Credentials []Credential `json:"credentials,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	AuditLogs   []AuditLog   `json:"audit_logs,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	UserRoles   []UserRole   `json:"user_roles,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
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

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user *User) error
	GetByID(id uuid.UUID) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByOIDC(provider, subject string) (*User, error)
	Update(user *User) error
	Delete(id uuid.UUID) error
	List(limit, offset int, filters map[string]interface{}) ([]*User, int64, error)
	Count() (int64, error)
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

// UserService defines the interface for user business logic
type UserService interface {
	CreateUser(ctx context.Context, req CreateUserRequest) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	GetUsers(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*User, int64, error)
	UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*User, error)

	// Admin-specific methods
	GetUserByID(id uuid.UUID) (*User, error)
	GetUsersWithFilters(filters UserFilters) ([]*User, int64, error)
	UpdateUserDirect(user *User) (*User, error)
	DeleteUserByID(id uuid.UUID) error
	GetUserStats() (*UserStats, error)
	Authenticate(ctx context.Context, email, password string) (*User, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
}

// AuthService defines the interface for authentication business logic
type AuthService interface {
	Register(req CreateUserRequest) (*User, string, error)                               // Returns user and JWT token
	Login(email, password string) (*User, string, error)                                 // Returns user and JWT token
	LoginWithContext(email, password, clientIP, userAgent string) (*User, string, error) // Returns user and JWT token with context
	ValidateToken(token string) (*User, error)
	Logout(userID uuid.UUID, token string) error
}

// OIDCService defines the interface for OIDC authentication
type OIDCService interface {
	GetAuthURL(ctx context.Context, provider, state string) (string, error)
	ExchangeCode(ctx context.Context, provider, code, state string) (*User, string, error)
	EndSession(ctx context.Context, userID uuid.UUID, provider, idToken, postLogoutRedirectURI string) error
	GetLogoutURL(ctx context.Context, provider, postLogoutRedirectURI string) (string, error)
}

// LogoutService defines the interface for logout operations
type LogoutService interface {
	Logout(userID uuid.UUID, token string) error
	BatchLogout(ctx context.Context, userID uuid.UUID, tokens []string) error
	GetLogoutStats(ctx context.Context) (map[string]interface{}, error)
	CleanupExpiredTokens(ctx context.Context) error
}

// PluginActivationService defines the interface for plugin activation
type PluginActivationService interface {
	ActivatePlugin(ctx context.Context, userID uuid.UUID, provider string) error
	DeactivatePlugin(ctx context.Context, userID uuid.UUID, provider string) error
	GetActivePlugins(ctx context.Context, userID uuid.UUID) ([]string, error)
}

// EventService defines the interface for event operations
type EventService interface {
	Publish(ctx context.Context, eventType string, data interface{}) error
	Subscribe(ctx context.Context, eventType string, handler func(data interface{}) error) error
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

// CacheService defines the cache service interface
type CacheService interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	GetOrSet(ctx context.Context, key string, setter func() (interface{}, error), ttl time.Duration) (interface{}, error)
	InvalidatePattern(ctx context.Context, pattern string) error
	GetStats(ctx context.Context) (map[string]interface{}, error)
	Health(ctx context.Context) error
}
