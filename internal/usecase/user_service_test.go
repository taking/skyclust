package usecase

import (
	"context"
	"skyclust/internal/domain"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

// MockAuditLogRepository is a mock implementation of AuditLogRepository
type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) Create(log *domain.AuditLog) error {
	args := m.Called(log)
	return args.Error(0)
}

func (m *MockAuditLogRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]*domain.AuditLog, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) GetByAction(action string, limit, offset int) ([]*domain.AuditLog, error) {
	args := m.Called(action, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) GetByDateRange(start, end time.Time, limit, offset int) ([]*domain.AuditLog, error) {
	args := m.Called(start, end, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) CountByUserID(userID uuid.UUID) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAuditLogRepository) CountByAction(action string) (int64, error) {
	args := m.Called(action)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAuditLogRepository) DeleteOldLogs(before time.Time) error {
	args := m.Called(before)
	return args.Error(0)
}

func (m *MockUserRepository) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*domain.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByOIDC(provider, subject string) (*domain.User, error) {
	args := m.Called(provider, subject)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(limit, offset int, filters map[string]interface{}) ([]*domain.User, int64, error) {
	args := m.Called(limit, offset, filters)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.User), args.Get(1).(int64), args.Error(2)
}

// MockPasswordHasher is a mock implementation of PasswordHasher
// MockPasswordHasher is defined in auth_service_test.go

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
				userRepo.On("GetByEmail", "test@example.com").Return((*domain.User)(nil), nil)
				userRepo.On("GetByUsername", "testuser").Return((*domain.User)(nil), nil)
				hasher.On("HashPassword", "password123").Return("hashed_password", nil)
				userRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)
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
				existingUser := &domain.User{ID: uuid.New(), Email: "test@example.com"}
				userRepo.On("GetByEmail", "test@example.com").Return(existingUser, nil)
			},
			expectError: true,
			errorType:   "ALREADY_EXISTS",
		},
		{
			name: "username already exists",
			request: domain.CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher) {
				userRepo.On("GetByEmail", "test@example.com").Return((*domain.User)(nil), nil)
				existingUser := &domain.User{ID: uuid.New(), Username: "testuser"}
				userRepo.On("GetByUsername", "testuser").Return(existingUser, nil)
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
			auditLogRepo := new(MockAuditLogRepository)
			service := NewUserService(userRepo, hasher, auditLogRepo)

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
					ID:           uuid.New(),
					Email:        "test@example.com",
					PasswordHash: "hashed_password",
					IsActive:     true,
				}
				userRepo.On("GetByEmail", "test@example.com").Return(user, nil)
				hasher.On("VerifyPassword", "password123", "hashed_password").Return(true)
			},
			expectError: false,
		},
		{
			name:     "user not found",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher) {
				userRepo.On("GetByEmail", "test@example.com").Return((*domain.User)(nil), nil)
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
					ID:           uuid.New(),
					Email:        "test@example.com",
					PasswordHash: "hashed_password",
					IsActive:     true,
				}
				userRepo.On("GetByEmail", "test@example.com").Return(user, nil)
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
					ID:           uuid.New(),
					Email:        "test@example.com",
					PasswordHash: "hashed_password",
					IsActive:     false,
				}
				userRepo.On("GetByEmail", "test@example.com").Return(user, nil)
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
			auditLogRepo := new(MockAuditLogRepository)
			service := NewUserService(userRepo, hasher, auditLogRepo)

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
			userID: "11111111-1111-1111-1111-111111111111",
			setupMocks: func(userRepo *MockUserRepository) {
				userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
				user := &domain.User{
					ID:       userID,
					Username: "testuser",
					Email:    "test@example.com",
				}
				userRepo.On("GetByID", userID).Return(user, nil)
			},
			expectError: false,
		},
		{
			name:   "user not found",
			userID: "00000000-0000-0000-0000-000000000000",
			setupMocks: func(userRepo *MockUserRepository) {
				userID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
				userRepo.On("GetByID", userID).Return((*domain.User)(nil), nil)
			},
			expectError: true,
			errorType:   "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := new(MockUserRepository)
			hasher := new(MockPasswordHasher)
			tt.setupMocks(userRepo)

			// Create service
			auditLogRepo := new(MockAuditLogRepository)
			service := NewUserService(userRepo, hasher, auditLogRepo)

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
				expectedID := uuid.MustParse(tt.userID)
				assert.Equal(t, expectedID, user.ID)
			}

			// Verify mocks
			userRepo.AssertExpectations(t)
		})
	}
}
