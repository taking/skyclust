package domain

import (
	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user *User) error
	GetByID(id uuid.UUID) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByOIDC(provider, subject string) (*User, error)
	Update(user *User) error
	Delete(id uuid.UUID) error
	List(limit, offset int, filters map[string]interface{}) ([]*User, int64, error)
	Count() (int64, error)
}
