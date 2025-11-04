package oidc

import "time"

// OIDC Service Constants
// These constants are specific to OIDC authentication operations

// HTTP client constants
const (
	// DefaultHTTPClientTimeout is the default timeout for HTTP client requests
	DefaultHTTPClientTimeout = 30 * time.Second
)

// State validation constants
const (
	// StateCacheTTL is the time-to-live for OIDC state cache entries
	StateCacheTTL = 10 * time.Minute

	// StateMaxAge is the maximum age for OIDC state before expiration
	StateMaxAge = 10 * time.Minute
)

