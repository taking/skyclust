package services

import (
	"skyclust/internal/domain"
)

// OIDCService defines the interface for OpenID Connect operations
type OIDCService interface {
	// GetOIDCProviders retrieves available OIDC providers
	GetOIDCProviders() ([]interface{}, error)

	// GetOIDCProvider retrieves a specific OIDC provider
	GetOIDCProvider(providerID string) (interface{}, error)

	// RegisterOIDCProvider registers a new OIDC provider
	RegisterOIDCProvider(provider interface{}) error

	// UpdateOIDCProvider updates an existing OIDC provider
	UpdateOIDCProvider(provider interface{}) error

	// DeleteOIDCProvider deletes an OIDC provider
	DeleteOIDCProvider(providerID string) error

	// AuthenticateWithOIDC authenticates a user with OIDC
	AuthenticateWithOIDC(providerID, code, state string) (*domain.User, string, error)

	// GetOIDCAuthURL generates OIDC authentication URL
	GetOIDCAuthURL(providerID, state string) (string, error)

	// ValidateOIDCToken validates an OIDC token
	ValidateOIDCToken(token string) (*domain.User, error)
}
