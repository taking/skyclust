package usecase

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"skyclust/internal/domain"
)

// MockPasswordHasher for testing
type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) VerifyPassword(password, hashedPassword string) bool {
	args := m.Called(password, hashedPassword)
	return args.Bool(0)
}

// MockTokenBlacklist for testing
type MockTokenBlacklist struct {
	mock.Mock
}

func (m *MockTokenBlacklist) Add(token string, expiry time.Duration) error {
	args := m.Called(token, expiry)
	return args.Error(0)
}

func (m *MockTokenBlacklist) Contains(token string) (bool, error) {
	args := m.Called(token)
	return args.Bool(0), args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name        string
		request     domain.CreateUserRequest
		setupMocks  func(*MockUserRepository, *MockPasswordHasher, *MockAuditLogRepository)
		expectError bool
		errorType   string
	}{
		{
			name: "successful registration",
			request: domain.CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher, auditRepo *MockAuditLogRepository) {
				userRepo.On("GetByEmail", "test@example.com").Return((*domain.User)(nil), nil)
				userRepo.On("GetByUsername", "testuser").Return((*domain.User)(nil), nil)
				hasher.On("HashPassword", "password123").Return("hashed_password", nil)
				userRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)
				auditRepo.On("Create", mock.AnythingOfType("*domain.AuditLog")).Return(nil)
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
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher, auditRepo *MockAuditLogRepository) {
				existingUser := &domain.User{ID: uuid.New(), Email: "test@example.com"}
				userRepo.On("GetByEmail", "test@example.com").Return(existingUser, nil)
				userRepo.On("GetByUsername", "testuser").Return((*domain.User)(nil), nil)
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
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher, auditRepo *MockAuditLogRepository) {
				existingUser := &domain.User{ID: uuid.New(), Username: "testuser"}
				userRepo.On("GetByUsername", "testuser").Return(existingUser, nil)
			},
			expectError: true,
			errorType:   "ALREADY_EXISTS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := new(MockUserRepository)
			hasher := new(MockPasswordHasher)
			auditRepo := new(MockAuditLogRepository)
			tt.setupMocks(userRepo, hasher, auditRepo)

			// Create service
			service := NewAuthService(userRepo, auditRepo, hasher, nil, "test-secret", 24*time.Hour)

			// Execute
			user, token, err := service.Register(tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Contains(t, err.Error(), tt.errorType)
				}
				assert.Nil(t, user)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, token)
				assert.Equal(t, tt.request.Username, user.Username)
				assert.Equal(t, tt.request.Email, user.Email)
			}

			// Verify mocks
			userRepo.AssertExpectations(t)
			hasher.AssertExpectations(t)
			auditRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		password    string
		setupMocks  func(*MockUserRepository, *MockPasswordHasher, *MockAuditLogRepository)
		expectError bool
		errorType   string
	}{
		{
			name:     "successful login",
			username: "testuser",
			password: "password123",
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher, auditRepo *MockAuditLogRepository) {
				user := &domain.User{
					ID:           uuid.New(),
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: "hashed_password",
					IsActive:     true,
				}
				userRepo.On("GetByUsername", "testuser").Return(user, nil)
				hasher.On("VerifyPassword", "password123", "hashed_password").Return(true)
				auditRepo.On("Create", mock.AnythingOfType("*domain.AuditLog")).Return(nil)
			},
			expectError: false,
		},
		{
			name:     "user not found",
			username: "nonexistent",
			password: "password123",
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher, auditRepo *MockAuditLogRepository) {
				userRepo.On("GetByUsername", "nonexistent").Return((*domain.User)(nil), nil)
				userRepo.On("GetByEmail", "nonexistent").Return((*domain.User)(nil), nil)
			},
			expectError: true,
			errorType:   "INVALID_CREDENTIALS",
		},
		{
			name:     "invalid password",
			username: "testuser",
			password: "wrongpassword",
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher, auditRepo *MockAuditLogRepository) {
				user := &domain.User{
					ID:           uuid.New(),
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: "hashed_password",
					IsActive:     true,
				}
				userRepo.On("GetByUsername", "testuser").Return(user, nil)
				hasher.On("VerifyPassword", "wrongpassword", "hashed_password").Return(false)
			},
			expectError: true,
			errorType:   "INVALID_CREDENTIALS",
		},
		{
			name:     "inactive user",
			username: "testuser",
			password: "password123",
			setupMocks: func(userRepo *MockUserRepository, hasher *MockPasswordHasher, auditRepo *MockAuditLogRepository) {
				user := &domain.User{
					ID:           uuid.New(),
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: "hashed_password",
					IsActive:     false,
				}
				userRepo.On("GetByUsername", "testuser").Return(user, nil)
			},
			expectError: true,
			errorType:   "UNAUTHORIZED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := new(MockUserRepository)
			hasher := new(MockPasswordHasher)
			auditRepo := new(MockAuditLogRepository)
			tt.setupMocks(userRepo, hasher, auditRepo)

			// Create service
			service := NewAuthService(userRepo, auditRepo, hasher, nil, "test-secret", 24*time.Hour)

			// Execute
			user, token, err := service.Login(tt.username, tt.password)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Contains(t, err.Error(), tt.errorType)
				}
				assert.Nil(t, user)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, token)
			}

			// Verify mocks
			userRepo.AssertExpectations(t)
			hasher.AssertExpectations(t)
			auditRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		setupMocks  func(*MockUserRepository, *MockTokenBlacklist)
		expectError bool
		errorType   string
	}{
		{
			name:  "invalid token format",
			token: "invalid-token",
			setupMocks: func(userRepo *MockUserRepository, blacklist *MockTokenBlacklist) {
				// No mock setup needed for invalid token format
			},
			expectError: true,
			errorType:   "UNAUTHORIZED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			userRepo := new(MockUserRepository)
			blacklist := new(MockTokenBlacklist)
			tt.setupMocks(userRepo, blacklist)

			// Create service
			service := NewAuthService(userRepo, nil, nil, nil, "test-secret", 24*time.Hour)

			// Execute
			user, err := service.ValidateToken(tt.token)

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
			}

			// Verify mocks
			userRepo.AssertExpectations(t)
			blacklist.AssertExpectations(t)
		})
	}
}
