package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

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
	GetUserCount() (int64, error)
	Authenticate(ctx context.Context, email, password string) (*User, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
	HashPassword(password string) (string, error)
}

// AuthService defines the interface for authentication business logic
type AuthService interface {
	Register(req CreateUserRequest) (*User, string, string, error)                               // Returns user, access token, and refresh token
	Login(email, password string) (*User, string, string, error)                                 // Returns user, access token, and refresh token
	LoginWithContext(email, password, clientIP, userAgent string) (*User, string, string, error) // Returns user, access token, and refresh token with context
	ValidateToken(token string) (*User, error)
	RefreshToken(refreshToken string) (string, string, error) // Returns new access token and refresh token
	RevokeRefreshToken(refreshToken string) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error
	Logout(userID uuid.UUID, token string) error
}

// OIDCService defines the interface for OIDC authentication
type OIDCService interface {
	GetAuthURL(ctx context.Context, provider, state string) (string, error)
	ExchangeCode(ctx context.Context, provider, code, state string) (*User, string, error)
	EndSession(ctx context.Context, userID uuid.UUID, provider, idToken, postLogoutRedirectURI string) error
	GetLogoutURL(ctx context.Context, provider, postLogoutRedirectURI string) (string, error)

	// OIDC Provider management
	CreateProvider(ctx context.Context, userID uuid.UUID, provider *OIDCProvider) (*OIDCProvider, error)
	GetUserProviders(ctx context.Context, userID uuid.UUID) ([]*OIDCProvider, error)
	GetProvider(ctx context.Context, userID uuid.UUID, providerID uuid.UUID) (*OIDCProvider, error)
	UpdateProvider(ctx context.Context, userID uuid.UUID, providerID uuid.UUID, provider *OIDCProvider) (*OIDCProvider, error)
	DeleteProvider(ctx context.Context, userID uuid.UUID, providerID uuid.UUID) error
}

// LogoutService defines the interface for logout operations
type LogoutService interface {
	Logout(userID uuid.UUID, token string) error
	BatchLogout(ctx context.Context, userID uuid.UUID, tokens []string) error
	GetLogoutStats(ctx context.Context) (map[string]interface{}, error)
	CleanupExpiredTokens(ctx context.Context) error
}

// EventService defines the interface for event operations
type EventService interface {
	Publish(ctx context.Context, eventType string, data interface{}) error
	Subscribe(ctx context.Context, eventType string, handler func(data interface{}) error) error
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
