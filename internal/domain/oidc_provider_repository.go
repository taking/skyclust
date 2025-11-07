package domain

import (
	"github.com/google/uuid"
)

// OIDCProviderRepository defines the interface for OIDC provider persistence
type OIDCProviderRepository interface {
	Create(provider *OIDCProvider) error
	GetByID(id uuid.UUID) (*OIDCProvider, error)
	GetByUserID(userID uuid.UUID) ([]*OIDCProvider, error)
	GetByUserIDAndName(userID uuid.UUID, name string) (*OIDCProvider, error)
	Update(provider *OIDCProvider) error
	Delete(id uuid.UUID) error
	GetEnabledByUserID(userID uuid.UUID) ([]*OIDCProvider, error)
}
