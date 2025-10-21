package services

import (
	"skyclust/internal/domain"
)

// CredentialService defines the interface for credential management operations
type CredentialService interface {
	// CreateCredential creates a new credential
	CreateCredential(credential *domain.Credential) error

	// GetCredential retrieves a credential by ID
	GetCredential(id string) (*domain.Credential, error)

	// GetCredentials retrieves all credentials for a user
	GetCredentials(userID string) ([]*domain.Credential, error)

	// UpdateCredential updates an existing credential
	UpdateCredential(credential *domain.Credential) error

	// DeleteCredential deletes a credential
	DeleteCredential(id string) error

	// GetCredentialsByProvider retrieves credentials by cloud provider
	GetCredentialsByProvider(provider string) ([]*domain.Credential, error)

	// ValidateCredential validates credential access
	ValidateCredential(id string) error
}
