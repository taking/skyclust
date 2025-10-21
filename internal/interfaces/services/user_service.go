package services

import (
	"context"
	"github.com/google/uuid"
	"skyclust/internal/domain"
)

// UserService defines the interface for user business operations
type UserService interface {
	// User management
	CreateUser(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
	GetUsers(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.User, int64, error)
	UpdateUser(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.User, error)

	// Admin-specific methods
	GetUserByID(id uuid.UUID) (*domain.User, error)
	GetUsersWithFilters(filters domain.UserFilters) ([]*domain.User, int64, error)
	UpdateUserDirect(user *domain.User) (*domain.User, error)
	DeleteUserByID(id uuid.UUID) error
	GetUserStats() (*domain.UserStats, error)
	Authenticate(ctx context.Context, email, password string) (*domain.User, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
}
