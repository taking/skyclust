package repositories

import (
	"skyclust/internal/domain"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Basic CRUD operations
	Create(user *domain.User) error
	GetByID(id uuid.UUID) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	GetByUsername(username string) (*domain.User, error)
	Update(user *domain.User) error
	Delete(id uuid.UUID) error

	// List operations
	List(limit, offset int) ([]*domain.User, error)
	Count() (int64, error)

	// Search operations
	Search(query string, limit, offset int) ([]*domain.User, error)
	GetByRole(role domain.Role, limit, offset int) ([]*domain.User, error)

	// OIDC operations
	GetByOIDCSubject(provider, subject string) (*domain.User, error)
	UpdateOIDCInfo(userID uuid.UUID, provider, subject string) error

	// Status operations
	Activate(id uuid.UUID) error
	Deactivate(id uuid.UUID) error
	IsActive(id uuid.UUID) (bool, error)
}
