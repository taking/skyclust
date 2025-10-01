package domain

import (
	"context"
	"time"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"` // Hidden from JSON
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
}

// UserService defines the business logic interface for users
type UserService interface {
	CreateUser(ctx context.Context, req CreateUserRequest) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	Authenticate(ctx context.Context, email, password string) (*User, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// Validate performs validation on the CreateUserRequest
func (r *CreateUserRequest) Validate() error {
	if len(r.Username) < 3 || len(r.Username) > 50 {
		return NewValidationError("username must be between 3 and 50 characters")
	}
	if len(r.Email) == 0 {
		return NewValidationError("email is required")
	}
	if len(r.Password) < 8 {
		return NewValidationError("password must be at least 8 characters")
	}
	return nil
}

// Validate performs validation on the UpdateUserRequest
func (r *UpdateUserRequest) Validate() error {
	if r.Username != nil && (len(*r.Username) < 3 || len(*r.Username) > 50) {
		return NewValidationError("username must be between 3 and 50 characters")
	}
	if r.Email != nil && len(*r.Email) == 0 {
		return NewValidationError("email cannot be empty")
	}
	return nil
}
