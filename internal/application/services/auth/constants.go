package auth

import "time"

// Auth Service Constants
// These constants are specific to authentication operations

// Token expiry constants
const (
	// DefaultTokenExpiry is the default token expiry time (24 hours)
	DefaultTokenExpiry = 24 * time.Hour

	// BlacklistTokenExpiry is the expiry time for tokens in blacklist (24 hours)
	BlacklistTokenExpiry = 24 * time.Hour
)

