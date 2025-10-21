package repositories

import (
	"context"
	"skyclust/internal/domain"

	"github.com/google/uuid"
)

// CredentialRepository defines the interface for credential data operations
type CredentialRepository interface {
	// Basic CRUD operations
	Create(credential *domain.Credential) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Credential, error)
	Update(credential *domain.Credential) error
	Delete(ctx context.Context, id uuid.UUID) error

	// List operations
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Credential, error)
	List(limit, offset int) ([]*domain.Credential, error)
	Count() (int64, error)

	// Search operations
	Search(query string, limit, offset int) ([]*domain.Credential, error)
	GetByProvider(provider string, limit, offset int) ([]*domain.Credential, error)

	// Status operations
	Activate(id uuid.UUID) error
	Deactivate(id uuid.UUID) error
	IsActive(id uuid.UUID) (bool, error)
}
