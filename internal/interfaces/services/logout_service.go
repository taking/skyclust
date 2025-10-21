package services

import "github.com/google/uuid"

// LogoutService defines the interface for logout operations
type LogoutService interface {
	// Logout invalidates a user's session
	Logout(userID uuid.UUID, token string) error

	// LogoutAllSessions logs out a user from all sessions
	LogoutAllSessions(userID string) error

	// IsTokenValid checks if a token is still valid
	IsTokenValid(token string) bool

	// RevokeToken revokes a specific token
	RevokeToken(token string) error

	// GetActiveSessions retrieves active sessions for a user
	GetActiveSessions(userID string) ([]string, error)
}
