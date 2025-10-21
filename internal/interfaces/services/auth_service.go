package services

import (
	"github.com/google/uuid"
	"skyclust/internal/domain"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Register creates a new user account
	Register(req domain.CreateUserRequest) (*domain.User, string, error)

	// Login authenticates a user and returns a token
	Login(username, password string) (*domain.User, string, error)

	// LoginWithContext authenticates with additional context
	LoginWithContext(username, password, clientIP, userAgent string) (*domain.User, string, error)

	// ValidateToken validates a JWT token
	ValidateToken(token string) (*domain.User, error)

	// Logout invalidates a user's session
	Logout(userID uuid.UUID, token string) error
}
