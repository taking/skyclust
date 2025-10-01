package usecase

import (
	"context"
	"testing"

	"cmp/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

// MockPasswordHasher is a mock implementation of PasswordHasher
type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) VerifyPassword(password, hash string) bool {
	args := m.Called(password, hash)
	return args.Bool(0)
}

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name        string
		request     domain.CreateUserRequest
		setupMocks  func(*MockUserRepository, *MockPasswordHasher)
		expectError bool
		errorType   string
	}{
		{
			name: "successful user creation",
			request: domain.CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher) {
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
				userRepo.On("GetByUsername", mock.Anything, "testuser").Return(nil, nil)
				hasher.On("HashPassword", "password123").Return("hashed_password", nil)
				userRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "user already exists by email",
			request: domain.CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher) {
				existingUser := &domain.User{ID: "1", Email: "test@example.com"}
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)
			},
			expectError: true,
			errorType:   "USER_ALREADY_EXISTS",
		},
		{
			name: "username already exists",
			request: domain.CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher) {
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
				existingUser := &domain.User{ID: "1", Username: "testuser"}
				userRepo.On("GetByUsername", mock.Anything, "testuser").Return(existingUser, nil)
			},
			expectError: true,
			errorType:   "username already exists",
		},
		{
			name: "invalid request - short username",
			request: domain.CreateUserRequest{
				Username: "ab", // Too short
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher) {
				// No mocks needed for validation error
			},
			expectError: true,
			errorType:   "validation failed",
		},
		{
			name: "invalid request - short password",
			request: domain.CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "123", // Too short
			},
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher) {
				// No mocks needed for validation error
			},
			expectError: true,
			errorType:   "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := new(MockUserRepository)
			hasher := new(MockPasswordHasher)
			tt.setupMocks(userRepo, hasher)

			// Create service
			service := NewUserService(userRepo, hasher)

			// Execute
			user, err := service.CreateUser(context.Background(), tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Contains(t, err.Error(), tt.errorType)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.request.Username, user.Username)
				assert.Equal(t, tt.request.Email, user.Email)
				assert.NotEmpty(t, user.ID)
				assert.True(t, user.IsActive)
			}

			// Verify mocks
			userRepo.AssertExpectations(t)
			hasher.AssertExpectations(t)
		})
	}
}

func TestUserService_Authenticate(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		setupMocks  func(*MockUserRepository, *MockPasswordHasher)
		expectError bool
		errorType   string
	}{
		{
			name:     "successful authentication",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher) {
				user := &domain.User{
					ID:       "1",
					Email:    "test@example.com",
					Password: "hashed_password",
					IsActive: true,
				}
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
				hasher.On("VerifyPassword", "password123", "hashed_password").Return(true)
			},
			expectError: false,
		},
		{
			name:     "user not found",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher) {
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
			},
			expectError: true,
			errorType:   "INVALID_CREDENTIALS",
		},
		{
			name:     "invalid password",
			email:    "test@example.com",
			password: "wrongpassword",
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher) {
				user := &domain.User{
					ID:       "1",
					Email:    "test@example.com",
					Password: "hashed_password",
					IsActive: true,
				}
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
				hasher.On("VerifyPassword", "wrongpassword", "hashed_password").Return(false)
			},
			expectError: true,
			errorType:   "INVALID_CREDENTIALS",
		},
		{
			name:     "inactive user",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher) {
				user := &domain.User{
					ID:       "1",
					Email:    "test@example.com",
					Password: "hashed_password",
					IsActive: false,
				}
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
			},
			expectError: true,
			errorType:   "user account is disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := new(MockUserRepository)
			hasher := new(MockPasswordHasher)
			tt.setupMocks(userRepo, hasher)

			// Create service
			service := NewUserService(userRepo, hasher)

			// Execute
			user, err := service.Authenticate(context.Background(), tt.email, tt.password)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Contains(t, err.Error(), tt.errorType)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
			}

			// Verify mocks
			userRepo.AssertExpectations(t)
			hasher.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		setupMocks  func(*MockUserRepository)
		expectError bool
		errorType   string
	}{
		{
			name:   "successful user retrieval",
			userID: "1",
			setupMocks: func(userRepo *MockUserRepository) {
				user := &domain.User{
					ID:       "1",
					Username: "testuser",
					Email:    "test@example.com",
				}
				userRepo.On("GetByID", mock.Anything, "1").Return(user, nil)
			},
			expectError: false,
		},
		{
			name:   "user not found",
			userID: "1",
			setupMocks: func(userRepo *MockUserRepository) {
				userRepo.On("GetByID", mock.Anything, "1").Return(nil, nil)
			},
			expectError: true,
			errorType:   "USER_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := new(MockUserRepository)
			hasher := new(MockPasswordHasher)
			tt.setupMocks(userRepo)

			// Create service
			service := NewUserService(userRepo, hasher)

			// Execute
			user, err := service.GetUser(context.Background(), tt.userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Contains(t, err.Error(), tt.errorType)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.userID, user.ID)
			}

			// Verify mocks
			userRepo.AssertExpectations(t)
		})
	}
}
