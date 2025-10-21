package dto

import (
	"time"

	"github.com/google/uuid"
)

// UserDTO represents a user data transfer object
type UserDTO struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Active       bool      `json:"is_active"`
	OIDCProvider string    `json:"oidc_provider,omitempty"`
	OIDCSubject  string    `json:"oidc_subject,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Username string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Active   *bool  `json:"is_active,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	User         *UserDTO `json:"user"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int      `json:"expires_in"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// CreateOIDCUserRequest represents a request to create an OIDC user
type CreateOIDCUserRequest struct {
	Username     string `json:"username" validate:"required,min=3,max=50"`
	Email        string `json:"email" validate:"required,email"`
	OIDCProvider string `json:"oidc_provider" validate:"required"`
	OIDCSubject  string `json:"oidc_subject" validate:"required"`
}

// LinkOIDCRequest represents a request to link an OIDC account
type LinkOIDCRequest struct {
	OIDCProvider string `json:"oidc_provider" validate:"required"`
	OIDCSubject  string `json:"oidc_subject" validate:"required"`
}
